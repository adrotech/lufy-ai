package installer

import (
	"fmt"
	"io"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

func printPlan(plan Plan, noEngram bool, stdout io.Writer) {
	fmt.Fprintf(stdout, "Plan de instalación para %s\n", plan.TargetRoot)
	fmt.Fprintf(stdout, "Source root: %s\n", plan.SourceRoot)
	fmt.Fprintf(stdout, "Scope: %s projectRoot=%s", plan.Scope, plan.TargetRoot)
	if plan.GlobalRoot != "" {
		fmt.Fprintf(stdout, " globalRoot=%s", plan.GlobalRoot)
	}
	fmt.Fprintln(stdout)
	for _, action := range plan.Actions {
		fmt.Fprintf(stdout, "- [%s] %s (%s)", action.Kind, action.Target, action.Reason)
		if action.SourceHash != "" {
			fmt.Fprintf(stdout, " source=%s", shortHash(action.SourceHash))
		}
		if action.CurrentHash != "" {
			fmt.Fprintf(stdout, " current=%s", shortHash(action.CurrentHash))
		}
		fmt.Fprintln(stdout)
	}
	for _, conflict := range plan.Conflicts {
		fmt.Fprintf(stdout, "- [warn-conflict] %s (%s) current=%s source=%s\n", conflict.Path, conflict.Reason, shortHash(conflict.CurrentHash), shortHash(conflict.SourceHash))
	}
	if noEngram {
		fmt.Fprintln(stdout, "Engram: omitido por --no-engram")
	} else if path, ok := platform.ResolveEngram(false, platform.OSResolver{}); ok {
		fmt.Fprintf(stdout, "Engram: detectado en PATH (%s)\n", path)
	} else {
		fmt.Fprintln(stdout, "Engram: no encontrado en PATH (instalación base continúa)")
	}
}
