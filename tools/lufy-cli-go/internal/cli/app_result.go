package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/resultcontract"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/resultcontractledger"
	"gopkg.in/yaml.v3"
)

const (
	resultValidationSchemaVersion   = "lufy-result-validation/v1"
	transitionDecisionSchemaVersion = "lufy-result-transition-decision/v1"
)

type resultDecisionOutput struct {
	SchemaVersion string   `json:"schema_version"`
	Status        string   `json:"status"`
	Fingerprint   string   `json:"fingerprint,omitempty"`
	Reason        string   `json:"reason,omitempty"`
	Recovery      string   `json:"recovery,omitempty"`
	Missing       []string `json:"missing_evidence,omitempty"`
	NextOwner     string   `json:"next_owner,omitempty"`
	DecisionID    string   `json:"decision_id,omitempty"`
	NextVersion   uint64   `json:"next_version,omitempty"`
}

func runResult(args []string, deps Dependencies) int {
	if len(args) == 0 {
		printResultHelp(deps.Stderr)
		return ExitUsageErr
	}
	switch args[0] {
	case "validate":
		return runResultValidate(args[1:], deps)
	case "normalize":
		return runResultNormalize(args[1:], deps)
	case "transition":
		return runResultTransition(args[1:], deps)
	case "-h", "--help", "help":
		printResultHelp(deps.Stdout)
		return ExitOK
	default:
		fmt.Fprintf(deps.Stderr, "Subcomando result desconocido: %s\n\n", args[0])
		printResultHelp(deps.Stderr)
		return ExitUsageErr
	}
}

