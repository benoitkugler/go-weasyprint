package fontconfig

import (
	"fmt"
)

// ported from fontconfig/src/fccfg.c Copyright © 2000 Keith Packard

const (
	FcQualAny uint8 = iota
	FcQualAll
	FcQualFirst
	FcQualNotFirst
)

type FcTest struct {
	kind   FcMatchKind
	qual   uint8
	object FcObject
	op     FcOp
	expr   *FcExpr
}

type FcEdit struct {
	object  FcObject
	op      FcOp
	expr    *FcExpr
	binding FcValueBinding
}

type FcRule interface{} // *Test Or Edit

type FcRuleSet struct {
	name        string
	description string
	domain      string
	enabled     bool
	subst       [FcMatchKindEnd][][]FcRule
}

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
	 * for performing a whole set of substitutions
	 *
	 * 0.. substitutions for patterns
	 * 1.. substitutions for fonts
	 * 2.. substitutions for scanned fonts
	 */
	subst      [FcMatchKindEnd][]*FcRuleSet
	maxObjects int /* maximum number of tests in all substs */

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

// FcConfigSubstituteWithPat performs the sequence of pattern modification operations. If `kind` is
// FcMatchPattern, then those tagged as pattern operations are applied, else
// if `kind` is FcMatchFont, those tagged as font operations are applied and
// `pPat` is used for test; elements with target=pattern. Returns `false`
// if the substitution cannot be performed.
// If `config` is nil, the current configuration is used.
func (config *FcConfig) FcConfigSubstituteWithPat(p, pPat *FcPattern, kind FcMatchKind) bool {
	if kind < FcMatchKindBegin || kind >= FcMatchKindEnd {
		return false
	}

	config = fallbackConfig(config)

	var v interface{}

	s := config.subst[kind]
	if kind == FcMatchPattern {
		strs := FcGetDefaultLangs()
		var lsund FcLangSet
		lsund.add("und")

		for lang := range strs {
			e := p.elts[FC_LANG]

			for _, ll := range e {
				vvL := ll.value

				if vv, ok := vvL.(FcLangSet); ok {
					var ls FcLangSet
					ls.add(lang)

					b := vv.FcLangSetContains(ls)
					if b {
						goto bail_lang
					}
					if vv.FcLangSetContains(lsund) {
						goto bail_lang
					}
				} else {
					vv, _ := vvL.(string)
					if FcStrCmpIgnoreCase(vv, lang) == 0 {
						goto bail_lang
					}
					if FcStrCmpIgnoreCase(vv, "und") == 0 {
						goto bail_lang
					}
				}
			}
			v = lang
			p.addWithBinding(FC_LANG, v, FcValueBindingWeak, true)
		}
	bail_lang:
		var res FcResult
		v, res = p.FcPatternObjectGet(FC_PRGNAME, 0)
		if res == FcResultNoMatch {
			prgname := FcGetPrgname()
			if prgname != "" {
				p.Add(FC_PRGNAME, prgname, true)
			}
		}
	}

	nobjs := int(fcEnd) - 1 + config.maxObjects + 2
	valuePos := make([]int, nobjs)
	elt := make([]FcValueList, nobjs)
	tst := make([]*FcTest, nobjs)

	if debugMode {
		fmt.Println("FcConfigSubstitute ")
		fmt.Println(p.String())
	}

	data := newFamilyTable(p)

	var (
		m     *FcPattern
		table = &data
	)
	for _, rs := range s {
		if debugMode {
			fmt.Printf("\nRule Set: %s\n", rs.name)
		}
	subsLoop:
		for _, rules := range rs.subst[kind] {
			for i := range valuePos {
				elt[i] = nil
				valuePos[i] = -1
				tst[i] = nil
			}
			for _, r := range rules {
				switch r := r.(type) {
				case nil: // shouldn't be reached
					break
				case *FcTest:
					// Check the tests to see if they all match the pattern
					if debugMode {
						fmt.Println("FcConfigSubstitute test ", r)
					}
					if kind == FcMatchFont && r.kind == FcMatchPattern {
						m = pPat
						table = nil
					} else {
						m = p
						table = &data
					}
					var e FcValueList
					if m != nil {
						e = m.elts[r.object]
					}
					object := r.object
					// different 'kind' won't be the target of edit
					if elt[object] == nil && kind == r.kind {
						elt[object] = e
						tst[object] = r
					}
					// If there's no such field in the font, then FcQualAll matches for FcQualAny does not
					if e == nil {
						if r.qual == FcQualAll {
							valuePos[object] = -1
							continue
						} else {
							if debugMode {
								fmt.Println("No match")
							}
							continue subsLoop
						}
					}
					// Check to see if there is a match, mark the location to apply match-relative edits
					vlIndex := matchValueList(m, pPat, kind, r, e, table)
					// different 'kind' won't be the target of edit
					if valuePos[object] == -1 && kind == r.kind && vlIndex != -1 {
						valuePos[object] = vlIndex
					}
					if vlIndex == -1 || (r.qual == FcQualFirst && vlIndex != 0) ||
						(r.qual == FcQualNotFirst && vlIndex == 0) {
						if debugMode {
							fmt.Println("No match")
						}
						return true
					}
					break
				case FcEdit:
					object := r.object
					if debugMode {
						fmt.Println("Substitute ", r)
						fmt.Println()
					}
					// Evaluate the list of expressions
					l := r.expr.FcConfigValues(p, pPat, kind, r.binding)
					if tst[object] != nil && (tst[object].kind == FcMatchFont || kind == FcMatchPattern) {
						elt[object] = p.elts[tst[object].object]
					}

					switch r.op {
					case FcOpAssign:
						// If there was a test, then replace the matched value with the newList list of values
						if valuePos[object] != -1 {
							thisValue := valuePos[object]

							// Append the newList list of values after the current value
							elt[object].insert(thisValue, true, l, r.object, table)

							//  Delete the marked value
							if thisValue != -1 {
								elt[object].del(thisValue, object, table)
							}

							// Adjust a pointer into the value list to ensure future edits occur at the same place
							break
						}
						fallthrough
					case FcOpAssignReplace:
						// Delete all of the values and insert the newList set
						p.FcConfigPatternDel(r.object, table)
						p.FcConfigPatternAdd(r.object, l, true, table)
						// Adjust a pointer into the value list as they no longer point to anything valid
						valuePos[object] = -1
					case FcOpPrepend:
						if valuePos[object] != -1 {
							elt[object].insert(valuePos[object], false, l, r.object, table)
							break
						}
						fallthrough
					case FcOpPrependFirst:
						p.FcConfigPatternAdd(r.object, l, false, table)
					case FcOpAppend:
						if valuePos[object] != -1 {
							elt[object].insert(valuePos[object], true, l, r.object, table)
							break
						}
						fallthrough
					case FcOpAppendLast:
						p.FcConfigPatternAdd(r.object, l, true, table)
					case FcOpDelete:
						if valuePos[object] != -1 {
							elt[object].del(valuePos[object], object, table)
							break
						}
						fallthrough
					case FcOpDeleteAll:
						p.FcConfigPatternDel(r.object, table)
					}
					// Now go through the pattern and eliminate any properties without data
					p.canon(r.object)

					if debugMode {
						fmt.Println("FcConfigSubstitute edit", p.String())
					}
				}
			}
		}
	}
	if debugMode {
		fmt.Println("FcConfigSubstitute done", p.String())
	}

	return true
}

/* Objects MT-safe for readonly access. */

// #if defined (_WIN32) && !defined (R_OK)
// #define R_OK 4
// #endif

// #if defined(_WIN32) && !defined(S_ISFIFO)
// #define S_ISFIFO(m) 0
// #endif

