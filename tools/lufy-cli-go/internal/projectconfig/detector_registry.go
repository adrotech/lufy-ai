package projectconfig

type stackDetector struct {
	applies func(string) bool
	scan    func(string) []Stack
}

type surfaceDetector struct {
	matches func(Stack) bool
	detect  func(string, Stack) []ProjectSurface
}

func scanRootStacks(root string) []Stack {
	var stacks []Stack
	for _, detector := range defaultStackDetectors() {
		if !detector.applies(root) {
			continue
		}
		stacks = append(stacks, detector.scan(root)...)
	}
	return stacks
}

func defaultStackDetectors() []stackDetector {
	unsupportedNotes := "Soporte oficial pendiente. Completar manualmente o esperar release."
	return []stackDetector{
		{
			applies: func(root string) bool { return exists(root, "go.mod") },
			scan:    func(root string) []Stack { return []Stack{scanGo(root)} },
		},
		{
			applies: func(root string) bool { return exists(root, "package.json") },
			scan:    func(root string) []Stack { return []Stack{scanJS(root)} },
		},
		{
			applies: func(root string) bool { return existsAny(root, "pyproject.toml", "requirements.txt", "setup.py") },
			scan:    func(root string) []Stack { return []Stack{scanPython(root)} },
		},
		{
			applies: func(root string) bool { return existsAny(root, "pom.xml", "build.gradle", "build.gradle.kts") },
			scan:    func(root string) []Stack { return []Stack{scanJVM(root)} },
		},
		{
			applies: func(root string) bool { return exists(root, "Cargo.toml") },
			scan:    func(root string) []Stack { return []Stack{unsupportedStack("rust", "cargo", ".rs", unsupportedNotes)} },
		},
		{
			applies: func(string) bool { return true },
			scan:    scanUnsupportedStacks,
		},
	}
}

func detectStackSurfaces(root string, stacks []Stack) []ProjectSurface {
	var surfaces []ProjectSurface
	for _, stack := range stacks {
		for _, detector := range defaultSurfaceDetectors() {
			if !detector.matches(stack) {
				continue
			}
			surfaces = append(surfaces, detector.detect(root, stack)...)
		}
	}
	return surfaces
}

func defaultSurfaceDetectors() []surfaceDetector {
	return []surfaceDetector{
		{
			matches: func(stack Stack) bool { return stack.ID == "typescript" || stack.ID == "javascript" },
			detect:  func(root string, stack Stack) []ProjectSurface { return []ProjectSurface{detectJSSurface(root, stack)} },
		},
		{
			matches: func(stack Stack) bool { return stack.ID == "go" },
			detect:  func(root string, stack Stack) []ProjectSurface { return []ProjectSurface{detectGoSurface(root, stack)} },
		},
		{
			matches: func(stack Stack) bool { return stack.ID == "python" },
			detect: func(root string, stack Stack) []ProjectSurface {
				return []ProjectSurface{detectServerSurface("python-app", root, stack)}
			},
		},
		{
			matches: func(stack Stack) bool { return stack.ID == "java" || stack.ID == "kotlin" },
			detect: func(root string, stack Stack) []ProjectSurface {
				return []ProjectSurface{detectServerSurface(stack.ID+"-app", root, stack)}
			},
		},
	}
}
