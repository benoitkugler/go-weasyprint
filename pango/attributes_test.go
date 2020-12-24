package pango

import (
	"strings"
	"testing"
)

func testCopy(t *testing.T, attr *Attribute) {
	a := attr.pango_attribute_copy()
	assertTrue(t, attr.pango_attribute_equal(*a), "cloned values")
}

func TestAttributesBasic(t *testing.T) {
	testCopy(t, pango_attr_language_new(pango_language_from_string("ja-JP")))
	testCopy(t, pango_attr_family_new("Times"))
	testCopy(t, pango_attr_foreground_new(100, 200, 300))
	testCopy(t, pango_attr_background_new(100, 200, 300))
	testCopy(t, pango_attr_size_new(1024))
	testCopy(t, pango_attr_size_new_absolute(1024))
	testCopy(t, pango_attr_style_new(PANGO_STYLE_ITALIC))
	testCopy(t, pango_attr_weight_new(PANGO_WEIGHT_ULTRALIGHT))
	testCopy(t, pango_attr_variant_new(PANGO_VARIANT_SMALL_CAPS))
	testCopy(t, pango_attr_stretch_new(PANGO_STRETCH_SEMI_EXPANDED))
	testCopy(t, pango_attr_font_desc_new(pango_font_description_from_string("Computer Modern 12")))
	testCopy(t, pango_attr_underline_new(PANGO_UNDERLINE_LOW))
	testCopy(t, pango_attr_underline_new(PANGO_UNDERLINE_ERROR_LINE))
	testCopy(t, pango_attr_underline_color_new(100, 200, 300))
	testCopy(t, pango_attr_overline_new(PANGO_OVERLINE_SINGLE))
	testCopy(t, pango_attr_overline_color_new(100, 200, 300))
	testCopy(t, pango_attr_strikethrough_new(true))
	testCopy(t, pango_attr_strikethrough_color_new(100, 200, 300))
	testCopy(t, pango_attr_rise_new(256))
	testCopy(t, pango_attr_scale_new(2.56))
	testCopy(t, pango_attr_fallback_new(false))
	testCopy(t, pango_attr_letter_spacing_new(1024))

	rect := Rectangle{x: 0, y: 0, width: 10, height: 10}
	testCopy(t, pango_attr_shape_new(rect, rect))
	testCopy(t, pango_attr_gravity_new(PANGO_GRAVITY_SOUTH))
	testCopy(t, pango_attr_gravity_hint_new(PANGO_GRAVITY_HINT_STRONG))
	testCopy(t, pango_attr_allow_breaks_new(false))
	testCopy(t, pango_attr_show_new(PANGO_SHOW_SPACES))
	testCopy(t, pango_attr_insert_hyphens_new(false))
}

/* check that pango_attribute_equal compares values, but not ranges */
func TestAttributesEqual(t *testing.T) {
	attr1 := pango_attr_size_new(10)
	attr2 := pango_attr_size_new(20)
	attr3 := pango_attr_size_new(20)
	attr3.StartIndex = 1
	attr3.EndIndex = 2

	assertFalse(t, attr1.pango_attribute_equal(*attr2), "attribute equality")
	assertTrue(t, attr2.pango_attribute_equal(*attr3), "attribute equality")
}

// void
// print_attr_list (PangoAttrList *attrs, GString *string)
// {
//   PangoAttrIterator *iter;

//   if (!attrs)
//     return;

//   iter = pango_attr_list_get_iterator (attrs);
//   do {
//     gint start, end;
//     GSList *list, *l;

//     pango_attr_iterator_range (iter, &start, &end);
//     g_string_append_printf (string, "range %d %d\n", start, end);
//     list = pango_attr_iterator_get_attrs (iter);
//     for (l = list; l; l = l.next)
//       {
//         PangoAttribute *attr = l.data;
//         print_attribute (attr, string);
//         g_string_append (string, "\n");
//       }
//     g_slist_free_full (list, (GDestroyNotify)pango_attribute_destroy);
//   } while (pango_attr_iterator_next (iter));

//   pango_attr_iterator_destroy (iter);
// }

func print_attributes(attrs AttrList) string {
	chunks := make([]string, len(attrs))
	for i, attr := range attrs {
		chunks[i] = attr.String() + "\n"
	}
	return strings.Join(chunks, "")
}

func assert_attributes(t *testing.T, attrs AttrList, expected string) {
	s := print_attributes(attrs)
	if s != expected {
		t.Errorf("-----\nattribute list mismatch\nexpected:\n%s-----\nreceived:\n%s-----\n",
			expected, s)
	}
}

