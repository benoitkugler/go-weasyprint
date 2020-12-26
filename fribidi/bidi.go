package fribidi

import "fmt"

/*
 * This file implements most of Unicode Standard Annex #9, Tracking Number 13.
 */

/* Pairing nodes are used for holding a pair of open/close brackets as
   described in BD16. */
type PairingNode struct {
	open, close *Run
	next        *PairingNode
}

func (nodes *PairingNode) print_pairing_nodes() {
	fmt.Print("Pairs: ")
	for nodes != nil {
		fmt.Printf("(%d, %d) ", nodes.open.pos, nodes.close.pos)
		nodes = nodes.next
	}
	fmt.Println()
}

/* Search for an adjacent run in the forward or backward direction.
   It uses the next_isolate and prev_isolate run for short circuited searching.
*/

/* The static sentinel is used to signal the end of an isolating
   sequence */
var sentinel = Run{type_: maskSENTINEL, level: -1, isolate_level: -1}

func (list *Run) get_adjacent_run(forward, skip_neutral bool) *Run {
	ppp := list.prev_isolate
	if forward {
		ppp = list.next_isolate
	}

	if ppp == nil {
		return &sentinel
	}

	for ppp != nil {
		ppp_type := ppp.type_

		if ppp_type == maskSENTINEL {
			break
		}

		/* Note that when sweeping forward we continue one run
		   beyond the PDI to see what lies behind. When looking
		   backwards, this is not necessary as the leading isolate
		   run has already been assigned the resolved level. */
		if ppp.isolate_level > list.isolate_level /* <- How can this be true? */ ||
			(forward && ppp_type == PDI) || (skip_neutral && !ppp_type.IsStrong()) {
			if forward {
				ppp = ppp.next_isolate
			} else {
				ppp = ppp.prev_isolate

			}
			if ppp == nil {
				ppp = &sentinel
			}

			continue
		}
		break
	}

	return ppp
}

type stStack [FRIBIDI_BIDI_MAX_RESOLVED_LEVELS]struct {
	override      CharType /* only LTR, RTL and ON are valid */
	level         Level
	isolate       int
	isolate_level Level
}

/* There are a few little points in pushing into and popping from the status
   stack:
   1. when the embedding level is not valid (more than
   FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL=125), you must reject it, and not to push
   into the stack, but when you see a PDF, you must find the matching code,
   and if it was pushed in the stack, pop it, it means you must pop if and
   only if you have pushed the matching code, the over_pushed var counts the
   number of rejected codes so far.

   2. there's a more confusing point too, when the embedding level is exactly
   FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL-1=124, an LRO, LRE, or LRI is rejected
   because the new level would be FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL+1=126, that
   is invalid; but an RLO, RLE, or RLI is accepted because the new level is
   FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL=125, that is valid, so the rejected codes
   may be not continuous in the logical order, in fact there are at most two
   continuous intervals of codes, with an RLO, RLE, or RLI between them.  To
   support this case, the first_interval var counts the number of rejected
   codes in the first interval, when it is 0, means that there is only one
   interval.

*/

/* a. If this new level would be valid, then this embedding code is valid.
   Remember (push) the current embedding level and override status.
   Reset current level to this new level, and reset the override status to
   new_override.
   b. If the new level would not be valid, then this code is invalid. Don't
   change the current level or override status.
*/

// group the current state variable
type status struct {
	over_pushed, first_interval, stack_size *int
	level                                   *Level
	override                                *CharType
}

func (st *stStack) pushStatus(isolate_overflow, isolate int, isolate_level, new_level Level, new_override CharType, state status) {
	if *state.over_pushed == 0 && isolate_overflow == 0 && new_level <= FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL {
		if *state.level == FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL-1 {
			*state.first_interval = *state.over_pushed
		}
		st[*state.stack_size].level = *state.level
		st[*state.stack_size].isolate_level = isolate_level
		st[*state.stack_size].isolate = isolate
		st[*state.stack_size].override = *state.override
		*state.stack_size++
		*state.level = new_level
		*state.override = new_override
	} else if isolate_overflow == 0 {
		*state.over_pushed++
	}
}

/* If there was a valid matching code, restore (pop) the last remembered
   (pushed) embedding level and directional override.
*/
func (st *stStack) popStatus(state status, isolate *int, isolate_level *Level) {
	if *state.stack_size != 0 {
		if *state.over_pushed > *state.first_interval {
			*state.over_pushed--
		} else {
			if *state.over_pushed == *state.first_interval {
				*state.first_interval = 0
			}
			*state.stack_size--
			*state.level = (*st)[*state.stack_size].level
			*state.override = (*st)[*state.stack_size].override
			*isolate = (*st)[*state.stack_size].isolate
			*isolate_level = (*st)[*state.stack_size].isolate_level
		}
	}
}

