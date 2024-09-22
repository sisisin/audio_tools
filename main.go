package main

import (
	"context"
	"flag"
	"fmt"
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
	dryRun := flag.Bool("dry-run", false, "dry run mode")
	flag.Parse()
	ctx = lib.WithVerboseFlag(ctx, *verbose)
	ctx = lib.WithDryRunFlag(ctx, *dryRun)

	subCommand := flag.Arg(0)

	if subCommand == "" {
		log.Println("No subcommand specified")
		os.Exit(1)
	}

	fmt.Printf("start %s\n", subCommand)
	switch subCommand {
	case "convertPlaylist":
		converttomacvlcplaylist.Run(*config)
	case "syncPlaylist":
		err := syncplaylistfiles.Run(ctx, *config)
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Println("No subcommand matched")
	}
}
