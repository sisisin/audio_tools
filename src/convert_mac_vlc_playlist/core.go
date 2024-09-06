package convertmacvlcplaylist

import (
	"bufio"
	"flag"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/sisisin/audio_tools/src/lib"
	"golang.org/x/text/unicode/norm"
)

var replaceTo string

var sourcePlaylistFile = os.Getenv("SOURCE_PLAYLIST_FILE")
var sourceMusicDirPath = os.Getenv("SOURCE_MUSIC_DIR_PATH")
var outDir = os.Getenv("OUT_DIR")

func prepare() {
	if sourcePlaylistFile == "" {
		panic("The SOURCE_PLAYLIST_FILE environment variable is required")
	}
	if sourceMusicDirPath == "" {
		panic("The SOURCE_MUSIC_DIR_PATH environment variable is required")
	}
	if outDir == "" {
		panic("The OUT_DIR environment variable is required")
	}

	flag.StringVar(&replaceTo, "replace-to", "", "Replacement string for the playlist")
	flag.Parse()

	if replaceTo == "" {
		panic("The -replace-to flag is required")
	}
}

func Run() {
	log.Println("Start")

	prepare()
	readSourcePlaylist()
}

func readSourcePlaylist() {

	file := lib.Must1(os.Open(sourcePlaylistFile))
	defer file.Close()

	var m3uString []string = []string{"#EXTM3U"}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, sourceMusicDirPath) {
			name, dest := encodeSourceToDest(line)
			m3uString = append(m3uString, "#EXTINF:-1,"+name, dest)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	outFileName := filepath.Join(outDir, "output.m3u8")
	log.Printf("Writing to %s", outFileName)
	err := os.WriteFile(outFileName, []byte(strings.Join(m3uString, "\n")), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func encodeSourceToDest(source string) (name, dest string) {
	prefixRemoved := strings.Replace(source, sourceMusicDirPath, "", 1)
	slashed := filepath.ToSlash(prefixRemoved)
	normalized := norm.NFD.String(slashed)

	name = filepath.Base(normalized)

	encoded := ""
	for _, seg := range strings.Split(normalized, "/") {
		if seg == "" {
			continue
		}
		encoded += "/" + url.PathEscape(seg)
	}

	dest = replaceTo + encoded
	return
}