func runResultValidate(args []string, deps Dependencies) int {
	if resultHelpRequested(args) {
		printResultValidateHelp(deps.Stdout)
		return ExitOK
	}
	fs := flag.NewFlagSet("result validate", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	stdin := fs.Bool("stdin", false, "Leer contrato desde stdin")
	file := fs.String("file", "", "Leer contrato desde archivo")
	role := fs.String("role", "", "Validar status para un rol registrado")
	jsonOutput := fs.Bool("json", false, "Emitir decisión JSON")
	if err := fs.Parse(args); err != nil || len(fs.Args()) > 0 || *stdin == (*file != "") {
		printResultValidateHelp(deps.Stderr)
		return ExitUsageErr
	}
	input, closeInput, err := resultInput(*stdin, *file, deps)
	if err != nil {
		return writeResultInvalid(deps, *jsonOutput, resultValidationSchemaVersion, err)
	}
	defer closeInput()
	contract, err := resultcontract.Decode(input)
	if err != nil {
		return writeResultInvalid(deps, *jsonOutput, resultValidationSchemaVersion, err)
	}
	canonical, err := resultcontract.Canonicalize(contract)
	if err != nil {
		return writeResultInvalid(deps, *jsonOutput, resultValidationSchemaVersion, err)
	}
	roles := resultRolePolicy()
	evidence := resultEvidencePolicy()
	claimRole := strings.TrimSpace(*role)
	if claimRole == "" {
		claimRole = "__structural__"
		roles[claimRole] = allResultStatuses()
	}
	claim := resultcontract.EvaluateClaim(contract, claimRole, roles, evidence)
	if !claim.Accepted {
		output := resultDecisionOutput{
			SchemaVersion: resultValidationSchemaVersion, Status: "rejected", Fingerprint: canonical.Fingerprint,
			Reason: claim.Reason, Recovery: claim.Recovery, Missing: evidenceStrings(claim.MissingEvidence), NextOwner: claim.NextOwner,
		}
		writeResultOutput(deps, *jsonOutput, output, fmt.Sprintf("result validate: rejected reason=%s recovery=%s", output.Reason, output.Recovery))
		return ExitResultRejected
	}
	output := resultDecisionOutput{SchemaVersion: resultValidationSchemaVersion, Status: "valid", Fingerprint: canonical.Fingerprint}
	writeResultOutput(deps, *jsonOutput, output, fmt.Sprintf("result validate: valid fingerprint=%s", canonical.Fingerprint))
	return ExitOK
}

func runResultNormalize(args []string, deps Dependencies) int {
	if resultHelpRequested(args) {
		printResultNormalizeHelp(deps.Stdout)
		return ExitOK
	}
	fs := flag.NewFlagSet("result normalize", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	stdin := fs.Bool("stdin", false, "Leer legacy result desde stdin")
	file := fs.String("file", "", "Leer legacy result desde archivo")
	jsonOutput := fs.Bool("json", false, "Emitir contrato JSON")
	if err := fs.Parse(args); err != nil || len(fs.Args()) > 0 || *stdin == (*file != "") {
		printResultNormalizeHelp(deps.Stderr)
		return ExitUsageErr
	}
	input, closeInput, err := resultInput(*stdin, *file, deps)
	if err != nil {
		return writeResultInvalid(deps, *jsonOutput, resultValidationSchemaVersion, err)
	}
	defer closeInput()
	contract, err := resultcontract.NormalizeLegacy(input)
	if err != nil {
		return writeResultInvalid(deps, *jsonOutput, resultValidationSchemaVersion, err)
	}
	if *jsonOutput {
		body, marshalErr := json.MarshalIndent(contract, "", "  ")
		if marshalErr != nil {
			return writeResultInvalid(deps, true, resultValidationSchemaVersion, marshalErr)
		}
		fmt.Fprintln(deps.Stdout, string(body))
		return ExitOK
	}
	body, err := yaml.Marshal(contract)
	if err != nil {
		fmt.Fprintln(deps.Stderr, "no se pudo renderizar contrato normalizado")
		return ExitRuntimeErr
	}
	fmt.Fprint(deps.Stdout, string(body))
	return ExitOK
}

func runResultTransition(args []string, deps Dependencies) int {
	if resultHelpRequested(args) {
		printResultTransitionHelp(deps.Stdout)
		return ExitOK
	}
	fs := flag.NewFlagSet("result transition", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)
	stdin := fs.Bool("stdin", false, "Leer request desde stdin")
	file := fs.String("file", "", "Leer request desde archivo")
	target := fs.String("target", ".", "Repositorio target")
	record := fs.Bool("record", false, "Persistir receipt content-free")
	jsonOutput := fs.Bool("json", false, "Emitir decisión JSON")
	if err := fs.Parse(args); err != nil || len(fs.Args()) > 0 || *stdin == (*file != "") {
		printResultTransitionHelp(deps.Stderr)
		return ExitUsageErr
	}
	input, closeInput, err := resultInput(*stdin, *file, deps)
	if err != nil {
		return writeResultInvalid(deps, *jsonOutput, transitionDecisionSchemaVersion, err)
	}
	defer closeInput()
	request, err := resultcontract.DecodeTransitionRequest(input)
	if err != nil {
		return writeResultInvalid(deps, *jsonOutput, transitionDecisionSchemaVersion, err)
	}
	policy := resultcontract.TransitionPolicy{Roles: resultRolePolicy(), Evidence: resultEvidencePolicy()}
	var ledgerBridge resultcontractledger.Bridge
	if *record {
		store, _, storeErr := newRunStore(*target)
		if storeErr != nil {
			return writeResultUnavailable(deps, *jsonOutput, "ledger_unavailable")
		}
		policy.Receipts = newFileResultReceiptStore(*target)
		ledgerBridge = resultcontractledger.New(store)
	}
	decision := resultcontract.EvaluateTransition(request.Prior, request.Intent, policy)
	if *record && (decision.Status == resultcontract.DecisionAccepted || decision.Status == resultcontract.DecisionDuplicateNoop) {
		ledgerDecision := decision
		if ledgerDecision.Status == resultcontract.DecisionDuplicateNoop {
			ledgerDecision.Status = resultcontract.DecisionAccepted
		}
		ledgerResult := ledgerBridge.Record(backgroundContext(), resultcontractledger.RecordRequest{
			RunID: request.Prior.RunID, IdempotencyKey: request.Intent.IdempotencyKey,
			Contract: request.Intent.NextContract, Decision: ledgerDecision,
		})
		switch ledgerResult.Status {
		case resultcontractledger.StatusRecorded, resultcontractledger.StatusDuplicateNoop:
		case resultcontractledger.StatusConflict:
			output := resultDecisionOutput{
				SchemaVersion: transitionDecisionSchemaVersion, Status: "conflict",
				Reason: "ledger_idempotency_conflict", Recovery: "usar una idempotency key nueva o reintentar el intent original",
			}
			writeResultOutput(deps, *jsonOutput, output, "result transition: conflict reason=ledger_idempotency_conflict")
			return ExitResultConflict
		default:
			return writeResultUnavailable(deps, *jsonOutput, "ledger_unavailable")
		}
	}
	status := string(decision.Status)
	exitCode := ExitOK
	if decision.Reason == "receipt_unavailable" {
		return writeResultUnavailable(deps, *jsonOutput, "receipt_unavailable")
	} else if decision.Status == resultcontract.DecisionRejected {
		exitCode = ExitResultRejected
	} else if decision.Status == resultcontract.DecisionConflict {
		exitCode = ExitResultConflict
	}
	output := resultDecisionOutput{
		SchemaVersion: transitionDecisionSchemaVersion, Status: status, Reason: decision.Reason,
		Recovery: decision.Recovery, Missing: evidenceStrings(decision.MissingEvidence), NextOwner: decision.NextOwner,
		DecisionID: decision.DecisionID, NextVersion: decision.NextVersion, Fingerprint: decision.NextFingerprint,
	}
	writeResultOutput(deps, *jsonOutput, output, fmt.Sprintf("result transition: %s reason=%s recovery=%s", status, output.Reason, output.Recovery))
	return exitCode
}

func writeResultUnavailable(deps Dependencies, jsonOutput bool, reason string) int {
	output := resultDecisionOutput{
		SchemaVersion: transitionDecisionSchemaVersion,
		Status:        "unavailable",
		Reason:        reason,
		Recovery:      "reintentar cuando el almacenamiento durable esté disponible",
	}
	writeResultOutput(deps, jsonOutput, output, fmt.Sprintf("result transition: unavailable reason=%s recovery=%s", output.Reason, output.Recovery))
	return ExitResultUnavailable
}

func resultInput(stdin bool, path string, deps Dependencies) (io.Reader, func(), error) {
	if stdin {
		input := deps.Stdin
		if input == nil {
			input = os.Stdin
		}
		return input, func() {}, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, func() {}, errors.New("input no disponible")
	}
	return file, func() { _ = file.Close() }, nil
}

func resultRolePolicy() resultcontract.RolePolicy {
	policy := resultcontract.RolePolicy{}
	for _, role := range domain.DefaultRoleContracts() {
		statuses := make([]resultcontract.Status, 0, len(role.Output.AllowedStatus))
		for _, status := range role.Output.AllowedStatus {
			statuses = append(statuses, resultcontract.Status(status))
		}
		policy[string(role.ID)] = statuses
	}
	return policy
}

func resultEvidencePolicy() resultcontract.EvidencePolicy {
	return resultcontract.EvidencePolicy{ByStatus: map[resultcontract.Status]resultcontract.EvidenceRequirement{
		resultcontract.StatusValidated: {Categories: []resultcontract.EvidenceCategory{resultcontract.EvidenceCategoryPassedCommand}, NextOwner: "validator"},
		resultcontract.StatusDelivered: {Categories: []resultcontract.EvidenceCategory{resultcontract.EvidenceCategoryPassedCommand, resultcontract.EvidenceCategoryArtifactReference}, NextOwner: "delivery"},
		resultcontract.StatusClosed:    {Categories: []resultcontract.EvidenceCategory{resultcontract.EvidenceCategoryPassedCommand, resultcontract.EvidenceCategoryArtifactReference, resultcontract.EvidenceCategoryStatic}, NextOwner: "orchestrator"},
	}}
}

func allResultStatuses() []resultcontract.Status {
	return []resultcontract.Status{
		resultcontract.StatusReady, resultcontract.StatusImplemented, resultcontract.StatusValidated,
		resultcontract.StatusDeliveryPending, resultcontract.StatusSyncPending, resultcontract.StatusBlocked,
		resultcontract.StatusEscalated, resultcontract.StatusDelivered, resultcontract.StatusClosed,
	}
}

func evidenceStrings(values []resultcontract.EvidenceCategory) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = string(value)
	}
	return result
}