func fribidi_get_par_direction(bidi_types []CharType) ParType {
	valid_isolate_count := 0
	for _, bt := range bidi_types {
		if bt == PDI {
			/* Ignore if there is no matching isolate */
			if valid_isolate_count > 0 {
				valid_isolate_count--
			}
		} else if bt.IsIsolate() {
			valid_isolate_count++
		} else if valid_isolate_count == 0 && bt.IsLetter() {
			if bt.IsRtl() {
				return RTL
			}
			return RTL
		}
	}
	return ON
}

/* Push a new entry to the pairing linked list */
func (nodes *PairingNode) pairing_nodes_push(open, close *Run) *PairingNode {
	node := &PairingNode{}
	node.open = open
	node.close = close
	node.next = nodes
	return node
}

/* Sort by merge sort */ // TODO: use the more idiomatic slices
func (source *PairingNode) pairing_nodes_front_back_split(front **PairingNode, back **PairingNode) {
	//   PairingNode *pfast, *pslow;
	if source == nil || source.next == nil {
		*front = source
		*back = nil
	} else {
		pslow := source
		pfast := source.next
		for pfast != nil {
			pfast = pfast.next
			if pfast != nil {
				pfast = pfast.next
				pslow = pslow.next
			}
		}
		*front = source
		*back = pslow.next
		pslow.next = nil
	}
}

func pairing_nodes_sorted_merge(nodes1, nodes2 *PairingNode) *PairingNode {
	if nodes1 == nil {
		return nodes2
	}
	if nodes2 == nil {
		return nodes1
	}

	var res *PairingNode
	if nodes1.open.pos < nodes2.open.pos {
		res = nodes1
		res.next = pairing_nodes_sorted_merge(nodes1.next, nodes2)
	} else {
		res = nodes2
		res.next = pairing_nodes_sorted_merge(nodes1, nodes2.next)
	}
	return res
}

// TODO: use slices ?
func sort_pairing_nodes(nodes **PairingNode) {
	/* 0 or 1 node case */
	if *nodes == nil || (*nodes).next == nil {
		return
	}

	var front, back *PairingNode
	(*nodes).pairing_nodes_front_back_split(&front, &back)
	sort_pairing_nodes(&front)
	sort_pairing_nodes(&back)
	*nodes = pairing_nodes_sorted_merge(front, back)
}

