package main

import (
	"flag"
	"mp3bak2/audioplayer"
	"mp3bak2/globals"
	"mp3bak2/tui"
	"os"
	"strings"
)

var (
	mode  string
	path  string
	help  bool
	modes = []string{"filesystem", "database", "index"}
)

func init() {
	var (
		defaultPath, _ = os.Getwd()
		defaultMode    = "filesystem"
		defaultHelp    = false

		modeUsage = "specifies what mode to run. [" + strings.Join(modes, ", ") + "]"
		pathUsage = "specifies the root path.(not used in databse mode) (default: current working directory)"
		helpUsage = "print this help message"
	)

	flag.StringVar(&mode, "m", defaultMode, modeUsage)
	flag.StringVar(&path, "p", defaultPath, pathUsage)
	flag.BoolVar(&help, "h", defaultHelp, helpUsage)

}

func main() {
	// parse command line arguments
	flag.Parse()

	// check for help flag
	if help {
		flag.PrintDefaults()
		return
	}

	// check if mode is correct
	if !globals.Contains(modes, mode) {
		println("please use one of the availble modes\n")
		flag.PrintDefaults()
		return
	}

	// check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		println("specified path does not exist\n")
		flag.PrintDefaults()
		return
	}

	// initialize audio player
	go audioplayer.Controller()

	// start user interface
	// (on current thread as to not immediately exit)
	tui.Start()

}
