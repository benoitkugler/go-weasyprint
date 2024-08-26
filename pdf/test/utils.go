package test

import (
	"fmt"
	"path/filepath"
	"sync"

	fc "github.com/benoitkugler/textprocessing/fontconfig"
	"github.com/benoitkugler/textprocessing/pango/fcfonts"

	"github.com/benoitkugler/webrender/text"
)

var lock sync.Mutex

// LoadTestFontConfig loads the font index in [cacheDir],
// creating it if needed.
func LoadTestFontConfig(cacheDir string) (text.FontConfiguration, error) {
	lock.Lock()
	defer lock.Unlock()

	const cacheFile = "cache.fc"
	cachePath := filepath.Join(cacheDir, cacheFile)

	_, err := fc.LoadFontsetFile(cachePath)
	if err != nil {
		// build the index
		fmt.Println("Scanning fonts...")
		_, err := fc.ScanAndCache(cachePath)
		if err != nil {
			return nil, fmt.Errorf("scanning fonts: %s", err)
		}
		fmt.Println("Font index written in", cachePath)
	}

	fs, err := fc.LoadFontsetFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("reading font index: %s", err)
	}

	return text.NewFontConfigurationPango(fcfonts.NewFontMap(fc.Standard.Copy(), fs)), nil
}
