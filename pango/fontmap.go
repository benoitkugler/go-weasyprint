package pango

// FontMap represents the set of fonts available for a
// particular rendering system.
type FontMap interface {
	// Loads the font in the fontmap that is the closest match for `desc`.
	// Returns nil if no font matched.
	load_font(context *Context, desc *FontDescription) Font
	// List all available families
	list_families() []*FontFamily
	// Load a set of fonts in the fontmap that can be used to render a font matching `desc`.
	// Returns nil if no font matched.
	load_fontset(context *Context, desc *FontDescription, language Language) Fontset

	// const char     *shape_engine_type; the type of rendering-system-dependent engines that can handle fonts of this fonts loaded with this fontmap.

	// Returns the current serial number of the fontmap.  The serial number is
	// initialized to an small number larger than zero when a new fontmap
	// is created and is increased whenever the fontmap is changed. It may
	// wrap, but will never have the value 0. Since it can wrap, never compare
	// it with "less than", always use "not equals".
	//
	// The fontmap can only be changed using backend-specific API, like changing
	// fontmap resolution.
	get_serial() uint

	// Forces a change in the context, which will cause any Context
	// using this fontmap to change.
	//
	// This function is only useful when implementing a new backend
	// for Pango, something applications won't do. Backends should
	// call this function if they have attached extra data to the context
	// and such data is changed.
	changed()

	// Gets a font family by name.
	get_family(name string) *FontFamily

	// Gets the FontFace to which `font` belongs.
	get_face(font Font) *FontFace
}