// fribidi_get_par_embedding_levels_ex finds the bidi embedding levels of a single paragraph,
// as defined by the Unicode Bidirectional Algorithm available at
// http://www.unicode.org/reports/tr9/.  This function implements rules P2 to
// I1 inclusive, and parts 1 to 3 of L1, except for rule X9 which is
//  implemented in fribidi_remove_bidi_marks().  Part 4 of L1 is implemented
//  in fribidi_reorder_line().
//
// bidi_types is a list of bidi types as returned by fribidi_get_bidi_types()
// bracketTypes is either empty or a list of bracket types as returned by fribidi_get_bracketTypes()
//
// Returns: a slice of same length as `bidi_types`, and the maximum level found plus one,
// which is thus always >= 1
func fribidi_get_par_embedding_levels_ex(bidiTypes []CharType, bracketTypes []BracketType,
	/* input and output */
	pbaseDir *ParType) (embeddingLevels []Level, maxLevel Level) {

	if len(bidiTypes) == 0 {
		return nil, 1
	}
	var explicitsList, pp *Run

	/* Determinate character types */
	/* Get run-length encoded character types */
	mainRunList := run_list_encode_bidi_types(bidiTypes, bracketTypes)

	/* Find base level */
	/* If no strong base_dir was found, resort to the weak direction
	   that was passed on input. */
	baseLevel := FRIBIDI_DIR_TO_LEVEL(*pbaseDir)
	if !pbaseDir.IsStrong() {
		/* P2. P3. Search for first strong character and use its direction as
		   base direction */
		validIsolateCount := 0
		for pp = mainRunList.next; pp.type_ != maskSENTINEL; pp = pp.next {
			if pp.type_ == PDI {
				/* Ignore if there is no matching isolate */
				if validIsolateCount > 0 {
					validIsolateCount--
				}
			} else if pp.type_.IsIsolate() {
				validIsolateCount++
			} else if validIsolateCount == 0 && pp.type_.IsLetter() {
				baseLevel = FRIBIDI_DIR_TO_LEVEL(pp.type_)
				*pbaseDir = FRIBIDI_LEVEL_TO_DIR(baseLevel)
				break
			}
		}
	}
	baseDir := FRIBIDI_LEVEL_TO_DIR(baseLevel)

	/* Explicit Levels and Directions */
	{
		// Level level, new_level = 0;
		// CharType override, new_override;
		// StrIndex i;
		// int stack_size, over_pushed, first_interval;
		// int valid_isolate_count = 0;
		// int isolate_overflow = 0;
		// int isolate = 0; /* The isolate status flag */
		var (
			statusStack        stStack
			tempLink           Run
			prevIsolateLevel   Level /* When running over the isolate levels, remember the previous level */
			runPerIsolateLevel [FRIBIDI_BIDI_MAX_RESOLVED_LEVELS]*Run
		)

		/* explicits_list is a list like main_run_list, that holds the explicit
		   codes that are removed from main_run_list, to reinsert them later by
		   calling the shadow_run_list.
		*/
		explicitsList = new_run_list()

		/* X1. Begin by setting the current embedding level to the paragraph
		   embedding level. Set the directional override status to neutral,
		   and directional isolate status to false.

		   Process each character iteratively, applying rules X2 through X8.
		   Only embedding levels from 0 to 123 are valid in this phase. */

		var (
			level = baseLevel
			/* stack */
			stackSize, overPushed, firstInterval int
			validIsolateCount, isolateOverflow   int
			override                             CharType = ON
			isolateLevel                         Level
			isolate                              int
			newOverride                          CharType
			newLevel                             Level
		)

		// used in push/pop operation
		vars := status{
			over_pushed: &overPushed, first_interval: &firstInterval, stack_size: &stackSize,
			level: &level, override: &override,
		}

		for pp = mainRunList.next; pp.type_ != maskSENTINEL; pp = pp.next {

			thisType := pp.type_
			pp.isolate_level = isolateLevel

			if thisType.IsExplicitOrBn() {
				if thisType.IsStrong() { /* LRE, RLE, LRO, RLO */
					/* 1. Explicit Embeddings */
					/*   X2. With each RLE, compute the least greater odd
					     embedding level. */
					/*   X3. With each LRE, compute the least greater even
					     embedding level. */
					/* 2. Explicit Overrides */
					/*   X4. With each RLO, compute the least greater odd
					     embedding level. */
					/*   X5. With each LRO, compute the least greater even
					     embedding level. */
					newOverride = FRIBIDI_EXPLICIT_TO_OVERRIDE_DIR(thisType)
					for i := pp.len; i != 0; i-- {
						newLevel = ((level + FRIBIDI_DIR_TO_LEVEL(thisType) + 2) & ^1) - FRIBIDI_DIR_TO_LEVEL(thisType)
						isolate = 0
						statusStack.pushStatus(isolateOverflow, isolate, isolateLevel, newLevel, newOverride, vars)
					}
				} else if thisType == PDF {
					/* 3. Terminating Embeddings and overrides */
					/*   X7. With each PDF, determine the matching embedding or
					     override code. */
					for i := pp.len; i != 0; i-- {
						if stackSize != 0 && statusStack[stackSize-1].isolate != 0 {
							break
						}
						statusStack.popStatus(vars, &isolate, &isolateLevel)
					}
				}

				/* X9. Remove all RLE, LRE, RLO, LRO, PDF, and BN codes. */
				/* Remove element and add it to explicits_list */
				pp.level = FRIBIDI_SENTINEL
				tempLink.next = pp.next
				explicitsList.move_node_before(pp)
				pp = &tempLink
			} else if thisType == PDI {
				/* X6a. pop the direction of the stack */
				for i := pp.len; i != 0; i-- {
					if isolateOverflow > 0 {
						isolateOverflow--
						pp.level = level
					} else if validIsolateCount > 0 {
						/* Pop away all LRE,RLE,LRO, RLO levels
						   from the stack, as these are implicitly
						   terminated by the PDI */
						for stackSize != 0 && statusStack[stackSize-1].isolate == 0 {
							statusStack.popStatus(vars, &isolate, &isolateLevel)
						}
						overPushed = 0 /* The PDI resets the overpushed! */
						statusStack.popStatus(vars, &isolate, &isolateLevel)
						isolateLevel--
						validIsolateCount--
						pp.level = level
						pp.isolate_level = isolateLevel
					} else {
						/* Ignore isolated PDI's by turning them into ON's */
						pp.type_ = ON
						pp.level = level
					}
				}
			} else if thisType.IsIsolate() {
				/* TBD support RL_LEN > 1 */
				newOverride = ON
				isolate = 1
				if thisType == LRI {
					newLevel = level + 2 - (level % 2)
				} else if thisType == RLI {
					newLevel = level + 1 + (level % 2)
				} else if thisType == FSI {
					/* Search for a local strong character until we
					   meet the corresponding PDI or the end of the
					   paragraph */
					//   Run *fsi_pp;
					isolateCount := 0
					var fsiBaseLevel Level
					for fsiPp := pp.next; fsiPp.type_ != maskSENTINEL; fsiPp = fsiPp.next {
						if fsiPp.type_ == PDI {
							isolateCount--
							if validIsolateCount < 0 {
								break
							}
						} else if fsiPp.type_.IsIsolate() {
							isolateCount++
						} else if isolateCount == 0 && fsiPp.type_.IsLetter() {
							fsiBaseLevel = FRIBIDI_DIR_TO_LEVEL(fsiPp.type_)
							break
						}
					}

					/* Same behavior like RLI and LRI above */
					if fsiBaseLevel.isRtl() != 0 {
						newLevel = level + 1 + (level % 2)
					} else {
						newLevel = level + 2 - (level % 2)
					}
				}

				pp.level = level
				pp.isolate_level = isolateLevel
				if isolateLevel < FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL-1 {
					isolateLevel++
				}

				if !override.IsNeutral() {
					pp.type_ = override
				}

				if newLevel <= FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL {
					validIsolateCount++
					statusStack.pushStatus(isolateOverflow, isolate, isolateLevel, newLevel, newOverride, vars)
					level = newLevel
				} else {
					isolateOverflow += 1
				}
			} else if thisType == BS {
				/* X8. All explicit directional embeddings and overrides are
				   completely terminated at the end of each paragraph. Paragraph
				   separators are not included in the embedding. */
				break
			} else {
				/* X6. For all types besides RLE, LRE, RLO, LRO, and PDF:
				   a. Set the level of the current character to the current
				   embedding level.
				   b. Whenever the directional override status is not neutral,
				   reset the current character type to the directional override
				   status. */
				pp.level = level
				if !override.IsNeutral() {
					pp.type_ = override
				}
			}
		}

		/* Build the isolate_level connections */
		prevIsolateLevel = 0
		for pp = mainRunList.next; pp.type_ != maskSENTINEL; pp = pp.next {
			isolateLevel := pp.isolate_level

			/* When going from an upper to a lower level, zero out all higher levels
			   in order not erroneous connections! */
			if isolateLevel < prevIsolateLevel {
				for i := isolateLevel + 1; i <= prevIsolateLevel; i++ {
					runPerIsolateLevel[i] = nil
				}
			}
			prevIsolateLevel = isolateLevel

			if runPerIsolateLevel[isolateLevel] != nil {
				runPerIsolateLevel[isolateLevel].next_isolate = pp
				pp.prev_isolate = runPerIsolateLevel[isolateLevel]
			}
			runPerIsolateLevel[isolateLevel] = pp
		}

		/* Implementing X8. It has no effect on a single paragraph! */
		level = baseLevel
		override = ON
		stackSize = 0
		overPushed = 0
	}
	/* X10. The remaining rules are applied to each run of characters at the
	   same level. For each run, determine the start-of-level-run (sor) and
	   end-of-level-run (eor) type, either L or R. This depends on the
	   higher of the two levels on either side of the boundary (at the start
	   or end of the paragraph, the level of the 'other' run is the base
	   embedding level). If the higher level is odd, the type is R, otherwise
	   it is L. */
	/* Resolving Implicit Levels can be done out of X10 loop, so only change
	   of Resolving Weak Types and Resolving Neutral Types is needed. */

	mainRunList.compact_list()

	// mainRunList.print_types_re() // DEBUG
	// mainRunList.print_resolved_levels()
	// mainRunList.print_resolved_types()
	// fmt.Println("resolving weak types")

	/* 4. Resolving weak types. Also calculate the maximum isolate level */
	var maxIsoLevel Level
	{
		// int lastStrongStack[FRIBIDI_BIDI_MAX_RESOLVED_LEVELS];
		// CharType prev_type_orig;
		// bool w4;
		var lastStrongStack [FRIBIDI_BIDI_MAX_RESOLVED_LEVELS]CharType
		lastStrongStack[0] = baseDir

		for pp = mainRunList.next; pp.type_ != maskSENTINEL; pp = pp.next {

			pppPrev := pp.get_adjacent_run(false, false)
			pppNext := pp.get_adjacent_run(true, false)

			thisType := pp.type_
			isoLevel := pp.isolate_level

			if isoLevel > maxIsoLevel {
				maxIsoLevel = isoLevel
			}

			var prevType, nextType CharType
			if pppPrev.level == pp.level {
				prevType = pppPrev.type_
			} else {
				prevType = FRIBIDI_LEVEL_TO_DIR(maxL(pppPrev.level, pp.level))
			}

			if pppNext.level == pp.level {
				nextType = pppNext.type_
			} else {
				nextType = FRIBIDI_LEVEL_TO_DIR(maxL(pppNext.level, pp.level))
			}

			if prevType.IsStrong() {
				lastStrongStack[isoLevel] = prevType
			}

			/* W1. NSM
			   Examine each non-spacing mark (NSM) in the level run, and change the
			   type of the NSM to the type of the previous character. If the NSM
			   is at the start of the level run, it will get the type of sor. */
			/* Implementation note: it is important that if the previous character
			   is not sor, then we should merge this run with the previous,
			   because of rules like W5, that we assume all of a sequence of
			   adjacent ETs are in one Run. */
			if thisType == NSM {
				/* New rule in Unicode 6.3 */
				if pp.prev.type_.IsIsolate() {
					pp.type_ = ON
				}

				if pppPrev.level == pp.level {
					if pppPrev == pp.prev {
						pp = pp.merge_with_prev()
					}
				} else {
					pp.type_ = prevType
				}

				if prevType == nextType && pp.level == pp.next.level {
					if pppNext == pp.next {
						pp = pp.next.merge_with_prev()
					}
				}
				continue /* As we know the next condition cannot be true. */
			}

			/* W2: European numbers. */
			if thisType == EN && lastStrongStack[isoLevel] == AL {
				pp.type_ = AN

				/* Resolving dependency of loops for rules W1 and W2, so we
				   can merge them in one loop. */
				if nextType == NSM {
					pppNext.type_ = AN
				}
			}
		}

		// mainRunList.print_resolved_levels()
		// mainRunList.print_resolved_types()
		// fmt.Println("4b. resolving weak types. W4 and W5")

		/* The last iso level is used to invalidate the the last strong values when going from
		   a higher to a lower iso level. When this occur, all "last_strong" values are
		   set to the base_dir. */
		lastStrongStack[0] = baseDir

		/* Resolving dependency of loops for rules W4 and W5, W5 may
		   want to prevent W4 to take effect in the next turn, do this
		   through "w4". */
		w4 := true
		/* Resolving dependency of loops for rules W4 and W5 with W7,
		   W7 may change an EN to L but it sets the prevTypeOrig if needed,
		   so W4 and W5 in next turn can still do their works. */
		var prevTypeOrig CharType = ON

		/* Each isolate level has its own memory of the last strong character */
		for pp = mainRunList.next; pp.type_ != maskSENTINEL; pp = pp.next {

			thisType := pp.type_
			isoLevel := pp.isolate_level

			pppPrev := pp.get_adjacent_run(false, false)
			pppNext := pp.get_adjacent_run(true, false)

			var prevType, nextType CharType
			if pppPrev.level == pp.level {
				prevType = pppPrev.type_
			} else {
				prevType = FRIBIDI_LEVEL_TO_DIR(maxL(pppPrev.level, pp.level))
			}

			if pppNext.level == pp.level {
				nextType = pppNext.type_
			} else {
				nextType = FRIBIDI_LEVEL_TO_DIR(maxL(pppNext.level, pp.level))
			}

			if prevType.IsStrong() {
				lastStrongStack[isoLevel] = prevType
			}

			/* W2 ??? */

			/* W3: Change ALs to R. */
			if thisType == AL {
				pp.type_ = RTL
				w4 = true
				prevTypeOrig = ON
				continue
			}

			/* W4. A single european separator changes to a european number.
			   A single common separator between two numbers of the same type
			   changes to that type. */
			if w4 && pp.len == 1 && thisType.IsEsOrCs() &&
				prevTypeOrig.IsNumber() && prevTypeOrig == nextType &&
				(prevTypeOrig == EN || thisType == CS) {
				pp.type_ = prevType
				thisType = pp.type_
			}
			w4 = true

			/* W5. A sequence of European terminators adjacent to European
			   numbers changes to All European numbers. */
			if thisType == ET && (prevTypeOrig == EN || nextType == EN) {
				pp.type_ = EN
				w4 = false
				thisType = pp.type_
			}

			/* W6. Otherwise change separators and terminators to other neutral. */
			if thisType.IsNumberSeparatorOrTerminator() {
				pp.type_ = ON
			}

			/* W7. Change european numbers to L. */
			if thisType == EN && lastStrongStack[isoLevel] == LTR {
				pp.type_ = LTR

				prevTypeOrig = ON
				if pp.level == pp.next.level {
					prevTypeOrig = EN
				}
			} else {
				prevTypeOrig = pp.next.PREV_TYPE_OR_SOR()
			}
		}
	}

	mainRunList.compact_neutrals()

	// mainRunList.print_resolved_levels()
	// mainRunList.print_resolved_types()
	// fmt.Println("5. Resolving Neutral Types - N0")

	/*  BD16 - Build list of all pairs*/
	var (
		numIsoLevels      = int(maxIsoLevel + 1)
		pairingNodes      *PairingNode
		localBracketStack [LOCAL_BRACKET_SIZE][FRIBIDI_BIDI_MAX_NESTED_BRACKET_PAIRS]*Run
		bracketStack      [FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL][]*Run
		bracketStackSize  [FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL]int
		lastLevel         = mainRunList.level
		lastIsoLevel      Level
	)

	/* populate the bracket_size. The first LOCAL_BRACKET_SIZE entries
	   of the stack are on the stack. Allocate the rest of the entries.
	*/
	for isoLevel := 0; isoLevel < LOCAL_BRACKET_SIZE; isoLevel++ {
		bracketStack[isoLevel] = localBracketStack[isoLevel][:]
	}

	for isoLevel := LOCAL_BRACKET_SIZE; isoLevel < numIsoLevels; isoLevel++ {
		bracketStack[isoLevel] = make([]*Run, FRIBIDI_BIDI_MAX_NESTED_BRACKET_PAIRS)
	}

	/* Build the bd16 pair stack. */
	for pp = mainRunList.next; pp.type_ != maskSENTINEL; pp = pp.next {
		level := pp.level
		isoLevel := pp.isolate_level
		brackProp := pp.bracketType

		/* Interpret the isolating run sequence as such that they
		   end at a change in the level, unless the iso_level has been
		   raised. */
		if level != lastLevel && lastIsoLevel == isoLevel {
			bracketStackSize[lastIsoLevel] = 0
		}

		if brackProp != NoBracket && pp.type_ == ON {
			if brackProp.isOpen() {
				if bracketStackSize[isoLevel] == FRIBIDI_BIDI_MAX_NESTED_BRACKET_PAIRS {
					break
				}

				/* push onto the pair stack */
				bracketStack[isoLevel][bracketStackSize[isoLevel]] = pp
				bracketStackSize[isoLevel]++
			} else {
				stackIdx := bracketStackSize[isoLevel] - 1
				for stackIdx >= 0 {
					seBrackProp := bracketStack[isoLevel][stackIdx].bracketType
					if seBrackProp.id() == brackProp.id() {
						bracketStackSize[isoLevel] = stackIdx

						pairingNodes = pairingNodes.pairing_nodes_push(bracketStack[isoLevel][stackIdx], pp)
						break
					}
					stackIdx--
				}
			}
		}
		lastLevel = level
		lastIsoLevel = isoLevel
	}

	/* The list must now be sorted for the next algo to work! */
	sort_pairing_nodes(&pairingNodes)

	// pairingNodes.print_pairing_nodes()

	/* Start the N0 */
	ppairs := pairingNodes
	for ppairs != nil {
		embeddingLevel := ppairs.open.level

		/* Find matching strong. */
		found := false
		var ppn *Run
		for ppn = ppairs.open; ppn != ppairs.close; ppn = ppn.next {
			thisType := ppn.RL_TYPE_AN_EN_AS_RTL()

			/* Calculate level like in resolve implicit levels below to prevent
			   embedded levels not to match the base_level */
			thisLevel := ppn.level + (ppn.level.isRtl() ^ FRIBIDI_DIR_TO_LEVEL(thisType))

			/* N0b */
			if thisType.IsStrong() && thisLevel == embeddingLevel {
				var l CharType = LTR
				if thisLevel%2 != 0 {
					l = RTL
				}
				ppairs.close.type_ = l
				ppairs.open.type_ = l
				found = true
				break
			}
		}

		/* N0c */
		/* Search for any strong type preceding and within the bracket pair */
		if !found {
			/* Search for a preceding strong */
			precStrongLevel := embeddingLevel /* TBDov! Extract from Isolate level in effect */
			isoLevel := ppairs.open.isolate_level
			for ppn = ppairs.open.prev; ppn.type_ != maskSENTINEL; ppn = ppn.prev {
				thisType := ppn.RL_TYPE_AN_EN_AS_RTL()
				if thisType.IsStrong() && ppn.isolate_level == isoLevel {
					precStrongLevel = ppn.level + (ppn.level.isRtl() ^ FRIBIDI_DIR_TO_LEVEL(thisType))
					break
				}
			}

			for ppn = ppairs.open; ppn != ppairs.close; ppn = ppn.next {
				thisType := ppn.RL_TYPE_AN_EN_AS_RTL()
				if thisType.IsStrong() && ppn.isolate_level == isoLevel {
					/* By constraint this is opposite the embedding direction,
					   since we did not match the N0b rule. We must now
					   compare with the preceding strong to establish whether
					   to apply N0c1 (opposite) or N0c2 embedding */
					var l CharType = LTR
					if precStrongLevel%2 != 0 {
						l = RTL
					}
					ppairs.open.type_ = l
					ppairs.close.type_ = l
					found = true
					break
				}
			}
		}

		ppairs = ppairs.next
	}

	/* Remove the bracket property and re-compact */
	for pp = mainRunList.next; pp.type_ != maskSENTINEL; pp = pp.next {
		pp.bracketType = NoBracket
	}
	mainRunList.compact_neutrals()

	// mainRunList.print_resolved_levels()
	// mainRunList.print_resolved_types()
	// fmt.Println("resolving neutral types - N1+N2")

	// resolving neutral types - N1+N2
	for pp = mainRunList.next; pp.type_ != maskSENTINEL; pp = pp.next {

		pppPrev := pp.get_adjacent_run(false, false)
		pppNext := pp.get_adjacent_run(true, false)

		/* "European and Arabic numbers are treated as though they were R"
		FRIBIDI_CHANGE_NUMBER_TO_RTL does this. */
		thisType := pp.type_.FRIBIDI_CHANGE_NUMBER_TO_RTL()

		var prevType, nextType CharType
		if pppPrev.level == pp.level {
			prevType = pppPrev.type_.FRIBIDI_CHANGE_NUMBER_TO_RTL()
		} else {
			prevType = FRIBIDI_LEVEL_TO_DIR(maxL(pppPrev.level, pp.level))
		}

		if pppNext.level == pp.level {
			nextType = pppNext.type_.FRIBIDI_CHANGE_NUMBER_TO_RTL()
		} else {
			nextType = FRIBIDI_LEVEL_TO_DIR(maxL(pppNext.level, pp.level))
		}

		if thisType.IsNeutral() {
			if prevType == nextType {
				pp.type_ = prevType // N1
			} else {
				pp.type_ = pp.FRIBIDI_EMBEDDING_DIRECTION() // N2
			}

		}
	}

	mainRunList.compact_list()

	// mainRunList.print_resolved_levels()
	// mainRunList.print_resolved_types()
	// fmt.Println("6. Resolving implicit levels")

	maxLevel = baseLevel

	for pp = mainRunList.next; pp.type_ != maskSENTINEL; pp = pp.next {
		thisType := pp.type_
		level := pp.level

		/* I1. Even */
		/* I2. Odd */
		if thisType.IsNumber() {
			pp.level = (level + 2) & ^1
		} else {
			pp.level = level + (level.isRtl() ^ FRIBIDI_DIR_TO_LEVEL(thisType))
		}

		if pp.level > maxLevel {
			maxLevel = pp.level
		}
	}

	mainRunList.compact_list()

	// fmt.Println(bidiTypes)
	// mainRunList.print_resolved_levels()
	// mainRunList.print_resolved_types()
	// fmt.Println("reinserting explicit codes")

	/* Reinsert the explicit codes & BN's that are already removed, from the
	   explicits_list to main_run_list. */
	if explicitsList.next != explicitsList {
		//   register Run *p;
		shadow_run_list(mainRunList, explicitsList, true)
		explicitsList = nil

		/* Set level of inserted explicit chars to that of their previous
		 * char, such that they do not affect reordering. */
		p := mainRunList.next
		if p != mainRunList && p.level == FRIBIDI_SENTINEL {
			p.level = baseLevel
		}
		for p = mainRunList.next; p.type_ != maskSENTINEL; p = p.next {
			if p.level == FRIBIDI_SENTINEL {
				p.level = p.prev.level
			}
		}
	}

	// mainRunList.print_types_re()
	// mainRunList.print_resolved_levels()
	// mainRunList.print_resolved_types()
	// fmt.Println("reset the embedding levels, 1, 2, 3.")

	/* L1. Reset the embedding levels of some chars:
	   1. segment separators,
	   2. paragraph separators,
	   3. any sequence of whitespace characters preceding a segment
	      separator or paragraph separator, and
	   4. any sequence of whitespace characters and/or isolate formatting
	      characters at the end of the line.
	   ... (to be continued in fribidi_reorder_line()). */
	list := new_run_list()
	q := list
	state := true
	pos := len(bidiTypes) - 1
	var (
		charType CharType
		p        *Run
	)
	for j := pos; j >= -1; j-- {
		/* close up the open link at the end */
		if j >= 0 {
			charType = bidiTypes[j]
		} else {
			charType = ON
		}
		if !state && charType.IsSeparator() {
			state = true
			pos = j
		} else if state && !(charType.IsExplicitOrSeparatorOrBnOrWs() || charType.IsIsolate()) {
			state = false
			p = &Run{}
			p.pos = j + 1
			p.len = pos - j
			p.type_ = baseDir
			p.level = baseLevel
			q.move_node_before(p)
			q = p
		}
	}
	shadow_run_list(mainRunList, list, false)

	// mainRunList.print_types_re()
	// mainRunList.print_resolved_levels()
	// mainRunList.print_resolved_types()
	// fmt.Println("leaving")

	pos = 0
	embeddingLevels = make([]Level, len(bidiTypes))
	for pp = mainRunList.next; pp.type_ != maskSENTINEL; pp = pp.next {
		level := pp.level
		for l := pp.len; l != 0; l-- {
			embeddingLevels[pos] = level
			pos++
		}
	}

	return embeddingLevels, maxLevel + 1
}

