package syncplaylistfiles

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sisisin/audio_tools/src/lib"
)

func instantiateSyncClient(config SyncPlaylistFilesConfig) (syncClient, error) {
	switch config.Mode {
	case "local":
		return &syncClientLocal{}, nil
	case "adb":
		return &syncClientAdb{}, nil
	default:
		return nil, fmt.Errorf("unsupported mode: %s", config.Mode)
	}
}

type syncClientLocal struct{}

func (*syncClientLocal) CopyFile(ctx context.Context, src, dest string) error {
	if lib.IsDryRun(ctx) {
		fmt.Println("CopyFile", src, dest)
		return nil
	}

	err := os.MkdirAll(filepath.Dir(dest), os.ModePerm)
	if err != nil {
		return errors.Wrapf(err, "failed to create dir: %s", filepath.Dir(dest))
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return errors.Wrapf(err, "failed to open file: %s", src)
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return errors.Wrapf(err, "failed to create file: %s", dest)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return errors.Wrapf(err, "failed to copy file: %s -> %s", src, dest)
	}

	err = destFile.Sync()
	if err != nil {
		return errors.Wrapf(err, "failed to sync file: %s", dest)
	}

	return nil
}
func (*syncClientLocal) ReadDestDir(ctx context.Context, destDir string) (map[string]bool, error) {
	config := getConfig(ctx)
	paths := make(map[string]bool)
	err := os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create dir: %s", destDir)
	}

	err = filepath.Walk(destDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "failed to walk: %s", p)
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(config.DestDir, p)
		if err != nil {
			return errors.Wrapf(err, "failed to get relative path: %s", p)
		}
		paths[filepath.ToSlash(rel)] = true
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to walk: %s", destDir)
	}
	return paths, nil
}

func (*syncClientLocal) RemoveFile(ctx context.Context, path string) error {
	if lib.IsDryRun(ctx) {
		fmt.Println("RemoveFile", path)
		return nil
	}
	err := os.Remove(path)
	if err != nil {
		return errors.Wrapf(err, "failed to remove file: %s", path)
	}
	return nil
}

func (*syncClientLocal) ToDestPath(ctx context.Context, p string) string {
	config := getConfig(ctx)
	destPath := filepath.Join(config.DestDir, strings.Replace(p, config.SourceBaseDir, "", 1))

	return destPath
}
