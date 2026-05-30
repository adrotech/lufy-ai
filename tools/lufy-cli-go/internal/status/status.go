package status

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

type Options struct {
	Target  string
	JSON    bool
	Verbose bool
	Scope   assets.Scope
}

type Report struct {
	OK                    bool          `json:"ok"`
	TargetRoot            string        `json:"targetRoot"`
	Scope                 string        `json:"scope,omitempty"`
	GlobalRoot            string        `json:"globalRoot,omitempty"`
	Installed             bool          `json:"installed"`
	SchemaVersion         int           `json:"schemaVersion,omitempty"`
	ToolVersion           string        `json:"toolVersion,omitempty"`
	ToolCommit            string        `json:"toolCommit,omitempty"`
	ToolBuildDate         string        `json:"toolBuildDate,omitempty"`
	Tool                  string        `json:"tool,omitempty"`
	MethodologyByTier     any           `json:"methodologyByTier,omitempty"`
	SourceRootFingerprint string        `json:"sourceRootFingerprint,omitempty"`
	InstalledAt           string        `json:"installedAt,omitempty"`
	UpdatedAt             string        `json:"updatedAt,omitempty"`
	Assets                int           `json:"assets"`
	Missing               int           `json:"missing"`
	Drifted               int           `json:"drifted"`
	Errors                int           `json:"errors"`
	AssetDetails          []AssetDetail `json:"assetDetails,omitempty"`
}

type AssetDetail struct {
	TargetRel         string `json:"targetRel"`
	Status            string `json:"status"`
	Policy            string `json:"policy,omitempty"`
	Scope             string `json:"scope,omitempty"`
	Tool              string `json:"tool,omitempty"`
	Methodology       string `json:"methodology,omitempty"`
	Component         string `json:"component,omitempty"`
	Expected          string `json:"expected,omitempty"`
	Actual            string `json:"actual,omitempty"`
	LufyNew           string `json:"lufyNew,omitempty"`
	RecommendedAction string `json:"recommendedAction,omitempty"`
	Error             string `json:"error,omitempty"`
}

type Service struct{}

func NewService() Service {
	return Service{}
}

func (s Service) Run(opts Options, stdout io.Writer) error {
	report, err := s.Build(opts.Target, opts.Verbose, opts.Scope)
	if err != nil {
		return err
	}
	if opts.JSON {
		body, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(stdout, "%s\n", body)
		return err
	}
	if !report.Installed {
		fmt.Fprintf(stdout, "Status para %s: no instalado\n", report.TargetRoot)
		return nil
	}
	fmt.Fprintf(stdout, "Status para %s\n", report.TargetRoot)
	fmt.Fprintf(stdout, "Scope: %s", report.Scope)
	if report.GlobalRoot != "" {
		fmt.Fprintf(stdout, " globalRoot=%s", report.GlobalRoot)
	}
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "Instalado: sí schema=%d tool=%s\n", report.SchemaVersion, report.ToolVersion)
	fmt.Fprintf(stdout, "Assets gestionados: %d\n", report.Assets)
	fmt.Fprintf(stdout, "Drift local: %d\n", report.Drifted)
	fmt.Fprintf(stdout, "Faltantes: %d\n", report.Missing)
	fmt.Fprintf(stdout, "Errores: %d\n", report.Errors)
	fmt.Fprintf(stdout, "Última actualización: %s\n", report.UpdatedAt)
	if opts.Verbose {
		for _, detail := range report.AssetDetails {
			fmt.Fprintf(stdout, "- [%s] %s", detail.Status, detail.TargetRel)
			if detail.Expected != "" {
				fmt.Fprintf(stdout, " expected=%s", shortHash(detail.Expected))
			}
			if detail.Actual != "" {
				fmt.Fprintf(stdout, " actual=%s", shortHash(detail.Actual))
			}
			if detail.Error != "" {
				fmt.Fprintf(stdout, " error=%s", detail.Error)
			}
			fmt.Fprintln(stdout)
		}
	}
	return nil
}

func (s Service) Build(target string, verbose bool, rawScope assets.Scope) (Report, error) {
	resolved, err := platform.ResolveTargetPath(target)
	if err != nil {
		return Report{}, err
	}
	scope, err := assets.ParseScope(string(rawScope))
	if err != nil {
		return Report{}, err
	}
	report := Report{TargetRoot: resolved, OK: true, Scope: string(scope)}
	if scope == assets.ScopeGlobal || scope == assets.ScopeBoth {
		globalRoot, err := platform.ResolveOpenCodeConfigRoot()
		if err != nil {
			return Report{}, err
		}
		report.GlobalRoot = globalRoot
	}
	st, err := state.Load(resolved)
	if err != nil {
		return Report{}, err
	}
	if st == nil {
		report.Installed = false
		return report, nil
	}
	report.Installed = true
	report.SchemaVersion = st.SchemaVersion
	report.ToolVersion = st.ToolVersion
	report.ToolCommit = st.ToolCommit
	report.ToolBuildDate = st.ToolBuildDate
	report.Tool = string(st.Tool)
	report.MethodologyByTier = st.MethodologyByTier
	report.SourceRootFingerprint = st.SourceRootFingerprint
	report.InstalledAt = st.InstalledAt
	report.UpdatedAt = st.UpdatedAt
	report.Assets = len(st.Assets)
	for _, asset := range st.Assets {
		detail := AssetDetail{TargetRel: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Tool: asset.Tool, Methodology: asset.Methodology, Component: asset.Component, Expected: asset.TargetSHA256}
		path, err := platform.SafeJoin(resolved, asset.TargetRel)
		if err != nil {
			report.Errors++
			detail.Status = "error"
			detail.Error = err.Error()
			if verbose {
				report.AssetDetails = append(report.AssetDetails, detail)
			}
			continue
		}
		actual, err := assets.FileSHA256(path)
		if os.IsNotExist(err) {
			report.Missing++
			detail.Status = "missing"
			if verbose {
				report.AssetDetails = append(report.AssetDetails, detail)
			}
			continue
		}
		if err != nil {
			report.Errors++
			detail.Status = "error"
			detail.Error = err.Error()
			if verbose {
				report.AssetDetails = append(report.AssetDetails, detail)
			}
			continue
		}
		detail.Actual = actual
		if actual != asset.TargetSHA256 {
			if asset.Policy == string(assets.PolicyNoReplace) && regularStatusFile(path+".lufy-new") {
				detail.Status = "lufy-new"
				detail.LufyNew = asset.TargetRel + ".lufy-new"
				detail.RecommendedAction = "review-lufy-new"
				if verbose {
					report.AssetDetails = append(report.AssetDetails, detail)
				}
				continue
			}
			report.Drifted++
			detail.Status = "drift"
		} else {
			detail.Status = "ok"
		}
		if verbose {
			report.AssetDetails = append(report.AssetDetails, detail)
		}
	}
	report.OK = report.Missing == 0 && report.Drifted == 0 && report.Errors == 0
	return report, nil
}

func regularStatusFile(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0
}

func shortHash(hash string) string {
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}
