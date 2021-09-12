package hyphen

import (
	"embed"
	"log"
)

//go:embed dictionaries
var dictionaries embed.FS

var languages map[string]string

func init() {
	var err error
	languages, err = getLanguages(dictionaries)
	if err != nil {
		log.Fatal(err)
	}
}
