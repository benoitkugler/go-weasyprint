package fribidi

const FRIBIDI_SENTINEL = -1

type Run struct {
	prev *Run
	next *Run

	pos, len             int
	type_                CharType
	level, isolate_level Level
	bracket_type         BracketType

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

//  void
//  free_run_list (
//    Run *run_list
//  )
//  {
//    if (!run_list)
// 	 return;

//    validate (run_list);

//    {
// 	 register Run *pp;

// 	 pp = run_list;
// 	 pp.prev.next = NULL;
// 	 for LIKELY
// 	   (pp)
// 	   {
// 	 register Run *p;

// 	 p = pp;
// 	 pp = pp.next;
// 	 fribidi_free (p);
// 	   };
//    }
//  }

func run_list_encode_bidi_types(bidiTypes []CharType, bracketTypes []BracketType, len int) *Run {
	//    Run *list, *last;
	//    register Run *run = NULL;
	//    int i;

	//    fribidi_assert (bidiTypes);

	/* Create the list sentinel */
	list := new_run_list()
	last := list

	/* Scan over the character types */
	for i, char_type := range bidiTypes {
		bracket_type := NoBracket
		if bracketTypes != nil {
			bracket_type = bracketTypes[i]
		}

		if char_type != last.type_ || bracket_type != NoBracket || // Always separate bracket into single char runs!
			last.bracket_type != NoBracket || char_type.IsIsolate() {
			run := &Run{}
			run.type_ = char_type
			run.pos = i
			last.len = run.pos - last.pos
			last.next = run
			run.prev = last
			run.bracket_type = bracket_type
			last = run
		}
	}

	/* Close the circle */
	last.len = len - last.pos
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

func (run_list *Run) validate() {} // only used to debug TODO: include in test ?