func assert_attr_iterator(t *testing.T, iter *AttrIterator, expected string) {
	attrs := iter.pango_attr_iterator_get_attrs()
	assert_attributes(t, attrs, expected)
}

func TestList(t *testing.T) {
	var list AttrList

	list.pango_attr_list_insert(pango_attr_size_new(10))
	list.pango_attr_list_insert(pango_attr_size_new(20))
	list.pango_attr_list_insert(pango_attr_size_new(30))

	assert_attributes(t, list, "[0,-1]size=10\n"+
		"[0,-1]size=20\n"+
		"[0,-1]size=30\n")

	list = nil

	/* test that insertion respects StartIndex */
	list.pango_attr_list_insert(pango_attr_size_new(10))
	attr := pango_attr_size_new(20)
	attr.StartIndex = 10
	attr.EndIndex = 20
	list.pango_attr_list_insert(attr)
	list.pango_attr_list_insert(pango_attr_size_new(30))
	attr = pango_attr_size_new(40)
	attr.StartIndex = 10
	attr.EndIndex = 40
	list.pango_attr_list_insert_before(attr)

	assert_attributes(t, list, "[0,-1]size=10\n"+
		"[0,-1]size=30\n"+
		"[10,40]size=40\n"+
		"[10,20]size=20\n")
}

func TestListChange(t *testing.T) {
	var list AttrList

	attr := pango_attr_size_new(10)
	attr.StartIndex = 0
	attr.EndIndex = 10
	list.pango_attr_list_insert(attr)
	attr = pango_attr_size_new(20)
	attr.StartIndex = 20
	attr.EndIndex = 30
	list.pango_attr_list_insert(attr)
	attr = pango_attr_weight_new(PANGO_WEIGHT_BOLD)
	attr.StartIndex = 0
	attr.EndIndex = 30
	list.pango_attr_list_insert(attr)

	assert_attributes(t, list, "[0,10]size=10\n"+
		"[0,30]weight=700\n"+
		"[20,30]size=20\n")

	/* simple insertion with pango_attr_list_change */
	attr = pango_attr_variant_new(PANGO_VARIANT_SMALL_CAPS)
	attr.StartIndex = 10
	attr.EndIndex = 20
	list.pango_attr_list_change(*attr)

	assert_attributes(t, list, "[0,10]size=10\n"+
		"[0,30]weight=700\n"+
		"[10,20]variant=1\n"+
		"[20,30]size=20\n")

	/* insertion with splitting */
	attr = pango_attr_weight_new(PANGO_WEIGHT_LIGHT)
	attr.StartIndex = 15
	attr.EndIndex = 20
	list.pango_attr_list_change(*attr)

	assert_attributes(t, list, "[0,10]size=10\n"+
		"[0,15]weight=700\n"+
		"[10,20]variant=1\n"+
		"[15,20]weight=300\n"+
		"[20,30]size=20\n"+
		"[20,30]weight=700\n")

	/* insertion with joining */
	attr = pango_attr_size_new(20)
	attr.StartIndex = 5
	attr.EndIndex = 20
	list.pango_attr_list_change(*attr)

	assert_attributes(t, list, "[0,5]size=10\n"+
		"[0,15]weight=700\n"+
		"[5,30]size=20\n"+
		"[10,20]variant=1\n"+
		"[15,20]weight=300\n"+
		"[20,30]weight=700\n")
}

// func TestListSplice (t *testing.T,void) {
//    PangoAttrList *base;
//    PangoAttrList *list;
//    PangoAttrList *other;
//    PangoAttribute *attr;

//    base = pango_attr_list_new ();
//    attr = pango_attr_size_new (10);
//    attr.StartIndex = 0;
//    attr.EndIndex = -1;
//    pango_attr_list_insert (base, attr);
//    attr = pango_attr_weight_new (PANGO_WEIGHT_BOLD);
//    attr.StartIndex = 10;
//    attr.EndIndex = 15;
//    pango_attr_list_insert (base, attr);
//    attr = pango_attr_variant_new (PANGO_VARIANT_SMALL_CAPS);
//    attr.StartIndex = 20;
//    attr.EndIndex = 30;
//    pango_attr_list_insert (base, attr);

//    assert_attributes (t,base, "[0,-1]size=10\n"
// 						   "[10,15]weight=700\n"
// 						   "[20,30]variant=1\n");

//    /* splice in an empty list */
//    list = pango_attr_list_copy (base);
//    other = pango_attr_list_new ();
//    pango_attr_list_splice (list, other, 11, 5);

//    assert_attributes (t,list, "[0,-1]size=10\n"
// 						   "[10,20]weight=700\n"
// 						   "[25,35]variant=1\n");

//    pango_attr_list_unref (list);
//    pango_attr_list_unref (other);

