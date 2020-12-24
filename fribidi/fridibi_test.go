package fribidi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

//  #define appname "fribidi"

//  #define MAX_STR_LEN 65000

//  #define ALLOCATE(tp,ln) ((tp *) fribidi_malloc (sizeof (tp) * (ln)))

//  bool do_break, do_pad, do_mirror, do_reorder_nsm, do_clean;
//  bool show_input, show_visual, show_basedir;
//  bool show_ltov, show_vtol, show_levels;
//  const int default_text_width = 80;
//  int text_width;
//  const char *char_set;
//  const char *bol_text, *eol_text;
//  FriBidiParType input_base_direction;
//  int char_set_num;

//    /* Break help string into little ones, to assure ISO C89 conformance */
//    printf ("Usage: " appname " [OPTION]... [FILE]...\n"
// 	   "A command line interface for the " FRIBIDI_NAME " library.\n"
// 	   "Convert a logical string to visual.\n"
// 	   "\n"
// 	   "  -h, --help            Display this information and exit\n"
// 	   "  -V, --version         Display version information and exit\n"
// 	   "  -v, --verbose         Verbose mode, same as --basedir --ltov --vtol\n"
// 	   "                        --levels\n");
//    printf ("  -d, --debug           Output debug information\n"
// 	   "  -t, --test            Test " FRIBIDI_NAME
// 	   ", same as --clean --nobreak\n"
// 	   "                        --showinput --reordernsm --width %d\n",
// 	   default_text_width);
//    printf ("  -c, --charset CS      Specify character set, default is %s\n"
// 	   "      --charsetdesc CS  Show descriptions for character set CS and exit\n"
// 	   "      --caprtl          Old style: set character set to CapRTL\n",
// 	   char_set);
//    printf ("      --showinput       Output the input string too\n"
// 	   "      --nopad           Do not right justify RTL lines\n"
// 	   "      --nobreak         Do not break long lines\n"
// 	   "  -w, --width W         Screen width for padding, default is %d, but if\n"
// 	   "                        environment variable COLUMNS is defined, its value\n"
// 	   "                        will be used, --width overrides both of them.\n",
// 	   default_text_width);
//    printf
// 	 ("  -B, --bol BOL         Output string BOL before the visual string\n"
// 	  "  -E, --eol EOL         Output string EOL after the visual string\n"
// 	  "      --rtl             Force base direction to RTL\n"
// 	  "      --ltr             Force base direction to LTR\n"
// 	  "      --wrtl            Set base direction to RTL if no strong character found\n");
//    printf
// 	 ("      --wltr            Set base direction to LTR if no strong character found\n"
// 	  "                        (default)\n"
// 	  "      --nomirror        Turn mirroring off, to do it later\n"
// 	  "      --reordernsm      Reorder NSM sequences to follow their base character\n"
// 	  "      --clean           Remove explicit format codes in visual string\n"
// 	  "                        output, currently does not affect other outputs\n"
// 	  "      --basedir         Output Base Direction\n");
//    printf ("      --ltov            Output Logical to Visual position map\n"
// 	   "      --vtol            Output Visual to Logical position map\n"
// 	   "      --levels          Output Embedding Levels\n"
// 	   "      --novisual        Do not output the visual string, to be used with\n"
// 	   "                        --basedir, --ltov, --vtol, --levels\n");
//    printf ("  All string indexes are zero based\n" "\n" "Output:\n"
// 	   "  For each line of input, output something like this:\n"
// 	   "    [input-str` => '][BOL][[padding space]visual-str][EOL]\n"
// 	   "    [\\n base-dir][\\n ltov-map][\\n vtol-map][\\n levels]\n");

const default_text_width = 80

func decode(dec *encoding.Decoder, input []byte) []CharType {
	u, err := dec.Bytes(input)
	if err != nil {
		log.Fatal(err)
	}
	runes := []rune(string(u))
	out := make([]CharType, len(runes))
	for i, r := range runes {
		out[i] = CharType(r)
	}
	return out
}

func parseCharset(filename string) func([]byte) []CharType {
	switch enc := strings.Split(filename, "_")[1]; enc {
	case "UTF-8":
		return func(b []byte) []CharType { return decode(unicode.UTF8.NewDecoder(), b) }
	case "ISO8859-8":
		return func(b []byte) []CharType { return decode(charmap.ISO8859_8.NewDecoder(), b) }
	case "CapRTL":
		return fribidi_cap_rtl_to_unicode
	default:
		panic("unsupported encoding " + enc)
	}
}