// static FcConfig    *_fcConfig; /* MT-safe */
// static FcMutex	   *_lock;

// static void
// lock_config (void)
// {
//     FcMutex *lock;
// retry:
//     lock = fc_atomic_ptr_get (&_lock);
//     if (!lock)
//     {
// 	lock = (FcMutex *) malloc (sizeof (FcMutex));
// 	FcMutexInit (lock);
// 	if (!fc_atomic_ptr_cmpexch (&_lock, nil, lock))
// 	{
// 	    FcMutexFinish (lock);
// 	    goto retry;
// 	}
// 	FcMutexLock (lock);
// 	/* Initialize random state */
// 	FcRandom ();
// 	return;
//     }
//     FcMutexLock (lock);
// }

// static void
// unlock_config (void)
// {
//     FcMutex *lock;
//     lock = fc_atomic_ptr_get (&_lock);
//     FcMutexUnlock (lock);
// }

// static void
// free_lock (void)
// {
//     FcMutex *lock;
//     lock = fc_atomic_ptr_get (&_lock);
//     if (lock && fc_atomic_ptr_cmpexch (&_lock, lock, nil))
//     {
// 	FcMutexFinish (lock);
// 	free (lock);
//     }
// }

// static FcConfig *
// FcConfigEnsure (void)
// {
//     FcConfig	*config;
// retry:
//     config = fc_atomic_ptr_get (&_fcConfig);
//     if (!config)
//     {
// 	config = FcInitLoadConfigAndFonts ();

// 	if (!config || !fc_atomic_ptr_cmpexch (&_fcConfig, nil, config)) {
// 	    if (config)
// 		FcConfigDestroy (config);
// 	    goto retry;
// 	}
//     }
//     return config;
// }

// static void
// FcDestroyAsRule (void *data)
// {
//     FcRuleDestroy (data);
// }

// static void
// FcDestroyAsRuleSet (void *data)
// {
//     FcRuleSetDestroy (data);
// }

// FcBool
// FcConfigInit (void)
// {
//   return FcConfigEnsure () ? true : false;
// }

// void
// FcConfigFini (void)
// {
//     FcConfig *cfg = fc_atomic_ptr_get (&_fcConfig);
//     if (cfg && fc_atomic_ptr_cmpexch (&_fcConfig, cfg, nil))
// 	FcConfigDestroy (cfg);
//     free_lock ();
// }

// static FcChar8 *
// FcConfigRealPath(const FcChar8 *path)
// {
//     char	resolved_name[FC_PATH_MAX+1];
//     char	*resolved_ret;

//     if (!path)
// 	return nil;

// #ifndef _WIN32
//     resolved_ret = realpath((const char *) path, resolved_name);
// #else
//     if (GetFullPathNameA ((LPCSTR) path, FC_PATH_MAX, resolved_name, nil) == 0)
//     {
//         fprintf (stderr, "Fontconfig warning: GetFullPathNameA failed.\n");
//         return nil;
//     }
//     resolved_ret = resolved_name;
// #endif
//     if (resolved_ret)
// 	path = (FcChar8 *) resolved_ret;
//     return FcStrCopyFilename(path);
// }

// FcConfig *
// FcConfigCreate (void)
// {
//     FcSetName	set;
//     FcConfig	*config;
//     FcMatchKind	k;
//     FcBool	err = false;

//     config = malloc (sizeof (FcConfig));
//     if (!config)
// 	goto bail0;

//     config.configDirs = FcStrSetCreate ();
//     if (!config.configDirs)
// 	goto bail1;

//     config.configMapDirs = FcStrSetCreate();
//     if (!config.configMapDirs)
// 	goto bail1_5;

//     config.configFiles = FcStrSetCreate ();
//     if (!config.configFiles)
// 	goto bail2;

//     config.fontDirs = FcStrSetCreate ();
//     if (!config.fontDirs)
// 	goto bail3;

//     config.acceptGlobs = FcStrSetCreate ();
//     if (!config.acceptGlobs)
// 	goto bail4;

//     config.rejectGlobs = FcStrSetCreate ();
//     if (!config.rejectGlobs)
// 	goto bail5;

//     config.acceptPatterns = FcFontSetCreate ();
//     if (!config.acceptPatterns)
// 	goto bail6;

//     config.rejectPatterns = FcFontSetCreate ();
//     if (!config.rejectPatterns)
// 	goto bail7;

//     config.cacheDirs = FcStrSetCreate ();
//     if (!config.cacheDirs)
// 	goto bail8;

//     for (k = FcMatchKindBegin; k < FcMatchKindEnd; k++)
//     {
// 	config.subst[k] = FcPtrListCreate (FcDestroyAsRuleSet);
// 	if (!config.subst[k])
// 	    err = true;
//     }
//     if (err)
// 	goto bail9;

//     config.maxObjects = 0;
//     for (set = FcSetSystem; set <= FcSetApplication; set++)
// 	config.fonts[set] = 0;

//     config.rescanTime = time(0);
//     config.rescanInterval = 30;

//     config.expr_pool = nil;

//     config.sysRoot = FcConfigRealPath((const FcChar8 *) getenv("FONTCONFIG_SYSROOT"));

//     config.rulesetList = FcPtrListCreate (FcDestroyAsRuleSet);
//     if (!config.rulesetList)
// 	goto bail9;
//     config.availConfigFiles = FcStrSetCreate ();
//     if (!config.availConfigFiles)
// 	goto bail10;

//     FcRefInit (&config.ref, 1);

//     return config;

// bail10:
//     FcPtrListDestroy (config.rulesetList);
// bail9:
//     for (k = FcMatchKindBegin; k < FcMatchKindEnd; k++)
// 	if (config.subst[k])
// 	    FcPtrListDestroy (config.subst[k]);
//     FcStrSetDestroy (config.cacheDirs);
// bail8:
//     FcFontSetDestroy (config.rejectPatterns);
// bail7:
//     FcFontSetDestroy (config.acceptPatterns);
// bail6:
//     FcStrSetDestroy (config.rejectGlobs);
// bail5:
//     FcStrSetDestroy (config.acceptGlobs);
// bail4:
//     FcStrSetDestroy (config.fontDirs);
// bail3:
//     FcStrSetDestroy (config.configFiles);
// bail2:
//     FcStrSetDestroy (config.configMapDirs);
// bail1_5:
//     FcStrSetDestroy (config.configDirs);
// bail1:
//     free (config);
// bail0:
//     return 0;
// }

// static FcFileTime
// FcConfigNewestFile (FcStrSet *files)
// {
//     FcStrList	    *list = FcStrListCreate (files);
//     FcFileTime	    newest = { 0, false };
//     FcChar8	    *file;
//     struct  stat    statb;

//     if (list)
//     {
// 	for ((file = FcStrListNext (list)))
// 	    if (FcStat (file, &statb) == 0)
// 		if (!newest.set || statb.st_mtime - newest.time > 0)
// 		{
// 		    newest.set = true;
// 		    newest.time = statb.st_mtime;
// 		}
// 	FcStrListDone (list);
//     }
//     return newest;
// }

// FcBool
// FcConfigUptoDate (FcConfig *config)
// {
//     FcFileTime	config_time, config_dir_time, font_time;
//     time_t	now = time(0);
//     FcBool	ret = true;

//     config = fallbackConfig (config);
//     if (!config)
// 	return false;

