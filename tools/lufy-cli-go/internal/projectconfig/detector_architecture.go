package projectconfig

func detectArchitecture(root, surfaceType string) ArchitectureProfile {
	profile := DefaultArchitectureProfile(surfaceType)
	detected := detectedArchitectures(root, surfaceType)
	if len(detected) == 0 {
		return profile
	}
	profile.Detected = detected
	profile.Preferred = detected[0]
	profile.ReviewRequired = true
	profile.StructuralExpectations = defaultArchitectureStructuralExpectations(surfaceType, profile.Preferred)
	return profile
}

func detectedArchitectures(root, surfaceType string) []string {
	switch surfaceType {
	case "frontend":
		if existsAny(root, "src/features", "features") {
			return []string{"feature_driven"}
		}
	case "backend":
		return detectBackendArchitectures(root)
	case "fullstack":
		if existsAny(root, "src/features", "features") {
			return []string{"feature_driven_frontend"}
		}
	}
	return nil
}

func detectBackendArchitectures(root string) []string {
	var detected []string
	if existsAny(root, "internal/adapters", "adapters") && existsAny(root, "internal/ports", "ports") {
		detected = append(detected, "hexagonal")
	}
	if existsAny(root, "internal/domain", "domain") && existsAny(root, "internal/usecase", "internal/usecases", "internal/application", "usecases", "application") && existsAny(root, "internal/infrastructure", "infrastructure") {
		detected = append(detected, "clean_architecture")
	}
	if hasControllerServiceRepository(root) {
		detected = append(detected, "controller_service_repository")
	}
	return unique(detected)
}

func hasControllerServiceRepository(root string) bool {
	hasController := existsAny(root, "controllers", "controller", "api/controllers", "internal/controllers", "internal/controller", "src/controllers", "src/controller")
	hasService := existsAny(root, "services", "service", "internal/services", "internal/service", "src/services", "src/service")
	hasRepository := existsAny(root, "repositories", "repository", "repos", "internal/repositories", "internal/repository", "internal/repos", "src/repositories", "src/repository")
	return hasController && hasService && hasRepository
}