//    /* splice in some attributes */
//    list = pango_attr_list_copy (base);
//    other = pango_attr_list_new ();
//    attr = pango_attr_size_new (20);
//    attr.StartIndex = 0;
//    attr.EndIndex = 3;
//    pango_attr_list_insert (other, attr);
//    attr = pango_attr_stretch_new (PANGO_STRETCH_CONDENSED);
//    attr.StartIndex = 2;
//    attr.EndIndex = 4;
//    pango_attr_list_insert (other, attr);

//    pango_attr_list_splice (list, other, 11, 5);

//    assert_attributes (t,list, "[0,11]size=10\n"
// 						   "[10,20]weight=700\n"
// 						   "[11,14]size=20\n"
// 						   "[13,15]stretch=2\n"
// 						   "[14,-1]size=10\n"
// 						   "[25,35]variant=1\n");

//    pango_attr_list_unref (list);
//    pango_attr_list_unref (other);

//    pango_attr_list_unref (base);
//  }

//  /* Test that empty lists work in pango_attr_list_splice */
// func TestListSplice2 (t *testing.T,void) {
//    PangoAttrList *list;
//    PangoAttrList *other;
//    PangoAttribute *attr;

//    var list AttrList
//    other = pango_attr_list_new ();

//    pango_attr_list_splice (list, other, 11, 5);

//    g_assert_null (pango_attr_list_get_attributes (list));

//    attr = pango_attr_size_new (10);
//    attr.StartIndex = 0;
//    attr.EndIndex = -1;
//    pango_attr_list_insert (other, attr);

//    pango_attr_list_splice (list, other, 11, 5);

//    assert_attributes (t,list, "[11,-1]size=10\n");

//    pango_attr_list_unref (other);
//    other = pango_attr_list_new ();

//    pango_attr_list_splice (list, other, 11, 5);

//    assert_attributes (t,list, "[11,-1]size=10\n");

//    pango_attr_list_unref (other);
//    pango_attr_list_unref (list);
//  }

//  static gboolean
//  just_weight (PangoAttribute *attribute, gpointer user_data)
//  {
//    if (attribute.klass.type == ATTR_WEIGHT)
// 	 return true;
//    else
// 	 return false;
//  }

func TestListFilter(t *testing.T) {

	var list AttrList
	list.pango_attr_list_insert(pango_attr_size_new(10))
	attr := pango_attr_stretch_new(PANGO_STRETCH_CONDENSED)
	attr.StartIndex = 10
	attr.EndIndex = 20
	list.pango_attr_list_insert(attr)
	attr = pango_attr_weight_new(PANGO_WEIGHT_BOLD)
	attr.StartIndex = 20
	list.pango_attr_list_insert(attr)

	assert_attributes(t, list, "[0,-1]size=10\n"+
		"[10,20]stretch=2\n"+
		"[20,-1]weight=700\n")

	out := list.pango_attr_list_filter(func(attr *Attribute) bool { return false })
	if len(out) != 0 {
		t.Errorf("expected empty list, got %v", out)
	}

	out = list.pango_attr_list_filter(func(attr *Attribute) bool { return attr.Type == ATTR_WEIGHT })
	if len(out) == 0 {
		t.Error("expected list, got 0 elements")
	}

	assert_attributes(t, list, "[0,-1]size=10\n"+
		"[10,20]stretch=2\n")
	assert_attributes(t, out, "[20,-1]weight=700\n")
}

// TODO: add copy test once it's implemented
func TestIter(t *testing.T) {
	var list AttrList
	iter := list.pango_attr_list_get_iterator()

	assertFalse(t, iter.pango_attr_iterator_next(), "empty iterator")
	if L := iter.pango_attr_iterator_get_attrs(); len(L) != 0 {
		t.Errorf("expected empty list, got %v", L)
	}

	list = nil
	list.pango_attr_list_insert(pango_attr_size_new(10))
	attr := pango_attr_stretch_new(PANGO_STRETCH_CONDENSED)
	attr.StartIndex = 10
	attr.EndIndex = 30
	list.pango_attr_list_insert(attr)
	attr = pango_attr_weight_new(PANGO_WEIGHT_BOLD)
	attr.StartIndex = 20
	list.pango_attr_list_insert(attr)

	iter = list.pango_attr_list_get_iterator()
	// copy = pango_attr_iterator_copy(iter)
	assertEquals(t, int(iter.StartIndex), 0)
	assertEquals(t, int(iter.EndIndex), 10)
	assertTrue(t, iter.pango_attr_iterator_next(), "iterator has a next element")
	assertEquals(t, int(iter.StartIndex), 10)
	assertEquals(t, int(iter.EndIndex), 20)
	assertTrue(t, iter.pango_attr_iterator_next(), "iterator has a next element")
	assertEquals(t, int(iter.StartIndex), 20)
	assertEquals(t, int(iter.EndIndex), 30)
	assertTrue(t, iter.pango_attr_iterator_next(), "iterator has a next element")
	assertEquals(t, int(iter.StartIndex), 30)
	assertEquals(t, int(iter.EndIndex), maxInt)
	assertTrue(t, iter.pango_attr_iterator_next(), "iterator has a next element")
	assertEquals(t, int(iter.StartIndex), maxInt)
	assertEquals(t, int(iter.EndIndex), maxInt)
	assertTrue(t, !iter.pango_attr_iterator_next(), "iterator has no more element")

	// pango_attr_iterator_range(copy, &start, &end)
	// assertEquals(t, start, 0)
	// assertEquals(t, end, 10)
	// pango_attr_iterator_destroy(copy)
}

