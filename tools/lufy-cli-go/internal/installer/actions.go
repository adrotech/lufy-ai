package installer

type ActionKind string

const (
	ActionMkdir                 ActionKind = "mkdir"
	ActionBackup                ActionKind = "backup"
	ActionCopy                  ActionKind = "copy"
	ActionUpdateManaged         ActionKind = "update-managed"
	ActionMergeBlock            ActionKind = "merge-block"
	ActionAdoptMergeBlock       ActionKind = "adopt-merge-block"
	ActionWriteLufyNew          ActionKind = "write-lufy-new"
	ActionAgentsReferenceCreate ActionKind = "agents-reference-create"
	ActionAgentsReferenceInsert ActionKind = "agents-reference-insert"
	ActionVerify                ActionKind = "verify"
	ActionAgentsReferenceSkip   ActionKind = "agents-reference-skip"
	ActionSkip                  ActionKind = "skip"
)

var actionOrder = map[ActionKind]int{
	ActionMkdir:                 0,
	ActionBackup:                1,
	ActionCopy:                  2,
	ActionUpdateManaged:         3,
	ActionMergeBlock:            4,
	ActionAdoptMergeBlock:       5,
	ActionWriteLufyNew:          6,
	ActionAgentsReferenceCreate: 7,
	ActionAgentsReferenceInsert: 8,
	ActionVerify:                10,
	ActionAgentsReferenceSkip:   11,
	ActionSkip:                  12,
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

func hasContentMutation(actions []Action) bool {
	for _, action := range actions {
		switch action.Kind {
		case ActionCopy, ActionUpdateManaged, ActionMergeBlock, ActionWriteLufyNew, ActionAgentsReferenceCreate, ActionAgentsReferenceInsert:
			return true
		}
	}
	return false
}

func requiresConfirmation(actions []Action) bool {
	for _, action := range actions {
		switch action.Kind {
		case ActionMkdir, ActionCopy, ActionUpdateManaged, ActionMergeBlock, ActionAdoptMergeBlock, ActionWriteLufyNew, ActionAgentsReferenceCreate, ActionAgentsReferenceInsert, ActionBackup:
			return true
		}
	}
	return false
}
