package resultcontract

type StateVector struct {
	Work      string `json:"work" yaml:"work"`
	Delivery  string `json:"delivery" yaml:"delivery"`
	Sync      string `json:"sync" yaml:"sync"`
	Attention string `json:"attention" yaml:"attention"`
	Terminal  bool   `json:"terminal" yaml:"terminal"`
}

func DeriveState(status Status, previous *StateVector) (StateVector, error) {
	switch status {
	case StatusReady:
		return StateVector{Work: "ready", Delivery: "not_required", Sync: "not_required", Attention: "none"}, nil
	case StatusImplemented:
		return StateVector{Work: "implemented", Delivery: "not_required", Sync: "not_required", Attention: "none"}, nil
	case StatusValidated:
		return StateVector{Work: "validated", Delivery: "not_required", Sync: "not_required", Attention: "none"}, nil
	case StatusDeliveryPending:
		return StateVector{Work: "validated", Delivery: "pending", Sync: "not_required", Attention: "none"}, nil
	case StatusSyncPending:
		return StateVector{Work: "validated", Delivery: "not_required", Sync: "pending", Attention: "none"}, nil
	case StatusDelivered:
		return StateVector{Work: "validated", Delivery: "delivered", Sync: "not_required", Attention: "none"}, nil
	case StatusClosed:
		return StateVector{Work: "terminal", Delivery: "delivered", Sync: "synced", Attention: "none", Terminal: true}, nil
	case StatusBlocked, StatusEscalated:
		if previous == nil {
			return StateVector{}, diagnostic("previous_state_required", "status", "proveer el vector previo para preservar delivery y sync")
		}
		attention := "recovery_required"
		if status == StatusEscalated {
			attention = "escalated"
		}
		return StateVector{
			Work:      "blocked",
			Delivery:  previous.Delivery,
			Sync:      previous.Sync,
			Attention: attention,
		}, nil
	default:
		return StateVector{}, diagnostic("invalid_enum", "status", "usar un status documentado")
	}
}

var explicitTransitionTable = map[Status]map[Status]bool{
	StatusReady: {
		StatusReady: false, StatusImplemented: true, StatusValidated: false,
		StatusDeliveryPending: false, StatusSyncPending: false, StatusBlocked: true,
		StatusEscalated: true, StatusDelivered: false, StatusClosed: false,
	},
	StatusImplemented: {
		StatusReady: false, StatusImplemented: false, StatusValidated: true,
		StatusDeliveryPending: false, StatusSyncPending: false, StatusBlocked: true,
		StatusEscalated: true, StatusDelivered: false, StatusClosed: false,
	},
	StatusValidated: {
		StatusReady: false, StatusImplemented: false, StatusValidated: false,
		StatusDeliveryPending: true, StatusSyncPending: true, StatusBlocked: true,
		StatusEscalated: true, StatusDelivered: true, StatusClosed: true,
	},
	StatusDeliveryPending: {
		StatusReady: false, StatusImplemented: false, StatusValidated: false,
		StatusDeliveryPending: false, StatusSyncPending: true, StatusBlocked: true,
		StatusEscalated: true, StatusDelivered: true, StatusClosed: false,
	},
	StatusSyncPending: {
		StatusReady: false, StatusImplemented: false, StatusValidated: false,
		StatusDeliveryPending: true, StatusSyncPending: false, StatusBlocked: true,
		StatusEscalated: true, StatusDelivered: true, StatusClosed: false,
	},
	StatusBlocked: {
		StatusReady: true, StatusImplemented: true, StatusValidated: true,
		StatusDeliveryPending: true, StatusSyncPending: true, StatusBlocked: false,
		StatusEscalated: true, StatusDelivered: false, StatusClosed: false,
	},
	StatusEscalated: {
		StatusReady: true, StatusImplemented: true, StatusValidated: true,
		StatusDeliveryPending: true, StatusSyncPending: true, StatusBlocked: true,
		StatusEscalated: false, StatusDelivered: false, StatusClosed: false,
	},
	StatusDelivered: {
		StatusReady: false, StatusImplemented: false, StatusValidated: false,
		StatusDeliveryPending: false, StatusSyncPending: true, StatusBlocked: true,
		StatusEscalated: true, StatusDelivered: false, StatusClosed: true,
	},
	StatusClosed: {
		StatusReady: false, StatusImplemented: false, StatusValidated: false,
		StatusDeliveryPending: false, StatusSyncPending: false, StatusBlocked: false,
		StatusEscalated: false, StatusDelivered: false, StatusClosed: false,
	},
}

func IsAllowedTransition(from, to Status) bool {
	row, known := explicitTransitionTable[from]
	if !known {
		return false
	}
	allowed, known := row[to]
	return known && allowed
}
