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
	mode  string
	root  string
	debug bool
	help  bool
	modes = []string{"filesystem", "database", "index"}
)

func init() {
	var (
		defaultMode = "filesystem"
		defaultHelp = false
		debugHelp   = false

		modeUsage  = "specifies what mode to run. [" + strings.Join(modes, ", ") + "]"
		pathUsage  = "specifies the root path.(not used in databse mode) (default: current working directory)"
		helpUsage  = "print this help message"
		debugUsage = "used by the developer to access debug features"
	)

	flag.StringVar(&mode, "m", defaultMode, modeUsage)
	flag.StringVar(&root, "p", "", pathUsage)
	flag.BoolVar(&help, "h", defaultHelp, helpUsage)
	flag.BoolVar(&debug, "d", debugHelp, debugUsage)

}

func main() {
	// parse command line arguments
	flag.Parse()

	defaultPath, _ := os.Getwd()
	if root == "" {
		root = defaultPath
	} else {
		root = path.Clean(path.Join(defaultPath, root))
	}

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
		println("chosen path: " + root)
		println("specified path does not exist\n")
		flag.PrintDefaults()
		return
	}

	// index filesystem at specified path
	if mode == "index" {
		database.Index(root)
		return
	}

	if debug {
		println("debug mode")
		folder := database.GetFolderByID(1)
		tracks := database.GetTracksByFolderID(folder.Id)
		for _, track := range tracks {
			println(" - " + track.Title.String)
		}
		return
	}

	// initialize audio player
	go audioplayer.Controller()

	// start user interface
	// (on current thread as to not immediately exit)
	tui.Start(root)

}
