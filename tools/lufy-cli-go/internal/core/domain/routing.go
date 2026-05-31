package domain

import "fmt"

type RoutingPolicyOptions struct {
	AllowT2None  bool
	T2NoneReason string
}

func (m MethodologyByTier) ValidateRoutingPolicy(opts RoutingPolicyOptions) error {
	if err := m.ValidateSupported(); err != nil {
		return err
	}
	if selection, ok := m[TierT1]; ok && selection.ID == MethodologyNone {
		return fmt.Errorf("T1 no permite metodologia none sin una metodologia full")
	}
	if selection, ok := m[TierT2]; ok && selection.ID == MethodologyNone {
		if !opts.AllowT2None || opts.T2NoneReason == "" {
			return fmt.Errorf("T2 con metodologia none requiere justificacion explicita")
		}
	}
	return nil
}