func main() {
	//    int exit_val;
	//    bool file_found;
	//    char *s;
	//    FILE *IN;

	do_clean := true
	do_reorder_nsm := true
	show_input := true
	do_break := false
	text_width := default_text_width

	do_pad := true
	do_mirror := true
	show_visual := true
	show_basedir := false
	show_ltov := false
	show_vtol := false
	show_levels := false
	char_set := "UTF-8"
	bol_text := nil
	eol_text := nil
	input_base_direction := ON

	// s = getenv("COLUMNS")
	// if s {
	// 	i = atoi(s)
	// 	if i > 0 {
	// 		text_width = i
	// 	}
	// }

	const CHARSETDESC = 257
	const CAPRTL = 258

	charset := parseCharset(filename)

	// fribidi_set_mirroring(do_mirror)
	// fribidi_set_reorder_nsm(do_reorder_nsm)

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	lines := bytes.Split(b, []byte{'\n'})

	/* Read and process input one line at a time */

	padding_width := text_width
	if show_input {
		padding_width = (text_width - 10) / 2
	}
	break_width := 3 * MAX_STR_LEN
	if do_break {
		break_width = padding_width
	}

	for _, line := range lines {
		//  const char *new_line, *nl_found;
		//  FriBidiChar logical[MAX_STR_LEN];
		//  char outstring[MAX_STR_LEN];
		//  FriBidiParType base;
		//  FriBidiStrIndex len;

		// len = fribidi_charset_to_unicode(char_set_num, S_, len, logical)
		logical := charset(line)
		//    FriBidiChar *visual;
		//    FriBidiStrIndex *ltov, *vtol;
		//    FriBidiLevel *levels;
		//    bool log2vis;

		var (
			visual     []rune
			ltov, vtol []int
			levels     []Level
		)
		if show_visual {
			visual = make([]rune, len(logical)+1)
		}
		if show_ltov {
			ltov = make([]int, len(logical)+1)
		}
		if show_vtol {
			vtol = make([]int, len(logical)+1)
		}
		if show_levels {
			levels = make([]Level, len(logical)+1)
		}

		/* Create a bidi string. */
		base := input_base_direction

		log2vis = fribidi_log2vis(logical, len, &base,
			/* output */
			visual, ltov, vtol, levels)

		if log2vis {

			if show_input {
				printf("%-*s => ", padding_width, S_)
			}

			/* Remove explicit marks, if asked for. */

			if do_clean {
				len = fribidi_remove_bidi_marks(visual, len, ltov, vtol, levels)
			}

			if show_visual {
				printf("%s", nl_found)

				if bol_text {
					printf("%s", bol_text)
				}

				/* Convert it to input charset and print. */
				{
					var idx, st int
					for idx = 0; idx < len; {
						var inlen int

						wid := break_width
						st = idx
						if char_set_num != FRIBIDI_CHAR_SET_CAP_RTL {
							for wid > 0 && idx < len {
								if FRIBIDI_IS_EXPLICIT_OR_ISOLATE_OR_BN_OR_NSM(GetBidiType(visual[idx])) {
									wid -= 0
								} else {
									wid -= 1
								}
								idx++
							}
						} else {
							for wid > 0 && idx < len {
								wid--
								idx++
							}
						}
						if wid < 0 && idx-st > 1 {
							idx--
						}
						inlen = idx - st

						fribidi_unicode_to_charset(char_set_num, visual+st, inlen, outstring)
						if FRIBIDI_IS_RTL(base) {
							var w int
							if do_pad {
								w = (padding_width +
									strlen(outstring) -
									(break_width -
										wid))
							} else {
								w = 0
							}
							printf("%*s", w, outstring)
						} else {
							printf("%s", outstring)
						}
						if idx < len {
							printf("\n")
						}
					}
				}
				if eol_text {
					printf("%s", eol_text)
				}

				nl_found = "\n"
			}
			if show_basedir {
				printf("%s", nl_found)
				if FRIBIDI_DIR_TO_LEVEL(base) {
					printf("Base direction: %s", "R")
				} else {
					printf("Base direction: %s", "L")
				}
				nl_found = "\n"
			}
			if show_ltov {
				printf("%s", nl_found)
				for i := 0; i < len; i++ {
					printf("%ld ", ltov[i])
				}
				nl_found = "\n"
			}
			if show_vtol {
				printf("%s", nl_found)
				for i := 0; i < len; i++ {
					printf("%ld ", vtol[i])
				}
				nl_found = "\n"
			}
			if show_levels {
				printf("%s", nl_found)
				for i := 0; i < len; i++ {
					printf("%d ", levels[i])
				}
				nl_found = "\n"
			}
		} else {
			exit_val = 2
		}

		if *nl_found {
			printf("%s", new_line)
		}
	}
}

