package pango

// GlyphItem is a pair of a Item and the glyphs
// resulting from shaping the text corresponding to an item.
// As an example of the usage of GlyphItem, the results
// of shaping text with Layout is a list of LayoutLine,
// each of which contains a list of GlyphItem.
type GlyphItem struct {
	item   *Item
	glyphs *GlyphString
}

// @text: text that @glyph_item corresponds to
//   (glyph_item.item.offset is an offset from the
//    start of @text)
// @log_attrs: (array): logical attributes for the item
//   (the first logical attribute refers to the position
//   before the first character in the item)
// pango_glyph_item_letter_space adds spacing between the graphemes of `glyph_item` to
// give the effect of typographic letter spacing.
// `letter_spacing` is specified in Pango units and may be negative, though too large
//   negative values will give ugly result
func (glyph_item *GlyphItem) pango_glyph_item_letter_space(text []rune, log_attrs []CharAttr, letter_spacing int) {
	//    GlyphItemIter iter;
	//    PangoGlyphInfo *glyphs = glyph_item.glyphs.glyphs;
	//    gboolean have_cluster;
	//    int space_left, space_right;

	space_left := letter_spacing / 2

	// hinting
	if (letter_spacing & (PANGO_SCALE - 1)) == 0 {
		space_left = PANGO_UNITS_ROUND(space_left)
	}

	space_right := letter_spacing - space_left
	have_cluster := iter.pango_glyph_item_iter_init_start(glyph_item, text)
	for ; have_cluster; have_cluster = iter.pango_glyph_item_iter_next_cluster() {
		if !log_attrs[iter.start_char].is_cursor_position {
			continue
		}

		if iter.start_glyph < iter.end_glyph { // LTR
			if iter.start_char > 0 {
				glyphs[iter.start_glyph].geometry.width += space_left
				glyphs[iter.start_glyph].geometry.x_offset += space_left
			}
			if iter.end_char < glyph_item.item.num_chars {
				glyphs[iter.end_glyph-1].geometry.width += space_right
			}
		} else { // RTL
			if iter.start_char > 0 {
				glyphs[iter.start_glyph].geometry.width += space_right
			}
			if iter.end_char < glyph_item.item.num_chars {
				glyphs[iter.end_glyph+1].geometry.x_offset += space_left
				glyphs[iter.end_glyph+1].geometry.width += space_left
			}
		}
	}
}

// GlyphItemIter is an iterator over the clusters in a
// `GlyphItem`. The forward direction of the
// iterator is the logical direction of text. That is, with increasing
// `start_index` and `start_char` values. If `glyph_item` is right-to-left
// (that is, if `glyph_item.item.analysis.level` is odd),
// then `start_glyph` decreases as the iterator moves forward. Moreover,
// in right-to-left cases, `start_glyph` is greater than `end_glyph`.
//
// An iterator should be initialized using either of
// `pango_glyph_item_iter_init_start()` and
// `pango_glyph_item_iter_init_end()`, for forward and backward iteration
// respectively, and walked over using any desired mixture of
// `pango_glyph_item_iter_next_cluster()` and
// `pango_glyph_item_iter_prev_cluster()`.
//
// Note that `text` is the start of the text for layout, which is then
// indexed by `glyph_item.item.offset` to get to the
// text of `glyph_item`. The `start_index` and `end_index` values can directly
// index into `text`. The `start_glyph`, `end_glyph`, `start_char`, and `end_char`
// values however are zero-based for the `glyph_item`. For each cluster, the
// item pointed at by the start variables is included in the cluster while
// the one pointed at by end variables is not.
type GlyphItemIter struct {
	glyph_item *GlyphItem
	text       []rune

	start_glyph int
	start_index int
	start_char  int

	end_glyph int
	end_index int
	end_char  int
}

// pango_glyph_item_iter_next_cluster advances the iterator to the next cluster in the glyph item.
func (iter *GlyphItemIter) pango_glyph_item_iter_next_cluster() bool {
	//    int glyph_index = iter.end_glyph;
	//    PangoGlyphString *glyphs = iter.glyph_item.glyphs;
	//    int cluster;
	//    PangoItem *item = iter.glyph_item.item;

	if LTR(iter.glyph_item) {
		if glyph_index == glyphs.num_glyphs {
			return false
		}
	} else {
		if glyph_index < 0 {
			return false
		}
	}

	iter.start_glyph = iter.end_glyph
	iter.start_index = iter.end_index
	iter.start_char = iter.end_char

	if LTR(iter.glyph_item) {
		cluster = glyphs.log_clusters[glyph_index]
		for {
			glyph_index++

			if glyph_index == glyphs.num_glyphs {
				iter.end_index = item.offset + item.length
				iter.end_char = item.num_chars
				break
			}

			if glyphs.log_clusters[glyph_index] > cluster {
				iter.end_index = item.offset + glyphs.log_clusters[glyph_index]
				iter.end_char += pango_utf8_strlen(iter.text+iter.start_index,
					iter.end_index-iter.start_index)
				break
			}
		}
	} else { /* RTL */
		cluster = glyphs.log_clusters[glyph_index]
		for {
			glyph_index--

			if glyph_index < 0 {
				iter.end_index = item.offset + item.length
				iter.end_char = item.num_chars
				break
			}

			if glyphs.log_clusters[glyph_index] > cluster {
				iter.end_index = item.offset + glyphs.log_clusters[glyph_index]
				iter.end_char += pango_utf8_strlen(iter.text+iter.start_index,
					iter.end_index-iter.start_index)
				break
			}
		}
	}

	iter.end_glyph = glyph_index

	return true
}

