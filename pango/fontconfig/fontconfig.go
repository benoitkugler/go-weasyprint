package fontconfig

type FcPattern struct {
	num         int
	size        int
	elts_offset int
}

type FcFontSet struct {
	nfont int
	sfont int
	fonts []*FcPattern
}

type FcStrSet map[string]bool

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
	// FcFontSet	*fonts[FcSetApplication + 1];
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
