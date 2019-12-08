package text

import (
	"fmt"
	"testing"
)

func TestBreak(t *testing.T) {
	fmt.Println(pangoDefaultBreak("Je m'appelle Benoit KUGLER. HAHA!"))

	pangoDefaultBreak(`
	There is no single method for determining line breaks; the rules may differ based on user preference and document layout. The information in this annex, including the specification of the line breaking algorithm, allows for the necessary flexibility in determining line breaks according to different conventions. However, some characters have been encoded explicitly for their effect on line breaking. Because users adding such characters to a text expect that they will have the desired effect, these characters have been given required line breaking behavior.

To handle certain situations, some line breaking implementations use techniques that cannot be expressed within the framework of the Unicode Line Breaking Algorithm. Examples include using dictionaries of words for languages that do not use spaces, such as Thai; recognition of the language of the text in order to choose among different punctuation conventions; using dictionaries of common abbreviations or contractions to resolve ambiguities with periods or apostrophes; or a deeper analysis of common syntaxes for numbers or dates, and so on. The conformance requirements permit variations of this kind.

Processes which support multiple modes for determining line breaks are also accommodated. This situation can arise with marked-up text, rich text, style sheets, or other environments in which a higher-level protocol can carry formatting instructions that prevent or force line breaks in positions that differ from those specified by the Unicode Line Breaking Algorithm. The approach taken here requires that such processes have a conforming default line break behavior, and to disclose that they also include overrides or optional behaviors that are invoked via a higher-level protocol.

The methods by which a line layout process chooses optimal line breaks from among the available break opportunities is outside the scope of this specification. The behavior of a line layout process in situations where there are no suitable break opportunities is also outside of the scope of this specification.

Note: Locale-sensitive line break specifications can be expressed in LDML [UTS35]. Tailorings are available in the Common Locale Data Repository [CLDR].`)
}