//     config_time = FcConfigNewestFile (config.configFiles);
//     config_dir_time = FcConfigNewestFile (config.configDirs);
//     font_time = FcConfigNewestFile (config.fontDirs);
//     if ((config_time.set && config_time.time - config.rescanTime > 0) ||
// 	(config_dir_time.set && (config_dir_time.time - config.rescanTime) > 0) ||
// 	(font_time.set && (font_time.time - config.rescanTime) > 0))
//     {
// 	/* We need to check for potential clock problems here (OLPC ticket #6046) */
// 	if ((config_time.set && (config_time.time - now) > 0) ||
//     	(config_dir_time.set && (config_dir_time.time - now) > 0) ||
//         (font_time.set && (font_time.time - now) > 0))
// 	{
// 	    fprintf (stderr,
//                     "Fontconfig warning: Directory/file mtime in the future. New fonts may not be detected.\n");
// 	    config.rescanTime = now;
// 	    goto bail;
// 	}
// 	else
// 	{
// 	    ret = false;
// 	    goto bail;
// 	}
//     }
//     config.rescanTime = now;
// bail:
//     FcConfigDestroy (config);

//     return ret;
// }

// FcExpr *
// FcConfigAllocExpr (FcConfig *config)
// {
//     if (!config.expr_pool || config.expr_pool.next == config.expr_pool.end)
//     {
// 	FcExprPage *new_page;

// 	new_page = malloc (sizeof (FcExprPage));
// 	if (!new_page)
// 	    return 0;

// 	new_page.next_page = config.expr_pool;
// 	new_page.next = new_page.exprs;
// 	config.expr_pool = new_page;
//     }

//     return config.expr_pool.next++;
// }

// fallback to the current global configuratio if `config` is nil
func fallbackConfig(config *FcConfig) *FcConfig {
	if config != nil {
		return config
	}

	// TODO:
	/* lock during obtaining the value from _fcConfig and count up refcount there,
	 * there are the race between them.
	 */
	// lock_config ();
	// retry:
	// config = fc_atomic_ptr_get (&_fcConfig);
	// if (!config) 	{
	//     unlock_config ();

	//     config = FcInitLoadConfigAndFonts ();
	//     if (!config)
	// 	goto retry;
	//     lock_config ();
	//     if (!fc_atomic_ptr_cmpexch (&_fcConfig, nil, config))
	//     {
	// 	FcConfigDestroy (config);
	// 	goto retry;
	//     }
	// }
	// FcRefInc (&config.ref);
	// unlock_config ();

	return config
}

// void
// FcConfigDestroy (FcConfig *config)
// {
//     FcSetName	set;
//     FcExprPage	*page;
//     FcMatchKind	k;

//     if (FcRefDec (&config.ref) != 1)
// 	return;

//     (void) fc_atomic_ptr_cmpexch (&_fcConfig, config, nil);

//     FcStrSetDestroy (config.configDirs);
//     FcStrSetDestroy (config.configMapDirs);
//     FcStrSetDestroy (config.fontDirs);
//     FcStrSetDestroy (config.cacheDirs);
//     FcStrSetDestroy (config.configFiles);
//     FcStrSetDestroy (config.acceptGlobs);
//     FcStrSetDestroy (config.rejectGlobs);
//     FcFontSetDestroy (config.acceptPatterns);
//     FcFontSetDestroy (config.rejectPatterns);

//     for (k = FcMatchKindBegin; k < FcMatchKindEnd; k++)
// 	FcPtrListDestroy (config.subst[k]);
//     FcPtrListDestroy (config.rulesetList);
//     FcStrSetDestroy (config.availConfigFiles);
//     for (set = FcSetSystem; set <= FcSetApplication; set++)
// 	if (config.fonts[set])
// 	    FcFontSetDestroy (config.fonts[set]);

//     page = config.expr_pool;
//     for (page)
//     {
//       FcExprPage *next = page.next_page;
//       free (page);
//       page = next;
//     }
//     if (config.sysRoot)
// 	FcStrFree (config.sysRoot);

//     free (config);
// }

// /*
//  * Add cache to configuration, adding fonts and directories
//  */

// FcBool
// FcConfigAddCache (FcConfig *config, FcCache *cache,
// 		  FcSetName set, FcStrSet *dirSet, FcChar8 *forDir)
// {
//     FcFontSet	*fs;
//     intptr_t	*dirs;
//     int		i;
//     FcBool      relocated = false;

//     if (strcmp ((char *)FcCacheDir(cache), (char *)forDir) != 0)
//       relocated = true;

//     /*
//      * Add fonts
//      */
//     fs = FcCacheSet (cache);
//     if (fs)
//     {
// 	int	nref = 0;

// 	for (i = 0; i < fs.nfont; i++)
// 	{
// 	    FcPattern	*font = FcFontSetFont (fs, i);
// 	    FcChar8	*font_file;
// 	    FcChar8	*relocated_font_file = nil;

// 	    if (FcPatternObjectGetString (font, FC_FILE,
// 					  0, &font_file) == FcResultMatch)
// 	    {
// 		if (relocated)
// 		  {
// 		    FcChar8 *slash = FcStrLastSlash (font_file);
// 		    relocated_font_file = FcStrBuildFilename (forDir, slash + 1, nil);
// 		    font_file = relocated_font_file;
// 		  }

// 		/*
// 		 * Check to see if font is banned by filename
// 		 */
// 		if (!FcConfigAcceptFilename (config, font_file))
// 		{
// 		    free (relocated_font_file);
// 		    continue;
// 		}
// 	    }

// 	    /*
// 	     * Check to see if font is banned by pattern
// 	     */
// 	    if (!FcConfigAcceptFont (config, font))
// 	    {
// 		free (relocated_font_file);
// 		continue;
// 	    }

// 	    if (relocated_font_file)
// 	    {
// 	      font = FcPatternCacheRewriteFile (font, cache, relocated_font_file);
// 	      free (relocated_font_file);
// 	    }

// 	    if (FcFontSetAdd (config.fonts[set], font))
// 		nref++;
// 	}
// 	FcDirCacheReference (cache, nref);
//     }

//     /*
//      * Add directories
//      */
//     dirs = FcCacheDirs (cache);
//     if (dirs)
//     {
// 	for (i = 0; i < cache.dirs_count; i++)
// 	{
// 	    const FcChar8 *dir = FcCacheSubdir (cache, i);
// 	    FcChar8 *s = nil;

// 	    if (relocated)
// 	    {
// 		FcChar8 *base = FcStrBasename (dir);
// 		dir = s = FcStrBuildFilename (forDir, base, nil);
// 		FcStrFree (base);
// 	    }
// 	    if (FcConfigAcceptFilename (config, dir))
// 		FcStrSetAddFilename (dirSet, dir);
// 	    if (s)
// 		FcStrFree (s);
// 	}
//     }
//     return true;
// }

// static FcBool
// FcConfigAddDirList (FcConfig *config, FcSetName set, FcStrSet *dirSet)
// {
//     FcStrList	    *dirlist;
//     FcChar8	    *dir;
//     FcCache	    *cache;

//     dirlist = FcStrListCreate (dirSet);
//     if (!dirlist)
//         return false;

//     for ((dir = FcStrListNext (dirlist)))
//     {
// 	if (FcDebug () & FC_DBG_FONTSET)
// 	    printf ("adding fonts from %s\n", dir);
// 	cache = FcDirCacheRead (dir, false, config);
// 	if (!cache)
// 	    continue;
// 	FcConfigAddCache (config, cache, set, dirSet, dir);
// 	FcDirCacheUnload (cache);
//     }
//     FcStrListDone (dirlist);
//     return true;
// }

// /*
//  * Scan the current list of directories in the configuration
//  * and build the set of available fonts.
//  */

// FcBool
// FcConfigBuildFonts (FcConfig *config)
// {
//     FcFontSet	    *fonts;
//     FcBool	    ret = true;

//     config = fallbackConfig (config);
//     if (!config)
// 	return false;

