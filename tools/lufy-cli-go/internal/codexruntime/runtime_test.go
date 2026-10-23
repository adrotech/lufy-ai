package codexruntime

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

const fixtureFeatures = "multi_agent experimental false\nhooks under development false\napps stable true\nplugins experimental false\n"

func fixtureAbsolute(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

func fixtureService(t *testing.T, probe func(context.Context, string, []string, string) ([]byte, error)) Service {
	t.Helper()
	service := NewService()
	service.LookPath = func(name string) (string, error) {
		if name == "codex" {
			return "/fixture/bin/codex", nil
		}
		if name == "/fixture/App Folder/codex" {
			return name, nil
		}
		return "", errors.New("fixture executable absent")
	}
	service.Run = probe
	return service
}

func validFixtureProbe(_ context.Context, _ string, args []string, _ string) ([]byte, error) {
	if reflect.DeepEqual(args, []string{"--version"}) {
		return []byte("warning: synthetic warning\ncodex-cli 0.163.0-alpha.10.2\n"), nil
	}
	return []byte(fixtureFeatures), nil
}

func TestInspectDefaultUsesPATHOnlyVersionAndFeaturesInTarget(t *testing.T) {
	var calls [][]string
	var sharedDeadline time.Time
	service := fixtureService(t, func(ctx context.Context, executable string, args []string, cwd string) ([]byte, error) {
		if executable != fixtureAbsolute(t, "/fixture/bin/codex") || cwd != "/fixture/target" {
			t.Fatalf("incorrect executable/cwd: %q %q", executable, cwd)
		}
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) <= 0 || time.Until(deadline) > 5*time.Second {
			t.Fatal("default probe must have a live deadline bounded by five seconds")
		}
		if !sharedDeadline.IsZero() && !deadline.Equal(sharedDeadline) {
			t.Fatal("version and feature probes must share one total timeout")
		}
		sharedDeadline = deadline
		calls = append(calls, append([]string(nil), args...))
		return validFixtureProbe(ctx, executable, args, cwd)
	})
	report := service.Inspect(context.Background(), Options{Target: "/fixture/target"})
	if report.SchemaVersion != "codex-runtime-report/v1" || report.Status != "available" || report.Executable != fixtureAbsolute(t, "/fixture/bin/codex") || report.PathExecutable != fixtureAbsolute(t, "/fixture/bin/codex") {
		t.Fatalf("unexpected default report: %#v", report)
	}
	if report.Selection != "path" || !strings.Contains(report.Version, "0.163.0-alpha.10.2") || strings.Contains(report.Version, "warning") {
		t.Fatalf("selection/version must reflect parsed evidence: %#v", report)
	}
	if !reflect.DeepEqual(calls, [][]string{{"--version"}, {"features", "list"}}) {
		t.Fatalf("only read-only version/features probes are permitted: %#v", calls)
	}
}

func TestInspectExplicitExecutablePassesLiteralPathAndCustomDeadline(t *testing.T) {
	var executables []string
	service := fixtureService(t, func(ctx context.Context, executable string, args []string, cwd string) ([]byte, error) {
		executables = append(executables, executable)
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) <= 0 || time.Until(deadline) > 200*time.Millisecond {
			t.Fatal("explicit timeout must bound each probe")
		}
		return validFixtureProbe(ctx, executable, args, cwd)
	})
	report := service.Inspect(context.Background(), Options{Executable: "/fixture/App Folder/codex", Target: "/fixture/target", Timeout: 200 * time.Millisecond})
	if report.Executable != fixtureAbsolute(t, "/fixture/App Folder/codex") || report.Selection != "explicit" {
		t.Fatalf("explicit executable selection lost: %#v", report)
	}
	if !reflect.DeepEqual(executables, []string{fixtureAbsolute(t, "/fixture/App Folder/codex"), fixtureAbsolute(t, "/fixture/App Folder/codex")}) {
		t.Fatalf("explicit executable must remain a literal path without shell: %#v", executables)
	}
}

