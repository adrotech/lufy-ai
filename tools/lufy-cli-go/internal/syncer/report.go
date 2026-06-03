package syncer

import (
	"fmt"
	"io"
)

func printPlan(plan Plan, noEngram bool, stdout io.Writer) {
	fmt.Fprintf(stdout, "Plan de sync para %s\n", plan.TargetRoot)
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
		if action.RecordedSourceHash != "" {
			fmt.Fprintf(stdout, " recordedSource=%s", shortHash(action.RecordedSourceHash))
		}
		if action.RecordedTargetHash != "" {
			fmt.Fprintf(stdout, " recordedTarget=%s", shortHash(action.RecordedTargetHash))
		}
		fmt.Fprintln(stdout)
	}
	for _, conflict := range plan.Conflicts {
		fmt.Fprintf(stdout, "- [conflict] %s (%s) source=%s current=%s recordedSource=%s recordedTarget=%s\n", conflict.Path, conflict.Reason, shortHash(conflict.SourceHash), shortHash(conflict.CurrentHash), shortHash(conflict.RecordedSourceHash), shortHash(conflict.RecordedTargetHash))
	}
	if noEngram {
		fmt.Fprintln(stdout, "Engram: omitido por --no-engram")
	}
}
