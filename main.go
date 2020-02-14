package main

import (
	"flag"
	"fmt"
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
		helpUsage  = "print this help message"
		debugUsage = "used by the developer to access debug features"
	)

	flag.StringVar(&mode, "m", defaultMode, modeUsage)
	flag.BoolVar(&help, "h", defaultHelp, helpUsage)
	flag.BoolVar(&debug, "d", debugHelp, debugUsage)

}

func main() {

	// parse command line arguments
	flag.Parse()

	base, _ := os.Getwd()
	arg := flag.Arg(0)
	var root string

	// check that a path has been given
	if arg == "" && mode != "database" {
		fmt.Println("please specify a path\n")
		flag.PrintDefaults()
		return
	}

	// make absolute if not already
	if path.IsAbs(arg) {
		root = path.Clean(arg)
	} else {
		root = path.Clean(path.Join(base, arg))
	}

	// check for help flag
	if help {
		flag.PrintDefaults()
		return
	}

	// check if mode is correct
	if !globals.Contains(modes, mode) {
		fmt.Println("please use one of the availble modes\n")
		flag.PrintDefaults()
		return
	}

	// check if path exists
	if _, err := os.Stat(root); os.IsNotExist(err) {
		fmt.Println("chosen path: " + root)
		fmt.Println("specified path does not exist\n")
		flag.PrintDefaults()
		return
	}

	// start the databse connection pool
	db := database.Warmup()
	defer db.Close()

	// index filesystem at specified path
	if mode == "index" {
		database.Index(root)
		return
	}

	if debug {
		fmt.Println("debug mode")
		folder := database.GetFolderByID(1)
		tracks := database.GetTracksByFolderID(folder.Id)
		for _, track := range tracks {
			fmt.Println(" - " + track.Title.String)
		}
		return
	}

	// initialize audio player
	go audioplayer.Init()

	// start user interface
	// (on current thread as to not immediately exit)
	tui.Start(root, mode)

}
