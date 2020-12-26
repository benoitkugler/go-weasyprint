package fribidi

import (
	"fmt"
	"log"
)

const FRIBIDI_SENTINEL = -1

type Run struct {
	prev *Run
	next *Run

	pos, len             int
	type_                CharType
	level, isolate_level Level
	bracketType          BracketType

	/* Additional links for connecting the isolate tree */
	prev_isolate, next_isolate *Run
}

func new_run_list() *Run {
	var run Run

	run.type_ = maskSENTINEL
	run.level = FRIBIDI_SENTINEL
	run.pos = FRIBIDI_SENTINEL
	run.len = FRIBIDI_SENTINEL
	run.next = &run
	run.prev = &run

	return &run
}

func (x *Run) delete_node() {
	x.prev.next = x.next
	x.next.prev = x.prev
}

func (list *Run) insert_node_before(x *Run) {
	x.prev = list.prev
	list.prev.next = x
	x.next = list
	list.prev = x
}

func (list *Run) move_node_before(x *Run) {
	if x.prev != nil {
		x.delete_node()
	}
	list.insert_node_before(x)
}

/* Return the type of previous run or the SOR, if already at the start of
   a level run. */
func (pp *Run) PREV_TYPE_OR_SOR() CharType {
	if pp.prev.level == pp.level {
		return pp.prev.type_
	}
	return FRIBIDI_LEVEL_TO_DIR(maxL(pp.prev.level, pp.level))
}

/* "Within this scope, bidirectional types EN and AN are treated as R" */
func (list *Run) RL_TYPE_AN_EN_AS_RTL() CharType {
	if list.type_ == AN || list.type_ == EN {
		return RTL
	}
	return list.type_
}

/* Return the embedding direction of a link. */
func (link *Run) FRIBIDI_EMBEDDING_DIRECTION() CharType {
	return FRIBIDI_LEVEL_TO_DIR(link.level)
}

// bracketTypes is either empty or with same length as `bidiTypes`
func run_list_encode_bidi_types(bidiTypes []CharType, bracketTypes []BracketType) *Run {
	/* Create the list sentinel */
	list := new_run_list()
	last := list
	hasBrackets := len(bracketTypes) != 0

	/* Scan over the character types */
	for i, charType := range bidiTypes {
		bracketType := NoBracket
		if hasBrackets {
			bracketType = bracketTypes[i]
		}
		fmt.Println(i, bracketType)

		if charType != last.type_ || bracketType != NoBracket || // Always separate bracket into single char runs!
			last.bracketType != NoBracket || charType.IsIsolate() {
			run := &Run{}
			run.type_ = charType
			run.pos = i
			last.len = run.pos - last.pos
			last.next = run
			run.prev = last
			run.bracketType = bracketType
			last = run
		}
	}

	/* Close the circle */
	last.len = len(bidiTypes) - last.pos
	last.next = list
	list.prev = last

	list.validate()

	return list
}

/* override the run list 'base', with the runs in the list 'over', to
reinsert the previously-removed explicit codes (at X9) from
'explicits_list' back into 'type_rl_list' for example. This is used at the
end of I2 to restore the explicit marks, and also to reset the character
types of characters at L1.

it is assumed that the 'pos' of the first element in 'base' list is not
more than the 'pos' of the first element of the 'over' list, and the
'pos' of the last element of the 'base' list is not less than the 'pos'
of the last element of the 'over' list. these two conditions are always
satisfied for the two usages mentioned above.

Note:
  frees the over list.

Todo:
  use some explanatory names instead of p, q, ...
  rewrite comment above to remove references to special usage.
*/
func shadow_run_list(base, over *Run, preserveLength bool) {
	var (
		r, t      *Run
		p         = base
		pos, pos2 int
	)

	base.validate()
	over.validate()
	//    for_run_list (q, over)
	for q := over.next; q.type_ != maskSENTINEL; q = q.next {
		if q.len == 0 || q.pos < pos {
			continue
		}
		pos = q.pos
		for p.next.type_ != maskSENTINEL && p.next.pos <= pos {
			p = p.next
		}
		/* now p is the element that q must be inserted 'in'. */
		pos2 = pos + q.len
		r = p
		for r.next.type_ != maskSENTINEL && r.next.pos < pos2 {
			r = r.next
		}
		if preserveLength {
			r.len += q.len
		}
		/* now r is the last element that q affects. */
		if p == r {
			/* split p into at most 3 intervals, and insert q in the place of
			the second interval, set r to be the third part. */
			/* third part needed? */
			if p.pos+p.len > pos2 {
				r = &Run{}
				p.next.prev = r
				r.next = p.next
				r.level = p.level
				r.isolate_level = p.isolate_level
				r.type_ = p.type_
				r.len = p.pos + p.len - pos2
				r.pos = pos2
			} else {
				r = r.next
			}

			if p.pos+p.len >= pos {
				/* first part needed? */
				if p.pos < pos {
					/* cut the end of p. */
					p.len = pos - p.pos
				} else {
					t = p
					p = p.prev
				}
			}
		} else {
			if p.pos+p.len >= pos {
				/* p needed? */
				if p.pos < pos {
					/* cut the end of p. */
					p.len = pos - p.pos
				} else {
					p = p.prev
				}
			}

			/* r needed? */
			if r.pos+r.len > pos2 {
				/* cut the beginning of r. */
				r.len = r.pos + r.len - pos2
				r.pos = pos2
			} else {
				r = r.next
			}

			/* remove the elements between p and r. */
			for s := p.next; s != r; {
				t = s
				s = s.next
			}
		}
		/* before updating the next and prev runs to point to the inserted q,
		we must remember the next element of q in the 'over' list.
		*/
		t = q
		q = q.prev
		t.delete_node()
		p.next = t
		t.prev = p
		t.next = r
		r.prev = t
	}

	base.validate()
}

