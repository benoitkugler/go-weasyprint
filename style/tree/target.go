package tree

import (
	pr "github.com/benoitkugler/go-weasyprint/style/properties"
)

// Handle target-counter, target-counters && target-text.
//
// The TargetCollector is a structure providing required targets"
// counterValues && stuff needed to build Pending targets later,
// when the layout of all targetted anchors has been done.
//
// :copyright: Copyright 2011-2019 Simon Sapin && contributors, see AUTHORS.
// :license: BSD, see LICENSE for details.

type RemakeState struct {
	Anchors                     []string
	ContentLookups              []*CounterLookupItem
	ContentChanged, PagesWanted bool
}

type SkipStack struct {
	Skip  int
	Stack *SkipStack
}

type PageState struct {
	QuoteDepth    []int
	CounterValues CounterValues
	CounterScopes []pr.Set
}

// Copy returns a deep copy.
func (s PageState) Copy() PageState

func (s PageState) Equal(other PageState) bool

type PageBreak struct {
	Break string
	Page  pr.Page
}

type PageMaker struct {
	InitialResumeAt  *SkipStack
	InitialNextPage  PageBreak
	RightPage        bool
	InitialPageState PageState
	RemakeState      RemakeState
}

type Box interface {
	CachedCounterValues() CounterValues
	SetCachedCounterValues(cv CounterValues)
	MissingLink() Box
	SetMissingLink(b Box)
}

type CounterValues map[string][]int

// Copy performs a deep copy of c
func (c CounterValues) Copy() CounterValues

func (c CounterValues) Update(other CounterValues)

func equalInts(a, b []int) bool

// Equal deeply compare each elements of c and other
func (c CounterValues) Equal(other CounterValues) bool

type functionKey struct {
	sourceBox Box
	cssToken  string
}

func NewFunctionKey(sourceBox Box, cssToken string) functionKey

type funcStore = map[functionKey]ParseFunc

type ParseFunc = func(CounterValues)

// Item controlling Pending targets and page based target counters.
//
// Collected in the TargetCollector"s ``items``.
type TargetLookupItem struct {
	state string

	// Required by target-counter and target-counters to access the
	// target's .cachedCounterValues.
	// Needed for target-text via TEXTCONTENTEXTRACTORS.
	TargetBox Box

	// Functions that have to been called to check Pending targets.
	// Keys are (sourceBox, cssToken).
	parseAgainFunctions funcStore

	// Anchor position during pagination (pageNumber - 1)
	PageMakerIndex int

	// TargetBox's pageCounters during pagination
	CachedPageCounterValues CounterValues
}

func NewTargetLookupItem(state string) *TargetLookupItem

func (t TargetLookupItem) IsUpToDate() bool

type optionnalInt struct {
	int
	none bool
}

func NewOptionnalInt(i int) optionnalInt

// Item controlling page based counters.
//
// Collected in the TargetCollector's ``CounterLookupItems``.
type CounterLookupItem struct {
	// Function that have to been called to check Pending counter.
	ParseAgain ParseFunc

	// Missing counters and target counters
	MissingCounters       pr.Set
	MissingTargetCounters map[string]pr.Set

	// Box position during pagination (pageNumber - 1)
	PageMakerIndex optionnalInt

	// Marker for remakePage
	Pending bool

	// Targeting box's pageCounters during pagination
	CachedPageCounterValues CounterValues
}

func NewCounterLookupItem(parseAgain ParseFunc, missingCounters pr.Set, missingTargetCounters map[string]pr.Set) *CounterLookupItem

// Collector of HTML targets used by CSS content with ``target-*``.
type TargetCollector struct {
	// Lookup items for targets and page counters
	TargetLookupItems  map[string]*TargetLookupItem
	CounterLookupItems map[functionKey]*CounterLookupItem

	// When collecting is true, computeContentList() collects missing
	// page counters in CounterLookupItems. Otherwise, it mixes in the
	// TargetLookupItem's CachedPageCounterValues.
	// Is switched to false in CheckPendingTargets().
	collecting bool

	// hadPendingTargets is set to true when a target is needed but has
	// not been seen yet. CheckPendingTargets then uses this information
	// to call the needed ParseAgain functions.
	hadPendingTargets bool

	// List of anchors that have already been seen during parsing.
	existingAnchors pr.Set
}

func NewTargetCollector() TargetCollector

func (t TargetCollector) IsCollecting() bool

// Get anchor name from string or uri token.
func AnchorNameFromToken(anchorToken pr.ContentProperty) string

// Store ``anchorName`` in ``existingAnchors``.
func (tc TargetCollector) collectAnchor(anchorName string)

// Store a computed internal target"s ``anchorName``.
// ``anchorName`` must not start with "#" and be already unquoted.
func (tc TargetCollector) collectComputedTarget(anchorToken pr.ContentProperty)

// Get a TargetLookupItem corresponding to ``anchorToken``.
//
// If it is already filled by a previous anchor-Element, the status is
// "up-to-date". Otherwise, it is "Pending", we must parse the whole
// tree again.
func (tc *TargetCollector) LookupTarget(anchorToken pr.ContentProperty, sourceBox Box, cssToken string, parseAgain ParseFunc) *TargetLookupItem

// Store a target called ``anchorName``.
//
// If there is a Pending TargetLookupItem, it is updated. Only previously
// collected anchors are stored.
func (tc *TargetCollector) StoreTarget(anchorName string, targetCounterValues CounterValues, targetBox Box)

// Store the counterValues in the TargetBox like
// computeContentList does.
// TODO: remove attribute or set a default value in  Box type

// Collect missing (probably page-based) counters during formatting.
//
// The ``MissingCounters`` are re-used during pagination.
//
// The ``missingLink`` attribute added to the parentBox is required to
// connect the paginated boxes to their originating ``parentBox``.
func (tc TargetCollector) CollectMissingCounters(parentBox Box, cssToken string,
	parseAgainFunction ParseFunc, missingCounters pr.Set, missingTargetCounters map[string]pr.Set)

// No counter collection during pagination

// No need to add empty miss-lists

// TODO: remove attribute or set a default value in Box type

// Check Pending targets if needed.
func (tc *TargetCollector) CheckPendingTargets()

// Ready for pagination

// Store target's current ``PageMakerIndex`` and page counter values.
//
// Eventually update associated targeting boxes.
func (tc TargetCollector) CacheTargetPageCounters(anchorName string, pageCounterValues CounterValues, pageMakerIndex int,
	pageMaker []PageMaker)

// Only store page counters when paginating

// Spread the news: update boxes affected by a change in the
// anchor"s page counter values.

// (_, cssToken) = key
// Only update items that need counters in their content

// Don"t update if item has no missing target counter

// Pending marker for remakePage

// TODO: Is the item at all interested inthe new
// pageCounterValues? It probably is and this check is a
// brake.

// Hint: the box's own cached page counters trigger a
// separate "contentChanged".
