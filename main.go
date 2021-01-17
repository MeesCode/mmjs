// Package main is the main starting point of the program. It parses
// the command line arguments and initializes the mysql connection pool.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/MeesCode/mmjs/audioplayer"
	"github.com/MeesCode/mmjs/database"
	"github.com/MeesCode/mmjs/globals"
	"github.com/MeesCode/mmjs/plugins"
	"github.com/MeesCode/mmjs/tui"
)

var (
	mode      string
	webserver bool
	port      int
	debug     bool
	help      bool
	modes     = []string{"filesystem", "database", "index"}
)

func init() {
	var (
		defaultMode      = "filesystem"
		defaultHelp      = false
		defaultWebserver = false
		defaultPort      = 8080

		modeUsage      = "specifies what mode to run. [" + strings.Join(modes, ", ") + "]"
		webserverUsage = "a boolean to specify whether to run the webserver. (only in database mode)"
		webserverPort  = "on which port the run the web server"
		helpUsage      = "print this help message"
	)

	flag.BoolVar(&help, "h", defaultHelp, helpUsage)
	flag.StringVar(&mode, "m", defaultMode, modeUsage)
	flag.BoolVar(&webserver, "w", defaultWebserver, webserverUsage)
	flag.IntVar(&port, "p", defaultPort, webserverPort)
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

	// index filesystem at specified path
	if mode == "index" {
		database.Index()
		return
	}

	// start the databse connection pool
	if mode != "filesystem" {
		db := database.Warmup()
		defer db.Close()
	}

	// initialize audio player
	go audioplayer.Initialize()

	////////////////////////////////
	//     Start plugins here     //
	////////////////////////////////
	if webserver && mode == "database" {
		go plugins.Webserver(port)
	}

	// start user interface
	// (on current thread as to not immediately exit)
	tui.Start(mode)

}
