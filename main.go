package main

import (
	"flag"
	"log"
	"os"

	"github.com/sisisin/audio_tools/src/converttomacvlcplaylist"
	"github.com/sisisin/audio_tools/src/syncfilestoandroid"
)

func main() {
	if len(os.Args) < 2 {
		log.Println("No subcommand specified")
		os.Exit(1)
	}

	config := flag.String("config", "at.config.yaml", "config file")
	flag.Parse()
	subCommand := flag.Arg(0)

	if subCommand == "" {
		log.Println("No subcommand specified")
		os.Exit(1)
	}

	switch subCommand {
	case "convertPlaylist":
		converttomacvlcplaylist.Run(*config)
	case "syncToAndroid":
		syncfilestoandroid.Run(*config)
	default:
		log.Println("No subcommand matched")
	}
}
