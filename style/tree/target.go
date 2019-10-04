package tree

import (
	"log"
	"strings"

	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

// Handle target-counter, target-counters && target-text.
//
// The TargetCollector is a structure providing required targets"
// counterValues && stuff needed to build pending targets later,
// when the layout of all targetted anchors has been done.
//
// :copyright: Copyright 2011-2019 Simon Sapin && contributors, see AUTHORS.
// :license: BSD, see LICENSE for details.

type CounterValues map[string][]int

type Box interface {
	CachedCounterValues() CounterValues
	SetCachedCounterValues(cv CounterValues)
	MissingLink() Box
	SetMissingLink(b Box)
}

// Copy performs a deep copy of c
func (c CounterValues) Copy() CounterValues {
	out := make(CounterValues, len(c))
	for k, v := range c {
		out[k] = append([]int{}, v...)
	}
	return out
}

func (c CounterValues) Update(other CounterValues) {
	for k, v := range other {
		c[k] = v
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, ai := range a {
		if ai != b[i] {
			return false
		}
	}
	return true
}

// Equal deeply compare each elements of c and other
func (c CounterValues) Equal(other CounterValues) bool {
	if len(c) != len(other) {
		return false
	}
	for k, v := range c {
		if !equalInts(v, other[k]) {
			return false
		}
	}
	return true
}

type functionKey struct {
	sourceBox Box
	cssToken  string
}

type funcStore = map[functionKey]ParseFunc

type ParseFunc = func(CounterValues)

// Item controlling pending targets and page based target counters.
//
// Collected in the TargetCollector"s ``items``.
type TargetLookupItem struct {
	state string

	// Required by target-counter and target-counters to access the
	// target's .cachedCounterValues.
	// Needed for target-text via TEXTCONTENTEXTRACTORS.
	TargetBox Box

	// Functions that have to been called to check pending targets.
	// Keys are (sourceBox, cssToken).
	parseAgainFunctions funcStore

	// Anchor position during pagination (pageNumber - 1)
	pageMakerIndex int

	// TargetBox's pageCounters during pagination
	CachedPageCounterValues CounterValues
}

func NewTargetLookupItem(state string) *TargetLookupItem {
	if state == "" {
		state = "pending"
	}
	return &TargetLookupItem{state: state, parseAgainFunctions: funcStore{}, CachedPageCounterValues: CounterValues{}}
}

func (t TargetLookupItem) IsUpToDate() bool {
	return t.state == "up-to-date"
}

type optionnalInt struct {
	int
	none bool
}

// Item controlling page based counters.
//
// Collected in the TargetCollector's ``counterLookupItems``.
type counterLookupItem struct {
	// Function that have to been called to check pending counter.
	parseAgain ParseFunc

	// Missing counters and target counters
	missingCounters       pr.Set
	missingTargetCounters map[string]pr.Set

	// Box position during pagination (pageNumber - 1)
	pageMakerIndex optionnalInt

	// Marker for remakePage
	pending bool

	// Targeting box's pageCounters during pagination
	cachedPageCounterValues CounterValues
}

func NewCounterLookupItem(parseAgain ParseFunc, missingCounters pr.Set, missingTargetCounters map[string]pr.Set) *counterLookupItem {
	return &counterLookupItem{
		parseAgain:              parseAgain,
		missingCounters:         missingCounters,
		missingTargetCounters:   missingTargetCounters,
		pageMakerIndex:          optionnalInt{none: true},
		cachedPageCounterValues: CounterValues{},
	}
}

// Collector of HTML targets used by CSS content with ``target-*``.
type TargetCollector struct {
	// Lookup items for targets and page counters
	TargetLookupItems  map[string]*TargetLookupItem
	counterLookupItems map[functionKey]*counterLookupItem

	// When collecting is true, computeContentList() collects missing
	// page counters in CounterLookupItems. Otherwise, it mixes in the
	// TargetLookupItem's CachedPageCounterValues.
	// Is switched to false in CheckPendingTargets().
	collecting bool

	// hadPendingTargets is set to true when a target is needed but has
	// not been seen yet. CheckPendingTargets then uses this information
	// to call the needed parseAgain functions.
	hadPendingTargets bool

	// List of anchors that have already been seen during parsing.
	existingAnchors pr.Set
}

func NewTargetCollector() TargetCollector {
	return TargetCollector{
		TargetLookupItems:  map[string]*TargetLookupItem{},
		counterLookupItems: map[functionKey]*counterLookupItem{},
		collecting:         true,
	}
}

func (t TargetCollector) IsCollecting() bool {
	return t.collecting
}

// Get anchor name from string or uri token.
func AnchorNameFromToken(anchorToken pr.ContentProperty) string {
	asString, _ := anchorToken.Content.(pr.String)
	asUrl, ok := anchorToken.Content.(pr.NamedString)
	if anchorToken.Type == "string" && ok && strings.HasPrefix(string(asString), "#") {
		return string(asString[1:])
	} else if anchorToken.Type == "url" && asUrl.Name == "internal" {
		return asUrl.String
	}
	return ""
}

// Store ``anchorName`` in ``existingAnchors``.
func (tc TargetCollector) collectAnchor(anchorName string) {
	if anchorName != "" {
		if tc.existingAnchors.Has(anchorName) {
			log.Printf("Anchor defined twice: %s \n", anchorName)
		} else {
			tc.existingAnchors.Add(anchorName)
		}
	}
}

// Store a computed internal target"s ``anchorName``.
// ``anchorName`` must not start with "#" and be already unquoted.
func (tc TargetCollector) collectComputedTarget(anchorToken pr.ContentProperty) {
	anchorName := AnchorNameFromToken(anchorToken)
	if anchorName != "" {
		if _, in := tc.TargetLookupItems[anchorName]; !in {
			tc.TargetLookupItems[anchorName] = NewTargetLookupItem("")
		}
	}
}

// Get a TargetLookupItem corresponding to ``anchorToken``.
//
// If it is already filled by a previous anchor-Element, the status is
// "up-to-date". Otherwise, it is "pending", we must parse the whole
// tree again.
func (tc *TargetCollector) LookupTarget(anchorToken pr.ContentProperty, sourceBox Box, cssToken string, parseAgain ParseFunc) *TargetLookupItem {
	anchorName := AnchorNameFromToken(anchorToken)
	item, in := tc.TargetLookupItems[anchorName]
	if !in {
		item = NewTargetLookupItem("undefined")
	}

	if item.state == "pending" {
		if tc.existingAnchors.Has(anchorName) {
			tc.hadPendingTargets = true
			key := functionKey{sourceBox: sourceBox, cssToken: cssToken}
			if _, in := item.parseAgainFunctions[key]; !in {
				item.parseAgainFunctions[key] = parseAgain
			}
		} else {
			item.state = "undefined"
		}
	}

	if item.state == "undefined" {
		log.Printf("Content discarded: target points to undefined anchor '%s' \n", anchorToken)
	}

	return item
}

// Store a target called ``anchorName``.
//
// If there is a pending TargetLookupItem, it is updated. Only previously
// collected anchors are stored.
func (tc *TargetCollector) StoreTarget(anchorName string, targetCounterValues CounterValues, targetBox Box) {
	item := tc.TargetLookupItems[anchorName]
	if item != nil && item.state == "pending" {
		item.state = "up-to-date"
		item.TargetBox = targetBox
		// Store the counterValues in the TargetBox like
		// computeContentList does.
		// TODO: remove attribute or set a default value in  Box type
		if targetBox.CachedCounterValues() == nil {
			targetBox.SetCachedCounterValues(targetCounterValues.Copy())
		}
	}
}

// Collect missing (probably page-based) counters during formatting.
//
// The ``missingCounters`` are re-used during pagination.
//
// The ``missingLink`` attribute added to the parentBox is required to
// connect the paginated boxes to their originating ``parentBox``.
func (tc TargetCollector) CollectMissingCounters(parentBox Box, cssToken string,
	parseAgainFunction ParseFunc, missingCounters pr.Set, missingTargetCounters map[string]pr.Set) {

	// No counter collection during pagination
	if !tc.collecting {
		return
	}

	// No need to add empty miss-lists
	if len(missingCounters) > 0 || len(missingTargetCounters) > 0 {
		// TODO: remove attribute or set a default value in Box type
		if parentBox.MissingLink() == nil {
			parentBox.SetMissingLink(parentBox)
		}
		counterLookupItem := NewCounterLookupItem(
			parseAgainFunction, missingCounters,
			missingTargetCounters)
		key := functionKey{sourceBox: parentBox, cssToken: cssToken}
		if _, in := tc.counterLookupItems[key]; !in {
			tc.counterLookupItems[key] = counterLookupItem
		}

	}
}

// Check pending targets if needed.
func (tc *TargetCollector) CheckPendingTargets() {
	if tc.hadPendingTargets {
		for _, item := range tc.TargetLookupItems {
			for _, function := range item.parseAgainFunctions {
				function(nil)
			}
		}
		tc.hadPendingTargets = false
	}
	// Ready for pagination
	tc.collecting = false
}

// Store target's current ``pageMakerIndex`` and page counter values.
//
// Eventually update associated targeting boxes.
func (tc TargetCollector) cacheTargetPageCounters(anchorName string, pageCounterValues CounterValues, pageMakerIndex int,
	pageMaker [][]map[string]bool) {

	// Only store page counters when paginating
	if tc.collecting {
		return
	}

	item := tc.TargetLookupItems[anchorName]
	if item != nil && item.state == "up-to-date" {
		item.pageMakerIndex = pageMakerIndex
		if !item.CachedPageCounterValues.Equal(pageCounterValues) {
			item.CachedPageCounterValues = pageCounterValues.Copy()
		}
	}

	// Spread the news: update boxes affected by a change in the
	// anchor"s page counter values.
	for key, item := range tc.counterLookupItems {
		// (_, cssToken) = key
		// Only update items that need counters in their content
		if key.cssToken != "content" {
			continue
		}

		// Don"t update if item has no missing target counter
		missingCounters := item.missingTargetCounters[anchorName]
		if missingCounters == nil {
			continue
		}

		// Pending marker for remakePage
		if item.pageMakerIndex.none || item.pageMakerIndex.int >= len(pageMaker) {
			item.pending = true
			continue
		}

		// TODO: Is the item at all interested inthe new
		// pageCounterValues? It probably is and this check is a
		// brake.
		for counterName := range missingCounters {
			if _, in := pageCounterValues[counterName]; in {
				l := pageMaker[item.pageMakerIndex.int]
				remakeState := l[len(l)-1]
				remakeState["content_changed"] = true
				item.parseAgain(item.cachedPageCounterValues)
				break
			}
		}
		// Hint: the box's own cached page counters trigger a
		// separate "contentChanged".
	}
}