func TestIterGet(t *testing.T) {
	var list AttrList
	list.pango_attr_list_insert(pango_attr_size_new(10))
	attr := pango_attr_stretch_new(PANGO_STRETCH_CONDENSED)
	attr.StartIndex = 10
	attr.EndIndex = 30
	list.pango_attr_list_insert(attr)
	attr = pango_attr_weight_new(PANGO_WEIGHT_BOLD)
	attr.StartIndex = 20
	list.pango_attr_list_insert(attr)

	iter := list.pango_attr_list_get_iterator()
	iter.pango_attr_iterator_next()
	attr = iter.pango_attr_iterator_get(ATTR_SIZE)
	if attr == nil {
		t.Error("expected attribute")
	}
	assertEquals(t, attr.StartIndex, 0)
	assertEquals(t, attr.EndIndex, maxInt)
	attr = iter.pango_attr_iterator_get(ATTR_STRETCH)
	if attr == nil {
		t.Error("expected attribute")
	}
	assertEquals(t, attr.StartIndex, 10)
	assertEquals(t, attr.EndIndex, 30)
	attr = iter.pango_attr_iterator_get(ATTR_WEIGHT)
	if attr != nil {
		t.Errorf("expected no attribute, got %v", attr)
	}
	attr = iter.pango_attr_iterator_get(ATTR_GRAVITY)
	if attr != nil {
		t.Errorf("expected no attribute, got %v", attr)
	}
}

func TestIterGetFont(t *testing.T) {
	var list AttrList
	list.pango_attr_list_insert(pango_attr_size_new(10 * pangoScale))
	list.pango_attr_list_insert(pango_attr_family_new("Times"))
	attr := pango_attr_stretch_new(PANGO_STRETCH_CONDENSED)
	attr.StartIndex = 10
	attr.EndIndex = 30
	list.pango_attr_list_insert(attr)
	attr = pango_attr_language_new(pango_language_from_string("ja-JP"))
	attr.StartIndex = 10
	attr.EndIndex = 20
	list.pango_attr_list_insert(attr)
	attr = pango_attr_rise_new(100)
	attr.StartIndex = 20
	list.pango_attr_list_insert(attr)
	attr = pango_attr_fallback_new(false)
	attr.StartIndex = 20
	list.pango_attr_list_insert(attr)

	var (
		lang  Language
		attrs AttrList
	)
	iter := list.pango_attr_list_get_iterator()
	desc := pango_font_description_new()
	iter.pango_attr_iterator_get_font(&desc, &lang, &attrs)
	desc2 := pango_font_description_from_string("Times 10")
	assertTrue(t, desc.pango_font_description_equal(desc2), "same fonts")
	if lang != "" {
		t.Errorf("expected no language got %s", lang)
	}
	if len(attrs) != 0 {
		t.Errorf("expected no attributes, got %v", attrs)
	}

	iter.pango_attr_iterator_next()
	desc = pango_font_description_new()
	iter.pango_attr_iterator_get_font(&desc, &lang, &attrs)
	desc2 = pango_font_description_from_string("Times Condensed 10")
	assertTrue(t, desc.pango_font_description_equal(desc2), "same fonts")
	if lang == "" {
		t.Error("expected lang")
	}
	assertEquals(t, lang.String(), "ja-jp")
	if len(attrs) != 0 {
		t.Errorf("expected no attributes, got %v", attrs)
	}

	iter.pango_attr_iterator_next()
	desc = pango_font_description_new()
	iter.pango_attr_iterator_get_font(&desc, &lang, &attrs)
	desc2 = pango_font_description_from_string("Times Condensed 10")
	assertTrue(t, desc.pango_font_description_equal(desc2), "same fonts")
	if lang != "" {
		t.Errorf("expected no language got %s", lang)
	}
	assert_attributes(t, attrs, "[20,-1]rise=100\n"+
		"[20,-1]fallback=0\n")
}