//     fonts = FcFontSetCreate ();
//     if (!fonts)
//     {
// 	ret = false;
// 	goto bail;
//     }

//     FcConfigSetFonts (config, fonts, FcSetSystem);

//     if (!FcConfigAddDirList (config, FcSetSystem, config.fontDirs))
//     {
// 	ret = false;
// 	goto bail;
//     }
//     if (FcDebug () & FC_DBG_FONTSET)
// 	FcFontSetPrint (fonts);
// bail:
//     FcConfigDestroy (config);

//     return ret;
// }

// FcBool
// FcConfigSetCurrent (FcConfig *config)
// {
//     FcConfig *cfg;

//     if (config)
//     {
// 	if (!config.fonts[FcSetSystem])
// 	    if (!FcConfigBuildFonts (config))
// 		return false;
// 	FcRefInc (&config.ref);
//     }

//     lock_config ();
// retry:
//     cfg = fc_atomic_ptr_get (&_fcConfig);

//     if (config == cfg)
//     {
// 	unlock_config ();
// 	if (config)
// 	    FcConfigDestroy (config);
// 	return true;
//     }

//     if (!fc_atomic_ptr_cmpexch (&_fcConfig, cfg, config))
// 	goto retry;
//     unlock_config ();
//     if (cfg)
// 	FcConfigDestroy (cfg);

//     return true;
// }

// FcConfig *
// FcConfigGetCurrent (void)
// {
//     return FcConfigEnsure ();
// }

// FcBool
// FcConfigAddConfigDir (FcConfig	    *config,
// 		      const FcChar8 *d)
// {
//     return FcStrSetAddFilename (config.configDirs, d);
// }

// FcStrList *
// FcConfigGetConfigDirs (FcConfig   *config)
// {
//     FcStrList *ret;

//     config = fallbackConfig (config);
//     if (!config)
// 	return nil;
//     ret = FcStrListCreate (config.configDirs);
//     FcConfigDestroy (config);

//     return ret;
// }

// FcBool
// FcConfigAddFontDir (FcConfig	    *config,
// 		    const FcChar8   *d,
// 		    const FcChar8   *m,
// 		    const FcChar8   *salt)
// {
//     if (FcDebug() & FC_DBG_CACHE)
//     {
// 	if (m)
// 	{
// 	    printf ("%s . %s%s%s%s\n", d, m, salt ? " (salt: " : "", salt ? (const char *)salt : "", salt ? ")" : "");
// 	}
// 	else if (salt)
// 	{
// 	    printf ("%s%s%s%s\n", d, salt ? " (salt: " : "", salt ? (const char *)salt : "", salt ? ")" : "");
// 	}
//     }
//     return FcStrSetAddFilenamePairWithSalt (config.fontDirs, d, m, salt);
// }

// FcBool
// FcConfigResetFontDirs (FcConfig *config)
// {
//     if (FcDebug() & FC_DBG_CACHE)
//     {
// 	printf ("Reset font directories!\n");
//     }
//     return FcStrSetDeleteAll (config.fontDirs);
// }

// FcStrList *
// FcConfigGetFontDirs (FcConfig	*config)
// {
//     FcStrList *ret;

//     config = fallbackConfig (config);
//     if (!config)
// 	return nil;
//     ret = FcStrListCreate (config.fontDirs);
//     FcConfigDestroy (config);

//     return ret;
// }

// static FcBool
// FcConfigPathStartsWith(const FcChar8	*path,
// 		       const FcChar8	*start)
// {
//     int len = strlen((char *) start);

//     if (strncmp((char *) path, (char *) start, len) != 0)
// 	return false;

//     switch (path[len]) {
//     case '\0':
//     case FC_DIR_SEPARATOR:
// 	return true;
//     default:
// 	return false;
//     }
// }

// FcChar8 *
// FcConfigMapFontPath(FcConfig		*config,
// 		    const FcChar8	*path)
// {
//     FcStrList	*list;
//     FcChar8	*dir;
//     const FcChar8 *map, *rpath;
//     FcChar8     *retval;

//     list = FcConfigGetFontDirs(config);
//     if (!list)
// 	return 0;
//     for ((dir = FcStrListNext(list)))
// 	if (FcConfigPathStartsWith(path, dir))
// 	    break;
//     FcStrListDone(list);
//     if (!dir)
// 	return 0;
//     map = FcStrTripleSecond(dir);
//     if (!map)
// 	return 0;
//     rpath = path + strlen ((char *) dir);
//     for (*rpath == '/')
// 	rpath++;
//     retval = FcStrBuildFilename(map, rpath, nil);
//     if (retval)
//     {
// 	size_t len = strlen ((const char *) retval);
// 	for (len > 0 && retval[len-1] == '/')
// 	    len--;
// 	/* trim the last slash */
// 	retval[len] = 0;
//     }
//     return retval;
// }

// const FcChar8 *
// FcConfigMapSalt (FcConfig      *config,
// 		 const FcChar8 *path)
// {
//     FcStrList *list;
//     FcChar8 *dir;

//     list = FcConfigGetFontDirs (config);
//     if (!list)
// 	return nil;
//     for ((dir = FcStrListNext (list)))
// 	if (FcConfigPathStartsWith (path, dir))
// 	    break;
//     FcStrListDone (list);
//     if (!dir)
// 	return nil;

//     return FcStrTripleThird (dir);
// }

// FcBool
// FcConfigAddCacheDir (FcConfig	    *config,
// 		     const FcChar8  *d)
// {
//     return FcStrSetAddFilename (config.cacheDirs, d);
// }

// FcStrList *
// FcConfigGetCacheDirs (FcConfig *config)
// {
//     FcStrList *ret;

//     config = fallbackConfig (config);
//     if (!config)
// 	return nil;
//     ret = FcStrListCreate (config.cacheDirs);
//     FcConfigDestroy (config);

//     return ret;
// }

// FcBool
// FcConfigAddConfigFile (FcConfig	    *config,
// 		       const FcChar8   *f)
// {
//     FcBool	ret;
//     FcChar8	*file = FcConfigGetFilename (config, f);

//     if (!file)
// 	return false;

//     ret = FcStrSetAdd (config.configFiles, file);
//     FcStrFree (file);
//     return ret;
// }

// FcStrList *
// FcConfigGetConfigFiles (config *FcConfig)
// {
//     FcStrList *ret;

//     config = fallbackConfig (config);
//     if (!config)
// 	return nil;
//     ret = FcStrListCreate (config.configFiles);
//     FcConfigDestroy (config);

//     return ret;
// }

// FcChar8 *
// FcConfigGetCache (FcConfig  *config FC_UNUSED)
// {
//     return nil;
// }

// FcFontSet *
// FcConfigGetFonts (FcConfig	*config,
// 		  FcSetName	set)
// {
//     if (!config)
//     {
// 	config = FcConfigGetCurrent ();
// 	if (!config)
// 	    return 0;
//     }
//     return config.fonts[set];
// }

// void
// FcConfigSetFonts (FcConfig	*config,
// 		  FcFontSet	*fonts,
// 		  FcSetName	set)
// {
//     if (config.fonts[set])
// 	FcFontSetDestroy (config.fonts[set]);
//     config.fonts[set] = fonts;
// }

// FcBlanks *
// FcBlanksCreate (void)
// {
//     /* Deprecated. */
//     return nil;
// }

// void
// FcBlanksDestroy (FcBlanks *b FC_UNUSED)
// {
//     /* Deprecated. */
// }

// FcBool
// FcBlanksAdd (FcBlanks *b FC_UNUSED, FcChar32 ucs4 FC_UNUSED)
// {
//     /* Deprecated. */
//     return false;
// }

