package syncplaylistfiles

import (
	"fmt"
	"io"
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
}

func (c SyncPlaylistFilesConfig) MatchService() bool {
	return c.AtService == "sync_playlist_files"
}

func Run(configPath string) {
	config := lib.Load[SyncPlaylistFilesConfig](configPath)

	playlist, err := readPlaylist(config.SourcePlaylist, config.SourceBaseDir)
	if err != nil {
		panic(err)
	}
	destDirFiles, err := readDestDir(config.DestDir)
	if err != nil {
		panic(err)
	}

	{
		copyTargets := make([][]string, 0)
		for p := range playlist {
			destPath := path.Join(config.DestDir, strings.Replace(p, config.SourceBaseDir, "", 1))
			if !destDirFiles[destPath] {
				copyTargets = append(copyTargets, []string{p, destPath})
			}
		}

		fmt.Println("Syncing files", len(copyTargets), "files")
		for _, paths := range copyTargets {
			err := copyFile(paths[0], paths[1])
			if err != nil {
				panic(err)
			}
		}
	}
	{
		deleteTargets := make([]string, 0)
		for destPath := range destDirFiles {
			if !playlist[path.Join(config.SourceBaseDir, strings.Replace(destPath, config.DestDir, "", 1))] {
				deleteTargets = append(deleteTargets, destPath)
			}
		}

		fmt.Println("Deleting files", len(deleteTargets), "files")
		for _, p := range deleteTargets {
			err := os.Remove(p)
			if err != nil {
				panic(err)
			}
		}
	}
}

func copyFile(src, dest string) error {
	err := os.MkdirAll(filepath.Dir(dest), os.ModePerm)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	err = destFile.Sync()
	return err
}

func readDestDir(destDir string) (map[string]bool, error) {
	paths := make(map[string]bool)
	err := os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		paths[path] = true
		return nil
	})
	if err != nil {
		return nil, err
	}
	return paths, nil
}

func readPlaylist(playlistPath, sourceBaseDir string) (map[string]bool, error) {
	f, err := os.ReadFile(playlistPath)
	if err != nil {
		return nil, err
	}
	paths := make(map[string]bool)
	for _, line := range strings.Split(string(f), "\n") {
		fmt.Println(line)
		if strings.HasPrefix(line, sourceBaseDir) {
			paths[line] = true
		}
	}

	return paths, nil
}
