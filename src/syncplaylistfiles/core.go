package syncplaylistfiles

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sisisin/audio_tools/src/lib"
)

type SyncPlaylistFilesConfig struct {
	AtService      string `yaml:"at_service"`
	SourcePlaylist string `yaml:"source_playlist"`
	DestDir        string `yaml:"dest_dir"`
	SourceBaseDir  string `yaml:"source_base_dir"`
	Mode           string `yaml:"mode"`
}

func (c SyncPlaylistFilesConfig) MatchService() bool {
	return c.AtService == "sync_playlist_files"
}

type contextKey string

const configKey = contextKey("sync_playlist_files")

func withConfig(ctx context.Context, config SyncPlaylistFilesConfig) context.Context {
	return context.WithValue(ctx, configKey, config)
}
func getConfig(ctx context.Context) SyncPlaylistFilesConfig {
	c, ok := ctx.Value(configKey).(SyncPlaylistFilesConfig)
	if !ok {
		panic("config not found")
	}
	return c
}

type syncClient interface {
	CopyFile(ctx context.Context, src, dest string) error
	RemoveFile(ctx context.Context, path string) error
	ReadDestDir(ctx context.Context, destDir string) (map[string]bool, error)

	ToDestPath(ctx context.Context, path string) string
}

func Run(ctx context.Context, configPath string) error {
	verbose := lib.IsVerbose(ctx)
	config := lib.Load[SyncPlaylistFilesConfig](configPath)
	ctx = withConfig(ctx, config)
	client, err := instantiateSyncClient(config)
	if err != nil {
		return errors.Wrap(err, "failed to instantiate sync client")
	}
	playlist, err := readPlaylist(config.SourcePlaylist, config.SourceBaseDir)
	if err != nil {
		return errors.Wrap(err, "failed to read playlist")
	}
	if verbose {
		fmt.Println("====================")
		fmt.Printf("playlist: %v", playlist)
		fmt.Println("")
		fmt.Println("--------------------")
	}

	destDirFiles, err := client.ReadDestDir(ctx, config.DestDir)
	if err != nil {
		return errors.Wrap(err, "failed to read dest dir")
	}
	if verbose {
		fmt.Printf("destDirFiles: %v", destDirFiles)
		fmt.Println("")
		fmt.Println("====================")
	}

	{
		copyTargets := make([][]string, 0)
		for p := range playlist {
			if !destDirFiles[p] {
				sourcePath := filepath.Join(config.SourceBaseDir, p)
				copyTargets = append(copyTargets, []string{sourcePath, client.ToDestPath(ctx, p)})
			}
		}

		fmt.Println("Syncing files", len(copyTargets), "files")
		if verbose {
			fmt.Println("files to copy", copyTargets)
		}
		for _, paths := range copyTargets {
			err := client.CopyFile(ctx, paths[0], paths[1])
			if err != nil {
				log.Fatalf("%+v", err)
			}
		}
	}
	{
		deleteTargets := make([]string, 0)
		for p := range destDirFiles {
			if !playlist[p] {
				deleteTargets = append(deleteTargets, client.ToDestPath(ctx, p))
			}
		}

		fmt.Println("Deleting files", len(deleteTargets), "files")
		if verbose {
			fmt.Println("files to delete", deleteTargets)
		}
		for _, p := range deleteTargets {
			err := client.RemoveFile(ctx, p)
			if err != nil {
				return errors.Wrap(err, "failed to remove file")
			}
		}
	}
	return nil
}

func readPlaylist(playlistPath, sourceBaseDir string) (map[string]bool, error) {
	f, err := os.ReadFile(playlistPath)
	if err != nil {
		return nil, err
	}
	paths := make(map[string]bool)
	for _, line := range strings.Split(string(f), "\n") {
		line := strings.TrimSpace(line)
		if strings.HasPrefix(line, sourceBaseDir) {
			rel, err := filepath.Rel(sourceBaseDir, line)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get relative path: %s", line)
			}
			normalized := filepath.ToSlash(rel)
			paths[normalized] = true
		}
	}

	return paths, nil
}
