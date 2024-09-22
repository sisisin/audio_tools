package syncplaylistfiles

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

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
		return outWriter.String(), errors.Join(
			err,
			fmt.Errorf("command error occuerd\r\n  command: %s %s\r\n  stdout: %s\r\n  stderr: %s",
				cmd, strings.Join(args, " "), outWriter.String(), errWriter.String()),
		)
	}

	return outWriter.String(), nil
}

func (*syncClientAdb) ReadDestDir(ctx context.Context, destDir string) (map[string]bool, error) {
	verbose := lib.IsVerbose(ctx)
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
				normalized := filepath.ToSlash(strings.TrimSpace(strings.Replace(entry, config.DestDir, "", 1)))
				files[normalized] = true
			}
		}
	}

	if verbose {
		fmt.Println("ReadDestDir", files)
	}
	return files, nil
}
