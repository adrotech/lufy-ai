package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

func Evaluate(request EvaluationRequest, policy Policy) (Decision, error) {
	if err := ValidateEvaluation(request); err != nil {
		return Decision{}, err
	}
	if err := validatePolicy(policy); err != nil {
		return Decision{}, err
	}
	decision := Decision{
		SchemaVersion: DecisionSchemaVersion,
		Mode:          policy.Mode,
		PolicyVersion: policy.Version,
		Ranking:       []CandidateDecision{},
		GateAdvanced:  false,
	}
	if !policy.Enabled {
		decision.Action = ActionDisabled
		decision.Reason = "adaptive_routing_disabled"
		decision.Fingerprint = fingerprintDecision(request, policy, decision)
		return decision, nil
	}
	if len(request.Demand.ProtectedBoundaries) > 0 {
		decision.Action = ActionEscalate
		decision.Reason = "protected_boundary"
		decision.NextOwner = "user"
		decision.Fingerprint = fingerprintDecision(request, policy, decision)
		return decision, nil
	}

	profiles := append([]CapabilityProfile(nil), request.Profiles...)
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].ActorRef < profiles[j].ActorRef })
	if len(profiles) > policy.MaxCandidates {
		profiles = profiles[:policy.MaxCandidates]
	}
	for _, profile := range profiles {
		decision.Ranking = append(decision.Ranking, evaluateCandidate(request.Demand, profile))
	}
	sort.SliceStable(decision.Ranking, func(i, j int) bool {
		left, right := decision.Ranking[i], decision.Ranking[j]
		if left.Eligible != right.Eligible {
			return left.Eligible
		}
		if left.Breakdown.Total != right.Breakdown.Total {
			return left.Breakdown.Total > right.Breakdown.Total
		}
		if left.Breakdown.CapabilityMatch != right.Breakdown.CapabilityMatch {
			return left.Breakdown.CapabilityMatch > right.Breakdown.CapabilityMatch
		}
		if left.Breakdown.CapacityFit != right.Breakdown.CapacityFit {
			return left.Breakdown.CapacityFit > right.Breakdown.CapacityFit
		}
		return left.ActorRef < right.ActorRef
	})
	for _, candidate := range decision.Ranking {
		if !candidate.Eligible {
			continue
		}
		decision.Action = ActionRecommend
		decision.SelectedActorRef = candidate.ActorRef
		decision.SelectedRoleHint = candidate.RoleHint
		break
	}
	if decision.Action == "" {
		decision.Action = ActionNoCandidate
		decision.Reason = "no_eligible_candidate"
		decision.NextOwner = "orchestrator"
	}
	decision.Fingerprint = fingerprintDecision(request, policy, decision)
	return decision, nil
}

func evaluateCandidate(demand DemandSignal, profile CapabilityProfile) CandidateDecision {
	breakdown := ScoreBreakdown{
		DemandPriority:   demand.Priority,
		RiskGap:          max(0, demand.Risk-profile.RiskTolerance),
		CoordinationCost: demand.CoordinationCost + profile.CoordinationCost,
	}
	if profile.AvailableBudget >= demand.RequiredBudget {
		breakdown.CapacityFit = 100
	} else {
		breakdown.CapacityFit = profile.AvailableBudget * 100 / demand.RequiredBudget
	}

	levels := make(map[string]int, len(profile.Capabilities))
	for _, capability := range profile.Capabilities {
		levels[capability.Name] = capability.Level
	}
	weighted, weights := 0, 0
	missing := false
	for _, required := range demand.RequiredCapabilities {
		level, ok := levels[required.Name]
		if !ok {
			missing = true
			continue
		}
		match := min(level, required.Level) * 100 / required.Level
		weighted += match * required.Weight
		weights += required.Weight
	}
	if weights > 0 {
		breakdown.CapabilityMatch = weighted / weights
	}
	breakdown.Total = 5*breakdown.CapabilityMatch +
		4*breakdown.DemandPriority +
		2*breakdown.CapacityFit -
		6*breakdown.RiskGap -
		3*breakdown.CoordinationCost

	roleHints := append([]string(nil), profile.RoleHints...)
	sort.Strings(roleHints)
	candidate := CandidateDecision{ActorRef: profile.ActorRef, Eligible: true, Breakdown: breakdown}
	if len(roleHints) > 0 {
		candidate.RoleHint = roleHints[0]
	}
	if missing {
		candidate.Eligible = false
		candidate.Reason = "missing_capability"
	} else if profile.AvailableBudget < demand.RequiredBudget {
		candidate.Eligible = false
		candidate.Reason = "insufficient_budget"
	}
	return candidate
}

func fingerprintDecision(request EvaluationRequest, policy Policy, decision Decision) string {
	normalized := normalizeRequest(request)
	seed := struct {
		Request  EvaluationRequest `json:"request"`
		Policy   Policy            `json:"policy"`
		Decision Decision          `json:"decision"`
	}{Request: normalized, Policy: policy, Decision: decision}
	seed.Decision.Fingerprint = ""
	body, _ := json.Marshal(seed)
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func normalizeRequest(request EvaluationRequest) EvaluationRequest {
	request.Demand.RequiredCapabilities = append([]RequiredCapability(nil), request.Demand.RequiredCapabilities...)
	sort.Slice(request.Demand.RequiredCapabilities, func(i, j int) bool {
		return request.Demand.RequiredCapabilities[i].Name < request.Demand.RequiredCapabilities[j].Name
	})
	request.Demand.ProtectedBoundaries = append([]string(nil), request.Demand.ProtectedBoundaries...)
	sort.Strings(request.Demand.ProtectedBoundaries)
	request.Profiles = append([]CapabilityProfile(nil), request.Profiles...)
	for index := range request.Profiles {
		profile := &request.Profiles[index]
		profile.Capabilities = append([]Capability(nil), profile.Capabilities...)
		sort.Slice(profile.Capabilities, func(i, j int) bool { return profile.Capabilities[i].Name < profile.Capabilities[j].Name })
		profile.ActiveAssignments = append([]string(nil), profile.ActiveAssignments...)
		sort.Strings(profile.ActiveAssignments)
		profile.RoleHints = append([]string(nil), profile.RoleHints...)
		sort.Strings(profile.RoleHints)
	}
	sort.Slice(request.Profiles, func(i, j int) bool { return request.Profiles[i].ActorRef < request.Profiles[j].ActorRef })
	return request
}
