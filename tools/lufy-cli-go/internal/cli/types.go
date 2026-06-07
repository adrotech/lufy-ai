package cli

import "io"

const (
	ExitOK         = 0
	ExitRuntimeErr = 1
	ExitUsageErr   = 2
)

type Dependencies struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type InstallOptions struct {
	Target   string
	DryRun   bool
	Yes      bool
	NoEngram bool
	Backup   bool
}

type VerifyOptions struct {
	Target   string
	NoEngram bool
}

type BackupOptions struct {
	Target string
}

type RestoreOptions struct {
	Target string
	Backup string
	DryRun bool
	Yes    bool
}

type ActionableError struct {
	Message string
	Hint    string
	Code    int
}

func (e ActionableError) Error() string {
	if e.Hint == "" {
		return e.Message
	}
	return e.Message + "\nSugerencia: " + e.Hint
}