func TestIterGetAttrs(t *testing.T) {
	var list AttrList
	list.pango_attr_list_insert(pango_attr_size_new(10 * pangoScale))
	list.pango_attr_list_insert(pango_attr_family_new("Times"))
	attr := pango_attr_stretch_new(PANGO_STRETCH_CONDENSED)
	attr.StartIndex = 10
	attr.EndIndex = 30
	list.pango_attr_list_insert(attr)
	attr = pango_attr_language_new(pango_language_from_string("ja-JP"))
	attr.StartIndex = 10
	attr.EndIndex = 20
	list.pango_attr_list_insert(attr)
	attr = pango_attr_rise_new(100)
	attr.StartIndex = 20
	list.pango_attr_list_insert(attr)
	attr = pango_attr_fallback_new(false)
	attr.StartIndex = 20
	list.pango_attr_list_insert(attr)

	iter := list.pango_attr_list_get_iterator()
	assert_attr_iterator(t, iter, "[0,-1]size=10240\n"+
		"[0,-1]family=Times\n")

	iter.pango_attr_iterator_next()
	assert_attr_iterator(t, iter, "[0,-1]size=10240\n"+
		"[0,-1]family=Times\n"+
		"[10,30]stretch=2\n"+
		"[10,20]language=ja-jp\n")

	iter.pango_attr_iterator_next()
	assert_attr_iterator(t, iter, "[0,-1]size=10240\n"+
		"[0,-1]family=Times\n"+
		"[10,30]stretch=2\n"+
		"[20,-1]rise=100\n"+
		"[20,-1]fallback=0\n")

	iter.pango_attr_iterator_next()
	assert_attr_iterator(t, iter, "[0,-1]size=10240\n"+
		"[0,-1]family=Times\n"+
		"[20,-1]rise=100\n"+
		"[20,-1]fallback=0\n")

	iter.pango_attr_iterator_next()
	if l := iter.pango_attr_iterator_get_attrs(); len(l) != 0 {
		t.Errorf("expected no attributes, got %v", l)
	}
}

// TODO: enable when list_update is added
// func TestListUpdate(t *testing.T) {
// 	var list AttrList
// 	attr := pango_attr_size_new(10 * pangoScale)
// 	attr.StartIndex = 10
// 	attr.EndIndex = 11
// 	list.pango_attr_list_insert(attr)
// 	attr = pango_attr_rise_new(100)
// 	attr.StartIndex = 0
// 	attr.EndIndex = 200
// 	list.pango_attr_list_insert(attr)
// 	attr = pango_attr_family_new("Times")
// 	attr.StartIndex = 5
// 	attr.EndIndex = 15
// 	list.pango_attr_list_insert(attr)
// 	attr = pango_attr_fallback_new(false)
// 	attr.StartIndex = 11
// 	attr.EndIndex = 100
// 	list.pango_attr_list_insert(attr)
// 	attr = pango_attr_stretch_new(PANGO_STRETCH_CONDENSED)
// 	attr.StartIndex = 30
// 	attr.EndIndex = 60
// 	list.pango_attr_list_insert(attr)

// 	assert_attributes(t, list, "[0,200]rise=100\n"+
// 		"[5,15]family=Times\n"+
// 		"[10,11]size=10240\n"+
// 		"[11,100]fallback=0\n"+
// 		"[30,60]stretch=2\n")

// 	list.pango_attr_list_update(8, 10, 20)

// 	assert_attributes(t, list, "[0,210]rise=100\n"+
// 		"[5,8]family=Times\n"+
// 		"[28,110]fallback=0\n"+
// 		"[40,70]stretch=2\n")

// }

//  /* Test that empty lists work in pango_attr_list_update */
// func TestListUpdate2 (t *testing.T,void) {
//    PangoAttrList *list;

//    var list AttrList
//    pango_attr_list_update (list, 8, 10, 20);

//    g_assert_null (pango_attr_list_get_attributes (list));

//    pango_attr_list_unref (list);
//  }

// func TestListEqual (t *testing.T,void) {
//    PangoAttrList *list1, *list2;
//    PangoAttribute *attr;

//    list1 = pango_attr_list_new ();
//    list2 = pango_attr_list_new ();

//    assertTrue (t,pango_attr_list_equal (NULL, NULL));
//    assertFalse (t,pango_attr_list_equal (list1, NULL));
//    assertFalse (t,pango_attr_list_equal (NULL, list1));
//    assertTrue (t,pango_attr_list_equal (list1, list1));
//    assertTrue (t,pango_attr_list_equal (list1, list2));