//  /**
//   * pango_glyph_item_iter_prev_cluster:
//   * @iter: a #GlyphItemIter
//   *
//   * Moves the iterator to the preceding cluster in the glyph item.
//   * See #GlyphItemIter for details of cluster orders.
//   *
//   * Return value: %true if the iterator was moved, %false if we were already on the
//   *  first cluster.
//   *
//   * Since: 1.22
//   **/
// func (iter  *GlyphItemIter) pango_glyph_item_iter_prev_cluster ()  bool {
//    int glyph_index = iter.start_glyph;
//    PangoGlyphString *glyphs = iter.glyph_item.glyphs;
//    int cluster;
//    PangoItem *item = iter.glyph_item.item;

//    if (LTR (iter.glyph_item))
// 	 {
// 	   if (glyph_index == 0)
// 	 return false;
// 	 }
//    else
// 	 {
// 	   if (glyph_index == glyphs.num_glyphs - 1)
// 	 return false;

// 	 }

//    iter.end_glyph = iter.start_glyph;
//    iter.end_index = iter.start_index;
//    iter.end_char = iter.start_char;

//    if (LTR (iter.glyph_item))
// 	 {
// 	   cluster = glyphs.log_clusters[glyph_index - 1];
// 	   for
// 	 {
// 	   if (glyph_index == 0)
// 		 {
// 		   iter.start_index = item.offset;
// 		   iter.start_char = 0;
// 		   break;
// 		 }

// 	   glyph_index--;

// 	   if (glyphs.log_clusters[glyph_index] < cluster)
// 		 {
// 		   glyph_index++;
// 		   iter.start_index = item.offset + glyphs.log_clusters[glyph_index];
// 		   iter.start_char -= pango_utf8_strlen (iter.text + iter.start_index,
// 						  iter.end_index - iter.start_index);
// 		   break;
// 		 }
// 	 }
// 	 }
//    else			/* RTL */
// 	 {
// 	   cluster = glyphs.log_clusters[glyph_index + 1];
// 	   for
// 	 {
// 	   if (glyph_index == glyphs.num_glyphs - 1)
// 		 {
// 		   iter.start_index = item.offset;
// 		   iter.start_char = 0;
// 		   break;
// 		 }

// 	   glyph_index++;

// 	   if (glyphs.log_clusters[glyph_index] < cluster)
// 		 {
// 		   glyph_index--;
// 		   iter.start_index = item.offset + glyphs.log_clusters[glyph_index];
// 		   iter.start_char -= pango_utf8_strlen (iter.text + iter.start_index,
// 						  iter.end_index - iter.start_index);
// 		   break;
// 		 }
// 	 }
// 	 }

//    iter.start_glyph = glyph_index;

//    g_assert (iter.start_char <= iter.end_char);
//    g_assert (0 <= iter.start_char);

//    return true;
//  }

/**
 * pango_glyph_item_iter_init_start:
 * @iter: a #GlyphItemIter
 * @glyph_item: the glyph item to iterate over
 * @text: text corresponding to the glyph item
 *
 * Initializes a #GlyphItemIter structure to point to the
 * first cluster in a glyph item.
 * See #GlyphItemIter for details of cluster orders.
 *
 * Return value: %false if there are no clusters in the glyph item
 *
 * Since: 1.22
 **/
func (iter *GlyphItemIter) pango_glyph_item_iter_init_start(glyph_item *GlyphItem, text []rune) bool {
	iter.glyph_item = glyph_item
	iter.text = text

	if LTR(glyph_item) {
		iter.end_glyph = 0
	} else {
		iter.end_glyph = glyph_item.glyphs.num_glyphs - 1
	}

	iter.end_index = glyph_item.item.offset
	iter.end_char = 0

	iter.start_glyph = iter.end_glyph
	iter.start_index = iter.end_index
	iter.start_char = iter.end_char

	/* Advance onto the first cluster of the glyph item */
	return pango_glyph_item_iter_next_cluster(iter)
}

/**
 * pango_glyph_item_iter_init_end:
 * @iter: a #GlyphItemIter
 * @glyph_item: the glyph item to iterate over
 * @text: text corresponding to the glyph item
 *
 * Initializes a #GlyphItemIter structure to point to the
 * last cluster in a glyph item.
 * See #GlyphItemIter for details of cluster orders.
 *
 * Return value: %false if there are no clusters in the glyph item
 *
 * Since: 1.22
 **/
func (iter *GlyphItemIter) pango_glyph_item_iter_init_end(glyph_item *GlyphItem, text []rune) bool {
	iter.glyph_item = glyph_item
	iter.text = text

	if LTR(glyph_item) {
		iter.start_glyph = glyph_item.glyphs.num_glyphs
	} else {
		iter.start_glyph = -1
	}

	iter.start_index = glyph_item.item.offset + glyph_item.item.length
	iter.start_char = glyph_item.item.num_chars

	iter.end_glyph = iter.start_glyph
	iter.end_index = iter.start_index
	iter.end_char = iter.start_char

	/* Advance onto the first cluster of the glyph item */
	return pango_glyph_item_iter_prev_cluster(iter)
}
