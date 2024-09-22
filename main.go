package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/sisisin/audio_tools/src/converttomacvlcplaylist"
	"github.com/sisisin/audio_tools/src/lib"
	"github.com/sisisin/audio_tools/src/syncplaylistfiles"
)

func main() {
	ctx := context.Background()
	if len(os.Args) < 2 {
		log.Println("No subcommand specified")
		os.Exit(1)
	}

	config := flag.String("config", "at.config.yaml", "config file")
	verbose := flag.Bool("verbose", false, "verbose output")
	flag.Parse()
	ctx = lib.WithVerboseFlag(ctx, *verbose)

	subCommand := flag.Arg(0)

	if subCommand == "" {
		log.Println("No subcommand specified")
		os.Exit(1)
	}

	switch subCommand {
	case "convertPlaylist":
		converttomacvlcplaylist.Run(*config)
	case "syncPlaylist":
		syncplaylistfiles.Run(ctx, *config)
	default:
		log.Println("No subcommand matched")
	}
}
