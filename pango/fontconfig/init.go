package fontconfig

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ported from fontconfig/src/fcinit.c Copyright © 2001 Keith Packard

//  #if defined(FC_ATOMIC_INT_NIL)
//  #pragma message("Could not find any system to define atomic_int macros, library may NOT be thread-safe.")
//  #endif
//  #if defined(FC_MUTEX_IMPL_NIL)
//  #pragma message("Could not find any system to define mutex macros, library may NOT be thread-safe.")
//  #endif
//  #if defined(FC_ATOMIC_INT_NIL) || defined(FC_MUTEX_IMPL_NIL)
//  #pragma message("To suppress these warnings, define FC_NO_MT.")
//  #endif

const (
	CONFIGDIR        = "/usr/local/etc/fonts/conf.d"
	FC_CACHEDIR      = "/var/local/cache/fontconfig"
	FC_DEFAULT_FONTS = "<dir>/usr/share/fonts</dir>"
	FC_TEMPLATEDIR   = "/usr/local/share/fontconfig/conf.avail"
)

func initFallbackConfig(sysroot string) *FcConfig {
	fallback := fmt.Sprintf(`	
	 <fontconfig>
	  	%s
		<dir prefix="xdg">fonts</dir>
		<cachedir>%s</cachedir>
		<cachedir prefix="xdg">fontconfig</cachedir>
		<include ignore_missing="yes">%s</include>
		<include ignore_missing="yes" prefix="xdg">fontconfig/conf.d</include>
		<include ignore_missing="yes" prefix="xdg">fontconfig/fonts.conf</include>
	 </fontconfig>
	 `, FC_DEFAULT_FONTS, FC_CACHEDIR, CONFIGDIR)

	config := NewFcConfig()
	config.setSysRoot(sysroot)

	_ = config.ParseAndLoadFromMemory([]byte(fallback), os.Stdout)

	return config
}

//  int
//  FcGetVersion (void)
//  {
// 	 return FC_VERSION;
//  }

// Load the configuration files
func initLoadOwnConfig(logger io.Writer) *FcConfig {
	config := NewFcConfig()

	if err := config.parseConfig(logger, "", true); err != nil {
		sysroot := config.getSysRoot()
		fallback := initFallbackConfig(sysroot)
		return fallback
	}

	_ = config.parseConfig(logger, FC_TEMPLATEDIR, false)

	if len(config.cacheDirs) == 0 {
		//  FcChar8 *prefix, *p;
		//  size_t plen;
		haveOwn := false

		envFile := os.Getenv("FONTCONFIG_FILE")
		envPath := os.Getenv("FONTCONFIG_PATH")
		if envFile != "" || envPath != "" {
			haveOwn = true
		}

		if !haveOwn {
			fmt.Fprintf(logger, "fontconfig: no <cachedir> elements found. Check configuration.\n")
			fmt.Fprintf(logger, "fontconfig: adding <cachedir>%s</cachedir>\n", FC_CACHEDIR)
		}
		prefix := xdgCacheHome()
		if prefix == "" {
			return initFallbackConfig(config.getSysRoot())
		}
		prefix = filepath.Join(prefix, "fontconfig")
		if !haveOwn {
			fmt.Fprintf(logger, "fontconfig: adding <cachedir prefix=\"xdg\">fontconfig</cachedir>\n")
		}

		err := config.addCacheDir(FC_CACHEDIR)
		if err == nil {
			err = config.addCacheDir(prefix)
		}
		if err != nil {
			fmt.Fprintf(logger, "fontconfig: %s", err)
			return initFallbackConfig(config.getSysRoot())
		}
	}

	return config
}

//  FcConfig *
//  FcInitLoadConfig (void)
//  {
// 	 return initLoadOwnConfig (NULL);
//  }

// Loads the default configuration file and builds information about the
// available fonts.  Returns the resulting configuration.
func initLoadConfigAndFonts(logger io.Writer) *FcConfig {
	config := initLoadOwnConfig(logger)
	config.FcConfigBuildFonts()
	return config
}

//  /*
//   * Initialize the default library configuration
//   */
//  FcBool
//  FcInit (void)
//  {
// 	 return FcConfigInit ();
//  }

//  /*
//   * Free all library-allocated data structures.
//   */
//  void
//  FcFini (void)
//  {
// 	 FcConfigFini ();
// 	 FcConfigPathFini ();
// 	 FcDefaultFini ();
// 	 FcObjectFini ();
// 	 FcCacheFini ();
//  }

//  /*
//   * Reread the configuration and available font lists
//   */
//  FcBool
//  FcInitReinitialize (void)
//  {
// 	 FcConfig	*config;
// 	 FcBool	ret;

// 	 config = FcInitLoadConfigAndFonts ();
// 	 if (!config)
// 	 return FcFalse;
// 	 ret = FcConfigSetCurrent (config);
// 	 /* FcConfigSetCurrent() increases the refcount.
// 	  * decrease it here to avoid the memory leak.
// 	  */
// 	 FcConfigDestroy (config);

// 	 return ret;
//  }

//  FcBool
//  FcInitBringUptoDate (void)
//  {
// 	 FcConfig	*config = FcConfigReference (NULL);
// 	 FcBool	ret = FcTrue;
// 	 time_t	now;

// 	 if (!config)
// 	 return FcFalse;
// 	 /*
// 	  * rescanInterval == 0 disables automatic up to date
// 	  */
// 	 if (config.rescanInterval == 0)
// 	 goto bail;
// 	 /*
// 	  * Check no more often than rescanInterval seconds
// 	  */
// 	 now = time (0);
// 	 if (config.rescanTime + config.rescanInterval - now > 0)
// 	 goto bail;
// 	 /*
// 	  * If up to date, don't reload configuration
// 	  */
// 	 if (FcConfigUptoDate (0))
// 	 goto bail;
// 	 ret = FcInitReinitialize ();
//  bail:
// 	 FcConfigDestroy (config);

// 	 return ret;
//  }