func writeResultInvalid(deps Dependencies, jsonOutput bool, schema string, err error) int {
	reason, recovery := "invalid_input", "corregir el documento tipado"
	var diagnostic *resultcontract.DiagnosticError
	if errors.As(err, &diagnostic) {
		reason, recovery = diagnostic.Code, diagnostic.Recovery
	}
	output := resultDecisionOutput{SchemaVersion: schema, Status: "invalid", Reason: reason, Recovery: recovery}
	writeResultOutput(deps, jsonOutput, output, fmt.Sprintf("result: invalid reason=%s recovery=%s", reason, recovery))
	return ExitResultInvalid
}

func writeResultOutput(deps Dependencies, jsonOutput bool, output resultDecisionOutput, human string) {
	if jsonOutput {
		body, _ := json.Marshal(output)
		fmt.Fprintln(deps.Stdout, string(body))
		return
	}
	fmt.Fprintln(deps.Stdout, human)
}

func resultHelpRequested(args []string) bool {
	return len(args) == 1 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help")
}

func printResultHelp(out io.Writer) {
	fmt.Fprintln(out, "Uso: lufy-ai result <subcomando> [flags]")
	fmt.Fprintln(out, "Subcomandos: validate, normalize, transition")
}

func printResultValidateHelp(out io.Writer) {
	fmt.Fprintln(out, "Uso: lufy-ai result validate (--stdin|--file <path>) [--role <role>] [--json]")
}

func printResultNormalizeHelp(out io.Writer) {
	fmt.Fprintln(out, "Uso: lufy-ai result normalize (--stdin|--file <path>) [--json]")
}

func printResultTransitionHelp(out io.Writer) {
	fmt.Fprintln(out, "Uso: lufy-ai result transition (--stdin|--file <path>) [--target <dir>] [--record] [--json]")
}
