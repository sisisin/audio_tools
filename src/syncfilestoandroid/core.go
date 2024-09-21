package syncfilestoandroid

import (
	"fmt"

	"github.com/sisisin/audio_tools/src/lib"
)

type SyncFilesToAndroidConfig struct {
	AtService string `yaml:"at_service"`
}

func (c SyncFilesToAndroidConfig) MatchService() bool {
	return c.AtService == "sync_files_to_android"
}

func Run(configPath string) {
	config := lib.Load[SyncFilesToAndroidConfig](configPath)
	fmt.Printf("SyncFilesToAndroidConfig: %v\n", config)
}
