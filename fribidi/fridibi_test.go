package fribidi

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

const MAX_STR_LEN = 65000

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

type charsetI interface {
	decode(input []byte) []rune
	encode(input []rune) []byte
}

type stdCharset struct {
	encoding.Encoding
}

func (s stdCharset) decode(input []byte) []rune {
	u, err := s.Encoding.NewDecoder().Bytes(input)
	if err != nil {
		log.Fatal(err)
	}
	return []rune(string(u))
}

func (s stdCharset) encode(input []rune) []byte {
	u, err := s.Encoding.NewEncoder().Bytes([]byte(string(input)))
	if err != nil {
		log.Fatal(err)
	}
	return u
}

func parseCharset(filename string) charsetI {
	switch enc := strings.Split(filename, "_")[1]; enc {
	case "UTF-8":
		return stdCharset{unicode.UTF8}
	case "ISO8859-8":
		return stdCharset{charmap.ISO8859_8}
	case "CapRTL":
		return capRTLCharset{}
	default:
		panic("unsupported encoding " + enc)
	}
}

func processFile(filename string, fileOut io.Writer) error {

	do_clean, do_reorder_nsm, do_mirror := true, true, true
	show_input := true
	text_width := default_text_width
	// do_break := false

	do_pad := true
	show_visual := true
	show_basedir := false
	show_ltov := false
	show_vtol := false
	show_levels := false
	// char_set := "UTF-8"
	// bol_text := nil
	// eol_text := nil
	var input_base_direction ParType = ON

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

	flags := defaultFlags.adjust(FRIBIDI_FLAG_SHAPE_MIRRORING, do_mirror)
	flags = flags.adjust(FRIBIDI_FLAG_REORDER_NSM, do_reorder_nsm)

	padding_width := text_width
	if show_input {
		padding_width = (text_width - 10) / 2
	}
	break_width := 3 * MAX_STR_LEN

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	/* Read and process input one line at a time */
	for _, line := range bytes.Split(b, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		logical := charset.decode(line)

		var nl_found = ""

		/* Create a bidi string. */
		base := input_base_direction

		out, _ := fribidi_log2vis(flags, logical, &base)
		logToVis := out.LogicalToVisual()

		if show_input {
			fmt.Fprintf(fileOut, "%-*s => ", padding_width, line)
		}

		/* Remove explicit marks, if asked for. */

		if do_clean {
			out.Str = fribidi_remove_bidi_marks(out.Str, logToVis, out.VisualToLogical, out.EmbeddingLevels)
		}
		if show_visual {
			fmt.Fprintf(fileOut, "%s", nl_found)
			// if bol_text {
			// 	fmt.Fprintf(fileOut, "%s", bol_text)
			// }

			/* Convert it to input charset and print. */
			var st int
			for idx := 0; idx < len(out.Str); {
				var inlen int

				wid := break_width
				st = idx
				if _, isCapRTL := charset.(capRTLCharset); !isCapRTL {
					for wid > 0 && idx < len(out.Str) {
						if GetBidiType(out.Str[idx]).IsExplicitOrIsolateOrBnOrNsm() {
							wid -= 0
						} else {
							wid -= 1
						}
						idx++
					}
				} else {
					for wid > 0 && idx < len(out.Str) {
						wid--
						idx++
					}
				}
				if wid < 0 && idx-st > 1 {
					idx--
				}
				inlen = idx - st

				outstring := charset.encode(out.Str[st : inlen+st])
				if base.IsRtl() {
					var w int
					if do_pad {
						w = padding_width + len(outstring) - (break_width - wid)
					}
					fmt.Fprintf(fileOut, "%*s", w, outstring)
				} else {
					fmt.Fprintf(fileOut, "%s", outstring)
				}
				if idx < len(out.Str) {
					fmt.Fprintln(fileOut)
				}
			}
			// if eol_text {
			// 	fmt.Fprintf(fileOut, "%s", eol_text)
			// }

			nl_found = "\n"
		}
		if show_basedir {
			fmt.Fprintf(fileOut, "%s", nl_found)
			if FRIBIDI_DIR_TO_LEVEL(base) != 0 {
				fmt.Fprintf(fileOut, "Base direction: %s", "R")
			} else {
				fmt.Fprintf(fileOut, "Base direction: %s", "L")
			}
			nl_found = "\n"
		}
		if show_ltov {
			fmt.Fprintf(fileOut, "%s", nl_found)
			for _, ltov := range logToVis {
				fmt.Fprintf(fileOut, "%d ", ltov)
			}
			nl_found = "\n"
		}
		if show_vtol {
			fmt.Fprintf(fileOut, "%s", nl_found)
			for _, vtol := range out.VisualToLogical {
				fmt.Fprintf(fileOut, "%d ", vtol)
			}
			nl_found = "\n"
		}
		if show_levels {
			fmt.Fprintf(fileOut, "%s", nl_found)
			for _, level := range out.EmbeddingLevels {
				fmt.Fprintf(fileOut, "%d ", level)
			}
			nl_found = "\n"
		}

		if nl_found != "" {
			fmt.Fprintln(fileOut)
		}
	}
	return nil
}

func Test1(t *testing.T) {
	for _, file := range []string{
		"test/test_CapRTL_explicit.input",
		// "test/test_CapRTL_explicit.input",
		// "test/test_CapRTL_implicit.input",
		// "test/test_CapRTL_isolate.input",
		// "test/test_ISO8859-8_hebrew.input",
		// "test/test_UTF-8_persian.input",
		// "test/test_UTF-8_reordernsm.input",
	} {
		err := processFile(file, os.Stdout)
		if err != nil {
			t.Fatal("error in test file", file, err)
		}
	}
}
