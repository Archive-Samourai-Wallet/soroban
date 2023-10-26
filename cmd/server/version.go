package main

import (
	"fmt"
	"os"
	"runtime"
)

var (
	Commit  string
	Date    string
	Version string
)

func printVersionExit() {
	fmt.Printf("Version: %s\nCommit: %s\nRuntime: %s %s/%s\nDate: %s\n",
		Version,
		Commit,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		Date,
	)

	os.Exit(0)
}
