package lib

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type ConfigBase interface {
	MatchService() bool
}

func Load[T ConfigBase](configPath string) T {
	cfgYaml, err := os.ReadFile(configPath)
	if err != nil {
		log.Panicf("Failed to read config.yaml: %v", err)
		panic(err)
	}

	var res *T
	decoder := yaml.NewDecoder(bytes.NewReader(cfgYaml))
	for {
		var configBase T
		err := decoder.Decode(&configBase)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			panic(err)
		}
		if matched := configBase.MatchService(); !matched {
			continue
		}
		res = &configBase
	}

	if res == nil {
		log.Panicf("No config matched")
	}
	return *res
}

func Must1[T any](x T, err error) T {
	if err != nil {
		panic(err)
	}
	return x
}

type contextKey string

const verboseFlagKey contextKey = "verbose"

func WithVerboseFlag(ctx context.Context, verbose bool) context.Context {
	return context.WithValue(ctx, verboseFlagKey, verbose)
}
func IsVerbose(ctx context.Context) bool {
	v, _ := ctx.Value(verboseFlagKey).(bool)
	return v
}
