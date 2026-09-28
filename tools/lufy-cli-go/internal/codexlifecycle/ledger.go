package codexlifecycle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
)

const bindingSchemaVersion = "lufy-run-binding/v1"

type lifecycleBinding struct {
	SchemaVersion  string `json:"schema_version"`
	Adapter        string `json:"adapter"`
	Kind           string `json:"kind"`
	SourceRefHash  string `json:"source_ref_hash"`
	SessionRefHash string `json:"session_ref_hash"`
	RunID          string `json:"run_id"`
	ParentRunID    string `json:"parent_run_id,omitempty"`
	StartEventID   string `json:"start_event_id"`
}

type codexLedger struct {
	target       string
	bindingsRoot string
	store        *runledger.FileStore
}

func recordLifecycleEvent(target string, input Input) error {
	if strings.TrimSpace(input.SessionID) == "" {
		return fmt.Errorf("metadata estable incompleta")
	}
	store, runtimeRoot, enabled, err := lifecycleRunStore(target)
	if err != nil || !enabled {
		return err
	}
	bindingsRoot, err := platform.SafeJoin(target, filepath.Join(runtimeRoot, "bindings", "codex"))
	if err != nil {
		return err
	}
	adapter := codexLedger{target: target, bindingsRoot: bindingsRoot, store: store}
	return adapter.record(context.Background(), input)
}

func lifecycleRunStore(target string) (*runledger.FileStore, string, bool, error) {
	config := projectconfig.DefaultRunLedgerConfig()
	path, err := projectconfig.ExistingPath(target)
	if err != nil {
		return nil, config.Root, false, err
	}
	if _, statErr := os.Stat(path); statErr == nil {
		loaded, loadErr := projectconfig.Load(path)
		if loadErr != nil {
			return nil, config.Root, false, loadErr
		}
		config = loaded.RunLedger
	} else if !os.IsNotExist(statErr) {
		return nil, config.Root, false, statErr
	}
	if !config.IsEnabled() {
		return nil, config.Root, false, nil
	}
	store, err := runledger.NewFileStore(target, runledger.Options{RuntimeRoot: config.Root})
	return store, config.Root, true, err
}

func (a codexLedger) record(ctx context.Context, input Input) error {
	session, err := a.ensureSession(ctx, input.SessionID)
	if err != nil {
		return err
	}
	switch input.HookEventName {
	case "SessionStart":
		return nil
	case "SubagentStart":
		if strings.TrimSpace(input.AgentID) == "" {
			return fmt.Errorf("metadata estable incompleta")
		}
		_, err := a.ensureAgent(ctx, session, input)
		return err
	case "SubagentStop":
		if strings.TrimSpace(input.AgentID) == "" {
			return fmt.Errorf("metadata estable incompleta")
		}
		agent, found, err := a.loadBinding("agent", a.agentHash(input.SessionID, input.AgentID))
		if err != nil {
			return err
		}
		if !found {
			agent, err = a.ensureAgent(ctx, session, input)
			if err != nil {
				return err
			}
		}
		return a.appendLifecycle(ctx, input, agent, runledger.KindFinish, 0)
	case "Stop":
		return a.appendLifecycle(ctx, input, session, runledger.KindCheckpoint, 0)
	case "SessionEnd":
		return a.appendLifecycle(ctx, input, session, runledger.KindFinish, 0)
	default:
		return nil
	}
}

func (a codexLedger) ensureSession(ctx context.Context, sessionID string) (lifecycleBinding, error) {
	hash := a.pseudonym("session", sessionID)
	if binding, found, err := a.loadBinding("session", hash); err != nil || found {
		return binding, err
	}
	runID := localID("run-codex-session-", hash)
	draft := runledger.EventDraft{
		EventID: localID("evt-codex-session-start-", hash),
		RunID:   runID,
		Kind:    runledger.KindStart,
		Source: runledger.Source{
			Adapter:        "codex",
			EventName:      "SessionStart",
			SessionRefHash: hash,
		},
	}
	result, err := a.store.Append(ctx, runledger.AppendRequest{Draft: draft, IdempotencyKey: "codex:session-start:" + hash})
	if err != nil {
		return lifecycleBinding{}, err
	}
	binding := lifecycleBinding{
		SchemaVersion: bindingSchemaVersion, Adapter: "codex", Kind: "session", SourceRefHash: hash,
		SessionRefHash: hash, RunID: runID, StartEventID: result.Event.EventID,
	}
	if err := a.writeBinding(binding); err != nil {
		return lifecycleBinding{}, err
	}
	return binding, nil
}

