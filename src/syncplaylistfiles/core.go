package syncplaylistfiles

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

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
}

func Run(ctx context.Context, configPath string) {
	verbose := lib.IsVerbose(ctx)
	config := lib.Load[SyncPlaylistFilesConfig](configPath)
	ctx = withConfig(ctx, config)
	client, err := instantiateSyncClient(config)
	if err != nil {
		panic(err)
	}
	playlist, err := readPlaylist(config.SourcePlaylist, config.SourceBaseDir)
	if err != nil {
		panic(err)
	}
	destDirFiles, err := client.ReadDestDir(ctx, config.DestDir)
	if err != nil {
		panic(err)
	}

	{
		copyTargets := make([][]string, 0)
		for p := range playlist {
			if !destDirFiles[p] {
				destPath := path.Join(config.DestDir, filepath.ToSlash(strings.Replace(p, config.SourceBaseDir, "", 1)))
				sourcePath := filepath.Join(config.SourceBaseDir, p)
				copyTargets = append(copyTargets, []string{sourcePath, destPath})
			}
		}

		fmt.Println("Syncing files", len(copyTargets), "files")
		if verbose {
			fmt.Println("files to copy", copyTargets)
		}
		for _, paths := range copyTargets {
			err := client.CopyFile(ctx, paths[0], paths[1])
			if err != nil {
				panic(err)
			}
		}
	}
	{
		deleteTargets := make([]string, 0)
		for p := range destDirFiles {
			if !playlist[p] {
				destPath := path.Join(config.SourceBaseDir, filepath.ToSlash(strings.Replace(p, config.DestDir, "", 1)))
				deleteTargets = append(deleteTargets, destPath)
			}
		}

		fmt.Println("Deleting files", len(deleteTargets), "files")
		if verbose {
			fmt.Println("files to delete", deleteTargets)
		}
		for _, p := range deleteTargets {
			err := client.RemoveFile(ctx, p)
			if err != nil {
				panic(err)
			}
		}
	}
}

func readPlaylist(playlistPath, sourceBaseDir string) (map[string]bool, error) {
	f, err := os.ReadFile(playlistPath)
	if err != nil {
		return nil, err
	}
	paths := make(map[string]bool)
	for _, line := range strings.Split(string(f), "\n") {
		if strings.HasPrefix(line, sourceBaseDir) {
			normalized := filepath.ToSlash(strings.TrimSpace(strings.Replace(line, sourceBaseDir, "", 1)))
			paths[normalized] = true
		}
	}

	return paths, nil
}
