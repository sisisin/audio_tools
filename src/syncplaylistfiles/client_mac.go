package syncplaylistfiles

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func instantiateSyncClient(config SyncPlaylistFilesConfig) (syncClient, error) {
	switch config.Mode {
	case "mac":
		return &syncClientMac{}, nil
	case "adb":
		return &syncClientAdb{}, nil
	default:
		return nil, fmt.Errorf("unsupported mode: %s", config.Mode)
	}
}

type syncClientMac struct{}

func (*syncClientMac) CopyFile(src, dest string) error {
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
func (*syncClientMac) ReadDestDir(ctx context.Context, destDir string) (map[string]bool, error) {
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

func (*syncClientMac) RemoveFile(path string) error {
	return os.Remove(path)
}
