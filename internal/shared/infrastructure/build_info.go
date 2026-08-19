package infrastructure

import "runtime"

type BuildInfo struct {
	GoVersion string `json:"go_version"`
	GOOS      string `json:"goos"`
	GOARCH    string `json:"goarch"`
	Compiler  string `json:"compiler"`
}

func CurrentBuildInfo() BuildInfo {
	return BuildInfo{
		GoVersion: runtime.Version(),
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
		Compiler:  runtime.Compiler,
	}
}