// FcBool
// FcBlanksIsMember (FcBlanks *b FC_UNUSED, FcChar32 ucs4 FC_UNUSED)
// {
//     /* Deprecated. */
//     return false;
// }

// FcBlanks *
// FcConfigGetBlanks (FcConfig	*config FC_UNUSED)
// {
//     /* Deprecated. */
//     return nil;
// }

// FcBool
// FcConfigAddBlank (FcConfig	*config FC_UNUSED,
// 		  FcChar32    	blank FC_UNUSED)
// {
//     /* Deprecated. */
//     return false;
// }

// int
// FcConfigGetRescanInterval (FcConfig *config)
// {
//     int ret;

//     config = fallbackConfig (config);
//     if (!config)
// 	return 0;
//     ret = config.rescanInterval;
//     FcConfigDestroy (config);

//     return ret;
// }

// FcBool
// FcConfigSetRescanInterval (FcConfig *config, int rescanInterval)
// {
//     config = fallbackConfig (config);
//     if (!config)
// 	return false;
//     config.rescanInterval = rescanInterval;
//     FcConfigDestroy (config);

//     return true;
// }

// /*
//  * A couple of typos escaped into the library
//  */
// int
// FcConfigGetRescanInverval (FcConfig *config)
// {
//     return FcConfigGetRescanInterval (config);
// }

// FcBool
// FcConfigSetRescanInverval (FcConfig *config, int rescanInterval)
// {
//     return FcConfigSetRescanInterval (config, rescanInterval);
// }

// FcBool
// FcConfigAddRule (FcConfig	*config,
// 		 FcRule		*rule,
// 		 FcMatchKind	kind)
// {
//     /* deprecated */
//     return false;
// }

/* The bulk of the time in FcConfigSubstitute is spent walking
 * lists of family names. We speed this up with a hash table.
 * Since we need to take the ignore-blanks option into account,
 * we use two separate hash tables.
 */
// typedef struct
// {
//   int count;
// } FamilyTableEntry;

type FamilyTable struct {
	family_blank_hash familyBlankHash
	family_hash       familyHash
}

func newFamilyTable(p *FcPattern) FamilyTable {
	table := FamilyTable{
		family_blank_hash: make(familyBlankHash),
		family_hash:       make(familyHash),
	}

	e := p.elts[FC_FAMILY]
	table.add(e)
	return table
}

func (table FamilyTable) lookup(op FcOp, s string) bool {
	flags := op.getFlags()
	var has bool

	if (flags & FcOpFlagIgnoreBlanks) != 0 {
		_, has = table.family_blank_hash.lookup(s)
	} else {
		_, has = table.family_hash.lookup(s)
	}

	return has
}

func (table FamilyTable) add(values FcValueList) {
	for _, ll := range values {
		s := ll.value.(string)

		count, _ := table.family_hash.lookup(s)
		count++
		table.family_hash.add(s, count)

		count, _ = table.family_blank_hash.lookup(s)
		count++
		table.family_blank_hash.add(s, count)
	}
}

func (table FamilyTable) del(s string) {
	count, ok := table.family_hash.lookup(s)
	if ok {
		count--
		if count == 0 {
			table.family_hash.del(s)
		} else {
			table.family_hash.add(s, count)
		}
	}

	count, ok = table.family_blank_hash.lookup(s)
	if ok {
		count--
		if count == 0 {
			table.family_blank_hash.del(s)
		} else {
			table.family_blank_hash.add(s, count)
		}
	}
}

// static FcBool
// copy_string (const void *src, void **dest)
// {
//   *dest = strdup ((char *)src);
//   return true;
// }

// return the index into values, or -1
func matchValueList(p, pPat *FcPattern, kind FcMatchKind,
	t *FcTest, values FcValueList, table *FamilyTable) int {

	var (
		value FcValue
		e     = t.expr
		ret   = -1
	)

	for e != nil {
		// Compute the value of the match expression
		if e.op == FcOpComma {
			tree := e.u.(exprTree)
			value = tree.left.FcConfigEvaluate(p, pPat, kind)
			e = tree.right
		} else {
			value = e.FcConfigEvaluate(p, pPat, kind)
			e = nil
		}

		if t.object == FC_FAMILY && table != nil {
			if t.op == FcOpEqual || t.op == FcOpListing {
				if !table.lookup(t.op, value.(string)) {
					ret = -1
					continue
				}
			}
			if t.op == FcOpNotEqual && t.qual == FcQualAll {
				ret = -1
				if !table.lookup(t.op, value.(string)) {
					ret = 0
				}
				continue
			}
		}

		for i, v := range values {
			// Compare the pattern value to the match expression value
			if FcConfigCompareValue(v.value, t.op, value) {
				if ret == -1 {
					ret = i
				}
				if t.qual != FcQualAll {
					break
				}
			} else {
				if t.qual == FcQualAll {
					ret = -1
					break
				}
			}
		}
	}
	return ret
}

// FcBool
// FcConfigSubstitute (FcConfig	*config,
// 		    FcPattern	*p,
// 		    FcMatchKind	kind)
// {
//     return FcConfigSubstituteWithPat (config, p, 0, kind);
// }

// #if defined (_WIN32)

// static FcChar8 fontconfig_path[1000] = ""; /* MT-dontcare */
// FcChar8 fontconfig_instprefix[1000] = ""; /* MT-dontcare */

// #  if (defined (PIC) || defined (DLL_EXPORT))

// BOOL WINAPI
// DllMain (HINSTANCE hinstDLL,
// 	 DWORD     fdwReason,
// 	 LPVOID    lpvReserved);

// BOOL WINAPI
// DllMain (HINSTANCE hinstDLL,
// 	 DWORD     fdwReason,
// 	 LPVOID    lpvReserved)
// {
//   FcChar8 *p;

//   switch (fdwReason) {
//   case DLL_PROCESS_ATTACH:
//       if (!GetModuleFileName ((HMODULE) hinstDLL, (LPCH) fontconfig_path,
// 			      sizeof (fontconfig_path)))
// 	  break;

//       /* If the fontconfig DLL is in a "bin" or "lib" subfolder,
//        * assume it's a Unix-style installation tree, and use
//        * "etc/fonts" in there as FONTCONFIG_PATH. Otherwise use the
//        * folder where the DLL is as FONTCONFIG_PATH.
//        */
//       p = (FcChar8 *) strrchr ((const char *) fontconfig_path, '\\');
//       if (p)
//       {
// 	  *p = '\0';
// 	  p = (FcChar8 *) strrchr ((const char *) fontconfig_path, '\\');
// 	  if (p && (FcStrCmpIgnoreCase (p + 1, (const FcChar8 *) "bin") == 0 ||
// 		    FcStrCmpIgnoreCase (p + 1, (const FcChar8 *) "lib") == 0))
// 	      *p = '\0';
// 	  strcat ((char *) fontconfig_instprefix, (char *) fontconfig_path);
// 	  strcat ((char *) fontconfig_path, "\\etc\\fonts");
//       }
//       else
//           fontconfig_path[0] = '\0';

//       break;
//   }

//   return TRUE;
// }

// #  endif /* !PIC */

// #undef FONTCONFIG_PATH
// #define FONTCONFIG_PATH fontconfig_path

// #endif /* !_WIN32 */

// #ifndef FONTCONFIG_FILE
// #define FONTCONFIG_FILE	"fonts.conf"
// #endif

// static FcChar8 *
// FcConfigFileExists (const FcChar8 *dir, const FcChar8 *file)
// {
//     FcChar8    *path;
//     int         size, osize;

//     if (!dir)
// 	dir = (FcChar8 *) "";