//    attr = pango_attr_size_new (10 * pangoScale);
//    attr.StartIndex = 0;
//    attr.EndIndex = 7;
//    pango_attr_list_insert (list1, pango_attribute_copy (attr));
//    pango_attr_list_insert (list2, pango_attribute_copy (attr));
//    pango_attribute_destroy (attr);

//    assertTrue (t,pango_attr_list_equal (list1, list2));

//    attr = pango_attr_stretch_new (PANGO_STRETCH_CONDENSED);
//    attr.StartIndex = 0;
//    attr.EndIndex = 1;
//    pango_attr_list_insert (list1, pango_attribute_copy (attr));
//    assertTrue (t,!pango_attr_list_equal (list1, list2));

//    pango_attr_list_insert (list2, pango_attribute_copy (attr));
//    assertTrue (t,pango_attr_list_equal (list1, list2));
//    pango_attribute_destroy (attr);

//    attr = pango_attr_size_new (30 * pangoScale);
//    /* Same range as the first attribute */
//    attr.StartIndex = 0;
//    attr.EndIndex = 7;
//    pango_attr_list_insert (list2, pango_attribute_copy (attr));
//    assertTrue (t,!pango_attr_list_equal (list1, list2));
//    pango_attr_list_insert (list1, pango_attribute_copy (attr));
//    assertTrue (t,pango_attr_list_equal (list1, list2));
//    pango_attribute_destroy (attr);

//    pango_attr_list_unref (list1);
//    pango_attr_list_unref (list2);

//    /* Same range but different order */
//    {
// 	 PangoAttrList *list1, *list2;
// 	 PangoAttribute *attr1, *attr2;

// 	 list1 = pango_attr_list_new ();
// 	 list2 = pango_attr_list_new ();

// 	 attr1 = pango_attr_size_new (10 * pangoScale);
// 	 attr2 = pango_attr_stretch_new (PANGO_STRETCH_CONDENSED);

// 	 pango_attr_list_insert (list1, pango_attribute_copy (attr1));
// 	 pango_attr_list_insert (list1, pango_attribute_copy (attr2));

// 	 pango_attr_list_insert (list2, pango_attribute_copy (attr2));
// 	 pango_attr_list_insert (list2, pango_attribute_copy (attr1));

// 	 pango_attribute_destroy (attr1);
// 	 pango_attribute_destroy (attr2);

// 	 assertTrue (t,pango_attr_list_equal (list1, list2));
// 	 assertTrue (t,pango_attr_list_equal (list2, list1));

// 	 pango_attr_list_unref (list1);
// 	 pango_attr_list_unref (list2);
//    }
//  }

func TestInsert(t *testing.T) {
	var list AttrList
	attr := pango_attr_size_new(10 * pangoScale)
	attr.StartIndex = 10
	attr.EndIndex = 11
	list.pango_attr_list_insert(attr)
	attr = pango_attr_rise_new(100)
	attr.StartIndex = 0
	attr.EndIndex = 200
	list.pango_attr_list_insert(attr)
	attr = pango_attr_family_new("Times")
	attr.StartIndex = 5
	attr.EndIndex = 15
	list.pango_attr_list_insert(attr)
	attr = pango_attr_fallback_new(false)
	attr.StartIndex = 11
	attr.EndIndex = 100
	list.pango_attr_list_insert(attr)
	attr = pango_attr_stretch_new(PANGO_STRETCH_CONDENSED)
	attr.StartIndex = 30
	attr.EndIndex = 60
	list.pango_attr_list_insert(attr)

	assert_attributes(t, list, "[0,200]rise=100\n"+
		"[5,15]family=Times\n"+
		"[10,11]size=10240\n"+
		"[11,100]fallback=0\n"+
		"[30,60]stretch=2\n")

	attr = pango_attr_family_new("Times")
	attr.StartIndex = 10
	attr.EndIndex = 25
	list.pango_attr_list_change(*attr)

	assert_attributes(t, list, "[0,200]rise=100\n"+
		"[5,25]family=Times\n"+
		"[10,11]size=10240\n"+
		"[11,100]fallback=0\n"+
		"[30,60]stretch=2\n")

	attr = pango_attr_family_new("Futura")
	attr.StartIndex = 11
	attr.EndIndex = 25
	list.pango_attr_list_insert(attr)

	assert_attributes(t, list, "[0,200]rise=100\n"+
		"[5,25]family=Times\n"+
		"[10,11]size=10240\n"+
		"[11,100]fallback=0\n"+
		"[11,25]family=Futura\n"+
		"[30,60]stretch=2\n")

}

