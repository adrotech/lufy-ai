package projectconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func scanGo(root string) Stack {
	version := ""
	if data, err := os.ReadFile(filepath.Join(root, "go.mod")); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "go ") {
				version = strings.TrimSpace(strings.TrimPrefix(line, "go "))
				break
			}
		}
	}
	return Stack{ID: "go", Supported: true, Version: version, PackageManager: "go modules", Frameworks: []string{}, TestRunner: CommandConfig{Command: "go test ./...", CoverageCommand: "go test -coverprofile=coverage.out ./...", CoverageThreshold: 85}, Linter: CommandConfig{Command: goLinter(root), AutoFix: goLinterFix(root)}, Formatter: Formatter{Command: "gofmt -w", FileExtensions: []string{".go"}}, StaticAnalysis: CommandConfig{Command: "go vet ./..."}, AntiPatterns: []string{"mock.Anything", "gomonkey", "fmt.Println", "panic("}, ObservabilityLibs: detectInFiles(root, []string{"go.mod"}, []string{"go.opentelemetry.io", "datadog/dd-trace-go", "meli/fury_go-toolkit-otel"})}
}

func scanJS(root string) Stack {
	pkg := readPackageJSON(root)
	id := "javascript"
	if exists(root, "tsconfig.json") {
		id = "typescript"
	}
	frameworks := []string{}
	if hasDep(pkg, "react") {
		frameworks = append(frameworks, "react")
	}
	if hasDep(pkg, "next") {
		frameworks = append(frameworks, "react", "next")
	}
	if hasDepPrefix(pkg, "@remix-run/") {
		frameworks = append(frameworks, "react", "remix")
	}
	for _, fw := range []string{"vue", "svelte"} {
		if hasDep(pkg, fw) {
			frameworks = append(frameworks, fw)
		}
	}
	frameworks = unique(frameworks)
	test := "npm test"
	coverage := "npm test -- --coverage"
	for _, candidate := range []string{"vitest", "jest", "mocha"} {
		if hasDep(pkg, candidate) {
			test = candidate + " run"
			if candidate == "jest" || candidate == "mocha" {
				test = candidate
			}
			coverage = test + " --coverage"
			break
		}
	}
	linter := "TODO: eslint ."
	if hasDep(pkg, "eslint") || existsGlob(root, ".eslintrc*") {
		linter = "eslint ."
	}
	formatter := "TODO: prettier --write"
	if hasDep(pkg, "prettier") || existsGlob(root, ".prettierrc*") {
		formatter = "prettier --write"
	}
	static := ""
	if id == "typescript" {
		static = "tsc --noEmit"
	}
	return Stack{ID: id, Supported: true, PackageManager: jsPackageManager(root), Frameworks: frameworks, TestRunner: CommandConfig{Command: test, CoverageCommand: coverage, CoverageThreshold: 80}, Linter: CommandConfig{Command: linter, AutoFix: linter + " --fix"}, Formatter: Formatter{Command: formatter, FileExtensions: []string{".ts", ".tsx", ".js", ".jsx", ".json", ".md", ".css"}}, StaticAnalysis: CommandConfig{Command: static}, AntiPatterns: []string{"console.log", "// @ts-ignore", "as any", "useEffect(() => {}, [])"}, ObservabilityLibs: detectPackageLibs(pkg, []string{"@opentelemetry/api", "pino", "winston"})}
}

func scanPython(root string) Stack {
	text := readKnown(root, "pyproject.toml", "requirements.txt", "setup.py")
	test := "python -m unittest"
	if strings.Contains(text, "pytest") {
		test = "pytest"
	}
	linter := "TODO: ruff check"
	fix := "TODO: ruff check --fix"
	formatter := "TODO: ruff format"
	if strings.Contains(text, "ruff") {
		linter, fix, formatter = "ruff check", "ruff check --fix", "ruff format"
	} else if strings.Contains(text, "flake8") {
		linter = "flake8"
	} else if strings.Contains(text, "pylint") {
		linter = "pylint"
	}
	if strings.Contains(text, "black") && !strings.Contains(text, "ruff") {
		formatter = "black ."
	}
	static := "TODO: mypy ."
	if strings.Contains(text, "mypy") {
		static = "mypy ."
	}
	pm := "pip"
	if exists(root, "uv.lock") {
		pm = "uv"
	} else if exists(root, "poetry.lock") {
		pm = "poetry"
	}
	return Stack{ID: "python", Supported: true, PackageManager: pm, Frameworks: pythonFrameworks(text), TestRunner: CommandConfig{Command: test, CoverageCommand: "pytest --cov", CoverageThreshold: 80}, Linter: CommandConfig{Command: linter, AutoFix: fix}, Formatter: Formatter{Command: formatter, FileExtensions: []string{".py"}}, StaticAnalysis: CommandConfig{Command: static}, AntiPatterns: []string{"print(", "except:", "MagicMock()", "unittest.skip"}, ObservabilityLibs: containsAny(text, []string{"opentelemetry", "structlog"})}
}