// charset CapRTL used only for testing purposes

//  enum
//  {
//  # define _FRIBIDI_ADD_TYPE(TYPE,SYMBOL) TYPE = FRIBIDI_TYPE_##TYPE,
//  # include "fribidi-bidi-types-list.h"
//  # undef _FRIBIDI_ADD_TYPE
//    _FRIBIDI_MAX_TYPES_VALUE
//  };

//  enum
//  {
//  # define _FRIBIDI_ADD_TYPE(TYPE,SYMBOL) DUMMY_##TYPE,
//  # include "fribidi-bidi-types-list.h"
//  # undef _FRIBIDI_ADD_TYPE
//    numTypes
//  };

var capRTLCharTypes = [...]CharType{
	ON, ON, ON, ON, LTR, RTL, ON, ON, ON, ON, ON, ON, ON, BS, RLO, RLE, /* 00-0f */
	LRO, LRE, PDF, WS, LRI, RLI, FSI, PDI, ON, ON, ON, ON, ON, ON, ON, ON, /* 10-1f */
	WS, ON, ON, ON, ET, ON, ON, ON, ON, ON, ON, ET, CS, ON, ES, ES, /* 20-2f */
	EN, EN, EN, EN, EN, EN, AN, AN, AN, AN, CS, ON, ON, ON, ON, ON, /* 30-3f */
	RTL, AL, AL, AL, AL, AL, AL, RTL, RTL, RTL, RTL, RTL, RTL, RTL, RTL, RTL, /* 40-4f */
	RTL, RTL, RTL, RTL, RTL, RTL, RTL, RTL, RTL, RTL, RTL, ON, BS, ON, BN, ON, /* 50-5f */
	NSM, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, /* 60-6f */
	LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, LTR, ON, SS, ON, WS, ON, /* 70-7f */
}

var caprtl_to_unicode = make([]CharType, len(capRTLCharTypes))

/* We do not support surrogates yet */
const FRIBIDI_UNICODE_CHARS = 0x110000

const numTypes = 23

func init() {
	//    int request[numTypes];
	//    int num_types = 0, count = 0;
	//    FriBidiCharType i;
	var (
		mark             [len(capRTLCharTypes)]byte
		num_types, count int
		to_type          [numTypes]CharType
		request          [numTypes]int
	)

	for i, ct := range capRTLCharTypes {
		if ct == GetBidiType(rune(i)) {
			caprtl_to_unicode[i] = CharType(i)
			mark[i] = 1
		} else {
			var j int

			caprtl_to_unicode[i] = FRIBIDI_UNICODE_CHARS
			mark[i] = 0
			if _, ok := fribidi_get_mirror_char(rune(i)); ok {
				fmt.Println("warning: I could not map mirroring character map to itself in CapRTL")
			}

			for j = 0; j < num_types; j++ {
				if to_type[j] == ct {
					break
				}
			}
			if j == num_types {
				num_types++
				to_type[j] = ct
				request[j] = 0
			}
			request[j]++
			count++
		}
		for i = 0; i < 0x10000 && count != 0; i++ { /* Assign BMP chars to CapRTL entries */
			if _, ok := fribidi_get_mirror_char(rune(i)); !ok && !(i < len(capRTLCharTypes) && mark[i] != 0) {
				var j, k int
				t := GetBidiType(rune(i))
				for j = 0; j < num_types; j++ {
					if to_type[j] == t {
						break
					}
				}
				if j >= num_types || request[j] == 0 { /* Do not need this type */
					continue
				}
				for k = 0; k < len(capRTLCharTypes); k++ {
					if caprtl_to_unicode[k] == FRIBIDI_UNICODE_CHARS && to_type[j] == capRTLCharTypes[k] {
						request[j]--
						count--
						caprtl_to_unicode[k] = CharType(i)
						break
					}
				}
			}
		}
		if count != 0 {
			var j int

			fmt.Println("warning: could not find a mapping for CapRTL to Unicode:")
			for j = 0; j < num_types; j++ {
				if request[j] != 0 {
					fmt.Println("  need this type: %d", to_type[j])
				}
			}
		}
	}
}

//  static char
//  fribidi_unicode_to_cap_rtl_c (
//    /* input */
//    FriBidiChar uch
//  )
//  {
//    int i;

//    if (!caprtl_to_unicode)
// 	 init_cap_rtl ();