/* test something that gtk does */
func TestMerge(t *testing.T) {
	var list AttrList
	attr := pango_attr_size_new(10 * pangoScale)
	attr.StartIndex = 10
	attr.EndIndex = 11
	list.pango_attr_list_insert(attr)
	attr = pango_attr_rise_new(100)
	attr.StartIndex = 0
	attr.EndIndex = 200
	list.pango_attr_list_insert(attr)
	attr = pango_attr_family_new("Times")
	attr.StartIndex = 5
	attr.EndIndex = 15
	list.pango_attr_list_insert(attr)
	attr = pango_attr_fallback_new(false)
	attr.StartIndex = 11
	attr.EndIndex = 100
	list.pango_attr_list_insert(attr)
	attr = pango_attr_stretch_new(PANGO_STRETCH_CONDENSED)
	attr.StartIndex = 30
	attr.EndIndex = 60
	list.pango_attr_list_insert(attr)

	assert_attributes(t, list, "[0,200]rise=100\n"+
		"[5,15]family=Times\n"+
		"[10,11]size=10240\n"+
		"[11,100]fallback=0\n"+
		"[30,60]stretch=2\n")

	var list2 AttrList
	attr = pango_attr_size_new(10 * pangoScale)
	attr.StartIndex = 11
	attr.EndIndex = 13
	list2.pango_attr_list_insert(attr)
	attr = pango_attr_size_new(11 * pangoScale)
	attr.StartIndex = 13
	attr.EndIndex = 15
	list2.pango_attr_list_insert(attr)
	attr = pango_attr_size_new(12 * pangoScale)
	attr.StartIndex = 40
	attr.EndIndex = 50
	list2.pango_attr_list_insert(attr)

	assert_attributes(t, list2, "[11,13]size=10240\n"+
		"[13,15]size=11264\n"+
		"[40,50]size=12288\n")

	list2.pango_attr_list_filter(func(attr *Attribute) bool {
		list.pango_attr_list_change(*attr.pango_attribute_copy())
		return false
	})

	assert_attributes(t, list, "[0,200]rise=100\n"+
		"[5,15]family=Times\n"+
		"[10,13]size=10240\n"+
		"[11,100]fallback=0\n"+
		"[13,15]size=11264\n"+
		"[30,60]stretch=2\n"+
		"[40,50]size=12288\n")
}

// reproduce what the links example in gtk4-demo does
// with the colored Google link
func TestMerge2(t *testing.T) {
	var list AttrList
	attr := pango_attr_underline_new(PANGO_UNDERLINE_SINGLE)
	attr.StartIndex = 0
	attr.EndIndex = 10
	list.pango_attr_list_insert(attr)
	attr = pango_attr_foreground_new(0, 0, 0xffff)
	attr.StartIndex = 0
	attr.EndIndex = 10
	list.pango_attr_list_insert(attr)

	assert_attributes(t, list, "[0,10]underline=1\n"+
		"[0,10]foreground=#00000000ffff\n")

	attr = pango_attr_foreground_new(0xffff, 0, 0)
	attr.StartIndex = 2
	attr.EndIndex = 3

	list.pango_attr_list_change(*attr)

	assert_attributes(t, list, "[0,10]underline=1\n"+
		"[0,2]foreground=#00000000ffff\n"+
		"[2,3]foreground=#ffff00000000\n"+
		"[3,10]foreground=#00000000ffff\n")

	attr = pango_attr_foreground_new(0, 0xffff, 0)
	attr.StartIndex = 3
	attr.EndIndex = 4

	list.pango_attr_list_change(*attr)

	assert_attributes(t, list, "[0,10]underline=1\n"+
		"[0,2]foreground=#00000000ffff\n"+
		"[2,3]foreground=#ffff00000000\n"+
		"[3,4]foreground=#0000ffff0000\n"+
		"[4,10]foreground=#00000000ffff\n")

	attr = pango_attr_foreground_new(0, 0, 0xffff)
	attr.StartIndex = 4
	attr.EndIndex = 5

	list.pango_attr_list_change(*attr)

	assert_attributes(t, list, "[0,10]underline=1\n"+
		"[0,2]foreground=#00000000ffff\n"+
		"[2,3]foreground=#ffff00000000\n"+
		"[3,4]foreground=#0000ffff0000\n"+
		"[4,10]foreground=#00000000ffff\n")
}

//  /* This only prints rise, size and scale, which are the
//   * only relevant attributes in the test that uses this
//   * function.
//   */
// func print_tags_for_attributes (PangoAttrIterator *iter,							GString           *s)
//  {
//    PangoAttribute *attr;

