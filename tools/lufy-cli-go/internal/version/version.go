package version

import (
	"fmt"
	"runtime"
)

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

type Info struct {
	Version   string
	Commit    string
	BuildDate string
	GOOS      string
	GOARCH    string
	DevBuild  bool
}

func Current() Info {
	info := Info{Version: Version, Commit: Commit, BuildDate: BuildDate, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH}
	info.DevBuild = info.Version == "" || info.Version == "dev" || info.Commit == "" || info.Commit == "unknown" || info.BuildDate == "" || info.BuildDate == "unknown"
	if info.Version == "" {
		info.Version = "dev"
	}
	if info.Commit == "" {
		info.Commit = "unknown"
	}
	if info.BuildDate == "" {
		info.BuildDate = "unknown"
	}
	return info
}

func (i Info) String() string {
	label := i.Version
	if i.DevBuild {
		label = fmt.Sprintf("%s (development build)", label)
	}
	return fmt.Sprintf("lufy-ai %s\ncommit: %s\nbuildDate: %s\ngoos: %s\ngoarch: %s", label, i.Commit, i.BuildDate, i.GOOS, i.GOARCH)
}