//    for (i = 0; i < len(capRTLCharTypes); i++)
// 	 if (uch == caprtl_to_unicode[i])
// 	   return (unsigned char) i;
//    return '?';
//  }

const (
	FRIBIDI_CHAR_LRM = 0x200E
	FRIBIDI_CHAR_RLM = 0x200F
	FRIBIDI_CHAR_LRE = 0x202A
	FRIBIDI_CHAR_RLE = 0x202B
	FRIBIDI_CHAR_PDF = 0x202C
	FRIBIDI_CHAR_LRO = 0x202D
	FRIBIDI_CHAR_RLO = 0x202E
	FRIBIDI_CHAR_LRI = 0x2066
	FRIBIDI_CHAR_RLI = 0x2067
	FRIBIDI_CHAR_FSI = 0x2068
	FRIBIDI_CHAR_PDI = 0x2069
)

// Decode
func fribidi_cap_rtl_to_unicode(s []byte) []CharType {
	var us []CharType
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '_' {
			i++
			switch s[i] {
			case '>':
				us = append(us, FRIBIDI_CHAR_LRM)
			case '<':
				us = append(us, FRIBIDI_CHAR_RLM)
			case 'l':
				us = append(us, FRIBIDI_CHAR_LRE)
			case 'r':
				us = append(us, FRIBIDI_CHAR_RLE)
			case 'o':
				us = append(us, FRIBIDI_CHAR_PDF)
			case 'L':
				us = append(us, FRIBIDI_CHAR_LRO)
			case 'R':
				us = append(us, FRIBIDI_CHAR_RLO)
			case 'i':
				us = append(us, FRIBIDI_CHAR_LRI)
			case 'y':
				us = append(us, FRIBIDI_CHAR_RLI)
			case 'f':
				us = append(us, FRIBIDI_CHAR_FSI)
			case 'I':
				us = append(us, FRIBIDI_CHAR_PDI)
			case '_':
				us = append(us, '_')
			default:
				us = append(us, '_')
				i--
			}
		} else {
			us = append(us, caprtl_to_unicode[s[i]])
		}
	}
	return us
}

//  FriBidiStrIndex
//  fribidi_unicode_to_cap_rtl (
//    /* input */
//    const FriBidiChar *us,
//    FriBidiStrIndex len,
//    /* output */
//    char *s
//  )
//  {
//    FriBidiStrIndex i;
//    int j;

//    j = 0;
//    for (i = 0; i < len; i++)
// 	 {
// 	   FriBidiChar ch = us[i];
// 	   if (!FRIBIDI_IS_EXPLICIT (GetBidiType (ch))
// 		   && !FRIBIDI_IS_ISOLATE (GetBidiType (ch))
// 		   && ch != '_' && ch != FRIBIDI_CHAR_LRM && ch != FRIBIDI_CHAR_RLM)
// 	 s[j++] = fribidi_unicode_to_cap_rtl_c (ch);
// 	   else
// 	 {
// 	   s[j++] = '_';
// 	   switch (ch)
// 		 {
// 		 case FRIBIDI_CHAR_LRM:
// 		   s[j++] = '>';
// 		   break;
// 		 case FRIBIDI_CHAR_RLM:
// 		   s[j++] = '<';
// 		   break;
// 		 case FRIBIDI_CHAR_LRE:
// 		   s[j++] = 'l';
// 		   break;
// 		 case FRIBIDI_CHAR_RLE:
// 		   s[j++] = 'r';
// 		   break;
// 		 case FRIBIDI_CHAR_PDF:
// 		   s[j++] = 'o';
// 		   break;
// 		 case FRIBIDI_CHAR_LRO:
// 		   s[j++] = 'L';
// 		   break;
// 		 case FRIBIDI_CHAR_RLO:
// 		   s[j++] = 'R';
// 		   break;
// 		 case FRIBIDI_CHAR_LRI:
// 		   s[j++] = 'i';
// 		   break;
// 		 case FRIBIDI_CHAR_RLI:
// 		   s[j++] = 'y';
// 		   break;
// 		 case FRIBIDI_CHAR_FSI:
// 		   s[j++] = 'f';
// 		   break;
// 		 case FRIBIDI_CHAR_PDI:
// 		   s[j++] = 'I';
// 		   break;
// 		 case '_':
// 		   s[j++] = '_';
// 		   break;
// 		 default:
// 		   j--;
// 		   if (ch < 256)
// 		 s[j++] = fribidi_unicode_to_cap_rtl_c (ch);
// 		   else
// 		 s[j++] = '?';
// 		   break;
// 		 }
// 	 }
// 	 }
//    s[j] = 0;

//    return j;
//  }
