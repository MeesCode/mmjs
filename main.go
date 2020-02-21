// Package main is the main starting point of the program. It parses
// the command line arguments and initializes the mysql connection pool.
package main

import (
	"flag"
	"fmt"
	"log"
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

		modeUsage = "specifies what mode to run. [" + strings.Join(modes, ", ") + "]"
		helpUsage = "print this help message"
	)

	flag.StringVar(&mode, "m", defaultMode, modeUsage)
	flag.BoolVar(&help, "h", defaultHelp, helpUsage)
}

func main() {

	// parse command line arguments
	flag.Parse()

	base, err := os.Getwd()

	if err != nil {
		log.Fatalln("could not get working directory", err)
	}

	arg := flag.Arg(0)

	// check that a path has been given
	if arg == "" {
		fmt.Println("please specify a path")
		flag.PrintDefaults()
		return
	}

	// make absolute if not already
	if path.IsAbs(arg) {
		globals.Root = path.Clean(arg)
	} else {
		globals.Root = path.Clean(path.Join(base, arg))
	}

	// check for help flag
	if help {
		flag.PrintDefaults()
		return
	}

	// check if mode is correct
	if !globals.Contains(modes, mode) {
		fmt.Println("please use one of the availble modes")
		flag.PrintDefaults()
		return
	}

	// check if path exists
	if _, err := os.Stat(globals.Root); os.IsNotExist(err) {
		fmt.Println("chosen path: " + globals.Root)
		fmt.Println("specified path does not exist")
		flag.PrintDefaults()
		return
	}

	// start the databse connection pool
	if mode != "filesystem" {
		db := database.Warmup()
		defer db.Close()
	}

	// index filesystem at specified path
	if mode == "index" {
		database.Index()
		return
	}

	// initialize audio player
	audioplayer.Initialize()

	// start user interface
	// (on current thread as to not immediately exit)
	tui.Start(mode)

}
