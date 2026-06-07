package cli

import (
	"os"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/tui/projectprofile"
)

func surfaceProfilePrompt(deps Dependencies) projectconfig.ProfilePrompt {
	input := deps.Stdin
	if input == nil {
		input = os.Stdin
	}
	output := deps.Stdout
	if output == nil {
		output = os.Stdout
	}
	return projectprofile.NewPrompt(projectprofile.Options{
		Input:  input,
		Output: output,
	})
}
