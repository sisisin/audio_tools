package converttomacvlcplaylist

import (
	"bufio"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/sisisin/audio_tools/src/lib"
	"golang.org/x/text/unicode/norm"
)

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

func Run(configPath string) {
	config := lib.Load[ConvertToMacVlcPlaylistConfig](configPath)
	readSourcePlaylist(config)
}

func readSourcePlaylist(config ConvertToMacVlcPlaylistConfig) {
	for _, sourcePlaylist := range config.SourcePlaylists {
		file, err := os.Open(sourcePlaylist)
		if err != nil && os.IsNotExist(err) {
			log.Printf("File not found: %s", sourcePlaylist)
			continue
		}
		defer file.Close()

		var m3u8ForVLC []string = []string{"#EXTM3U"}
		var m3u8 []string = []string{"#EXTM3U"}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, config.PathReplacement.From) {
				res := encodeSourceToDest(config, line)
				m3u8ForVLC = append(m3u8ForVLC, res.ForVLC...)
				m3u8 = append(m3u8, res.ForOthers...)
			}
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}

		outForVLCFileName := filepath.Join(config.OutputDir, strings.Replace(filepath.Base(sourcePlaylist), filepath.Ext(sourcePlaylist), "_vlc"+filepath.Ext(sourcePlaylist), 1))
		outFileName := filepath.Join(config.OutputDir, filepath.Base(sourcePlaylist))
		log.Printf("Writing to %s", outFileName)
		err = os.WriteFile(outFileName, []byte(strings.Join(m3u8, "\n")), 0644)
		if err != nil {
			panic(err)
		}
		log.Printf("Writing to %s", outForVLCFileName)
		err = os.WriteFile(outForVLCFileName, []byte(strings.Join(m3u8ForVLC, "\n")), 0644)
		if err != nil {
			panic(err)
		}
	}
}

type M3U8Results struct {
	ForVLC    []string
	ForOthers []string
}

func encodeSourceToDest(config ConvertToMacVlcPlaylistConfig, source string) M3U8Results {
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

	return M3U8Results{
		ForVLC:    []string{"#EXTINF:-1," + name, "file://" + config.PathReplacement.To + encoded},
		ForOthers: []string{"#EXTINF:-1," + filepath.Base(slashed), config.PathReplacement.To + slashed},
	}
}
