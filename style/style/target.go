package style

import (
	"log"
	"strings"

	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

// Handle target-counter, target-counters && target-text.
//
// The targetCollector is a structure providing required targets"
// counterValues && stuff needed to build pending targets later,
// when the layout of all targetted anchors has been done.
//
// :copyright: Copyright 2011-2019 Simon Sapin && contributors, see AUTHORS.
// :license: BSD, see LICENSE for details.

type Box interface {
	MissingLink() Box
	SetMissingLink(b Box)
	CachedCounterValues() CounterValues
	SetCachedCounterValues(cv CounterValues)
}

type CounterValues map[string][]int

// Copy performs a deep copy of c
func (c CounterValues) Copy() CounterValues {
	out := make(CounterValues, len(c))
	for k, v := range c {
		out[k] = append([]int{}, v...)
	}
	return out
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

type funcStore = map[functionKey]checkFunc

type checkFunc func(CounterValues)

// Item controlling pending targets and page based target counters.
//
// Collected in the targetCollector"s ``items``.
type targetLookupItem struct {
	state string

	// Required by target-counter and target-counters to access the
	// target's .cachedCounterValues.
	// Needed for target-text via TEXTCONTENTEXTRACTORS.
	targetBox Box

	// Functions that have to been called to check pending targets.
	// Keys are (sourceBox, cssToken).
	parseAgainFunctions funcStore

	// Anchor position during pagination (pageNumber - 1)
	pageMakerIndex int

	// targetBox's pageCounters during pagination
	cachedPageCounterValues CounterValues
}

func NewTargetLookupItem(state string) *targetLookupItem {
	if state == "" {
		state = "pending"
	}
	return &targetLookupItem{state: state, parseAgainFunctions: funcStore{}, cachedPageCounterValues: CounterValues{}}
}

type optionnalInt struct {
	int
	none bool
}

// Item controlling page based counters.
//
// Collected in the targetCollector's ``counterLookupItems``.
type counterLookupItem struct {
	// Function that have to been called to check pending counter.
	parseAgain checkFunc

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

func NewCounterLookupItem(parseAgain checkFunc, missingCounters pr.Set, missingTargetCounters map[string]pr.Set) *counterLookupItem {
	return &counterLookupItem{
		parseAgain:              parseAgain,
		missingCounters:         missingCounters,
		missingTargetCounters:   missingTargetCounters,
		pageMakerIndex:          optionnalInt{none: true},
		cachedPageCounterValues: CounterValues{},
	}
}

// Collector of HTML targets used by CSS content with ``target-*``.
type targetCollector struct {
	// Lookup items for targets and page counters
	targetLookupItems  map[string]*targetLookupItem
	counterLookupItems map[functionKey]*counterLookupItem

	// When collecting is true, computeContentList() collects missing
	// page counters in CounterLookupItems. Otherwise, it mixes in the
	// targetLookupItem's cachedPageCounterValues.
	// Is switched to false in checkPendingTargets().
	collecting bool

	// hadPendingTargets is set to true when a target is needed but has
	// not been seen yet. checkPendingTargets then uses this information
	// to call the needed parseAgain functions.
	hadPendingTargets bool

	// List of anchors that have already been seen during parsing.
	existingAnchors pr.Set
}

func NewTargetCollector() targetCollector {
	return targetCollector{
		targetLookupItems:  map[string]*targetLookupItem{},
		counterLookupItems: map[functionKey]*counterLookupItem{},
		collecting:         true,
	}
}

// Get anchor name from string or uri token.
func anchorNameFromToken(anchorToken pr.ContentProperty) string {
	asString, ok := anchorToken.Content.(pr.String)
	asUrl, ok := anchorToken.Content.(pr.NamedString)
	if anchorToken.Type == "string" && ok && strings.HasPrefix(string(asString), "#") {
		return string(asString[1:])
	} else if anchorToken.Type == "url" && asUrl.Name == "internal" {
		return string(asUrl.String)
	}
	return ""
}

// Store ``anchorName`` in ``existingAnchors``.
func (tc targetCollector) collectAnchor(anchorName string) {
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
func (tc targetCollector) collectComputedTarget(anchorToken pr.ContentProperty) {
	anchorName := anchorNameFromToken(anchorToken)
	if anchorName != "" {
		if _, in := tc.targetLookupItems[anchorName]; !in {
			tc.targetLookupItems[anchorName] = NewTargetLookupItem("")
		}
	}
}

// Get a targetLookupItem corresponding to ``anchorToken``.
//
// If it is already filled by a previous anchor-element, the status is
// "up-to-date". Otherwise, it is "pending", we must parse the whole
// tree again.
func (tc *targetCollector) lookupTarget(anchorToken pr.ContentProperty, sourceBox Box, cssToken string, parseAgain checkFunc) *targetLookupItem {
	anchorName := anchorNameFromToken(anchorToken)
	item, in := tc.targetLookupItems[anchorName]
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
// If there is a pending targetLookupItem, it is updated. Only previously
// collected anchors are stored.
func (tc *targetCollector) storeTarget(anchorName string, targetCounterValues CounterValues, targetBox Box) {
	item := tc.targetLookupItems[anchorName]
	if item != nil && item.state == "pending" {
		item.state = "up-to-date"
		item.targetBox = targetBox
		// Store the counterValues in the targetBox like
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
func (tc targetCollector) collectMissingCounters(parentBox Box, cssToken string,
	parseAgainFunction checkFunc, missingCounters pr.Set, missingTargetCounters map[string]pr.Set) {

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
func (tc *targetCollector) checkPendingTargets() {
	if tc.hadPendingTargets {
		for _, item := range tc.targetLookupItems {
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
func (tc targetCollector) cacheTargetPageCounters(anchorName string, pageCounterValues CounterValues, pageMakerIndex int,
	pageMaker [][]map[string]bool) {

	// Only store page counters when paginating
	if tc.collecting {
		return
	}

	item := tc.targetLookupItems[anchorName]
	if item != nil && item.state == "up-to-date" {
		item.pageMakerIndex = pageMakerIndex
		if !item.cachedPageCounterValues.Equal(pageCounterValues) {
			item.cachedPageCounterValues = pageCounterValues.Copy()
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
