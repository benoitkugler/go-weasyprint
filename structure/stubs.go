package structure

// autogenerated from source_box.py

type InstanceBox interface {
	isBox()
}
type InstanceParentBox interface {
	isParentBox()
	isBox()
}

//
func (b *ParentBox) allChildren() {
	ParentBoxAllChildren(&b.BoxFields)
}

//
func (b *ParentBox) resetSpacing(side string) {
	ParentBoxResetSpacing(&b.BoxFields, side)
}

//
func (b *ParentBox) removeDecoration(start bool, end bool) {
	ParentBoxRemoveDecoration(&b.BoxFields, start, end)
}

// is_start=True is_end=True
func (b *ParentBox) copyWithChildren(newChildren []Box, isStart bool, isEnd bool) {
	ParentBoxCopyWithChildren(&b.BoxFields, newChildren, isStart, isEnd)
}

//
func (b *ParentBox) deepcopy() {
	ParentBoxDeepcopy(&b.BoxFields)
}

//
func (b *ParentBox) descendants() {
	ParentBoxDescendants(&b.BoxFields)
}

//
func (b *ParentBox) getWrappedTable() {
	ParentBoxGetWrappedTable(&b.BoxFields)
}

//
func (b *ParentBox) pageValues() {
	ParentBoxPageValues(&b.BoxFields)
}

type InstanceBlockLevelBox interface {
	isBlockLevelBox()
	isBox()
}
type InstanceBlockContainerBox interface {
	isBlockContainerBox()
	isBox()
	isParentBox()
}
type InstanceBlockBox interface {
	isBlockBox()
	isBlockLevelBox()
	isParentBox()
	isBox()
	isBlockContainerBox()
}
type InstanceLineBox interface {
	isLineBox()
	isBox()
	isParentBox()
}
type InstanceInlineLevelBox interface {
	isInlineLevelBox()
	isBox()
}

//
func (b *InlineLevelBox) removeDecoration(start bool, end bool) {
	InlineLevelBoxRemoveDecoration(&b.BoxFields, start, end)
}

type InstanceInlineBox interface {
	isInlineBox()
	isInlineLevelBox()
	isBox()
	isParentBox()
}
type InstanceTextBox interface {
	isTextBox()
	isInlineLevelBox()
	isBox()
}
type InstanceAtomicInlineLevelBox interface {
	isAtomicInlineLevelBox()
	isInlineLevelBox()
	isBox()
}
type InstanceInlineBlockBox interface {
	isInlineBlockBox()
	isAtomicInlineLevelBox()
	isParentBox()
	isBox()
	isBlockContainerBox()
	isInlineLevelBox()
}
type InstanceReplacedBox interface {
	isReplacedBox()
	isBox()
}
type InstanceBlockReplacedBox interface {
	isBlockReplacedBox()
	isBlockLevelBox()
	isBox()
	isReplacedBox()
}
type InstanceInlineReplacedBox interface {
	isInlineReplacedBox()
	isAtomicInlineLevelBox()
	isBox()
	isInlineLevelBox()
	isReplacedBox()
}
type InstanceTableBox interface {
	isTableBox()
	isBlockLevelBox()
	isBox()
	isParentBox()
}

//
func (b *TableBox) allChildren() {
	TableBoxAllChildren(&b.BoxFields)
}

// dx=0 dy=0 ignore_floats=False
func (b *TableBox) translate(dx float32, dy float32, ignoreFloats bool) {
	TableBoxTranslate(&b.BoxFields, dx, dy, ignoreFloats)
}

//
func (b *TableBox) pageValues() {
	TableBoxPageValues(&b.BoxFields)
}

type InstanceInlineTableBox interface {
	isInlineTableBox()
	isBlockLevelBox()
	isParentBox()
	isBox()
	isTableBox()
}
type InstanceTableRowGroupBox interface {
	isTableRowGroupBox()
	isBox()
	isParentBox()
}
type InstanceTableRowBox interface {
	isTableRowBox()
	isBox()
	isParentBox()
}
type InstanceTableColumnGroupBox interface {
	isTableColumnGroupBox()
	isBox()
	isParentBox()
}
type InstanceTableColumnBox interface {
	isTableColumnBox()
	isBox()
	isParentBox()
}
type InstanceTableCellBox interface {
	isTableCellBox()
	isBlockContainerBox()
	isParentBox()
	isBox()
}
type InstanceTableCaptionBox interface {
	isTableCaptionBox()
	isBlockLevelBox()
	isBlockBox()
	isParentBox()
	isBox()
	isBlockContainerBox()
}
type InstancePageBox interface {
	isPageBox()
	isBox()
	isParentBox()
}
type InstanceMarginBox interface {
	isMarginBox()
	isBlockContainerBox()
	isParentBox()
	isBox()
}
type InstanceFlexContainerBox interface {
	isFlexContainerBox()
	isBox()
	isParentBox()
}
type InstanceFlexBox interface {
	isFlexBox()
	isBlockLevelBox()
	isBox()
	isParentBox()
	isFlexContainerBox()
}
type InstanceInlineFlexBox interface {
	isInlineFlexBox()
	isBox()
	isParentBox()
	isInlineLevelBox()
	isFlexContainerBox()
}
type Box struct {
	BoxFields
	InstanceBox
}

type ParentBox struct {
	BoxFields
	InstanceParentBox
}

type BlockLevelBox struct {
	BoxFields
	InstanceBlockLevelBox

	clearance interface{} // None
}

type BlockContainerBox struct {
	BoxFields
	InstanceBlockContainerBox
}

type BlockBox struct {
	BoxFields
	InstanceBlockBox
}

type LineBox struct {
	BoxFields
	InstanceLineBox

	textOverflow string // "clip"
}

type InlineLevelBox struct {
	BoxFields
	InstanceInlineLevelBox
}

type InlineBox struct {
	BoxFields
	InstanceInlineBox
}

type TextBox struct {
	BoxFields
	InstanceTextBox

	justificationSpacing int // 0
}

type AtomicInlineLevelBox struct {
	BoxFields
	InstanceAtomicInlineLevelBox
}

type InlineBlockBox struct {
	BoxFields
	InstanceInlineBlockBox
}

type ReplacedBox struct {
	BoxFields
	InstanceReplacedBox
}

type BlockReplacedBox struct {
	BoxFields
	InstanceBlockReplacedBox
}

type InlineReplacedBox struct {
	BoxFields
	InstanceInlineReplacedBox
}

type TableBox struct {
	BoxFields
	InstanceTableBox

	TableFields
}

type InlineTableBox struct {
	BoxFields
	InstanceInlineTableBox

	TableFields
}

type TableRowGroupBox struct {
	BoxFields
	InstanceTableRowGroupBox
}

type TableRowBox struct {
	BoxFields
	InstanceTableRowBox
}

type TableColumnGroupBox struct {
	BoxFields
	InstanceTableColumnGroupBox

	span int // 1
}

type TableColumnBox struct {
	BoxFields
	InstanceTableColumnBox

	span int // 1
}

type TableCellBox struct {
	BoxFields
	InstanceTableCellBox
}

type TableCaptionBox struct {
	BoxFields
	InstanceTableCaptionBox
}

type PageBox struct {
	BoxFields
	InstancePageBox
}

type MarginBox struct {
	BoxFields
	InstanceMarginBox
}

type FlexContainerBox struct {
	BoxFields
	InstanceFlexContainerBox
}

type FlexBox struct {
	BoxFields
	InstanceFlexBox
}

type InlineFlexBox struct {
	BoxFields
	InstanceInlineFlexBox
}