//    attr = pango_attr_iterator_get (iter, ATTR_RISE);
//    if (attr)
// 	 g_string_append_printf (s, "[%d, %d]rise=%d\n",
// 							 attr.StartIndex, attr.EndIndex,
// 							 ((PangoAttrInt*)attr).value);

//    attr = pango_attr_iterator_get (iter, ATTR_SIZE);
//    if (attr)
// 	 g_string_append_printf (s, "[%d, %d]size=%d\n",
// 							 attr.StartIndex, attr.EndIndex,
// 							 ((PangoAttrInt*)attr).value);

//    attr = pango_attr_iterator_get (iter, ATTR_SCALE);
//    if (attr)
// 	 g_string_append_printf (s, "[%d, %d]scale=%f\n",
// 							 attr.StartIndex, attr.EndIndex,
// 							 ((PangoAttrFloat*)attr).value);
//  }

// func TestIterEpsilonZero (t *testing.T,void) {
//    const char *markup = "𝜀<span rise=\"-6000\" size=\"x-small\" font_desc=\"italic\">0</span> = 𝜔<span rise=\"8000\" size=\"smaller\">𝜔<span rise=\"14000\" size=\"smaller\">𝜔<span rise=\"20000\">.<span rise=\"23000\">.<span rise=\"26000\">.</span></span></span></span></span>";
//    PangoAttrList *attributes;
//    PangoAttrIterator *attr;
//    char *text;
//    GError *error = NULL;
//    GString *s;

//    s = g_string_new ("");

//    pango_parse_markup (markup, -1, 0, &attributes, &text, NULL, &error);
//    g_assert_no_error (error);
//    g_assert_cmpstr (text, ==, "𝜀0 = 𝜔𝜔𝜔...");

//    attr = pango_attr_list_get_iterator (attributes);
//    do
// 	 {
// 	   int start, end;

// 	   pango_attr_iterator_range (attr, &start, &end);

// 	   g_string_append_printf (s, "range: [%d, %d]\n", start, end);

// 	   print_tags_for_attributes (attr, s);
// 	 }
//    while (pango_attr_iterator_next (attr));

//    g_assert_cmpstr (s.str, ==,
// 					"range: [0, 4]\n"
// 					"range: [4, 5]\n"
// 					"[4, 5]rise=-6000\n"
// 					"[4, 5]scale=0.694444\n"
// 					"range: [5, 12]\n"
// 					"range: [12, 16]\n"
// 					"[12, 23]rise=8000\n"
// 					"[12, 23]scale=0.833333\n"
// 					"range: [16, 20]\n"
// 					"[16, 23]rise=14000\n"
// 					"[16, 23]scale=0.694444\n"
// 					"range: [20, 21]\n"
// 					"[20, 23]rise=20000\n"
// 					"[16, 23]scale=0.694444\n"
// 					"range: [21, 22]\n"
// 					"[21, 23]rise=23000\n"
// 					"[16, 23]scale=0.694444\n"
// 					"range: [22, 23]\n"
// 					"[22, 23]rise=26000\n"
// 					"[16, 23]scale=0.694444\n"
// 					"range: [23, 2147483647]\n");

//    g_free (text);
//    pango_attr_list_unref (attributes);
//    pango_attr_iterator_destroy (attr);

//    g_string_free (s, true);
//  }

//  int
//  main (int argc, char *argv[])
//  {
//    g_test_init (&argc, &argv, NULL);

//    g_test_add_func ("/attributes/basic", test_attributes_basic);
//    g_test_add_func ("/attributes/equal", test_attributes_equal);
//    g_test_add_func ("/attributes/list/basic", test_list);
//    g_test_add_func ("/attributes/list/change", test_list_change);
//    g_test_add_func ("/attributes/list/splice", test_list_splice);
//    g_test_add_func ("/attributes/list/splice2", test_list_splice2);
//    g_test_add_func ("/attributes/list/filter", test_list_filter);
//    g_test_add_func ("/attributes/list/update", test_list_update);
//    g_test_add_func ("/attributes/list/update2", test_list_update2);
//    g_test_add_func ("/attributes/list/equal", test_list_equal);
//    g_test_add_func ("/attributes/list/insert", test_insert);
//    g_test_add_func ("/attributes/list/merge", test_merge);
//    g_test_add_func ("/attributes/list/merge2", test_merge2);
//    g_test_add_func ("/attributes/iter/basic", test_iter);
//    g_test_add_func ("/attributes/iter/get", test_iter_get);
//    g_test_add_func ("/attributes/iter/get_font", test_iter_get_font);
//    g_test_add_func ("/attributes/iter/get_attrs", test_iter_get_attrs);
//    g_test_add_func ("/attributes/iter/epsilon_zero", test_iter_epsilon_zero);

//    return g_test_run ();
//  }