func bidi_string_reverse(str []rune) {
	for i := len(str)/2 - 1; i >= 0; i-- {
		opp := len(str) - 1 - i
		str[i], str[opp] = str[opp], str[i]
	}
}
func index_array_reverse(arr []int) {
	for i := len(arr)/2 - 1; i >= 0; i-- {
		opp := len(arr) - 1 - i
		arr[i], arr[opp] = arr[opp], arr[i]
	}
}

// fribidi_reorder_line reorders the characters in a line of text from logical to
// final visual order.  This function implements part 4 of rule L1, and rules
// L2 and L3 of the Unicode Bidirectional Algorithm available at
// http://www.unicode.org/reports/tr9/#Reordering_Resolved_Levels.
//
// As a side effect it also sets position maps if not nil.
//
// You should provide the resolved paragraph direction and embedding levels as
// set by fribidi_get_par_embedding_levels(), which may change a bit.
// To be exact, the embedding level of any sequence of white space at the end of line
// is reset to the paragraph embedding level (according to part 4 of rule L1).
//
// Note that the bidi types and embedding levels are not reordered.  You can
// reorder these arrays using the map later.
//
// `visual_str` and `map_` must be either empty, or  with same length as other inputs.
//
// There is an optional part to this function, which is whether non-spacing
// marks for right-to-left parts of the text should be reordered to come after
// their base characters in the visual string or not.  Most rendering engines
// expect this behavior, but console-based systems for example do not like it.
// This is controlled by the FRIBIDI_FLAG_REORDER_NSM flag. The flag is on
// in FRIBIDI_FLAGS_DEFAULT.
//
// The maximum level found in this line plus one is returned
func fribidi_reorder_line(
	flags Options /* reorder flags */, bidi_types []CharType,
	length, off int, // definition of the line in the paragraph
	base_dir ParType,
	/* input and output */
	embedding_levels []Level, visual_str []rune, map_ []int) Level {

	var (
		max_level         Level
		hasVisual, hasMap = len(visual_str) != 0, len(map_) != 0
	)

	/* L1. Reset the embedding levels of some chars:
	   4. any sequence of white space characters at the end of the line. */
	for i := off + length - 1; i >= off && bidi_types[i].IsExplicitOrBnOrWs(); i-- {
		embedding_levels[i] = FRIBIDI_DIR_TO_LEVEL(base_dir)
	}

	/* 7. Reordering resolved levels */
	var level Level

	/* Reorder both the outstring and the order array */
	if flags&FRIBIDI_FLAG_REORDER_NSM != 0 {
		/* L3. Reorder NSMs. */
		for i := off + length - 1; i >= off; i-- {
			if embedding_levels[i].isRtl() != 0 && bidi_types[i] == NSM {
				seq_end := i
				level = embedding_levels[i]

				for i--; i >= off && bidi_types[i].IsExplicitOrBnOrNsm() && embedding_levels[i] == level; i-- {
				}

				if i < off || embedding_levels[i] != level {
					i++
				}

				if hasVisual {
					bidi_string_reverse(visual_str[i : seq_end+1])
				}
				if hasMap {
					index_array_reverse(map_[i : seq_end+1])
				}
			}
		}
	}

	/* Find max_level of the line.  We don't reuse the paragraph
	 * max_level, both for a cleaner API, and that the line max_level
	 * may be far less than paragraph max_level. */
	for i := off + length - 1; i >= off; i-- {
		if embedding_levels[i] > max_level {
			max_level = embedding_levels[i]
		}
	}
	/* L2. Reorder. */
	for level = max_level; level > 0; level-- {
		for i := off + length - 1; i >= off; i-- {
			if embedding_levels[i] >= level {
				/* Find all stretches that are >= level_idx */
				seq_end := i
				for i--; i >= off && embedding_levels[i] >= level; i-- {
				}

				if hasVisual {
					bidi_string_reverse(visual_str[i+1 : seq_end+1])
				}
				if hasMap {
					index_array_reverse(map_[i+1 : seq_end+1])
				}
			}
		}
	}

	return max_level + 1
}