//     osize = strlen ((char *) dir) + 1 + strlen ((char *) file) + 1;
//     /*
//      * workaround valgrind warning because glibc takes advantage of how it knows memory is
//      * allocated to implement strlen by reading in groups of 4
//      */
//     size = (osize + 3) & ~3;

//     path = malloc (size);
//     if (!path)
// 	return 0;

//     strcpy ((char *) path, (const char *) dir);
//     /* make sure there's a single separator */
// #ifdef _WIN32
//     if ((!path[0] || (path[strlen((char *) path)-1] != '/' &&
// 		      path[strlen((char *) path)-1] != '\\')) &&
// 	!(file[0] == '/' ||
// 	  file[0] == '\\' ||
// 	  (isalpha (file[0]) && file[1] == ':' && (file[2] == '/' || file[2] == '\\'))))
// 	strcat ((char *) path, "\\");
// #else
//     if ((!path[0] || path[strlen((char *) path)-1] != '/') && file[0] != '/')
// 	strcat ((char *) path, "/");
//     else
// 	osize--;
// #endif
//     strcat ((char *) path, (char *) file);

//     if (access ((char *) path, R_OK) == 0)
// 	return path;

//     FcStrFree (path);

//     return 0;
// }

// static FcChar8 **
// FcConfigGetPath (void)
// {
//     FcChar8    **path;
//     FcChar8    *env, *e, *colon;
//     FcChar8    *dir;
//     int	    npath;
//     int	    i;

//     npath = 2;	/* default dir + null */
//     env = (FcChar8 *) getenv ("FONTCONFIG_PATH");
//     if (env)
//     {
// 	e = env;
// 	npath++;
// 	for (*e)
// 	    if (*e++ == FC_SEARCH_PATH_SEPARATOR)
// 		npath++;
//     }
//     path = calloc (npath, sizeof (FcChar8 *));
//     if (!path)
// 	goto bail0;
//     i = 0;

//     if (env)
//     {
// 	e = env;
// 	for (*e)
// 	{
// 	    colon = (FcChar8 *) strchr ((char *) e, FC_SEARCH_PATH_SEPARATOR);
// 	    if (!colon)
// 		colon = e + strlen ((char *) e);
// 	    path[i] = malloc (colon - e + 1);
// 	    if (!path[i])
// 		goto bail1;
// 	    strncpy ((char *) path[i], (const char *) e, colon - e);
// 	    path[i][colon - e] = '\0';
// 	    if (*colon)
// 		e = colon + 1;
// 	    else
// 		e = colon;
// 	    i++;
// 	}
//     }

// #ifdef _WIN32
// 	if (fontconfig_path[0] == '\0')
// 	{
// 		char *p;
// 		if(!GetModuleFileName(nil, (LPCH) fontconfig_path, sizeof(fontconfig_path)))
// 			goto bail1;
// 		p = strrchr ((const char *) fontconfig_path, '\\');
// 		if (p) *p = '\0';
// 		strcat ((char *) fontconfig_path, "\\fonts");
// 	}
// #endif
//     dir = (FcChar8 *) FONTCONFIG_PATH;
//     path[i] = malloc (strlen ((char *) dir) + 1);
//     if (!path[i])
// 	goto bail1;
//     strcpy ((char *) path[i], (const char *) dir);
//     return path;

// bail1:
//     for (i = 0; path[i]; i++)
// 	free (path[i]);
//     free (path);
// bail0:
//     return 0;
// }

// static void
// FcConfigFreePath (FcChar8 **path)
// {
//     FcChar8    **p;

//     for (p = path; *p; p++)
// 	free (*p);
//     free (path);
// }

// static FcBool	_FcConfigHomeEnabled = true; /* MT-goodenough */

// FcChar8 *
// FcConfigHome (void)
// {
//     if (_FcConfigHomeEnabled)
//     {
//         char *home = getenv ("HOME");

// #ifdef _WIN32
// 	if (home == nil)
// 	    home = getenv ("USERPROFILE");
// #endif

// 	return (FcChar8 *) home;
//     }
//     return 0;
// }

// FcChar8 *
// FcConfigXdgCacheHome (void)
// {
//     const char *env = getenv ("XDG_CACHE_HOME");
//     FcChar8 *ret = nil;

//     if (!_FcConfigHomeEnabled)
// 	return nil;
//     if (env && env[0])
// 	ret = FcStrCopy ((const FcChar8 *)env);
//     else
//     {
// 	const FcChar8 *home = FcConfigHome ();
// 	size_t len = home ? strlen ((const char *)home) : 0;

// 	ret = malloc (len + 7 + 1);
// 	if (ret)
// 	{
// 	    if (home)
// 		memcpy (ret, home, len);
// 	    memcpy (&ret[len], FC_DIR_SEPARATOR_S ".cache", 7);
// 	    ret[len + 7] = 0;
// 	}
//     }

//     return ret;
// }

// FcChar8 *
// FcConfigXdgConfigHome (void)
// {
//     const char *env = getenv ("XDG_CONFIG_HOME");
//     FcChar8 *ret = nil;

//     if (!_FcConfigHomeEnabled)
// 	return nil;
//     if (env)
// 	ret = FcStrCopy ((const FcChar8 *)env);
//     else
//     {
// 	const FcChar8 *home = FcConfigHome ();
// 	size_t len = home ? strlen ((const char *)home) : 0;

// 	ret = malloc (len + 8 + 1);
// 	if (ret)
// 	{
// 	    if (home)
// 		memcpy (ret, home, len);
// 	    memcpy (&ret[len], FC_DIR_SEPARATOR_S ".config", 8);
// 	    ret[len + 8] = 0;
// 	}
//     }

//     return ret;
// }

// FcChar8 *
// FcConfigXdgDataHome (void)
// {
//     const char *env = getenv ("XDG_DATA_HOME");
//     FcChar8 *ret = nil;

//     if (!_FcConfigHomeEnabled)
// 	return nil;
//     if (env)
// 	ret = FcStrCopy ((const FcChar8 *)env);
//     else
//     {
// 	const FcChar8 *home = FcConfigHome ();
// 	size_t len = home ? strlen ((const char *)home) : 0;

// 	ret = malloc (len + 13 + 1);
// 	if (ret)
// 	{
// 	    if (home)
// 		memcpy (ret, home, len);
// 	    memcpy (&ret[len], FC_DIR_SEPARATOR_S ".local" FC_DIR_SEPARATOR_S "share", 13);
// 	    ret[len + 13] = 0;
// 	}
//     }

//     return ret;
// }

// FcBool
// FcConfigEnableHome (FcBool enable)
// {
//     FcBool  prev = _FcConfigHomeEnabled;
//     _FcConfigHomeEnabled = enable;
//     return prev;
// }

// FcChar8 *
// FcConfigGetFilename (FcConfig      *config,
// 		     const FcChar8 *url)
// {
//     FcChar8    *file, *dir, **path, **p;
//     const FcChar8 *sysroot;

//     config = fallbackConfig (config);
//     if (!config)
// 	return nil;
//     sysroot = FcConfigGetSysRoot (config);
//     if (!url || !*url)
//     {
// 	url = (FcChar8 *) getenv ("FONTCONFIG_FILE");
// 	if (!url)
// 	    url = (FcChar8 *) FONTCONFIG_FILE;
//     }
//     file = 0;

//     if (FcStrIsAbsoluteFilename(url))
//     {
// 	if (sysroot)
// 	{
// 	    size_t len = strlen ((const char *) sysroot);

// 	    /* Workaround to avoid adding sysroot repeatedly */
// 	    if (strncmp ((const char *) url, (const char *) sysroot, len) == 0)
// 		sysroot = nil;
// 	}
// 	file = FcConfigFileExists (sysroot, url);
// 	goto bail;
//     }

