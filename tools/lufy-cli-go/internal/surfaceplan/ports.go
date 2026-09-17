package surfaceplan

type ProjectSource interface {
	Load(target string) (ProjectDefinition, error)
}

type ChangeSource interface {
	ChangedFiles(target, base string) ([]string, error)
}
