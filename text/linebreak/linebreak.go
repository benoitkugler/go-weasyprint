// Copyright 2013 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// gorilla/i18n/linebreak implements the Unicode line breaking algorithm.
//
// Line breaking, also known as word wrapping, is the process of breaking a
// section of text into lines such that it will fit in the available width
// of a page, window or other display area.
//
// As simple as it sounds, this is not a trivial task when support for
// multilingual texts is required. The particular algorithm used in this
// package is based on best practices defined in UAX #14:
//
//     http://www.unicode.org/reports/tr14/
//
// A similar package that served as inspiration for this one is Bram Stein's
// Unicode Tokenizer (for Node.js):
//
//     https://github.com/bramstein/unicode-tokenizer
package linebreak

import (
	"unicode"
)

func ResolveClass(r rune) GUnicodeBreakType {
	cls := Class(r)
	// LB1: Resolve AI, CB, CJ, SA, SG, and XX into other classes.
	// We are using the generic resolution proposed in UAX #14.
	switch cls {
	case G_UNICODE_BREAK_AI, G_UNICODE_BREAK_SG, G_UNICODE_BREAK_XX:
		cls = G_UNICODE_BREAK_AL
	case G_UNICODE_BREAK_CJ:
		cls = G_UNICODE_BREAK_NS
	case G_UNICODE_BREAK_SA:
		if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) {
			cls = G_UNICODE_BREAK_CM
		} else {
			cls = G_UNICODE_BREAK_AL
		}
	case G_UNICODE_BREAK_CB:
		// TODO: CB should be left to be resolved later, according to
		// LB9, LB10 and LB20.
		// For now we are using a placeholder; maybe not the best one.
		cls = G_UNICODE_BREAK_ID
	}
	return cls
}
