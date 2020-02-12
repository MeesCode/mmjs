package main

import (
	"flag"
	"mp3bak2/audioplayer"
	"mp3bak2/database"
	"mp3bak2/globals"
	"mp3bak2/tui"
	"os"
	"path"
	"strings"
)

var (
	mode           string
	root           string
	help           bool
	modes          = []string{"filesystem", "database", "index"}
	defaultPath, _ = os.Getwd()
)

func init() {
	var (
		defaultMode = "filesystem"
		defaultHelp = false

		modeUsage = "specifies what mode to run. [" + strings.Join(modes, ", ") + "]"
		pathUsage = "specifies the root path.(not used in databse mode) (default: current working directory)"
		helpUsage = "print this help message"
	)

	flag.StringVar(&mode, "m", defaultMode, modeUsage)
	flag.StringVar(&root, "p", defaultPath, pathUsage)
	flag.BoolVar(&help, "h", defaultHelp, helpUsage)

}

func main() {
	// parse command line arguments
	flag.Parse()

	root = path.Clean(path.Join(defaultPath, root))

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
	if _, err := os.Stat(root); os.IsNotExist(err) {
		println("specified path does not exist\n")
		flag.PrintDefaults()
		return
	}

	// index filesystem at specified path
	if mode == "index" {
		database.Index(root)
		return
	}

	// initialize audio player
	go audioplayer.Controller()

	// start user interface
	// (on current thread as to not immediately exit)
	tui.Start(root)

}
