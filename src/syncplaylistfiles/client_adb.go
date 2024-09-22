package syncplaylistfiles

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"strings"

	"github.com/sisisin/audio_tools/src/lib"
)

type syncClientAdb struct{}

func (*syncClientAdb) CopyFile(src, dest string) error {
	_, err := runCommand("adb", []string{"push", src, dest})
	return err
}
func (*syncClientAdb) RemoveFile(path string) error {
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
			errors.New(errWriter.String()),
			fmt.Errorf("command: %s %s", cmd, strings.Join(args, " ")),
		)
	}

	return outWriter.String(), nil
}

func (*syncClientAdb) ReadDestDir(ctx context.Context, destDir string) (map[string]bool, error) {
	verbose := lib.IsVerbose(ctx)
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
			entry := path.Join(dir, line)
			if line == "" {
				continue
			}
			if strings.HasSuffix(line, "/") {
				targetDirs = append(targetDirs, entry)
			} else {
				files[entry] = true
			}
		}
	}

	if verbose {
		fmt.Println("ReadDestDir", files)
		fmt.Println("ReadDestDir", targetDirs)
	}
	return files, nil
}
