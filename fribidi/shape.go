package fribidi

// fribidi_shape does all shaping work that depends on the resolved embedding
// levels of the characters.  Currently it does mirroring and Arabic shaping,
// but the list may grow in the future.  This function is a wrapper around
// fribidi_shape_mirroring and fribidi_shape_arabic.
//
// The flags parameter specifies which shapings are applied.  The only flags
// affecting the functionality of this function are those beginning with
// FRIBIDI_FLAG_SHAPE_.  Of these, only FRIBIDI_FLAG_SHAPE_MIRRORING is on
// in FRIBIDI_FLAGS_DEFAULT.  For details of the Arabic-specific flags see
// fribidi_shape_arabic. If ar_props is nil, no Arabic shaping is performed.
//
// Feel free to do your own shaping before or after calling this function,
// but you should take care of embedding levels yourself then.
func fribidi_shape(flags int, embedding_levels []Level,
	/* input and output */ ar_props []JoiningType, str []rune) {

	if len(ar_props) == 0 {
		fribidi_shape_arabic(flags, embedding_levels, ar_props, str)
	}

	if flags&FRIBIDI_FLAG_SHAPE_MIRRORING != 0 {
		fribidi_shape_mirroring(embedding_levels, str)
	}
}

// fribidi_shape_mirroring replaces mirroring characters on right-to-left embeddings in
// string with their mirrored equivalent as returned by
// fribidi_get_mirror_char().
//
// This function implements rule L4 of the Unicode Bidirectional Algorithm
// available at http://www.unicode.org/reports/tr9/#L4.
func fribidi_shape_mirroring(embedding_levels []Level, str []rune /* input and output */) {
	/* L4. Mirror all characters that are in odd levels and have mirrors. */
	for i, level := range embedding_levels {
		if level.isRtl() != 0 {
			if mirror, ok := fribidi_get_mirror_char(str[i]); ok {
				str[i] = mirror
			}
		}
	}
}
