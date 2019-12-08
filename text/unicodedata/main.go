package unicodedata

import (
	"unicode"
)

func IsEmojiExtendedPictographic(r rune) bool {
	return unicode.Is(_PangoExtended_PictographicTable, r)
}

func IsEmojiBaseCharacter(r rune) bool {
	return unicode.Is(_PangoEmojiTable, r)
}