//     if (*url == '~')
//     {
// 	dir = FcConfigHome ();
// 	if (dir)
// 	{
// 	    FcChar8 *s;

// 	    if (sysroot)
// 		s = FcStrBuildFilename (sysroot, dir, nil);
// 	    else
// 		s = dir;
// 	    file = FcConfigFileExists (s, url + 1);
// 	    if (sysroot)
// 		FcStrFree (s);
// 	}
// 	else
// 	    file = 0;
//     }
//     else
//     {
// 	path = FcConfigGetPath ();
// 	if (!path)
// 	{
// 	    file = nil;
// 	    goto bail;
// 	}
// 	for (p = path; *p; p++)
// 	{
// 	    FcChar8 *s;

// 	    if (sysroot)
// 		s = FcStrBuildFilename (sysroot, *p, nil);
// 	    else
// 		s = *p;
// 	    file = FcConfigFileExists (s, url);
// 	    if (sysroot)
// 		FcStrFree (s);
// 	    if (file)
// 		break;
// 	}
// 	FcConfigFreePath (path);
//     }
// bail:
//     FcConfigDestroy (config);

//     return file;
// }

// FcChar8 *
// FcConfigFilename (const FcChar8 *url)
// {
//     return FcConfigGetFilename (nil, url);
// }

// FcChar8 *
// FcConfigRealFilename (FcConfig		*config,
// 		      const FcChar8	*url)
// {
//     FcChar8 *n = FcConfigGetFilename (config, url);

//     if (n)
//     {
// 	FcChar8 buf[FC_PATH_MAX];
// 	ssize_t len;
// 	struct stat sb;

// 	if ((len = FcReadLink (n, buf, sizeof (buf) - 1)) != -1)
// 	{
// 	    buf[len] = 0;

// 	    /* We try to pick up a config from FONTCONFIG_FILE
// 	     * when url is null. don't try to address the real filename
// 	     * if it is a named pipe.
// 	     */
// 	    if (!url && FcStat (n, &sb) == 0 && S_ISFIFO (sb.st_mode))
// 		return n;
// 	    else if (!FcStrIsAbsoluteFilename (buf))
// 	    {
// 		FcChar8 *dirname = FcStrDirname (n);
// 		FcStrFree (n);
// 		if (!dirname)
// 		    return nil;

// 		FcChar8 *path = FcStrBuildFilename (dirname, buf, nil);
// 		FcStrFree (dirname);
// 		if (!path)
// 		    return nil;

// 		n = FcStrCanonFilename (path);
// 		FcStrFree (path);
// 	    }
// 	    else
// 	    {
// 		FcStrFree (n);
// 		n = FcStrdup (buf);
// 	    }
// 	}
//     }

//     return n;
// }

// /*
//  * Manage the application-specific fonts
//  */

// FcBool
// FcConfigAppFontAddFile (config *FcConfig,
// 			const FcChar8  *file)
// {
//     FcFontSet	*set;
//     FcStrSet	*subdirs;
//     FcStrList	*sublist;
//     FcChar8	*subdir;
//     FcBool	ret = true;

//     config = fallbackConfig (config);
//     if (!config)
// 	return false;

//     subdirs = FcStrSetCreateEx (FCSS_GROW_BY_64);
//     if (!subdirs)
//     {
// 	ret = false;
// 	goto bail;
//     }

//     set = FcConfigGetFonts (config, FcSetApplication);
//     if (!set)
//     {
// 	set = FcFontSetCreate ();
// 	if (!set)
// 	{
// 	    FcStrSetDestroy (subdirs);
// 	    ret = false;
// 	    goto bail;
// 	}
// 	FcConfigSetFonts (config, set, FcSetApplication);
//     }

//     if (!FcFileScanConfig (set, subdirs, file, config))
//     {
// 	FcStrSetDestroy (subdirs);
// 	ret = false;
// 	goto bail;
//     }
//     if ((sublist = FcStrListCreate (subdirs)))
//     {
// 	for ((subdir = FcStrListNext (sublist)))
// 	{
// 	    FcConfigAppFontAddDir (config, subdir);
// 	}
// 	FcStrListDone (sublist);
//     }
//     FcStrSetDestroy (subdirs);
// bail:
//     FcConfigDestroy (config);

//     return ret;
// }

// FcBool
// FcConfigAppFontAddDir (FcConfig	    *config,
// 		       const FcChar8   *dir)
// {
//     FcFontSet	*set;
//     FcStrSet	*dirs;
//     FcBool	ret = true;

//     config = fallbackConfig (config);
//     if (!config)
// 	return false;

//     dirs = FcStrSetCreateEx (FCSS_GROW_BY_64);
//     if (!dirs)
//     {
// 	ret = false;
// 	goto bail;
//     }

//     set = FcConfigGetFonts (config, FcSetApplication);
//     if (!set)
//     {
// 	set = FcFontSetCreate ();
// 	if (!set)
// 	{
// 	    FcStrSetDestroy (dirs);
// 	    ret = false;
// 	    goto bail;
// 	}
// 	FcConfigSetFonts (config, set, FcSetApplication);
//     }

//     FcStrSetAddFilename (dirs, dir);

//     if (!FcConfigAddDirList (config, FcSetApplication, dirs))
//     {
// 	FcStrSetDestroy (dirs);
// 	ret = false;
// 	goto bail;
//     }
//     FcStrSetDestroy (dirs);
// bail:
//     FcConfigDestroy (config);

//     return ret;
// }

// void
// FcConfigAppFontClear (FcConfig	    *config)
// {
//     config = fallbackConfig (config);
//     if (!config)
// 	return;

//     FcConfigSetFonts (config, 0, FcSetApplication);

//     FcConfigDestroy (config);
// }

// /*
//  * Manage filename-based font source selectors
//  */

// FcBool
// FcConfigGlobAdd (FcConfig	*config,
// 		 const FcChar8  *glob,
// 		 FcBool		accept)
// {
//     FcStrSet	*set = accept ? config.acceptGlobs : config.rejectGlobs;

//     return FcStrSetAdd (set, glob);
// }

// static FcBool
// FcConfigGlobsMatch (const FcStrSet	*globs,
// 		    const FcChar8	*string)
// {
//     int	i;

//     for (i = 0; i < globs.num; i++)
// 	if (FcStrGlobMatch (globs.strs[i], string))
// 	    return true;
//     return false;
// }

// FcBool
// FcConfigAcceptFilename (FcConfig	*config,
// 			const FcChar8	*filename)
// {
//     if (FcConfigGlobsMatch (config.acceptGlobs, filename))
// 	return true;
//     if (FcConfigGlobsMatch (config.rejectGlobs, filename))
// 	return false;
//     return true;
// }

// /*
//  * Manage font-pattern based font source selectors
//  */

// FcBool
// FcConfigPatternsAdd (FcConfig	*config,
// 		     FcPattern	*pattern,
// 		     FcBool	accept)
// {
//     FcFontSet	*set = accept ? config.acceptPatterns : config.rejectPatterns;

//     return FcFontSetAdd (set, pattern);
// }

// static FcBool
// FcConfigPatternsMatch (const FcFontSet	*patterns,
// 		       const FcPattern	*font)
// {
//     int i;

//     for (i = 0; i < patterns.nfont; i++)
// 	if (FcListPatternMatchAny (patterns.fonts[i], font))
// 	    return true;
//     return false;
// }

// FcBool
// FcConfigAcceptFont (FcConfig	    *config,
// 		    const FcPattern *font)
// {
//     if (FcConfigPatternsMatch (config.acceptPatterns, font))
// 	return true;
//     if (FcConfigPatternsMatch (config.rejectPatterns, font))
// 	return false;
//     return true;
// }