func TestInspectMissingExecutableDoesNotRunAnything(t *testing.T) {
	service := fixtureService(t, func(context.Context, string, []string, string) ([]byte, error) {
		t.Fatal("missing executable must not invoke a runner")
		return nil, nil
	})
	service.LookPath = func(string) (string, error) { return "", errors.New("SYNTHETIC_PRIVATE_RESOLVER_ERROR") }
	report := service.Inspect(context.Background(), Options{})
	if report.Status != "unavailable" || report.Executable != "" || len(report.Reasons) == 0 || report.NativeDelegation != "unknown" {
		t.Fatalf("unavailable executable must remain observable without tool inference: %#v", report)
	}
	if strings.Contains(strings.Join(report.Reasons, " "), "SYNTHETIC_PRIVATE") {
		t.Fatal("resolver diagnostics leaked into reason codes")
	}
}

func TestInspectInvalidVersionPreservesAvailableFeaturesAsPartial(t *testing.T) {
	service := fixtureService(t, func(_ context.Context, _ string, args []string, _ string) ([]byte, error) {
		if reflect.DeepEqual(args, []string{"--version"}) {
			return []byte("warning: please install codex-cli 0.163.0-alpha.10.2 before running\n"), nil
		}
		return []byte(fixtureFeatures), nil
	})
	report := service.Inspect(context.Background(), Options{})
	if report.Status != "partial" || report.Version != "" || report.FeaturesStatus != "available" || len(report.Reasons) == 0 {
		t.Fatalf("prose is not version evidence; valid features remain independently available: %#v", report)
	}
}

func TestInspectFeatureProbeErrorIsPartialAndRedactsOutput(t *testing.T) {
	service := fixtureService(t, func(ctx context.Context, exe string, args []string, cwd string) ([]byte, error) {
		if reflect.DeepEqual(args, []string{"features", "list"}) {
			return []byte("SYNTHETIC_PRIVATE_OUTPUT"), errors.New("SYNTHETIC_PRIVATE_ERROR")
		}
		return validFixtureProbe(ctx, exe, args, cwd)
	})
	report := service.Inspect(context.Background(), Options{})
	if report.Status != "partial" || report.FeaturesStatus != "unavailable" || len(report.Features) != 0 || len(report.Reasons) == 0 {
		t.Fatalf("failed feature probe must not fabricate feature values: %#v", report)
	}
	if strings.Contains(strings.Join(report.Reasons, " "), "SYNTHETIC_PRIVATE") {
		t.Fatal("stderr or process output leaked into reason codes")
	}
}

func TestInspectWhitelistFeaturesDoesNotInferNativeDelegation(t *testing.T) {
	service := fixtureService(t, func(ctx context.Context, exe string, args []string, cwd string) ([]byte, error) {
		if reflect.DeepEqual(args, []string{"features", "list"}) {
			return []byte("multi_agent experimental true\nhooks under development false\napps stable true\nplugins experimental false\nprivate_feature stable true\n"), nil
		}
		return validFixtureProbe(ctx, exe, args, cwd)
	})
	report := service.Inspect(context.Background(), Options{})
	want := map[string]bool{"multi_agent": true, "hooks": false, "apps": true, "plugins": false}
	if report.Status != "available" || report.NativeDelegation != "unknown" || !reflect.DeepEqual(report.Features, want) {
		t.Fatalf("feature flags are neither tool inventory nor native delegation proof: %#v", report)
	}
}

func TestInspectDuplicateKnownFeatureNeverOverwritesEvidence(t *testing.T) {
	service := fixtureService(t, func(ctx context.Context, exe string, args []string, cwd string) ([]byte, error) {
		if reflect.DeepEqual(args, []string{"features", "list"}) {
			return []byte(fixtureFeatures + "apps stable false\n"), nil
		}
		return validFixtureProbe(ctx, exe, args, cwd)
	})
	report := service.Inspect(context.Background(), Options{})
	if report.Status != "partial" || report.FeaturesStatus != "unavailable" || len(report.Features) != 0 || len(report.Reasons) == 0 {
		t.Fatalf("duplicate feature evidence must not silently overwrite a flag: %#v", report)
	}
}

func TestInspectHonorsEarlierParentDeadlineAndReturnsCodesOnly(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	parentDeadline, _ := ctx.Deadline()
	service := fixtureService(t, func(ctx context.Context, _ string, _ []string, _ string) ([]byte, error) {
		deadline, ok := ctx.Deadline()
		if !ok || deadline.After(parentDeadline) {
			t.Fatal("probe timeout must never extend the caller deadline")
		}
		return []byte("SYNTHETIC_PRIVATE_TIMEOUT_OUTPUT"), context.DeadlineExceeded
	})
	report := service.Inspect(ctx, Options{Timeout: 10 * time.Second})
	if report.Status != "unavailable" || report.NativeDelegation != "unknown" || len(report.Reasons) == 0 {
		t.Fatalf("timeout cannot become a complete runtime report: %#v", report)
	}
	for _, reason := range report.Reasons {
		if reason == "" || strings.ContainsAny(reason, " \t\r\n") || strings.Contains(reason, "SYNTHETIC_PRIVATE") {
			t.Fatal("reasons must contain safe codes, not raw diagnostics")
		}
	}
}

