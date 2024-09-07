package converttomacvlcplaylist

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/unicode/norm"
	"gopkg.in/yaml.v3"
)

var config = Load[ConvertToMacVlcPlaylistConfig]()

type ConfigBase interface {
	MatchService() bool
}

type ConvertToMacVlcPlaylistConfig struct {
	AtService       string   `yaml:"at_service"`
	SourcePlaylists []string `yaml:"source_playlists"`
	OutputDir       string   `yaml:"output_dir"`
	PathReplacement struct {
		From string `yaml:"from"`
		To   string `yaml:"to"`
	} `yaml:"path_replacement"`
}

func (c ConvertToMacVlcPlaylistConfig) MatchService() bool {
	return c.AtService == "convert_to_mac_vlc_playlist"
}

func Load[T ConfigBase]() T {
	// todo: config file path
	cfgYaml, err := os.ReadFile("at.config.yaml")
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

func Run() {
	log.Println("Start")

	readSourcePlaylist()
}

func readSourcePlaylist() {

	for _, sourcePlaylist := range config.SourcePlaylists {
		file, err := os.Open(sourcePlaylist)
		if err != nil && os.IsNotExist(err) {
			log.Printf("File not found: %s", sourcePlaylist)
			continue
		}
		defer file.Close()

		var m3uString []string = []string{"#EXTM3U"}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, config.PathReplacement.From) {
				m3uString = append(m3uString, encodeSourceToDest(line)...)
			}
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}

		outFileName := filepath.Join(config.OutputDir, filepath.Base(sourcePlaylist))
		log.Printf("Writing to %s", outFileName)
		err = os.WriteFile(outFileName, []byte(strings.Join(m3uString, "\n")), 0644)
		if err != nil {
			panic(err)
		}
	}
}

func encodeSourceToDest(source string) []string {
	prefixRemoved := strings.Replace(source, config.PathReplacement.From, "", 1)
	slashed := filepath.ToSlash(prefixRemoved)
	normalized := norm.NFD.String(slashed)

	name := filepath.Base(normalized)

	encoded := ""
	for _, seg := range strings.Split(normalized, "/") {
		if seg == "" {
			continue
		}
		encoded += "/" + url.PathEscape(seg)
	}

	dest := config.PathReplacement.To + encoded
	return []string{"#EXTINF:-1," + name, dest}
}
