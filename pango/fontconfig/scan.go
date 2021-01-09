package fontconfig

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ported from  fontconfig/src/fcdir.c 2000 Keith Packard

func scanFontConfig(set *FcFontSet, file string, config *FcConfig) bool {
	// int		i;
	// FcBool	ret = FcTrue;
	// int		old_nfont = set.nfont;
	sysroot := config.getSysRoot()

	if debugMode {
		fmt.Printf("\tScanning file %s...", file)
	}

	if _, ok := FcFreeTypeQueryAll(file, -1, set); !ok {
		return false
	}

	if debugMode {
		fmt.Println("done")
	}

	ret := true
	for _, font := range *set {

		/*
		 * Get rid of sysroot here so that targeting scan rule may contains FC_FILE pattern
		 * and they should usually expect without sysroot.
		 */
		if sysroot != "" {
			f, res := font.FcPatternObjectGetString(FC_FILE, 0)
			if res == FcResultMatch && strings.HasPrefix(f, sysroot) {
				font.del(FC_FILE)
				s := filepath.Clean(strings.TrimPrefix(f, sysroot))
				font.Add(FC_FILE, s, true)
			}
		}

		// Edit pattern with user-defined rules
		if config != nil && !config.FcConfigSubstitute(font, FcMatchScan) {
			ret = false
		}

		if !font.addFullname() {
			ret = false
		}

		if debugMode {
			fmt.Printf("Final font pattern:\n%s", font)
		}
	}
	return ret
}

func FcFileScanConfig(set *FcFontSet, dirs FcStrSet, file string, config *FcConfig) bool {
	if isDir(file) {
		sysroot := config.getSysRoot()
		d := file
		if sysroot != "" {
			if strings.HasPrefix(file, sysroot) {
				d = filepath.Clean(strings.TrimPrefix(file, sysroot))
			}
		}
		dirs[d] = true
		return true
	}

	return scanFontConfig(set, file, config)
}

// TODO:
func FcFreeTypeQueryAll(file string, id int, set *FcFontSet) (int, bool) {
	return 0, false
}