func (second *Run) merge_with_prev() *Run {
	first := second.prev
	first.next = second.next
	first.next.prev = first
	first.len += second.len
	if second.next_isolate != nil {
		second.next_isolate.prev_isolate = second.prev_isolate
		/* The following edge case typically shouldn't happen, but fuzz
		   testing shows it does, and the assignment protects against
		   a dangling pointer. */
	} else if second.next.prev_isolate == second {
		second.next.prev_isolate = second.prev_isolate
	}
	if second.prev_isolate != nil {
		second.prev_isolate.next_isolate = second.next_isolate
	}
	first.next_isolate = second.next_isolate

	return first
}

func (list *Run) compact_list() {
	if list.next != nil {
		for list = list.next; list.type_ != maskSENTINEL; list = list.next {
			/* Don't join brackets! */
			if list.prev.type_ == list.type_ && list.prev.level == list.level &&
				list.bracketType == NoBracket && list.prev.bracketType == NoBracket {
				list = list.merge_with_prev()
			}
		}
	}
}

func (list *Run) compact_neutrals() {
	if list.next != nil {
		for list = list.next; list.type_ != maskSENTINEL; list = list.next {
			if list.prev.level == list.level &&
				(list.prev.type_ == list.type_ ||
					(list.prev.type_.IsNeutral() && list.type_.IsNeutral())) &&
				list.bracketType == NoBracket /* Don't join brackets! */ &&
				list.prev.bracketType == NoBracket {
				list = list.merge_with_prev()
			}
		}
	}
}

func assertT(b bool) {
	if !b {
		log.Fatal("assertion error")
	}
}

// only used to debug TODO: include in test ?
func (run_list *Run) validate() {
	assertT(run_list != nil)
	assertT(run_list.next != nil)
	assertT(run_list.next.prev == run_list)
	assertT(run_list.type_ == maskSENTINEL)
	q := run_list
	for ; q.type_ != maskSENTINEL; q = q.next {
		assertT(q.next != nil)
		assertT(q.next.prev == q)
	}
	assertT(q == run_list)
}

// debug printing helpers

func (r Run) print_types_re() {
	fmt.Print("  Run types  : ")
	for pp := r.next; pp.type_ != maskSENTINEL; pp = pp.next {
		fmt.Printf("%d:%d(%s)[%d,%d] ", pp.pos, pp.len, pp.type_, pp.level, pp.isolate_level)
	}
	fmt.Println()
}

func (r Run) print_resolved_types() {
	fmt.Print("  Res. types: ")
	for pp := r.next; pp.type_ != maskSENTINEL; pp = pp.next {
		for i := pp.len; i != 0; i-- {
			fmt.Printf("%s ", pp.type_)
		}
	}
	fmt.Println()
}

func (r Run) print_resolved_levels() {
	fmt.Print("  Res. levels: ")
	for pp := r.next; pp.type_ != maskSENTINEL; pp = pp.next {
		for i := pp.len; i != 0; i-- {
			fmt.Printf("%d ", pp.level)
		}
	}
	fmt.Println()
}
