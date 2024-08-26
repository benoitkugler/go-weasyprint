package test

import (
	"fmt"
	"os"
	"path/filepath"

	fc "github.com/benoitkugler/textprocessing/fontconfig"
	"github.com/benoitkugler/textprocessing/pango/fcfonts"

	"github.com/benoitkugler/webrender/text"
)

// LoadTestFontConfig loads the font index in [cacheDir],
// creating it if needed.
func LoadTestFontConfig(cacheDir string) (text.FontConfiguration, error) {
	const cacheFile = "cache.fc"
	cachePath := filepath.Join(cacheDir, cacheFile)

	_, err := os.Stat(cachePath)
	if err != nil {
		// build the index
		fmt.Println("Scanning fonts...")
		_, err := fc.ScanAndCache(cachePath)
		if err != nil {
			return nil, err
		}
	}

	fs, err := fc.LoadFontsetFile(cachePath)
	if err != nil {
		return nil, err
	}

	return text.NewFontConfigurationPango(fcfonts.NewFontMap(fc.Standard.Copy(), fs)), nil
}
