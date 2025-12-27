package common

import (
	"fmt"
	"runtime"
)

var (
	Version   string
	BuildUser string
	BuildDate string
	CommitID  string
	GoVersion string
	GoOS      string
	GoArch    string
)

func getDefaultIfEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func GetVersion() string {
	version := getDefaultIfEmpty(Version, "0.3.4")
	buildUser := getDefaultIfEmpty(BuildUser, "unknown")
	buildDate := getDefaultIfEmpty(BuildDate, "unknown")
	commitID := getDefaultIfEmpty(CommitID, "unknown")
	goVersion := getDefaultIfEmpty(GoVersion, runtime.Version())
	goOS := getDefaultIfEmpty(GoOS, runtime.GOOS)
	goArch := getDefaultIfEmpty(GoArch, runtime.GOARCH)

	return fmt.Sprintf(
		"versão do kvm-compose: %s\n"+
			"commit ID: %s\n"+
			"build por: %s\n"+
			"data da versão: %s\n"+
			"Go: %s\n"+
			"GOOS: %s\n"+
			"GOARCH: %s\n",
		version, commitID, buildUser, buildDate, goVersion, goOS, goArch,
	)
}
