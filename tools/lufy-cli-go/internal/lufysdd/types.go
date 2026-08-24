package lufysdd

const ReportSchema = "lufy-sdd-report/v1"

type Mode string

const (
	ModeFull Mode = "full"
	ModeLite Mode = "lite"
)

type Level string

const (
	LevelError   Level = "error"
	LevelWarning Level = "warning"
	LevelInfo    Level = "info"
)

type Diagnostic struct {
	Level   Level  `json:"level"`
	Code    string `json:"code"`
	Path    string `json:"path,omitempty"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
}

type Progress struct {
	Complete int `json:"complete"`
	Total    int `json:"total"`
}

type Report struct {
	Schema      string       `json:"schema"`
	Action      string       `json:"action"`
	Change      string       `json:"change,omitempty"`
	Mode        Mode         `json:"mode,omitempty"`
	Status      string       `json:"status"`
	Root        string       `json:"root"`
	Progress    Progress     `json:"progress"`
	DeltaDigest string       `json:"deltaDigest,omitempty"`
	Actions     []Action     `json:"actions,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Action struct {
	Kind        string `json:"kind"`
	Path        string `json:"path"`
	Requirement string `json:"requirement,omitempty"`
}

func (r Report) Valid() bool {
	for _, diagnostic := range r.Diagnostics {
		if diagnostic.Level == LevelError {
			return false
		}
	}
	return true
}

type Requirement struct {
	Marker    string
	Title     string
	Line      int
	Body      string
	Scenarios []Scenario
}

type Scenario struct {
	Title   string
	Line    int
	HasWhen bool
	HasThen bool
}

type Delta struct {
	Capability   string
	Path         string
	Requirements []Requirement
}