func scanJVM(root string) Stack {
	id := "java"
	pm := "maven"
	test := "mvn test"
	coverage := "mvn test jacoco:report"
	lint := "mvn checkstyle:check"
	format := "mvn spotless:apply"
	static := "mvn spotbugs:check"
	if existsAny(root, "build.gradle", "build.gradle.kts") {
		pm, test, coverage, lint, format, static = "gradle", "gradle test", "gradle test jacocoTestReport", "gradle check", "gradle spotlessApply", "gradle spotbugsMain"
		if exists(root, "build.gradle.kts") {
			id = "kotlin"
		}
	}
	text := readKnown(root, "pom.xml", "build.gradle", "build.gradle.kts")
	frameworks := []string{}
	if strings.Contains(text, "spring-boot") || strings.Contains(text, "springframework.boot") {
		frameworks = append(frameworks, "spring-boot")
	}
	return Stack{ID: id, Supported: true, PackageManager: pm, Frameworks: frameworks, TestRunner: CommandConfig{Command: test, CoverageCommand: coverage, CoverageThreshold: 80}, Linter: CommandConfig{Command: lint}, Formatter: Formatter{Command: format, FileExtensions: []string{".java", ".kt"}}, StaticAnalysis: CommandConfig{Command: static}, AntiPatterns: []string{"System.out.println", "e.printStackTrace()", "@SuppressWarnings(\"unchecked\")"}, ObservabilityLibs: containsAny(text, []string{"io.opentelemetry", "io.micrometer", "logback", "log4j"})}
}

func unsupportedStack(id, pm, ext, notes string) Stack {
	return Stack{ID: id, Supported: false, PackageManager: pm, Frameworks: []string{}, TestRunner: CommandConfig{Command: "TODO", CoverageCommand: "TODO", CoverageThreshold: 80}, Linter: CommandConfig{Command: "TODO"}, Formatter: Formatter{Command: "TODO", FileExtensions: []string{ext}}, StaticAnalysis: CommandConfig{}, AntiPatterns: []string{}, ObservabilityLibs: []string{}, Notes: notes}
}

func scanUnsupportedStacks(root string) []Stack {
	var stacks []Stack
	notes := "Soporte oficial pendiente. Completar manualmente o esperar release."
	if exists(root, "composer.json") {
		stacks = append(stacks, unsupportedStack("php", "composer", ".php", notes))
	}
	if exists(root, "Gemfile") {
		stacks = append(stacks, unsupportedStack("ruby", "bundler", ".rb", notes))
	}
	if exists(root, "mix.exs") {
		stacks = append(stacks, unsupportedStack("elixir", "mix", ".exs", notes))
	}
	if existsGlob(root, "*.csproj") || existsGlob(root, "*.sln") {
		stacks = append(stacks, unsupportedStack("dotnet", "dotnet", ".cs", notes))
	}
	return stacks
}

func readPackageJSON(root string) map[string]string {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return map[string]string{}
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return map[string]string{}
	}
	deps := map[string]string{}
	for _, key := range []string{"dependencies", "devDependencies"} {
		if section, ok := raw[key].(map[string]any); ok {
			for name := range section {
				deps[name] = key
			}
		}
	}
	if section, ok := raw["scripts"].(map[string]any); ok {
		for _, value := range section {
			command, ok := value.(string)
			if !ok {
				continue
			}
			for _, tool := range []string{"vitest", "jest", "mocha", "eslint", "prettier"} {
				if strings.Contains(command, tool) {
					deps[tool] = "scripts"
				}
			}
		}
	}
	return deps
}

func readPackageScripts(root string) map[string]string {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return map[string]string{}
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return map[string]string{}
	}
	scripts := map[string]string{}
	if section, ok := raw["scripts"].(map[string]any); ok {
		for name, value := range section {
			command, ok := value.(string)
			if ok {
				scripts[name] = command
			}
		}
	}
	return scripts
}