// const FcChar8 *
// FcConfigGetSysRoot (const FcConfig *config)
// {
//     if (!config)
//     {
// 	config = FcConfigGetCurrent ();
// 	if (!config)
// 	    return nil;
//     }
//     return config.sysRoot;
// }

// void
// FcConfigSetSysRoot (FcConfig      *config,
// 		    const FcChar8 *sysroot)
// {
//     FcChar8 *s = nil;
//     FcBool init = false;
//     int nretry = 3;

// retry:
//     if (!config)
//     {
// 	/* We can't use FcConfigGetCurrent() here to ensure
// 	 * the sysroot is set prior to initialize FcConfig,
// 	 * to avoid loading caches from non-sysroot dirs.
// 	 * So postpone the initialization later.
// 	 */
// 	config = fc_atomic_ptr_get (&_fcConfig);
// 	if (!config)
// 	{
// 	    config = FcConfigCreate ();
// 	    if (!config)
// 		return;
// 	    init = true;
// 	}
//     }

//     if (sysroot)
//     {
// 	s = FcConfigRealPath(sysroot);
// 	if (!s)
// 	    return;
//     }

//     if (config.sysRoot)
// 	FcStrFree (config.sysRoot);

//     config.sysRoot = s;
//     if (init)
//     {
// 	config = FcInitLoadOwnConfigAndFonts (config);
// 	if (!config)
// 	{
// 	    /* Something failed. this is usually unlikely. so retrying */
// 	    init = false;
// 	    if (--nretry == 0)
// 	    {
// 		fprintf (stderr, "Fontconfig warning: Unable to initialize config and retry limit exceeded. sysroot functionality may not work as expected.\n");
// 		return;
// 	    }
// 	    goto retry;
// 	}
// 	FcConfigSetCurrent (config);
// 	/* FcConfigSetCurrent() increases the refcount.
// 	 * decrease it here to avoid the memory leak.
// 	 */
// 	FcConfigDestroy (config);
//     }
// }

// FcRuleSet *
// FcRuleSetCreate (const FcChar8 *name)
// {
//     FcRuleSet *ret = (FcRuleSet *) malloc (sizeof (FcRuleSet));
//     FcMatchKind k;
//     const FcChar8 *p;

//     if (!name)
// 	p = (const FcChar8 *)"";
//     else
// 	p = name;

//     if (ret)
//     {
// 	ret.name = FcStrdup (p);
// 	ret.description = nil;
// 	ret.domain = nil;
// 	for (k = FcMatchKindBegin; k < FcMatchKindEnd; k++)
// 	    ret.subst[k] = FcPtrListCreate (FcDestroyAsRule);
// 	FcRefInit (&ret.ref, 1);
//     }

//     return ret;
// }

// void
// FcRuleSetDestroy (FcRuleSet *rs)
// {
//     FcMatchKind k;

//     if (!rs)
// 	return;
//     if (FcRefDec (&rs.ref) != 1)
// 	return;

//     if (rs.name)
// 	FcStrFree (rs.name);
//     if (rs.description)
// 	FcStrFree (rs.description);
//     if (rs.domain)
// 	FcStrFree (rs.domain);
//     for (k = FcMatchKindBegin; k < FcMatchKindEnd; k++)
// 	FcPtrListDestroy (rs.subst[k]);

//     free (rs);
// }

// void
// FcRuleSetReference (FcRuleSet *rs)
// {
//     if (!FcRefIsConst (&rs.ref))
// 	FcRefInc (&rs.ref);
// }

// void
// FcRuleSetEnable (FcRuleSet	*rs,
// 		 FcBool		flag)
// {
//     if (rs)
//     {
// 	rs.enabled = flag;
// 	/* XXX: we may want to provide a feature
// 	 * to enable/disable rulesets through API
// 	 * in the future?
// 	 */
//     }
// }

// void
// FcRuleSetAddDescription (FcRuleSet	*rs,
// 			 const FcChar8	*domain,
// 			 const FcChar8	*description)
// {
//     if (rs.domain)
// 	FcStrFree (rs.domain);
//     if (rs.description)
// 	FcStrFree (rs.description);

//     rs.domain = domain ? FcStrdup (domain) : nil;
//     rs.description = description ? FcStrdup (description) : nil;
// }

// int
// FcRuleSetAdd (FcRuleSet		*rs,
// 	      FcRule		*rule,
// 	      FcMatchKind	kind)
// {
//     FcPtrListIter iter;
//     FcRule *r;
//     int n = 0, ret;

//     if (!rs ||
//        kind < FcMatchKindBegin || kind >= FcMatchKindEnd)
// 	return -1;
//     FcPtrListIterInitAtLast (rs.subst[kind], &iter);
//     if (!FcPtrListIterAdd (rs.subst[kind], &iter, rule))
// 	return -1;

//     for (r = rule; r; r = r.next)
//     {
// 	switch (r.type_)
// 	{
// 	case FcRuleTest:
// 	    if (r.u.test)
// 	    {
// 		if (r.u.test.kind == FcMatchDefault)
// 		    r.u.test.kind = kind;
// 		if (n < r.u.test.object)
// 		    n = r.u.test.object;
// 	    }
// 	    break;
// 	case FcRuleEdit:
// 	    if (n < r.u.edit.object)
// 		n = r.u.edit.object;
// 	    break;
// 	default:
// 	    break;
// 	}
//     }
//     if debugMode
//     {
// 	printf ("Add Rule(kind:%d, name: %s) ", kind, rs.name);
// 	FcRulePrint (rule);
//     }
//     ret = FC_OBJ_ID (n) - FC_MAX_BASE;

//     return ret < 0 ? 0 : ret;
// }

// void
// FcConfigFileInfoIterInit (FcConfig		*config,
// 			  FcConfigFileInfoIter	*iter)
// {
//     FcConfig *c;
//     FcPtrListIter *i = (FcPtrListIter *)iter;

//     if (!config)
// 	c = FcConfigGetCurrent ();
//     else
// 	c = config;
//     FcPtrListIterInit (c.rulesetList, i);
// }

// FcBool
// FcConfigFileInfoIterNext (FcConfig		*config,
// 			  FcConfigFileInfoIter	*iter)
// {
//     FcConfig *c;
//     FcPtrListIter *i = (FcPtrListIter *)iter;

//     if (!config)
// 	c = FcConfigGetCurrent ();
//     else
// 	c = config;
//     if (FcPtrListIterIsValid (c.rulesetList, i))
//     {
// 	FcPtrListIterNext (c.rulesetList, i);
//     }
//     else
// 	return false;

//     return true;
// }

// FcBool
// FcConfigFileInfoIterGet (FcConfig		*config,
// 			 FcConfigFileInfoIter	*iter,
// 			 FcChar8		**name,
// 			 FcChar8		**description,
// 			 FcBool			*enabled)
// {
//     FcConfig *c;
//     FcRuleSet *r;
//     FcPtrListIter *i = (FcPtrListIter *)iter;

//     if (!config)
// 	c = FcConfigGetCurrent ();
//     else
// 	c = config;
//     if (!FcPtrListIterIsValid (c.rulesetList, i))
// 	return false;
//     r = FcPtrListIterGetValue (c.rulesetList, i);
//     if (name)
// 	*name = FcStrdup (r.name && r.name[0] ? r.name : (const FcChar8 *) "fonts.conf");
//     if (description)
// 	*description = FcStrdup (!r.description ? _("No description") :
// 				 dgettext (r.domain ? (const char *) r.domain : GETTEXT_PACKAGE "-conf",
// 					   (const char *) r.description));
//     if (enabled)
// 	*enabled = r.enabled;

//     return true;
// }