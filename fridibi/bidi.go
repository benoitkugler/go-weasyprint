package fribidi

/* fribidi_get_par_embedding_levels_ex - get bidi embedding levels of a paragraph
 *
 * This function finds the bidi embedding levels of a single paragraph,
 * as defined by the Unicode Bidirectional Algorithm available at
 * http://www.unicode.org/reports/tr9/.  This function implements rules P2 to
 * I1 inclusive, and parts 1 to 3 of L1, except for rule X9 which is
 *  implemented in fribidi_remove_bidi_marks().  Part 4 of L1 is implemented
 *  in fribidi_reorder_line().
 *
 * There are a few macros defined in fribidi-bidi-types.h to work with this
 * embedding levels.
 *
 * Returns: Maximum level found plus one, or zero if any error occurred
 * (memory allocation failure most probably).




/*
 * This file implements most of Unicode Standard Annex #9, Tracking Number 13.
*/

// #ifndef MAX
// # define MAX(a,b) ((a) > (b) ? (a) : (b))
// #endif /* !MAX */

// /* Some convenience macros */
// #define RL_TYPE_(list) ((list).type)
// #define RL_LEN_(list) ((list).len)
// #define RL_LEVEL_(list) ((list).level)

// /* "Within this scope, bidirectional types EN and AN are treated as R" */
// #define RL_TYPE_AN_EN_AS_RTL(list) ( \
//  (((list).type == FRIBIDI_TYPE_AN) || ((list).type == FRIBIDI_TYPE_EN) | ((list).type == FRIBIDI_TYPE_RTL)) ? FRIBIDI_TYPE_RTL : (list).type)
// #define list.bracket_type ((list).bracket_type)
// #define RL_ISOLATE_LEVEL(list) ((list).isolate_level)

// #define LOCAL_BRACKET_SIZE 16

/* Pairing nodes are used for holding a pair of open/close brackets as
   described in BD16. */
type PairingNode struct {
	open, close *Run
	next        *PairingNode
}

func (second *Run) merge_with_prev() *Run {
	first := second.prev
	first.next = second.next
	first.next.prev = first
	first.len += second.len
	if second.next_isolate {
		second.next_isolate.prev_isolate = second.prev_isolate
		/* The following edge case typically shouldn't happen, but fuzz
		   testing shows it does, and the assignment protects against
		   a dangling pointer. */
	} else if second.next.prev_isolate == second {
		second.next.prev_isolate = second.prev_isolate
	}
	if second.prev_isolate {
		second.prev_isolate.next_isolate = second.next_isolate
	}
	first.next_isolate = second.next_isolate

	return first
}

func (list *Run) compact_list() {
	if list.next {
		for list = list.next; list.type_ != maskSENTINEL; list = list.next {
			/* Don't join brackets! */
			if list.prev.bracket_type == list.bracket_type && list.prev.level == list.level &&
				list.bracket_type == NoBracket && list.prev.bracket_type == NoBracket {
				list = merge_with_prev(list)
			}
		}
	}
}

func (list *Run) compact_neutrals() {
	if list.next {
		for list = list.next; list.type_ != maskSENTINEL; list = list.next {
			if list.prev.level == list.level &&
				(list.prev.bracket_type == list.bracket_type ||
					(list.prev.bracket_type.IsNeutral() && list.bracket_type.IsNeutral())) &&
				list.bracket_type == NoBracket /* Don't join brackets! */ &&
				list.prev.bracket_type == NoBracket {
				list = merge_with_prev(list)
			}
		}
	}
}

/* Search for an adjacent run in the forward or backward direction.
   It uses the next_isolate and prev_isolate run for short circuited searching.
*/

/* The static sentinel is used to signal the end of an isolating
   sequence */
var sentinel = Run{NULL, NULL, 0, 0, FRIBIDI_TYPE_SENTINEL, -1, -1, NoBrNoBracket, NULL, NULL}

func (list *Run) get_adjacent_run(forward, skip_neutral bool) *Run {
	ppp := list.prev_isolate
	if forward {
		ppp = list.next_isolate
	}

	if !ppp {
		return &sentinel
	}

	for ppp {
		ppp_type := ppp.bracket_type

		if ppp_type == FRIBIDI_TYPE_SENTINEL {
			break
		}

		/* Note that when sweeping forward we continue one run
		   beyond the PDI to see what lies behind. When looking
		   backwards, this is not necessary as the leading isolate
		   run has already been assigned the resolved level. */
		if ppp.isolate_level > list.isolate_level /* <- How can this be true? */ ||
			(forward && ppp_type == FRIBIDI_TYPE_PDI) || (skip_neutral && !FRIBIDI_IS_STRONG(ppp_type)) {
			if forward {
				ppp = ppp.next_isolate
			} else {
				ppp = ppp.prev_isolate

			}
			if !ppp {
				ppp = &sentinel
			}

			continue
		}
		break
	}

	return ppp
}

// #ifdef DEBUG
// /*======================================================================
//  *  For debugging, define some functions for printing the types and the
//  *  levels.
//  *----------------------------------------------------------------------*/

// static char char_from_level_array[] = {
//   '$',				/* -1 == FRIBIDI_SENTINEL, indicating
// 				 * start or end of string. */
//   /* 0-61 == 0-9,a-z,A-Z are the the only valid levels before resolving
//    * implicits.  after that the level @ may be appear too. */
//   '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
//   'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
//   'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't',
//   'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D',
//   'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N',
//   'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X',
//   'Y', 'Z',

//   /* TBD - insert another 125-64 levels */

//   '@',				/* 62 == only must appear after resolving
// 				 * implicits. */

