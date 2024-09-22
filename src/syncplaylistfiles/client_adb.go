package syncplaylistfiles

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	pkgerrors "github.com/pkg/errors"

	"github.com/sisisin/audio_tools/src/lib"
)

type syncClientAdb struct{}

func (*syncClientAdb) CopyFile(ctx context.Context, src, dest string) error {
	if lib.IsDryRun(ctx) {
		fmt.Println("CopyFile", src, dest)
		return nil
	}
	_, err := runCommand("adb", []string{"push", src, dest})
	return err
}

func (*syncClientAdb) RemoveFile(ctx context.Context, path string) error {
	if lib.IsDryRun(ctx) {
		fmt.Println("RemoveFile", path)
		return nil
	}
	_, err := runCommand("adb", []string{"shell", fmt.Sprintf("rm '%s'", path)})
	return err
}

type memoryWriter struct {
	data []byte
}

func (m *memoryWriter) Write(p []byte) (n int, err error) {
	m.data = append(m.data, p...)
	return len(p), nil
}

func (m *memoryWriter) String() string {
	return string(m.data)
}

func runCommand(cmd string, args []string) (string, error) {
	c := exec.Command(cmd, args...)
	outWriter := &memoryWriter{}
	errWriter := &memoryWriter{}

	c.Stdout = outWriter
	c.Stderr = errWriter
	if err := c.Run(); err != nil {
		nl := lib.GetNewline()
		return outWriter.String(), pkgerrors.Wrapf(
			err,
			"command error occuerd%s  command: %s %s%s  stdout: %s%s  stderr: %s",
			nl,
			cmd, strings.Join(args, " "), nl,
			outWriter.String(), nl,
			errWriter.String(),
		)
	}

	return outWriter.String(), nil
}

func (*syncClientAdb) ReadDestDir(ctx context.Context, destDir string) (map[string]bool, error) {
	config := getConfig(ctx)
	targetDirs := []string{destDir}
	files := make(map[string]bool)

	for len(targetDirs) > 0 {
		dir := targetDirs[0]
		targetDirs = targetDirs[1:]

		ls, err := runCommand("adb", []string{"shell", fmt.Sprintf("ls -p1 '%s'", dir)})
		if err != nil {
			return nil, err
		}
		for _, line := range strings.Split(ls, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			entry := filepath.ToSlash(filepath.Join(dir, line))
			if strings.HasSuffix(line, "/") {
				targetDirs = append(targetDirs, entry)
			} else {
				rel, err := filepath.Rel(config.DestDir, entry)
				if err != nil {
					return nil, pkgerrors.Wrapf(err, "failed to get relative path: %s", entry)
				}
				files[filepath.ToSlash(rel)] = true
			}
		}
	}

	return files, nil
}

func (*syncClientAdb) ToDestPath(ctx context.Context, p string) string {
	config := getConfig(ctx)
	destPath := filepath.ToSlash(filepath.Join(config.DestDir, filepath.ToSlash(strings.Replace(p, config.SourceBaseDir, "", 1))))

	return destPath
}