func TestInspectCanceledCallerNeverExecutesProbe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service := fixtureService(t, func(context.Context, string, []string, string) ([]byte, error) {
		t.Fatal("already canceled caller must not launch a version or feature probe")
		return nil, nil
	})
	report := service.Inspect(ctx, Options{})
	if report.Status != "unavailable" || len(report.Reasons) == 0 {
		t.Fatalf("canceled inspection must report missing evidence: %#v", report)
	}
}

func TestInspectOversizedOutputIsRejectedWithoutExposingContent(t *testing.T) {
	for _, probe := range []string{"version", "features"} {
		t.Run(probe, func(t *testing.T) {
			service := fixtureService(t, func(ctx context.Context, exe string, args []string, cwd string) ([]byte, error) {
				isVersion := reflect.DeepEqual(args, []string{"--version"})
				if isVersion == (probe == "version") {
					return []byte(strings.Repeat("SYNTHETIC_PRIVATE_OVERSIZED", 3000)), nil
				}
				return validFixtureProbe(ctx, exe, args, cwd)
			})
			report := service.Inspect(context.Background(), Options{})
			if report.Status != "partial" || len(report.Reasons) == 0 {
				t.Fatalf("oversized probe must degrade complete evidence: %#v", report)
			}
			if strings.Contains(report.Version+strings.Join(report.Reasons, " "), "SYNTHETIC_PRIVATE") {
				t.Fatal("oversized content leaked into report")
			}
			if probe == "features" && (report.FeaturesStatus != "unavailable" || len(report.Features) != 0) {
				t.Fatalf("oversized features must not yield feature evidence: %#v", report)
			}
		})
	}
}

// El helper ejecuta sólo el testbin sintético; nunca invoca Codex.
func TestRuntimeBoundedHelperProcess(t *testing.T) {
	if os.Getenv("LUFY_RUNTIME_BOUNDED_TEST_HELPER") != "1" {
		return
	}
	switch os.Args[len(os.Args)-1] {
	case "timeout":
		time.Sleep(5 * time.Second)
	case "overflow":
		_, _ = os.Stdout.Write([]byte(strings.Repeat("x", maxOutput+4096)))
	default:
		os.Exit(8)
	}
	os.Exit(0)
}

func TestRuntimeBoundedRunnerRealProcessCancelAndOverflow(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("LUFY_RUNTIME_BOUNDED_TEST_HELPER", "1")
	for _, mode := range []string{"timeout", "overflow"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			started := time.Now()
			out, err := run(ctx, executable, []string{"-test.run=^TestRuntimeBoundedHelperProcess$", "--", mode}, t.TempDir())
			if err == nil {
				t.Fatalf("synthetic %s process must fail", mode)
			}
			if len(out) > maxOutput {
				t.Fatal("real runner exceeded its output cap")
			}
			if time.Since(started) > 2*time.Second {
				t.Fatal("process cleanup exceeded bounded deadline tolerance")
			}
		})
	}
}

func TestRuntimeRepeatedVersionsNeverSelectLastEvidence(t *testing.T) {
	for _, output := range []string{
		"codex-cli 0.163.0-alpha.10.2\ncodex-cli 0.144.5\n",
		"codex-cli 0.163.0-alpha.10.2\ncodex-cli 0.144.5\ncodex-cli 0.163.0-alpha.10.2\n",
	} {
		t.Run(output, func(t *testing.T) {
			service := fixtureService(t, func(ctx context.Context, exe string, args []string, cwd string) ([]byte, error) {
				if reflect.DeepEqual(args, []string{"--version"}) {
					return []byte(output), nil
				}
				return validFixtureProbe(ctx, exe, args, cwd)
			})
			report := service.Inspect(context.Background(), Options{})
			if report.Status != "partial" || report.Version != "" || report.FeaturesStatus != "available" || len(report.Reasons) == 0 {
				t.Fatalf("ambiguous repeated versions must stay partial: %#v", report)
			}
		})
	}
}
