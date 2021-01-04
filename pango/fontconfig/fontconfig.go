package fontconfig

import "math"

const debugMode = false

type FcFontSet []*FcPattern // with length nfont, and cap sfont

type FcStrSet map[string]bool

const (
	FcSetSystem      = 0
	FcSetApplication = 1
)

type FcConfig struct {
	/*
	 * File names loaded from the configuration -- saved here as the
	 * cache file must be consulted before the directories are scanned,
	 * and those directives may occur in any order
	 */
	configDirs    FcStrSet /* directories to scan for fonts */
	configMapDirs FcStrSet /* mapped names to generate cache entries */
	/*
	 * List of directories containing fonts,
	 * built by recursively scanning the set
	 * of configured directories
	 */
	fontDirs FcStrSet
	/*
	 * List of directories containing cache files.
	 */
	cacheDirs FcStrSet
	/*
	 * Names of all of the configuration files used
	 * to create this configuration
	 */
	configFiles FcStrSet /* config files loaded */
	/*
	 * Substitution instructions for patterns and fonts;
	 * maxObjects is used to allocate appropriate intermediate storage
	 * while performing a whole set of substitutions
	 *
	 * 0.. substitutions for patterns
	 * 1.. substitutions for fonts
	 * 2.. substitutions for scanned fonts
	 */
	// FcPtrList	*subst[FcMatchKindEnd];
	// int		maxObjects;	    /* maximum number of tests in all substs */
	/*
	 * List of patterns used to control font file selection
	 */
	acceptGlobs    FcStrSet
	rejectGlobs    FcStrSet
	acceptPatterns *FcFontSet
	rejectPatterns *FcFontSet
	/*
	 * The set of fonts loaded from the listed directories; the
	 * order within the set does not determine the font selection,
	 * except in the case of identical matches in which case earlier fonts
	 * match preferrentially
	 */
	fonts [FcSetApplication + 1]FcFontSet
	/*
	 * Fontconfig can periodically rescan the system configuration
	 * and font directories.  This rescanning occurs when font
	 * listing requests are made, but no more often than rescanInterval
	 * seconds apart.
	 */
	// time_t rescanTime     /* last time information was scanned */
	// int    rescanInterval /* interval between scans */

	// FcExprPage *expr_pool /* pool of FcExpr's */

	sysRoot          string   /* override the system root directory */
	availConfigFiles FcStrSet /* config files available */
	// FcPtrList        *rulesetList /* List of rulesets being installed */
}

type FcResult uint8

const (
	FcResultMatch FcResult = iota
	FcResultNoMatch
	FcResultTypeMismatch
	FcResultNoId
	FcResultOutOfMemory
)

const (
	FC_SLANT_ROMAN   = 0
	FC_SLANT_ITALIC  = 100
	FC_SLANT_OBLIQUE = 110
)

const (
	FC_WIDTH_ULTRACONDENSED = 50
	FC_WIDTH_EXTRACONDENSED = 63
	FC_WIDTH_CONDENSED      = 75
	FC_WIDTH_SEMICONDENSED  = 87
	FC_WIDTH_NORMAL         = 100
	FC_WIDTH_SEMIEXPANDED   = 113
	FC_WIDTH_EXPANDED       = 125
	FC_WIDTH_EXTRAEXPANDED  = 150
	FC_WIDTH_ULTRAEXPANDED  = 200
)

const (
	FC_WEIGHT_THIN       = 0
	FC_WEIGHT_EXTRALIGHT = 40
	FC_WEIGHT_ULTRALIGHT = FC_WEIGHT_EXTRALIGHT
	FC_WEIGHT_LIGHT      = 50
	FC_WEIGHT_DEMILIGHT  = 55
	FC_WEIGHT_SEMILIGHT  = FC_WEIGHT_DEMILIGHT
	FC_WEIGHT_BOOK       = 75
	FC_WEIGHT_REGULAR    = 80
	FC_WEIGHT_NORMAL     = FC_WEIGHT_REGULAR
	FC_WEIGHT_MEDIUM     = 100
	FC_WEIGHT_DEMIBOLD   = 180
	FC_WEIGHT_SEMIBOLD   = FC_WEIGHT_DEMIBOLD
	FC_WEIGHT_BOLD       = 200
	FC_WEIGHT_EXTRABOLD  = 205
	FC_WEIGHT_ULTRABOLD  = FC_WEIGHT_EXTRABOLD
	FC_WEIGHT_BLACK      = 210
	FC_WEIGHT_HEAVY      = FC_WEIGHT_BLACK
	FC_WEIGHT_EXTRABLACK = 215
	FC_WEIGHT_ULTRABLACK = FC_WEIGHT_EXTRABLACK
)

var weightMap = [...]struct {
	ot, fc float64
}{
	{0, FC_WEIGHT_THIN},
	{100, FC_WEIGHT_THIN},
	{200, FC_WEIGHT_EXTRALIGHT},
	{300, FC_WEIGHT_LIGHT},
	{350, FC_WEIGHT_DEMILIGHT},
	{380, FC_WEIGHT_BOOK},
	{400, FC_WEIGHT_REGULAR},
	{500, FC_WEIGHT_MEDIUM},
	{600, FC_WEIGHT_DEMIBOLD},
	{700, FC_WEIGHT_BOLD},
	{800, FC_WEIGHT_EXTRABOLD},
	{900, FC_WEIGHT_BLACK},
	{1000, FC_WEIGHT_EXTRABLACK},
}

func lerp(x, x1, x2, y1, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	// assert(dx > 0 && dy >= 0 && x1 <= x && x <= x2)
	return y1 + (x-x1)*dy/dx
}

func FcWeightFromOpenTypeDouble(ot_weight float64) float64 {
	if ot_weight < 0 {
		return -1
	}

	ot_weight = math.Min(ot_weight, weightMap[len(weightMap)-1].ot)

	var i int
	for i = 1; ot_weight > weightMap[i].ot; i++ {
	}

	if ot_weight == weightMap[i].ot {
		return weightMap[i].fc
	}

	// interpolate between two items
	return lerp(ot_weight, weightMap[i-1].ot, weightMap[i].ot, weightMap[i-1].fc, weightMap[i].fc)
}
