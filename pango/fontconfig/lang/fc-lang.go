// Read a set of language orthographies and build C declarations for
// charsets which can then be used to identify which languages are
// supported by a given font.
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/exp/errors/fmt"
)

// port from fontconfig/fc-lang/fc-lang.py
//
// Copyright © 2001-2002 Keith Packard
// Copyright © 2019 Tim-Philipp Müller

func assert(b bool) {
	if !b {
		log.Fatal("assertion error")
	}
}

// we just store the leaves in a dict, we can order the leaves later if needed
type CharSet struct {
	leaves map[int][]uint32 // leaf_number -> leaf data (= 16 uint32)
}

// Build a single charset from a source file
//
// The file format is quite simple, either
// a single hex value or a pair separated with a dash
func parseOrthFile(fileName string, lines []lineData) CharSet {
	charset := CharSet{leaves: make(map[int][]uint32)}
	for _, l := range lines {
		fn, num, line := l.fileName, l.num, l.line
		deleteChar := strings.HasPrefix(line, "-")
		if deleteChar {
			line = line[1:]
		}
		var parts []string
		if strings.IndexByte(line, '-') != -1 {
			parts = strings.Split(line, "-")
		} else if strings.Index(line, "..") != -1 {
			parts = strings.Split(line, "..")
		} else {
			parts = []string{line}
		}

		var startString, endString string

		startString, parts = strings.TrimSpace(parts[0]), parts[1:]
		start, err := strconv.ParseInt(strings.TrimPrefix(startString, "0x"), 16, 32)
		if err != nil {
			log.Fatal("can't parse ", startString, " : ", err)
		}

		end := start
		if len(parts) != 0 {
			endString, parts = strings.TrimSpace(parts[0]), parts[1:]
			end, err = strconv.ParseInt(strings.TrimPrefix(endString, "0x"), 16, 32)
			if err != nil {
				log.Fatal("can't parse ", endString, " : ", err)
			}
		}
		if len(parts) != 0 {
			log.Fatalf("%s line %d: parse error (too many parts)", fn, num)
		}

		for ucs4 := start; ucs4 <= end; ucs4++ {
			if deleteChar {
				charset.delChar(int(ucs4))
			} else {
				charset.addChar(int(ucs4))
			}
		}
	}
	assert(charset.equals(charset)) // sanity check for the equals function

	return charset
}

func (cs CharSet) addChar(ucs4 int) {
	assert(ucs4 < 0x01000000)
	leafNum := ucs4 >> 8
	leaf, ok := cs.leaves[leafNum]
	if !ok {
		leaf = []uint32{0, 0, 0, 0, 0, 0, 0, 0} // 256/32 = 8
		cs.leaves[leafNum] = leaf
	}
	leaf[(ucs4&0xff)>>5] |= (1 << (ucs4 & 0x1f))
}

func (cs CharSet) delChar(ucs4 int) {
	assert(ucs4 < 0x01000000)
	leafNum := ucs4 >> 8
	if leaf, ok := cs.leaves[leafNum]; ok {
		leaf[(ucs4&0xff)>>5] &= ^(1 << (ucs4 & 0x1f))
		// We don't bother removing the leaf if it's empty
	}
}