//   '!',				/* 63 == FRIBIDI_LEVEL_INVALID, internal error,
// 				 * this level shouldn't be seen.  */

//   '*', '*', '*', '*', '*'	/* >= 64 == overflows, this levels and higher
// 				 * levels show a real bug!. */
// };

// #define fribidi_char_from_level(level) char_from_level_array[(level) + 1]

// static void
// print_types_re (
//   const Run *pp
// )
// {
//   fribidi_assert (pp);

//   MSG ("  Run types  : ");
//   for_run_list (pp, pp)
//   {
//     MSG6 ("%d:%d(%s)[%d,%d] ",
// 	  pp.pos, pp.len, fribidi_get_bidi_type_name (pp.type), pp.level, pp.isolate_level);
//   }
//   MSG ("\n");
// }

// static void
// print_resolved_levels (
//   const Run *pp
// )
// {
//   fribidi_assert (pp);

//   MSG ("  Res. levels: ");
//   for_run_list (pp, pp)
//   {
//     register StrIndex i;
//     for (i = RL_LEN (pp); i; i--)
//       MSG2 ("%c", fribidi_char_from_level (RL_LEVEL (pp)));
//   }
//   MSG ("\n");
// }

// static void
// print_resolved_types (
//   const Run *pp
// )
// {
//   fribidi_assert (pp);

//   MSG ("  Res. types : ");
//   for_run_list (pp, pp)
//   {
//     StrIndex i;
//     for (i = RL_LEN (pp); i; i--)
//       MSG2 ("%s ", fribidi_get_bidi_type_name (pp.type));
//   }
//   MSG ("\n");
// }

// static void
// print_bidi_string (
//   /* input */
//   const CharType *bidi_types,
//   const StrIndex len
// )
// {
//   register StrIndex i;

//   fribidi_assert (bidi_types);

//   MSG ("  Org. types : ");
//   for (i = 0; i < len; i++)
//     MSG2 ("%s ", fribidi_get_bidi_type_name (bidi_types[i]));
//   MSG ("\n");
// }

// static void print_pairing_nodes(PairingNode *nodes)
// {
//   MSG ("Pairs: ");
//   for (nodes)
//     {
//       MSG3 ("(%d, %d) ", nodes.open.pos, nodes.close.pos);
//       nodes = nodes.next;
//     }
//   MSG ("\n");
// }
// #endif /* DEBUG */

/*=========================================================================
 * define macros for push and pop the status in to / out of the stack
 *-------------------------------------------------------------------------*/

