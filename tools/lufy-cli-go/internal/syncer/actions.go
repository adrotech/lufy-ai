package syncer

import "fmt"

type ActionKind string

const (
	ActionBackup              ActionKind = "backup"
	ActionCreateManaged       ActionKind = "create-managed"
	ActionUpdateManaged       ActionKind = "update-managed"
	ActionMergeBlock          ActionKind = "merge-block"
	ActionWriteLufyNew        ActionKind = "write-lufy-new"
	ActionMergeJSON           ActionKind = "merge-json"
	ActionRetired             ActionKind = "retired"
	ActionWarnAgentsReference ActionKind = "warn-agents-reference"
	ActionPinnedSkip          ActionKind = "pinned-skip"
	ActionSkip                ActionKind = "skip"
)

var actionOrder = map[ActionKind]int{
	ActionBackup:              0,
	ActionCreateManaged:       1,
	ActionUpdateManaged:       2,
	ActionMergeBlock:          3,
	ActionWriteLufyNew:        4,
	ActionMergeJSON:           5,
	ActionRetired:             6,
	ActionWarnAgentsReference: 7,
	ActionPinnedSkip:          8,
	ActionSkip:                9,
}

func validateActionKinds(actions []Action) error {
	for _, action := range actions {
		switch action.Kind {
		case ActionBackup, ActionCreateManaged, ActionUpdateManaged, ActionMergeBlock, ActionWriteLufyNew, ActionMergeJSON, ActionRetired, ActionWarnAgentsReference, ActionPinnedSkip, ActionSkip:
			continue
		default:
			return fmt.Errorf("acción sync no soportada: %s", action.Kind)
		}
	}
	return nil
}

func requiresConfirmation(actions []Action) bool {
	for _, action := range actions {
		switch action.Kind {
		case ActionCreateManaged, ActionUpdateManaged, ActionMergeBlock, ActionWriteLufyNew, ActionBackup, ActionMergeJSON:
			return true
		}
	}
	return false
}

func targetsForKind(actions []Action, kind ActionKind) []string {
	var out []string
	for _, action := range actions {
		if action.Kind == kind {
			out = append(out, action.Target)
		}
	}
	return out
}

func hasActionKind(actions []Action, kind ActionKind) bool {
	for _, action := range actions {
		if action.Kind == kind {
			return true
		}
	}
	return false
}