func (a codexLedger) ensureAgent(ctx context.Context, session lifecycleBinding, input Input) (lifecycleBinding, error) {
	hash := a.agentHash(input.SessionID, input.AgentID)
	if binding, found, err := a.loadBinding("agent", hash); err != nil || found {
		return binding, err
	}
	runID := localID("run-codex-agent-", hash)
	source := a.source(input)
	source.EventName = "SubagentStart"
	draft := runledger.EventDraft{
		EventID:         localID("evt-codex-agent-start-", hash),
		RunID:           runID,
		ParentRunID:     session.RunID,
		CausedByEventID: session.StartEventID,
		Kind:            runledger.KindStart,
		Source:          source,
	}
	result, err := a.store.Append(ctx, runledger.AppendRequest{
		Draft: draft, IdempotencyKey: "codex:agent-start:" + hash, ProposedLamport: 1,
	})
	if err != nil {
		return lifecycleBinding{}, err
	}
	binding := lifecycleBinding{
		SchemaVersion: bindingSchemaVersion, Adapter: "codex", Kind: "agent", SourceRefHash: hash,
		SessionRefHash: session.SessionRefHash, RunID: runID, ParentRunID: session.RunID, StartEventID: result.Event.EventID,
	}
	if err := a.writeBinding(binding); err != nil {
		return lifecycleBinding{}, err
	}
	return binding, nil
}

func (a codexLedger) appendLifecycle(ctx context.Context, input Input, binding lifecycleBinding, kind runledger.EventKind, proposedLamport uint64) error {
	eventHash := a.pseudonym("event", strings.Join([]string{input.HookEventName, input.SessionID, input.TurnID, input.AgentID}, "\x00"))
	draft := runledger.EventDraft{
		EventID:     localID("evt-codex-", eventHash),
		RunID:       binding.RunID,
		ParentRunID: binding.ParentRunID,
		Kind:        kind,
		Source:      a.source(input),
	}
	_, err := a.store.Append(ctx, runledger.AppendRequest{
		Draft: draft, IdempotencyKey: "codex:event:" + eventHash, ProposedLamport: proposedLamport,
	})
	return err
}

func (a codexLedger) source(input Input) runledger.Source {
	source := runledger.Source{
		Adapter:        "codex",
		EventName:      input.HookEventName,
		SessionRefHash: a.pseudonym("session", input.SessionID),
		AgentType:      safeAgentType(input.AgentType),
	}
	if input.TurnID != "" {
		source.TurnRefHash = a.pseudonym("turn", input.SessionID+"\x00"+input.TurnID)
	}
	if input.AgentID != "" {
		source.AgentRefHash = a.agentHash(input.SessionID, input.AgentID)
	}
	return source
}

func (a codexLedger) pseudonym(kind, value string) string {
	return runledger.HashReference("lufy-codex-binding/v1\x00" + filepath.Clean(a.target) + "\x00" + kind + "\x00" + value)
}

func (a codexLedger) agentHash(sessionID, agentID string) string {
	return a.pseudonym("agent", sessionID+"\x00"+agentID)
}

func (a codexLedger) bindingPath(kind, hash string) (string, error) {
	return platform.SafeJoin(a.bindingsRoot, kind+"-"+hash+".json")
}

func (a codexLedger) loadBinding(kind, hash string) (lifecycleBinding, bool, error) {
	path, err := a.bindingPath(kind, hash)
	if err != nil {
		return lifecycleBinding{}, false, err
	}
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return lifecycleBinding{}, false, nil
	}
	if err != nil {
		return lifecycleBinding{}, false, err
	}
	defer file.Close()
	var binding lifecycleBinding
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&binding); err != nil {
		return lifecycleBinding{}, false, fmt.Errorf("binding inválido")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return lifecycleBinding{}, false, fmt.Errorf("binding inválido")
	}
	if binding.SchemaVersion != bindingSchemaVersion || binding.Adapter != "codex" || binding.Kind != kind || binding.SourceRefHash != hash {
		return lifecycleBinding{}, false, fmt.Errorf("binding inválido")
	}
	return binding, true, nil
}

func (a codexLedger) writeBinding(binding lifecycleBinding) error {
	path, err := a.bindingPath(binding.Kind, binding.SourceRefHash)
	if err != nil {
		return err
	}
	body, err := json.MarshalIndent(binding, "", "  ")
	if err != nil {
		return err
	}
	return platform.WriteFileAtomic(path, append(body, '\n'), 0o600)
}

func localID(prefix, hash string) string {
	const suffixLength = 32
	if len(hash) > suffixLength {
		hash = hash[:suffixLength]
	}
	return prefix + hash
}

func safeAgentType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 128 {
		return ""
	}
	for i := 0; i < len(value); i++ {
		char := value[i]
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || strings.ContainsRune("._:/-", rune(char)) {
			continue
		}
		return ""
	}
	return value
}