func (cs CharSet) sortedKeys() []int {
	var keys []int
	for k := range cs.leaves {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func (cs *CharSet) equals(otherCs CharSet) bool {
	keys := cs.sortedKeys()
	otherKeys := otherCs.sortedKeys()
	if len(keys) != len(otherKeys) {
		return false
	}

	for i := range keys {
		k1, k2 := keys[i], otherKeys[i]
		if k1 != k2 {
			return false
		}
		if !leavesEqual(cs.leaves[k1], otherCs.leaves[k2]) {
			return false
		}
	}
	return true
}

// Convert a file name into a name suitable for Go declarations
func getName(fileName string) string {
	return strings.Split(fileName, ".")[0]
}

// Convert a C name into a language name
func getLang(cName string) string {
	cName = strings.ReplaceAll(cName, "_", "-")
	cName = strings.ReplaceAll(cName, " ", "")
	return strings.ToLower(cName)
}

type lineData struct {
	fileName string
	num      int
	line     string
}

func readOrthFile(fileName string) []lineData {
	var lines []lineData

	b, err := ioutil.ReadFile("orth/" + fileName)
	if err != nil {
		log.Fatal("can't read ", fileName, err)
	}
	linesString := strings.Split(string(b), "\n")
	for num, line := range linesString {
		if strings.HasPrefix(line, "include ") {
			includeFn := strings.TrimSpace(line[8:])
			lines = append(lines, readOrthFile(includeFn)...)
		} else {
			// remove comments and strip whitespaces
			line = strings.TrimSpace(strings.Split(line, "#")[0])
			line = strings.TrimSpace(strings.Split(line, "\t")[0])
			// skip empty lines
			if line != "" {
				lines = append(lines, lineData{fileName, num, line})
			}
		}
	}
	return lines
}

func leavesEqual(leaf1, leaf2 []uint32) bool {
	if len(leaf1) != len(leaf2) {
		return false
	}
	for i, v1 := range leaf1 {
		if v1 != leaf2[i] {
			return false
		}
	}
	return true
}

func main() {
	output := flag.String("output", "", "output file")
	flag.Parse()
	var (
		err             error
		totalLeaves     = 0
		sets            []CharSet
		country         []int
		names, langs    []string
		langCountrySets = map[string][]int{}
	)

	outputFile := os.Stdout
	// Open output file
	if *output != "" {
		outputFile, err = os.Create(*output + ".go")
		if err != nil {
			log.Fatal(err)
		}
	}

	var sortedKeys []string
	orthEntries := map[string]int{}
	for i, fn := range orth_files {
		orthEntries[fn] = i
		sortedKeys = append(sortedKeys, fn)
	}
	sort.Strings(sortedKeys)

	for _, fn := range sortedKeys {
		lines := readOrthFile(fn)
		charset := parseOrthFile(fn, lines)

		sets = append(sets, charset)

		name := getName(fn)
		names = append(names, name)

		lang := getLang(name)
		langs = append(langs, lang)
		if strings.Index(lang, "-") != -1 {
			country = append(country, orthEntries[fn]) // maps to original index
			languageFamily := strings.Split(lang, "-")[0]
			langCountrySets[languageFamily] = append(langCountrySets[languageFamily], orthEntries[fn])
		}
		totalLeaves += len(charset.leaves)
	}
	// Find unique leaves
	var leaves [][]uint32
	for _, s := range sets {
		for _, leafNum := range s.sortedKeys() {
			leaf := s.leaves[leafNum]
			isUnique := true
			for _, existingLeaf := range leaves {
				if leavesEqual(leaf, existingLeaf) {
					isUnique = false
					break
				}
			}
			if isUnique {
				leaves = append(leaves, leaf)
			}
		}
	}

	// Find duplicate charsets
	var duplicate []int
	for i, s := range sets {
		var dup_num int // 0 means not duplicate
		if i >= 1 {
			for j, s_cmp := range sets {
				if j >= i {
					break
				}
				if s_cmp.equals(s) {
					dup_num = j
					break
				}
			}
		}

		duplicate = append(duplicate, dup_num)
	}

	var (
		tn  = 0
		off = map[int]int{}
	)
	for i, s := range sets {
		if duplicate[i] != 0 {
			continue
		}
		off[i] = tn
		tn += len(s.leaves)
	}

	// ----------------------------------- output -----------------------------------

	fmt.Fprintln(outputFile, "package fontconfig")
	fmt.Fprintln(outputFile, "// Code auto-generated by lang/fc-lang.go DO NOT EDIT")
	fmt.Fprintln(outputFile)
	fmt.Fprintf(outputFile, "// total size: %d unique leaves: %d \n\n", totalLeaves, len(leaves))

	fmt.Fprintf(outputFile, "//define LEAF0       (%d * sizeof (FcLangCharSet))\n", len(sets))
	fmt.Fprintf(outputFile, "//define OFF0        (LEAF0 + %d * sizeof (FcCharLeaf))\n", len(leaves))
	fmt.Fprintf(outputFile, "//define NUM0        (OFF0 + %d * sizeof (uintptr_t))\n", tn)
	fmt.Fprintln(outputFile, "//define SET(n)      (n * sizeof (FcLangCharSet) + offsetof (FcLangCharSet, charset))")
	fmt.Fprintln(outputFile, "//define OFF(s,o)    (OFF0 + o * sizeof (uintptr_t) - SET(s))")
	fmt.Fprintln(outputFile, "//define NUM(s,n)    (NUM0 + n * sizeof (FcChar16) - SET(s))")
	fmt.Fprintln(outputFile, "//define LEAF(o,l)   (LEAF0 + l * sizeof (FcCharLeaf) - (OFF0 + o * sizeof (intptr_t)))")
	fmt.Fprintln(outputFile, "var fcLangCharSets = fcLangData.langCharSets")

	assert(len(sets) < 256) // FIXME: need to change index type to 16-bit below then

	// leaf_offsets [%d]uintptr_t
	fmt.Fprintf(outputFile, `
	var fcLangData = struct {
        langCharSets [%d]FcLangCharSet
        leaves [%d]FcCharLeaf
        numbers [%d]uint16
	}{`, len(sets), len(leaves), tn)
	fmt.Fprintln(outputFile)

	// Dump sets
	fmt.Fprintln(outputFile, "[...]FcLangCharSet{")
	for i := range sets {
		j := i
		if duplicate[i] != 0 {
			j = duplicate[i]
		}
		fmt.Fprintf(outputFile, `    { %q,  }, // %d `, langs[i], len(sets[j].leaves))
		fmt.Fprintln(outputFile)
	}
	fmt.Fprintln(outputFile, "},")

	// Dump leaves
	fmt.Fprintln(outputFile, "[...]FcCharLeaf{")
	for l, leaf := range leaves {
		fmt.Fprintf(outputFile, "    { /* %d */", l)
		for i := 0; i < 8; i++ { // 256/32 = 8
			if i%4 == 0 {
				fmt.Fprint(outputFile, "\n   ")
			}
			fmt.Fprintf(outputFile, " 0x%08x,", leaf[i])
		}
		fmt.Fprintln(outputFile, "\n    },")
	}
	fmt.Fprintln(outputFile, "},")

	// Dump leaves
	// fmt.Fprintln(outputFile, "[...]uintptr_t{")
	// for i, s := range sets {
	// 	if duplicate[i] != 0 {
	// 		continue
	// 	}

	// 	fmt.Fprintf(outputFile, "    /* %s */\n", names[i])

	// 	for n, leafNum := range s.sortedKeys() {
	// 		leaf := s.leaves[leafNum]
	// 		if n%4 == 0 {
	// 			fmt.Fprint(outputFile, "   ")
	// 		}
	// 		var found []int
	// 		for k, unique_leaf := range leaves {
	// 			if leavesEqual(unique_leaf, leaf) {
	// 				found = append(found, k)
	// 			}
	// 		}
	// 		assert(len(found) != 0)
	// 		assert(len(found) == 1)
	// 		fmt.Fprintf(outputFile, " LEAF(%3d,%3d),", off[i], found[0])
	// 		if n%4 == 3 {
	// 			fmt.Fprintln(outputFile)
	// 		}
	// 	}
	// 	if len(s.leaves)%4 != 0 {
	// 		fmt.Fprintln(outputFile)
	// 	}
	// }
	// fmt.Fprintln(outputFile, "},")

	fmt.Fprintln(outputFile, "[...]uint16{")
	for i, s := range sets {
		if duplicate[i] != 0 {
			continue
		}
		fmt.Fprintf(outputFile, "    /* %s */\n", names[i])

		for n, leafNum := range s.sortedKeys() {
			// leaf := s.leaves[leafNum]
			if n%8 == 0 {
				fmt.Fprintf(outputFile, "   ")
			}
			fmt.Fprintf(outputFile, " 0x%04x,", leafNum)
			if n%8 == 7 {
				fmt.Fprintln(outputFile)
			}
		}
		if len(s.leaves)%8 != 0 {
			fmt.Fprintln(outputFile)
		}
	}
	fmt.Fprintln(outputFile, "},")
	fmt.Fprintln(outputFile, "};\n")

	// langIndices
	fmt.Fprintln(outputFile, "var fcLangCharSetIndices = [...]byte{")
	for i := range sets {
		fn := fmt.Sprintf("%s.orth", names[i])
		fmt.Fprintf(outputFile, "    %d, /* %s */\n", orthEntries[fn], names[i])
	}
	fmt.Fprintln(outputFile, "}")

	// langIndicesInv
	fmt.Fprintln(outputFile, "var fcLangCharSetIndicesInv = [...]byte{")
	for k := range orthEntries {
		name := getName(k)
		idx := -1
		for i, s := range names {
			if s == name {
				idx = i
				break
			}
		}
		fmt.Fprintf(outputFile, "    %d, /* %s */\n", idx, name)
	}
	fmt.Fprintln(outputFile, "}")

	fmt.Fprintf(outputFile, "const NUM_LANG_CHAR_SET = %d \n", len(sets))
	num_lang_set_map := (len(sets) + 31) / 32
	fmt.Fprintf(outputFile, "const NUM_LANG_SET_MAP	= %d \n", num_lang_set_map)

	// Dump indices with country codes
	assert(len(country) > 0)
	assert(len(langCountrySets) > 0)
	fmt.Fprintln(outputFile)
	fmt.Fprintln(outputFile, "var fcLangCountrySets = [][NUM_LANG_SET_MAP]uint32 {")
	var langCSSortedkeys []string
	for k := range langCountrySets {
		langCSSortedkeys = append(langCSSortedkeys, k)
	}
	sort.Strings(langCSSortedkeys)
	for _, k := range langCSSortedkeys {
		langset_map := make([]int, num_lang_set_map) // initialise all zeros
		for _, entries_id := range langCountrySets[k] {
			langset_map[entries_id>>5] |= (1 << (entries_id & 0x1f))
		}
		fmt.Fprintf(outputFile, "    {")
		for _, v := range langset_map {
			fmt.Fprintf(outputFile, " 0x%08x,", v)
		}
		fmt.Fprintf(outputFile, " }, /* %s */\n", k)
	}
	fmt.Fprintln(outputFile, "};\n")
	fmt.Fprintf(outputFile, "const NUM_COUNTRY_SET = %d\n", len(langCountrySets))

	// Find ranges for each letter for faster searching
	// Dump sets start/finish for the fastpath
	fmt.Fprintln(outputFile, "var fcLangCharSetRanges = []FcLangCharSetRange{")
	for c := 'a'; c <= 'z'; c++ {
		start := 9999
		stop := -1
		for i := range sets {
			if strings.HasPrefix(names[i], string(c)) {
				start = min(start, i)
				stop = max(stop, i)
			}
		}
		fmt.Fprintf(outputFile, "    { %d, %d }, /* %d */\n", start, stop, c)
	}
	fmt.Fprintln(outputFile, "};\n")

	if *output != "" {
		outputFile.Close()
		exec.Command("goimports", "-w", *output+".go").Run()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Do not reorder, magic
var orth_files = []string{
	"aa.orth",
	"ab.orth",
	"af.orth",
	"am.orth",
	"ar.orth",
	"as.orth",
	"ast.orth",
	"av.orth",
	"ay.orth",
	"az_az.orth",
	"az_ir.orth",
	"ba.orth",
	"bm.orth",
	"be.orth",
	"bg.orth",
	"bh.orth",
	"bho.orth",
	"bi.orth",
	"bin.orth",
	"bn.orth",
	"bo.orth",
	"br.orth",
	"bs.orth",
	"bua.orth",
	"ca.orth",
	"ce.orth",
	"ch.orth",
	"chm.orth",
	"chr.orth",
	"co.orth",
	"cs.orth",
	"cu.orth",
	"cv.orth",
	"cy.orth",
	"da.orth",
	"de.orth",
	"dz.orth",
	"el.orth",
	"en.orth",
	"eo.orth",
	"es.orth",
	"et.orth",
	"eu.orth",
	"fa.orth",
	"fi.orth",
	"fj.orth",
	"fo.orth",
	"fr.orth",
	"ff.orth",
	"fur.orth",
	"fy.orth",
	"ga.orth",
	"gd.orth",
	"gez.orth",
	"gl.orth",
	"gn.orth",
	"gu.orth",
	"gv.orth",
	"ha.orth",
	"haw.orth",
	"he.orth",
	"hi.orth",
	"ho.orth",
	"hr.orth",
	"hu.orth",
	"hy.orth",
	"ia.orth",
	"ig.orth",
	"id.orth",
	"ie.orth",
	"ik.orth",
	"io.orth",
	"is.orth",
	"it.orth",
	"iu.orth",
	"ja.orth",
	"ka.orth",
	"kaa.orth",
	"ki.orth",
	"kk.orth",
	"kl.orth",
	"km.orth",
	"kn.orth",
	"ko.orth",
	"kok.orth",
	"ks.orth",
	"ku_am.orth",
	"ku_ir.orth",
	"kum.orth",
	"kv.orth",
	"kw.orth",
	"ky.orth",
	"la.orth",
	"lb.orth",
	"lez.orth",
	"ln.orth",
	"lo.orth",
	"lt.orth",
	"lv.orth",
	"mg.orth",
	"mh.orth",
	"mi.orth",
	"mk.orth",
	"ml.orth",
	"mn_cn.orth",
	"mo.orth",
	"mr.orth",
	"mt.orth",
	"my.orth",
	"nb.orth",
	"nds.orth",
	"ne.orth",
	"nl.orth",
	"nn.orth",
	"no.orth",
	"nr.orth",
	"nso.orth",
	"ny.orth",
	"oc.orth",
	"om.orth",
	"or.orth",
	"os.orth",
	"pa.orth",
	"pl.orth",
	"ps_af.orth",
	"ps_pk.orth",
	"pt.orth",
	"rm.orth",
	"ro.orth",
	"ru.orth",
	"sa.orth",
	"sah.orth",
	"sco.orth",
	"se.orth",
	"sel.orth",
	"sh.orth",
	"shs.orth",
	"si.orth",
	"sk.orth",
	"sl.orth",
	"sm.orth",
	"sma.orth",
	"smj.orth",
	"smn.orth",
	"sms.orth",
	"so.orth",
	"sq.orth",
	"sr.orth",
	"ss.orth",
	"st.orth",
	"sv.orth",
	"sw.orth",
	"syr.orth",
	"ta.orth",
	"te.orth",
	"tg.orth",
	"th.orth",
	"ti_er.orth",
	"ti_et.orth",
	"tig.orth",
	"tk.orth",
	"tl.orth",
	"tn.orth",
	"to.orth",
	"tr.orth",
	"ts.orth",
	"tt.orth",
	"tw.orth",
	"tyv.orth",
	"ug.orth",
	"uk.orth",
	"ur.orth",
	"uz.orth",
	"ve.orth",
	"vi.orth",
	"vo.orth",
	"vot.orth",
	"wa.orth",
	"wen.orth",
	"wo.orth",
	"xh.orth",
	"yap.orth",
	"yi.orth",
	"yo.orth",
	"zh_cn.orth",
	"zh_hk.orth",
	"zh_mo.orth",
	"zh_sg.orth",
	"zh_tw.orth",
	"zu.orth",
	"ak.orth",
	"an.orth",
	"ber_dz.orth",
	"ber_ma.orth",
	"byn.orth",
	"crh.orth",
	"csb.orth",
	"dv.orth",
	"ee.orth",
	"fat.orth",
	"fil.orth",
	"hne.orth",
	"hsb.orth",
	"ht.orth",
	"hz.orth",
	"ii.orth",
	"jv.orth",
	"kab.orth",
	"kj.orth",
	"kr.orth",
	"ku_iq.orth",
	"ku_tr.orth",
	"kwm.orth",
	"lg.orth",
	"li.orth",
	"mai.orth",
	"mn_mn.orth",
	"ms.orth",
	"na.orth",
	"ng.orth",
	"nv.orth",
	"ota.orth",
	"pa_pk.orth",
	"pap_an.orth",
	"pap_aw.orth",
	"qu.orth",
	"quz.orth",
	"rn.orth",
	"rw.orth",
	"sc.orth",
	"sd.orth",
	"sg.orth",
	"sid.orth",
	"sn.orth",
	"su.orth",
	"ty.orth",
	"wal.orth",
	"za.orth",
	"lah.orth",
	"nqo.orth",
	"brx.orth",
	"sat.orth",
	"doi.orth",
	"mni.orth",
	"und_zsye.orth",
	"und_zmth.orth",
}
