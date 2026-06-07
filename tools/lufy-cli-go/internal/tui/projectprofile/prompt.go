package projectprofile

import (
	"fmt"
	"io"
	"os"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	tea "github.com/charmbracelet/bubbletea"
)

type TerminalDetector func(io.Reader, io.Writer) bool

type ProgramRunner func(Model, io.Reader, io.Writer) (projectconfig.ProjectProfile, error)

type Options struct {
	Input      io.Reader
	Output     io.Writer
	IsTerminal TerminalDetector
	RunProgram ProgramRunner
}

func NewPrompt(opts Options) projectconfig.ProfilePrompt {
	return func(cfg projectconfig.ProjectConfig) (projectconfig.ProjectProfile, error) {
		input := opts.Input
		if input == nil {
			input = os.Stdin
		}
		output := opts.Output
		if output == nil {
			output = os.Stdout
		}
		isTerminal := opts.IsTerminal
		if isTerminal == nil {
			isTerminal = DefaultIsTerminal
		}
		if !isTerminal(input, output) {
			fmt.Fprintln(output, "project_profile: modo no interactivo; se conserva la detección automática.")
			return cfg.ProjectProfile, nil
		}
		runProgram := opts.RunProgram
		if runProgram == nil {
			runProgram = runBubbleTea
		}
		return runProgram(NewModel(cfg), input, output)
	}
}

func DefaultIsTerminal(input io.Reader, output io.Writer) bool {
	if os.Getenv("CI") != "" || os.Getenv("LUFY_NO_TUI") != "" {
		return false
	}
	inFile, inOK := input.(*os.File)
	outFile, outOK := output.(*os.File)
	if !inOK || !outOK {
		return false
	}
	inInfo, inErr := inFile.Stat()
	outInfo, outErr := outFile.Stat()
	if inErr != nil || outErr != nil {
		return false
	}
	return inInfo.Mode()&os.ModeCharDevice != 0 && outInfo.Mode()&os.ModeCharDevice != 0
}

func runBubbleTea(model Model, input io.Reader, output io.Writer) (projectconfig.ProjectProfile, error) {
	program := tea.NewProgram(model, tea.WithInput(input), tea.WithOutput(output))
	finalModel, err := program.Run()
	if err != nil {
		return projectconfig.ProjectProfile{}, fmt.Errorf("ejecutar TUI project_profile: %w", err)
	}
	result, ok := finalModel.(Model)
	if !ok {
		return projectconfig.ProjectProfile{}, fmt.Errorf("resultado TUI project_profile inválido: %T", finalModel)
	}
	return result.Result()
}