type stStack [FRIBIDI_BIDI_MAX_RESOLVED_LEVELS]struct {
	override      CharType /* only LTR, RTL and ON are valid */
	level         Level
	isolate       int
	isolate_level int
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

func (st *stStack) PUSH_STATUS(over_pushed, stack_size int, level, new_level Level) (op, ss int, le Level, ov CharType) {
	if over_pushed == 0 && isolate_overflow == 0 && new_level <= FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL {
		if level == FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL-1 {
			first_interval = over_pushed
		}
		st[stack_size].level = level
		st[stack_size].isolate_level = isolate_level
		st[stack_size].isolate = isolate
		st[stack_size].override = override
		stack_size++
		level = new_level
		override = new_override
	} else if isolate_overflow == 0 {
		over_pushed++
	}
	return over_pushed, stack_size, level, overide
}

// #define PUSH_STATUS \
//     FRIBIDI_BEGIN_STMT \
//       if (over_pushed == 0 \
//                 && isolate_overflow == 0 \
//                 && new_level <= FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL)   \
//         { \
//           if (level == FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL - 1) \
//             first_interval = over_pushed; \
//           status_stack[stack_size].level = level; \
//           status_stack[stack_size].isolate_level = isolate_level; \
//           status_stack[stack_size].isolate = isolate; \
//           status_stack[stack_size].override = override; \
//           stack_size++; \
//           level = new_level; \
//           override = new_override; \
//         } else if (isolate_overflow == 0) \
// 	  over_pushed++; \
//     FRIBIDI_END_STMT

/* If there was a valid matching code, restore (pop) the last remembered
   (pushed) embedding level and directional override.
*/
func (st *stStack) PUSH_STATUS() {
	if stack_size {
		if over_pushed > first_interval {
			over_pushed--
		} else {
			if over_pushed == first_interval {
				first_interval = 0
			}
			stack_size--
			level = status_stack[stack_size].level
			override = status_stack[stack_size].override
			isolate = status_stack[stack_size].isolate
			isolate_level = status_stack[stack_size].isolate_level
		}
	}
}

// #define POP_STATUS \
//     FRIBIDI_BEGIN_STMT \
//       if (stack_size) \
//       { \
//         if (over_pushed > first_interval) \
//           over_pushed--; \
//         else \
//           { \
//             if (over_pushed == first_interval) \
//               first_interval = 0; \
//             stack_size--; \
//             level = status_stack[stack_size].level; \
//             override = status_stack[stack_size].override; \
//             isolate = status_stack[stack_size].isolate; \
//             isolate_level = status_stack[stack_size].isolate_level; \
//           } \
//       } \
//     FRIBIDI_END_STMT

// /* Return the type of previous run or the SOR, if already at the start of
//    a level run. */
// #define PREV_TYPE_OR_SOR(pp) \
//     ( \
//       pp.prev.level == pp.level ? \
//         pp.prev.bracket_type : \
//         FRIBIDI_LEVEL_TO_DIR(MAX(pp.prev.level, pp.level)) \
//     )

// /* Return the type of next run or the EOR, if already at the end of
//    a level run. */
// #define NEXT_TYPE_OR_EOR(pp) \
//     ( \
//       pp.next.level == pp.level ? \
//         pp.next.bracket_type : \
//         FRIBIDI_LEVEL_TO_DIR(MAX(pp.next.level, pp.level)) \
//     )

// /* Return the embedding direction of a link. */
// #define FRIBIDI_EMBEDDING_DIRECTION(link) \
//     FRIBIDI_LEVEL_TO_DIR(link.level)

func fribidi_get_par_direction(bidi_types []CharType) ParType {
	valid_isolate_count := 0

	for i, bt := range bidi_types {
		if bt == FRIBIDI_TYPE_PDI {
			/* Ignore if there is no matching isolate */
			if valid_isolate_count > 0 {
				valid_isolate_count--
			}
		} else if FRIBIDI_IS_ISOLATE(bt) {
			valid_isolate_count++
		} else if valid_isolate_count == 0 && FRIBIDI_IS_LETTER(bt) {
			if FRIBIDI_IS_RTL(bt) {
				return FRIBIDI_PAR_RTL
			}
			return FRIBIDI_PAR_LTR
		}
	}

	return FRIBIDI_PAR_ON
}

/* Push a new entry to the pairing linked list */
func (nodes *PairingNode) pairing_nodes_push(Run *open, Run *close) *PairingNode {
	node := &PairingNode{}
	node.open = open
	node.close = close
	node.next = nodes
	return node
}

/* Sort by merge sort */
func (source *PairingNode) pairing_nodes_front_back_split( /* output */ PairingNode **front, PairingNode **back) {
	//   PairingNode *pfast, *pslow;
	if !source || !source.next {
		*front = source
		*back = NULL
	} else {
		pslow = source
		pfast = source.next
		for pfast {
			pfast = pfast.next
			if pfast {
				pfast = pfast.next
				pslow = pslow.next
			}
		}
		*front = source
		*back = pslow.next
		pslow.next = NULL
	}
}

func pairing_nodes_sorted_merge(nodes1, nodes2 *PairingNode) *PairingNode {
	PairingNode * res = NULL
	if !nodes1 {
		return nodes2
	}
	if !nodes2 {
		return nodes1
	}

	if nodes1.open.pos < nodes2.open.pos {
		res = nodes1
		res.next = pairing_nodes_sorted_merge(nodes1.next, nodes2)
	} else {
		res = nodes2
		res.next = pairing_nodes_sorted_merge(nodes1, nodes2.next)
	}
	return res
}

func sort_pairing_nodes(PairingNode **nodes) {
	//   PairingNode *front, *back;

	/* 0 or 1 node case */
	if !*nodes || !(*nodes).next {
		return
	}

	pairing_nodes_front_back_split(*nodes, &front, &back)
	sort_pairing_nodes(&front)
	sort_pairing_nodes(&back)
	*nodes = pairing_nodes_sorted_merge(front, back)
}

// func (nodes *PairingNode) free_pairing_nodes() {
//   for (nodes)  {
//       PairingNode *p = nodes;
//       nodes = nodes.next;
//       fribidi_free(p);
//     }
// }

// #define for_run_list(x, list) \
// 	for ((x) = (list).next; (x).type_ != maskSENTINEL; (x) = (x).next)

func fribidi_get_par_embedding_levels_ex(
	/* input */
	bidi_types []CharType,
	bracket_types []BracketType,
	len int,
	/* input and output */
	pbase_dir *ParType,
	/* output */
	embedding_levels *Level,
) Level {
	//   Level base_level, max_level = 0;
	//   ParType base_dir;
	//   Run *main_run_list = NULL, *explicits_list = NULL, *pp;
	//   bool status = false;
	//   int max_iso_level = 0;

	if !len {
		status = true
		goto out
	}

	/* Determinate character types */
	/* Get run-length encoded character types */
	main_run_list := run_list_encode_bidi_types(bidi_types, bracket_types, len)
	if !main_run_list {
		goto out
	}

	/* Find base level */
	/* If no strong base_dir was found, resort to the weak direction
	   that was passed on input. */
	base_level = FRIBIDI_DIR_TO_LEVEL(*pbase_dir)
	if !FRIBIDI_IS_STRONG(*pbase_dir) {
		/* P2. P3. Search for first strong character and use its direction as
		   base direction */
		valid_isolate_count := 0
		for pp = main_run_list.next; pp.type_ != maskSENTINEL; pp = pp.next {
			if pp.bracket_type == FRIBIDI_TYPE_PDI {
				/* Ignore if there is no matching isolate */
				if valid_isolate_count > 0 {
					valid_isolate_count--
				}
			} else if FRIBIDI_IS_ISOLATE(pp.bracket_type) {
				valid_isolate_count++
			} else if valid_isolate_count == 0 && FRIBIDI_IS_LETTER(pp.bracket_type) {
				base_level = FRIBIDI_DIR_TO_LEVEL(pp.bracket_type)
				*pbase_dir = FRIBIDI_LEVEL_TO_DIR(base_level)
				break
			}
		}
	}
	base_dir = FRIBIDI_LEVEL_TO_DIR(base_level)

	/* Explicit Levels and Directions */
	{
		// Level level, new_level = 0;
		// int isolate_level = 0;
		// CharType override, new_override;
		// StrIndex i;
		// int stack_size, over_pushed, first_interval;
		// int valid_isolate_count = 0;
		// int isolate_overflow = 0;
		// int isolate = 0; /* The isolate status flag */
		var status_stack stStack
		// Run temp_link;
		// Run *run_per_isolate_level[FRIBIDI_BIDI_MAX_RESOLVED_LEVELS];
		// int prev_isolate_level = 0; /* When running over the isolate levels, remember the previous level */

		memset(run_per_isolate_level, 0, sizeof(run_per_isolate_level[0])*FRIBIDI_BIDI_MAX_RESOLVED_LEVELS)

		/* explicits_list is a list like main_run_list, that holds the explicit
		   codes that are removed from main_run_list, to reinsert them later by
		   calling the shadow_run_list.
		*/
		explicits_list = new_run_list()
		if !explicits_list {
			goto out
		}

		/* X1. Begin by setting the current embedding level to the paragraph
		   embedding level. Set the directional override status to neutral,
		   and directional isolate status to false.

		   Process each character iteratively, applying rules X2 through X8.
		   Only embedding levels from 0 to 123 are valid in this phase. */

		level = base_level
		override = FRIBIDI_TYPE_ON
		/* stack */
		stack_size = 0
		over_pushed = 0
		first_interval = 0
		valid_isolate_count = 0
		isolate_overflow = 0

		for pp = main_run_list.next; pp.type_ != maskSENTINEL; pp = pp.next {
			this_type := pp.bracket_type
			RL_ISOLATE_LEVEL(pp) = isolate_level

			if FRIBIDI_IS_EXPLICIT_OR_BN(this_type) {
				if FRIBIDI_IS_STRONG(this_type) { /* LRE, RLE, LRO, RLO */
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
					new_override = FRIBIDI_EXPLICIT_TO_OVERRIDE_DIR(this_type)
					for i = pp.len; i; i-- {
						new_level =
							((level + FRIBIDI_DIR_TO_LEVEL(this_type) + 2) & ^1) -
								FRIBIDI_DIR_TO_LEVEL(this_type)
						isolate = 0
						PUSH_STATUS
					}
				} else if this_type == FRIBIDI_TYPE_PDF {
					/* 3. Terminating Embeddings and overrides */
					/*   X7. With each PDF, determine the matching embedding or
					     override code. */
					for i = pp.len; i; i-- {
						if stack_size && status_stack[stack_size-1].isolate != 0 {
							break
						}
						POP_STATUS
					}
				}

				/* X9. Remove all RLE, LRE, RLO, LRO, PDF, and BN codes. */
				/* Remove element and add it to explicits_list */
				pp.level = FRIBIDI_SENTINEL
				temp_link.next = pp.next
				move_node_before(pp, explicits_list)
				pp = &temp_link
			} else if this_type == FRIBIDI_TYPE_PDI {
				/* X6a. pop the direction of the stack */
				for i = pp.len; i; i-- {
					if isolate_overflow > 0 {
						isolate_overflow--
						pp.level = level
					} else if valid_isolate_count > 0 {
						/* Pop away all LRE,RLE,LRO, RLO levels
						   from the stack, as these are implicitly
						   terminated by the PDI */
						for stack_size && !status_stack[stack_size-1].isolate {
							POP_STATUS
						}
						over_pushed = 0 /* The PDI resets the overpushed! */
						POP_STATUS
						isolate_level--
						valid_isolate_count--
						pp.level = level
						RL_ISOLATE_LEVEL(pp) = isolate_level
					} else {
						/* Ignore isolated PDI's by turning them into ON's */
						pp.bracket_type = FRIBIDI_TYPE_ON
						pp.level = level
					}
				}
			} else if FRIBIDI_IS_ISOLATE(this_type) {
				/* TBD support RL_LEN > 1 */
				new_override = FRIBIDI_TYPE_ON
				isolate = 1
				if this_type == FRIBIDI_TYPE_LRI {
					new_level = level + 2 - (level % 2)
				} else if this_type == FRIBIDI_TYPE_RLI {
					new_level = level + 1 + (level % 2)
				} else if this_type == FRIBIDI_TYPE_FSI {
					/* Search for a local strong character until we
					   meet the corresponding PDI or the end of the
					   paragraph */
					//   Run *fsi_pp;
					//   int isolate_count = 0;
					//   int fsi_base_level = 0;
					for fsi_pp = pp.next; fsi_pp.type_ != maskSENTINEL; fsi_pp = fsi_pp.next {
						if fsi_pp.bracket_type == FRIBIDI_TYPE_PDI {
							isolate_count--
							if valid_isolate_count < 0 {
								break
							}
						} else if FRIBIDI_IS_ISOLATE(fsi_pp.bracket_type) {
							isolate_count++
						} else if isolate_count == 0 && FRIBIDI_IS_LETTER(fsi_pp.bracket_type) {
							fsi_base_level = FRIBIDI_DIR_TO_LEVEL(fsi_pp.bracket_type)
							break
						}
					}

					/* Same behavior like RLI and LRI above */
					if FRIBIDI_LEVEL_IS_RTL(fsi_base_level) {
						new_level = level + 1 + (level % 2)
					} else {
						new_level = level + 2 - (level % 2)
					}
				}

				pp.level = level
				RL_ISOLATE_LEVEL(pp) = isolate_level
				if isolate_level < FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL-1 {
					isolate_level++
				}

				if !override.IsNeutral() {
					pp.bracket_type = override
				}

				if new_level <= FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL {
					valid_isolate_count++
					PUSH_STATUS
					level = new_level
				} else {
					isolate_overflow += 1
				}
			} else if this_type == FRIBIDI_TYPE_BS {
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
					pp.bracket_type = override
				}
			}
		}

		/* Build the isolate_level connections */
		prev_isolate_level = 0
		for pp = main_run_list.next; pp.type_ != maskSENTINEL; pp = pp.next {
			isolate_level := RL_ISOLATE_LEVEL(pp)
			//   int i;

			/* When going from an upper to a lower level, zero out all higher levels
			   in order not erroneous connections! */
			if isolate_level < prev_isolate_level {
				for i = isolate_level + 1; i <= prev_isolate_level; i++ {
					run_per_isolate_level[i] = 0
				}
			}
			prev_isolate_level = isolate_level

			if run_per_isolate_level[isolate_level] {
				run_per_isolate_level[isolate_level].next_isolate = pp
				pp.prev_isolate = run_per_isolate_level[isolate_level]
			}
			run_per_isolate_level[isolate_level] = pp
		}

		/* Implementing X8. It has no effect on a single paragraph! */
		level = base_level
		override = FRIBIDI_TYPE_ON
		stack_size = 0
		over_pushed = 0
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

	compact_list(main_run_list)

	/* 4. Resolving weak types. Also calculate the maximum isolate level */
	max_iso_level = 0
	{
		// int last_strong_stack[FRIBIDI_BIDI_MAX_RESOLVED_LEVELS];
		// CharType prev_type_orig;
		// bool w4;

		last_strong_stack[0] = base_dir

		for pp = main_run_list.next; pp.type_ != maskSENTINEL; pp = pp.next {
			//   register CharType prev_type, this_type, next_type;
			//   Run *ppp_prev, *ppp_next;
			//   int iso_level;

			ppp_prev = get_adjacent_run(pp, false, false)
			ppp_next = get_adjacent_run(pp, true, false)

			this_type = pp.bracket_type
			iso_level = RL_ISOLATE_LEVEL(pp)

			if iso_level > max_iso_level {
				max_iso_level = iso_level
			}

			if ppp_prev.level == pp.level {
				prev_type = ppp_prev.bracket_type
			} else {
				prev_type = FRIBIDI_LEVEL_TO_DIR(MAX(ppp_prev.level, pp.level))
			}

			if ppp_next.level == pp.level {
				next_type = ppp_next.bracket_type
			} else {
				next_type = FRIBIDI_LEVEL_TO_DIR(MAX(ppp_next.level, pp.level))
			}

			if FRIBIDI_IS_STRONG(prev_type) {
				last_strong_stack[iso_level] = prev_type
			}

			/* W1. NSM
			   Examine each non-spacing mark (NSM) in the level run, and change the
			   type of the NSM to the type of the previous character. If the NSM
			   is at the start of the level run, it will get the type of sor. */
			/* Implementation note: it is important that if the previous character
			   is not sor, then we should merge this run with the previous,
			   because of rules like W5, that we assume all of a sequence of
			   adjacent ETs are in one Run. */
			if this_type == FRIBIDI_TYPE_NSM {
				/* New rule in Unicode 6.3 */
				if FRIBIDI_IS_ISOLATE(pp.prev.bracket_type) {
					pp.bracket_type = FRIBIDI_TYPE_ON
				}

				if ppp_prev.level == pp.level {
					if ppp_prev == pp.prev {
						pp = merge_with_prev(pp)
					}
				} else {
					pp.bracket_type = prev_type
				}

				if prev_type == next_type && pp.level == pp.next.level {
					if ppp_next == pp.next {
						pp = merge_with_prev(pp.next)
					}
				}
				continue /* As we know the next condition cannot be true. */
			}

			/* W2: European numbers. */
			if this_type == FRIBIDI_TYPE_EN && last_strong_stack[iso_level] == FRIBIDI_TYPE_AL {
				pp.bracket_type = FRIBIDI_TYPE_AN

				/* Resolving dependency of loops for rules W1 and W2, so we
				   can merge them in one loop. */
				if next_type == FRIBIDI_TYPE_NSM {
					ppp_next.bracket_type = FRIBIDI_TYPE_AN
				}
			}
		}

		/* The last iso level is used to invalidate the the last strong values when going from
		   a higher to a lower iso level. When this occur, all "last_strong" values are
		   set to the base_dir. */
		last_strong_stack[0] = base_dir

		/* Resolving dependency of loops for rules W4 and W5, W5 may
		   want to prevent W4 to take effect in the next turn, do this
		   through "w4". */
		w4 = true
		/* Resolving dependency of loops for rules W4 and W5 with W7,
		   W7 may change an EN to L but it sets the prev_type_orig if needed,
		   so W4 and W5 in next turn can still do their works. */
		prev_type_orig = FRIBIDI_TYPE_ON

		/* Each isolate level has its own memory of the last strong character */
		for pp = main_run_list.next; pp.type_ != maskSENTINEL; pp = pp.next {
			//   register CharType prev_type, this_type, next_type;
			//   int iso_level;
			//   Run *ppp_prev, *ppp_next;

			this_type = pp.bracket_type
			iso_level = RL_ISOLATE_LEVEL(pp)

			ppp_prev = get_adjacent_run(pp, false, false)
			ppp_next = get_adjacent_run(pp, true, false)

			if ppp_prev.level == pp.level {
				prev_type = ppp_prev.bracket_type
			} else {
				prev_type = FRIBIDI_LEVEL_TO_DIR(MAX(ppp_prev.level, pp.level))
			}

			if ppp_next.level == pp.level {
				next_type = ppp_next.bracket_type
			} else {
				next_type = FRIBIDI_LEVEL_TO_DIR(MAX(ppp_next.level, pp.level))
			}

			if FRIBIDI_IS_STRONG(prev_type) {
				last_strong_stack[iso_level] = prev_type
			}

			/* W2 ??? */

			/* W3: Change ALs to R. */
			if this_type == FRIBIDI_TYPE_AL {
				pp.bracket_type = FRIBIDI_TYPE_RTL
				w4 = true
				prev_type_orig = FRIBIDI_TYPE_ON
				continue
			}

			/* W4. A single european separator changes to a european number.
			   A single common separator between two numbers of the same type
			   changes to that type. */
			if w4 && pp.len == 1 && FRIBIDI_IS_ES_OR_CS(this_type) &&
				FRIBIDI_IS_NUMBER(prev_type_orig) && prev_type_orig == next_type &&
				(prev_type_orig == FRIBIDI_TYPE_EN || this_type == FRIBIDI_TYPE_CS) {
				pp.bracket_type = prev_type
				this_type = pp.bracket_type
			}
			w4 = true

			/* W5. A sequence of European terminators adjacent to European
			   numbers changes to All European numbers. */
			if this_type == FRIBIDI_TYPE_ET && (prev_type_orig == FRIBIDI_TYPE_EN || next_type == FRIBIDI_TYPE_EN) {
				pp.bracket_type = FRIBIDI_TYPE_EN
				w4 = false
				this_type = pp.bracket_type
			}

			/* W6. Otherwise change separators and terminators to other neutral. */
			if FRIBIDI_IS_NUMBER_SEPARATOR_OR_TERMINATOR(this_type) {
				pp.bracket_type = FRIBIDI_TYPE_ON
			}

			/* W7. Change european numbers to L. */
			if this_type == FRIBIDI_TYPE_EN && last_strong_stack[iso_level] == FRIBIDI_TYPE_LTR {
				pp.bracket_type = FRIBIDI_TYPE_LTR

				prev_type_orig = FRIBIDI_TYPE_ON
				if pp.level == pp.next.level {
					prev_type_orig = FRIBIDI_TYPE_EN
				}
			} else {
				prev_type_orig = PREV_TYPE_OR_SOR(pp.next)
			}
		}
	}

	compact_neutrals(main_run_list)

	/* 5. Resolving Neutral Types */
	{
		/*  BD16 - Build list of all pairs*/
		// int num_iso_levels = max_iso_level + 1;
		// PairingNode *pairing_nodes = NULL;
		// Run *local_bracket_stack[FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL][LOCAL_BRACKET_SIZE];
		// Run **bracket_stack[FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL];
		// int bracket_stack_size[FRIBIDI_BIDI_MAX_EXPLICIT_LEVEL];
		// int last_level = main_run_list.level;
		// int last_iso_level = 0;

		memset(bracket_stack, 0, sizeof(bracket_stack[0])*num_iso_levels)
		memset(bracket_stack_size, 0, sizeof(bracket_stack_size[0])*num_iso_levels)

		/* populate the bracket_size. The first LOCAL_BRACKET_SIZE entries
		   of the stack are one the stack. Allocate the rest of the entries.
		*/
		{
			var iso_level int
			for iso_level = 0; iso_level < LOCAL_BRACKET_SIZE; iso_level++ {
				bracket_stack[iso_level] = local_bracket_stack[iso_level]
			}

			for iso_level = LOCAL_BRACKET_SIZE; iso_level < num_iso_levels; iso_level++ {
				bracket_stack[iso_level] = fribidi_malloc(sizeof(bracket_stack[0]) * FRIBIDI_BIDI_MAX_NESTED_BRACKET_PAIRS)
			}
		}

		/* Build the bd16 pair stack. */
		for pp = main_run_list.next; pp.type_ != maskSENTINEL; pp = pp.next {
			// int level = pp.level;
			// int iso_level = RL_ISOLATE_LEVEL(pp);
			// BracketType brack_prop = pp.bracket_type;

			/* Interpret the isolating run sequence as such that they
			   end at a change in the level, unless the iso_level has been
			   raised. */
			if level != last_level && last_iso_level == iso_level {
				bracket_stack_size[last_iso_level] = 0
			}

			if brack_prop != NoBracket && pp.bracket_type == FRIBIDI_TYPE_ON {
				if FRIBIDI_IS_BRACKET_OPEN(brack_prop) {
					if bracket_stack_size[iso_level] == FRIBIDI_BIDI_MAX_NESTED_BRACKET_PAIRS {
						break
					}

					/* push onto the pair stack */
					bracket_stack[iso_level][bracket_stack_size[iso_level]] = pp
					bracket_stack_size[iso_level]++
				} else {
					stack_idx := bracket_stack_size[iso_level] - 1
					for stack_idx >= 0 {
						se_brack_prop := bracket_stack[iso_level][stack_idx].bracket_type
						if FRIBIDI_BRACKET_ID(se_brack_prop) == FRIBIDI_BRACKET_ID(brack_prop) {
							bracket_stack_size[iso_level] = stack_idx

							pairing_nodes = pairing_nodes_push(pairing_nodes,
								bracket_stack[iso_level][stack_idx],
								pp)
							break
						}
						stack_idx--
					}
				}
			}
			last_level = level
			last_iso_level = iso_level
		}

		/* The list must now be sorted for the next algo to work! */
		sort_pairing_nodes(&pairing_nodes)

		/* Start the N0 */
		{
			ppairs := pairing_nodes
			for ppairs {
				embedding_level := ppairs.open.level

				/* Find matching strong. */
				found := false
				var ppn *Run
				for ppn = ppairs.open; ppn != ppairs.close; ppn = ppn.next {
					this_type := RL_TYPE_AN_EN_AS_RTL(ppn)

					/* Calculate level like in resolve implicit levels below to prevent
					   embedded levels not to match the base_level */
					this_level := ppn.level + (FRIBIDI_LEVEL_IS_RTL(ppn.level) ^ FRIBIDI_DIR_TO_LEVEL(this_type))

					/* N0b */
					if FRIBIDI_IS_STRONG(this_type) && this_level == embedding_level {
						l := FRIBIDI_TYPE_LTR
						if this_level % 2 {
							l = FRIBIDI_TYPE_RTL
						}
						ppairs.close.bracket_type = l
						ppairs.open.bracket_type = l
						found = true
						break
					}
				}

				/* N0c */
				/* Search for any strong type preceding and within the bracket pair */
				if !found {
					/* Search for a preceding strong */
					prec_strong_level := embedding_level /* TBDov! Extract from Isolate level in effect */
					iso_level := RL_ISOLATE_LEVEL(ppairs.open)
					for ppn = ppairs.open.prev; ppn.type_ != FRIBIDI_TYPE_SENTINEL; ppn = ppn.prev {
						this_type := RL_TYPE_AN_EN_AS_RTL(ppn)
						if FRIBIDI_IS_STRONG(this_type) && RL_ISOLATE_LEVEL(ppn) == iso_level {
							prec_strong_level = ppn.level +
								(FRIBIDI_LEVEL_IS_RTL(ppn.level) ^ FRIBIDI_DIR_TO_LEVEL(this_type))

							break
						}
					}

					for ppn = ppairs.open; ppn != ppairs.close; ppn = ppn.next {
						this_type := RL_TYPE_AN_EN_AS_RTL(ppn)
						if FRIBIDI_IS_STRONG(this_type) && RL_ISOLATE_LEVEL(ppn) == iso_level {
							/* By constraint this is opposite the embedding direction,
							   since we did not match the N0b rule. We must now
							   compare with the preceding strong to establish whether
							   to apply N0c1 (opposite) or N0c2 embedding */
							l := FRIBIDI_TYPE_LTR
							if prec_strong_level % 2 {
								l = FRIBIDI_TYPE_RTL
							}
							ppairs.open.bracket_type = l
							ppairs.close.bracket_type = l
							found = true
							break
						}
					}
				}

				ppairs = ppairs.next
			}

			free_pairing_nodes(pairing_nodes)

			/* Remove the bracket property and re-compact */
			{
				for pp = main_run_list.next; pp.type_ != maskSENTINEL; pp = pp.next {
					pp.bracket_type = NoBracket
				}
				compact_neutrals(main_run_list)
			}
		}
	}

	// resolving neutral types - N1+N2
	{
		for pp = main_run_list.next; pp.type_ != maskSENTINEL; pp = pp.next {

			//   CharType prev_type, this_type, next_type;
			//   Run *ppp_prev, *ppp_next;

			ppp_prev := get_adjacent_run(pp, false, false)
			ppp_next := get_adjacent_run(pp, true, false)

			/* "European and Arabic numbers are treated as though they were R"
			   FRIBIDI_CHANGE_NUMBER_TO_RTL does this. */
			this_type := FRIBIDI_CHANGE_NUMBER_TO_RTL(pp.bracket_type)

			if ppp_prev.level == pp.level {
				prev_type = FRIBIDI_CHANGE_NUMBER_TO_RTL(ppp_prev.bracket_type)
			} else {
				prev_type = FRIBIDI_LEVEL_TO_DIR(MAX(ppp_prev.level, pp.level))
			}

			if ppp_next.level == pp.level {
				next_type = FRIBIDI_CHANGE_NUMBER_TO_RTL(ppp_next.bracket_type)
			} else {
				next_type = FRIBIDI_LEVEL_TO_DIR(MAX(ppp_next.level, pp.level))
			}

			if this_type.IsNeutral() {
				if prev_type == next_type {
					pp.bracket_type = prev_type // N1
				} else {
					pp.bracket_type = FRIBIDI_EMBEDDING_DIRECTION(pp) // N2
				}

			}
		}
	}
	compact_list(main_run_list)

	/* 6. Resolving implicit levels */
	{
		max_level = base_level

		for pp = main_run_list.next; pp.type_ != maskSENTINEL; pp = pp.next {
			this_type := pp.bracket_type
			level := pp.level

			/* I1. Even */
			/* I2. Odd */
			if FRIBIDI_IS_NUMBER(this_type) {
				pp.level = (level + 2) & ^1
			} else {
				pp.level = level + (FRIBIDI_LEVEL_IS_RTL(level) ^ FRIBIDI_DIR_TO_LEVEL(this_type))
			}

			if pp.level > max_level {
				max_level = pp.level
			}
		}
	}

	compact_list(main_run_list)

	/* Reinsert the explicit codes & BN's that are already removed, from the
	   explicits_list to main_run_list. */
	if explicits_list.next != explicits_list {
		//   register Run *p;
		stat := shadow_run_list(main_run_list, explicits_list, true)
		explicits_list = NULL
		if !stat {
			goto out
		}

		/* Set level of inserted explicit chars to that of their previous
		 * char, such that they do not affect reordering. */
		p = main_run_list.next
		if p != main_run_list && p.level == FRIBIDI_SENTINEL {
			p.level = base_level
		}
		for p = main_run_list.next; p.type_ != maskSENTINEL; p = p.next {
			if p.level == FRIBIDI_SENTINEL {
				p.level = p.prev.level
			}
		}
	}

	{
		// register int j, state, pos;
		// register CharType char_type;
		// register Run *p, *q, *list;

		/* L1. Reset the embedding levels of some chars:
		   1. segment separators,
		   2. paragraph separators,
		   3. any sequence of whitespace characters preceding a segment
		      separator or paragraph separator, and
		   4. any sequence of whitespace characters and/or isolate formatting
		      characters at the end of the line.
		   ... (to be continued in fribidi_reorder_line()). */
		list = new_run_list()
		q = list
		state = 1
		pos = len - 1
		for j = len - 1; j >= -1; j-- {
			/* close up the open link at the end */
			if j >= 0 {
				char_type = bidi_types[j]
			} else {
				char_type = FRIBIDI_TYPE_ON
			}
			if !state && FRIBIDI_IS_SEPARATOR(char_type) {
				state = 1
				pos = j
			} else if state && !(FRIBIDI_IS_EXPLICIT_OR_SEPARATOR_OR_BN_OR_WS(char_type) ||
				FRIBIDI_IS_ISOLATE(char_type)) {
				state = 0
				p = new_run()
				p.pos = j + 1
				p.len = pos - j
				p.type_ = base_dir
				p.level = base_level
				move_node_before(p, q)
				q = p
			}
		}
		if !shadow_run_list(main_run_list, list, false) {
			goto out
		}
	}

	{
		pos := 0
		for pp = main_run_list.next; pp.type_ != maskSENTINEL; pp = pp.next {
			level := pp.level
			for l := pp.len; l; l-- {
				embedding_levels[pos] = level
				pos++
			}
		}
	}

	status = true

out:

	if status {
		return max_level + 1
	}
	return 0
}

// static void
// bidi_string_reverse (
//   Char *str,
//   const StrIndex len
// )
// {
//   StrIndex i;

//   fribidi_assert (str);

//   for (i = 0; i < len / 2; i++)
//     {
//       Char tmp = str[i];
//       str[i] = str[len - 1 - i];
//       str[len - 1 - i] = tmp;
//     }
// }

// static void
// index_array_reverse (
//   StrIndex *arr,
//   const StrIndex len
// )
// {
//   StrIndex i;

//   fribidi_assert (arr);

//   for (i = 0; i < len / 2; i++)
//     {
//       StrIndex tmp = arr[i];
//       arr[i] = arr[len - 1 - i];
//       arr[len - 1 - i] = tmp;
//     }
// }

// FRIBIDI_ENTRY Level
// fribidi_reorder_line (
//   /* input */
//   Flags flags, /* reorder flags */
//   const CharType *bidi_types,
//   const StrIndex len,
//   const StrIndex off,
//   const ParType base_dir,
//   /* input and output */
//   Level *embedding_levels,
//   Char *visual_str,
//   /* output */
//   StrIndex *map
// )
// {
//   bool status = false;
//   Level max_level = 0;

//   if
//     (len == 0)
//     {
//       status = true;
//       goto out;
//     }

//   DBG ("in fribidi_reorder_line");

//   fribidi_assert (bidi_types);
//   fribidi_assert (embedding_levels);

//   DBG ("reset the embedding levels, 4. whitespace at the end of line");
//   {
//     register StrIndex i;

//     /* L1. Reset the embedding levels of some chars:
//        4. any sequence of white space characters at the end of the line. */
//     for (i = off + len - 1; i >= off &&
// 	 FRIBIDI_IS_EXPLICIT_OR_BN_OR_WS (bidi_types[i]); i--)
//       embedding_levels[i] = FRIBIDI_DIR_TO_LEVEL (base_dir);
//   }

//   /* 7. Reordering resolved levels */
//   {
//     register Level level;
//     register StrIndex i;

//     /* Reorder both the outstring and the order array */
//     {
//       if (FRIBIDI_TEST_BITS (flags, FRIBIDI_FLAG_REORDER_NSM))
// 	{
// 	  /* L3. Reorder NSMs. */
// 	  for (i = off + len - 1; i >= off; i--)
// 	    if (FRIBIDI_LEVEL_IS_RTL (embedding_levels[i])
// 		&& bidi_types[i] == FRIBIDI_TYPE_NSM)
// 	      {
// 		register StrIndex seq_end = i;
// 		level = embedding_levels[i];

// 		for (i--; i >= off &&
// 		     FRIBIDI_IS_EXPLICIT_OR_BN_OR_NSM (bidi_types[i])
// 		     && embedding_levels[i] == level; i--)
// 		  ;

// 		if (i < off || embedding_levels[i] != level)
// 		  {
// 		    i++;
// 		    DBG ("warning: NSM(s) at the beginning of level run");
// 		  }

// 		if (visual_str)
// 		  {
// 		    bidi_string_reverse (visual_str + i, seq_end - i + 1);
// 		  }
// 		if (map)
// 		  {
// 		    index_array_reverse (map + i, seq_end - i + 1);
// 		  }
// 	      }
// 	}

//       /* Find max_level of the line.  We don't reuse the paragraph
//        * max_level, both for a cleaner API, and that the line max_level
//        * may be far less than paragraph max_level. */
//       for (i = off + len - 1; i >= off; i--)
// 	if (embedding_levels[i] > max_level)
// 	  max_level = embedding_levels[i];

//       /* L2. Reorder. */
//       for (level = max_level; level > 0; level--)
// 	for (i = off + len - 1; i >= off; i--)
// 	  if (embedding_levels[i] >= level)
// 	    {
// 	      /* Find all stretches that are >= level_idx */
// 	      register StrIndex seq_end = i;
// 	      for (i--; i >= off && embedding_levels[i] >= level; i--)
// 		;

// 	      if (visual_str)
// 		bidi_string_reverse (visual_str + i + 1, seq_end - i);
// 	      if (map)
// 		index_array_reverse (map + i + 1, seq_end - i);
// 	    }
//     }

//   }

//   status = true;

// out:

//   return status ? max_level + 1 : 0;
// }
