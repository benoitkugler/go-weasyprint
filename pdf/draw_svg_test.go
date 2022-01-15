package pdf

import (
	"fmt"
	"testing"

	"github.com/benoitkugler/webrender/utils/testutils"
)

func TestUse(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "use", `
        RRRRR_____
        RRRRR_____
        __________
        ___RRRRR__
        ___RRRRR__
        __________
        _____RRRRR
        _____RRRRR
        __________
        __________
    `, `
      <style>
        @page { size: 10px }
        svg { display: block }
      </style>
      <svg width="10px" height="10px" xmlns="http://www.w3.org/2000/svg"
           xlink="http://www.w3.org/1999/xlink">
        <defs>
          <rect id="rectangle" width="5" height="2" fill="red" />
        </defs>
        <use xlink:href="#rectangle" />
        <use xlink:href="#rectangle" x="3" y="3" />
        <use xlink:href="#rectangle" x="5" y="6" />
      </svg>
    `)
}

// Test how SVG simple patterns are drawn.

func TestPattern(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "pattern", `
        BBrrBBrr
        BBrrBBrr
        rrBBrrBB
        rrBBrrBB
        BBrrBBrr
        BBrrBBrr
        rrBBrrBB
        rrBBrrBB
    `, `
      <style>
        @page { size: 8px }
        svg { display: block }
      </style>
      <svg width="8px" height="8px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <pattern id="pat" x="0" y="0" width="4" height="4"
            patternUnits="userSpaceOnUse"
            patternContentUnits="userSpaceOnUse">
            <rect x="0" y="0" width="2" height="2" fill="blue" />
            <rect x="0" y="2" width="2" height="2" fill="red" />
            <rect x="2" y="0" width="2" height="2" fill="red" />
            <rect x="2" y="2" width="2" height="2" fill="blue" />
          </pattern>
        </defs>
        <rect x="0" y="0" width="8" height="8" fill="url(#pat)" />
      </svg>
    `)
}

func TestPattern_2(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "pattern_2", `
        BBrrBBrr
        BBrrBBrr
        rrBBrrBB
        rrBBrrBB
        BBrrBBrr
        BBrrBBrr
        rrBBrrBB
        rrBBrrBB
    `, `
      <style>
        @page { size: 8px }
        svg { display: block }
      </style>
      <svg width="8px" height="8px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <pattern id="pat" x="0" y="0" width="50%" height="50%"
            patternUnits="objectBoundingBox"
            patternContentUnits="userSpaceOnUse">
            <rect x="0" y="0" width="2" height="2" fill="blue" />
            <rect x="0" y="2" width="2" height="2" fill="red" />
            <rect x="2" y="0" width="2" height="2" fill="red" />
            <rect x="2" y="2" width="2" height="2" fill="blue" />
          </pattern>
        </defs>
        <rect x="0" y="0" width="8" height="8" fill="url(#pat)" />
      </svg>
    `)
}

func TestPattern_3(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "pattern_3", `
        BBrrBBrr
        BBrrBBrr
        rrBBrrBB
        rrBBrrBB
        BBrrBBrr
        BBrrBBrr
        rrBBrrBB
        rrBBrrBB
    `, `
      <style>
        @page { size: 8px }
        svg { display: block }
      </style>
      <svg width="8px" height="8px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <pattern id="pat" x="0" y="0" width="4" height="4"
            patternUnits="userSpaceOnUse"
            patternContentUnits="userSpaceOnUse">
            <rect x="0" y="0" width="2" height="2" fill="blue" />
            <rect x="0" y="2" width="2" height="2" fill="red" />
            <rect x="2" y="0" width="2" height="2" fill="red" />
            <rect x="2" y="2" width="2" height="2" fill="blue" />
          </pattern>
        </defs>
        <rect x="0" y="0" width="8" height="8" fill="url(#pat)" />
      </svg>
    `)
}

func TestPattern_4(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "pattern_4", `
        BBrrBBrr
        BBrrBBrr
        rrBBrrBB
        rrBBrrBB
        BBrrBBrr
        BBrrBBrr
        rrBBrrBB
        rrBBrrBB
    `, `
      <style>
        @page { size: 8px }
        svg { display: block }
      </style>
      <svg width="8px" height="8px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <pattern id="pat" x="0" y="0" width="4" height="4"
            patternUnits="userSpaceOnUse"
            patternContentUnits="objectBoundingBox">
            <rect x="0" y="0" width="50%" height="50%" fill="blue" />
            <rect x="0" y="50%" width="50%" height="50%" fill="red" />
            <rect x="50%" y="0" width="50%" height="50%" fill="red" />
            <rect x="50%" y="50%" width="50%" height="50%" fill="blue" />
          </pattern>
        </defs>
        <rect x="0" y="0" width="8" height="8" fill="url(#pat)" />
      </svg>
    `)
}

// Test how opacity is handled for SVG.

// TODO: xfail tests fail because of GhostScript and are supposed to work with
// real PDF files.

const svgOpacitySource = `
  <style>
    @page { size: 9px }
    svg { display: block }
  </style>
  <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">%s</svg>`

func TestOpacity(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertSameRendering(t, "opacity", fmt.Sprintf(svgOpacitySource, `
            <rect x="2" y="2" width="5" height="5" stroke-width="2"
                  stroke="rgb(127, 255, 127)" fill="rgb(127, 127, 255)" />
        `), fmt.Sprintf(svgOpacitySource, `
            <rect x="2" y="2" width="5" height="5" stroke-width="2"
                  stroke="lime" fill="blue" opacity="0.5" />
        `), 0)
}

func TestFillOpacity(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertSameRendering(t, "fill_opacity", fmt.Sprintf(svgOpacitySource, `
            <rect x="2" y="2" width="5" height="5"
                  fill="blue" opacity="0.5" />
            <rect x="2" y="2" width="5" height="5" stroke-width="2"
                  stroke="lime" fill="transparent" />
        `), fmt.Sprintf(svgOpacitySource, `
            <rect x="2" y="2" width="5" height="5" stroke-width="2"
                  stroke="lime" fill="blue" fill-opacity="0.5" />
        `), 0)
}

// @pytest.mark.xfail

// func TestStrokeOpacity(t *testing.T) {
// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     assertSameRendering(t,  (
//         ("stroke_opacity_reference", svgOpacitySource % `
//             <rect x="2" y="2" width="5" height="5"
//                   fill="blue" />
//             <rect x="2" y="2" width="5" height="5" stroke-width="2"
//                   stroke="lime" fill="transparent" opacity="0.5" />
//         `),
//         ("stroke_opacity", svgOpacitySource % `
//             <rect x="2" y="2" width="5" height="5" stroke-width="2"
//                   stroke="lime" fill="blue" stroke-opacity="0.5" />
//         `),
//     ))

// @pytest.mark.xfail

// func TestStrokeFillOpacity(t *testing.T) {
// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     assertSameRendering(t,  (
//         ("stroke_fill_opacity_reference", svgOpacitySource % `
//             <rect x="2" y="2" width="5" height="5"
//                   fill="blue" opacity="0.5" />
//             <rect x="2" y="2" width="5" height="5" stroke-width="2"
//                   stroke="lime" fill="transparent" opacity="0.5" />
//         `),
//         ("stroke_fill_opacity", svgOpacitySource % `
//             <rect x="2" y="2" width="5" height="5" stroke-width="2"
//                   stroke="lime" fill="blue"
//                   stroke-opacity="0.5" fill-opacity="0.5" />
//         `),
//     ))

// @pytest.mark.xfail

// func TestPatternGradientStrokeFillOpacity(t *testing.T) {
// capt := testutils.CaptureLogs()
// defer capt.AssertNoLogs(t)

//     assertSameRendering(t,  (
//         ("pattern_gradient_stroke_fill_opacity_reference", svgOpacitySource % `
//             <defs>
//               <linearGradient id="grad" x1="0" y1="0" x2="0" y2="1"
//                               gradientUnits="objectBoundingBox">
//                 <stop stop-color="black" offset="42.86%"></stop>
//                 <stop stop-color="green" offset="42.86%"></stop>
//               </linearGradient>
//               <pattern id="pat" x="0" y="0" width="2" height="2"
//                        patternUnits="userSpaceOnUse"
//                        patternContentUnits="userSpaceOnUse">
//                 <rect x="0" y="0" width="1" height="1" fill="blue" />
//                 <rect x="0" y="1" width="1" height="1" fill="red" />
//                 <rect x="1" y="0" width="1" height="1" fill="red" />
//                 <rect x="1" y="1" width="1" height="1" fill="blue" />
//               </pattern>
//             </defs>
//             <rect x="2" y="2" width="5" height="5"
//                   fill="url(#pat)" opacity="0.5" />
//             <rect x="2" y="2" width="5" height="5" stroke-width="2"
//                   stroke="url(#grad)" fill="transparent" opacity="0.5" />
//         `),
//         ("pattern_gradient_stroke_fill_opacity", svgOpacitySource % `
//             <defs>
//               <linearGradient id="grad" x1="0" y1="0" x2="0" y2="1"
//                               gradientUnits="objectBoundingBox">
//                 <stop stop-color="black" offset="42.86%"></stop>
//                 <stop stop-color="green" offset="42.86%"></stop>
//               </linearGradient>
//               <pattern id="pat" x="0" y="0" width="2" height="2"
//                        patternUnits="userSpaceOnUse"
//                        patternContentUnits="userSpaceOnUse">
//                 <rect x="0" y="0" width="1" height="1" fill="blue" />
//                 <rect x="0" y="1" width="1" height="1" fill="red" />
//                 <rect x="1" y="0" width="1" height="1" fill="red" />
//                 <rect x="1" y="1" width="1" height="1" fill="blue" />
//               </pattern>
//             </defs>
//             <rect x="2" y="2" width="5" height="5" stroke-width="2"
//                   stroke="url(#grad)" fill="url(#pat)"
//                   stroke-opacity="0.5" fill-opacity="0.5" />
//         `),
//     ))

// Test how SVG simple shapes are drawn.

func TestRectStroke(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "rect_stroke", `
        _________
        _RRRRRRR_
        _RRRRRRR_
        _RR___RR_
        _RR___RR_
        _RR___RR_
        _RRRRRRR_
        _RRRRRRR_
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <rect x="2" y="2" width="5" height="5"
              stroke-width="2" stroke="red" fill="none" />
      </svg>
    `)
}

func TestRectFill(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "rect_fill", `
        _________
        _________
        __RRRRR__
        __RRRRR__
        __RRRRR__
        __RRRRR__
        __RRRRR__
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <rect x="2" y="2" width="5" height="5" fill="red" />
      </svg>
    `)
}

func TestRectStrokeFill(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "rect_stroke_fill", `
        _________
        _RRRRRRR_
        _RRRRRRR_
        _RRBBBRR_
        _RRBBBRR_
        _RRBBBRR_
        _RRRRRRR_
        _RRRRRRR_
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <rect x="2" y="2" width="5" height="5"
              stroke-width="2" stroke="red" fill="blue" />
      </svg>
    `)
}

func TestRectRound(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "rect_round", `
        _zzzzzzz_
        zzzzzzzzz
        zzRRRRRzz
        zzRRRRRzz
        zzRRRRRzz
        zzRRRRRzz
        zzRRRRRzz
        zzzzzzzzz
        _zzzzzzz_
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <rect width="9" height="9" fill="red" rx="4" ry="4" />
      </svg>
    `)
}

func TestRectRoundZero(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "rect_round_zero", `
        RRRRRRRRR
        RRRRRRRRR
        RRRRRRRRR
        RRRRRRRRR
        RRRRRRRRR
        RRRRRRRRR
        RRRRRRRRR
        RRRRRRRRR
        RRRRRRRRR
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <rect width="9" height="9" fill="red" rx="0" ry="4" />
      </svg>
    `)
}

func TestLine(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "line", `
        _________
        _________
        _________
        _________
        RRRRRR___
        RRRRRR___
        _________
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <line x1="0" y1="5" x2="6" y2="5"
          stroke="red" stroke-width="2"/>
      </svg>
    `)
}

func TestPolyline(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "polyline", `
        _________
        RRRRRR___
        RRRRRR___
        RR__RR___
        RR__RR___
        RR__RR___
        _________
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <polyline points="1,6, 1,2, 5,2, 5,6"
          stroke="red" stroke-width="2" fill="none"/>
      </svg>
    `)
}

func TestPolylineFill(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "polyline_fill", `
        _________
        RRRRRR___
        RRRRRR___
        RRBBRR___
        RRBBRR___
        RRBBRR___
        _________
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <polyline points="1,6, 1,2, 5,2, 5,6"
          stroke="red" stroke-width="2" fill="blue"/>
      </svg>
    `)
}

func TestPolygon(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "polygon", `
        _________
        RRRRRR___
        RRRRRR___
        RR__RR___
        RR__RR___
        RRRRRR___
        RRRRRR___
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <polygon points="1,6, 1,2, 5,2, 5,6"
          stroke="red" stroke-width="2" fill="none"/>
      </svg>
    `)
}

func TestPolygonFill(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "polygon_fill", `
        _________
        RRRRRR___
        RRRRRR___
        RRBBRR___
        RRBBRR___
        RRRRRR___
        RRRRRR___
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <polygon points="1,6, 1,2, 5,2, 5,6"
          stroke="red" stroke-width="2" fill="blue"/>
      </svg>
    `)
}

func TestCircleStroke(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "circle_stroke", `
        __________
        __RRRRRR__
        _RRRRRRRR_
        _RRRRRRRR_
        _RRR__RRR_
        _RRR__RRR_
        _RRRRRRRR_
        _RRRRRRRR_
        __RRRRRR__
        __________
    `, `
      <style>
        @page { size: 10px }
        svg { display: block }
      </style>
      <svg width="10px" height="10px" xmlns="http://www.w3.org/2000/svg">
        <circle cx="5" cy="5" r="3"
          stroke="red" stroke-width="2" fill="none"/>
      </svg>
    `)
}

func TestCircleFill(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "circle_fill", `
        __________
        __RRRRRR__
        _RRRRRRRR_
        _RRRRRRRR_
        _RRRBBRRR_
        _RRRBBRRR_
        _RRRRRRRR_
        _RRRRRRRR_
        __RRRRRR__
        __________
    `, `
      <style>
        @page { size: 10px }
        svg { display: block }
      </style>
      <svg width="10px" height="10px" xmlns="http://www.w3.org/2000/svg">
        <circle cx="5" cy="5" r="3"
          stroke="red" stroke-width="2" fill="blue"/>
      </svg>
    `)
}

func TestEllipseStroke(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "ellipse_stroke", `
        __________
        __RRRRRR__
        _RRRRRRRR_
        _RRRRRRRR_
        _RRR__RRR_
        _RRR__RRR_
        _RRRRRRRR_
        _RRRRRRRR_
        __RRRRRR__
        __________
    `, `
      <style>
        @page { size: 10px }
        svg { display: block }
      </style>
      <svg width="10px" height="10px" xmlns="http://www.w3.org/2000/svg">
        <ellipse cx="5" cy="5" rx="3" ry="3"
          stroke="red" stroke-width="2" fill="none"/>
      </svg>
    `)
}

func TestEllipseFill(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "ellipse_fill", `
        __________
        __RRRRRR__
        _RRRRRRRR_
        _RRRRRRRR_
        _RRRBBRRR_
        _RRRBBRRR_
        _RRRRRRRR_
        _RRRRRRRR_
        __RRRRRR__
        __________
    `, `
      <style>
        @page { size: 10px }
        svg { display: block }
      </style>
      <svg width="10px" height="10px" xmlns="http://www.w3.org/2000/svg">
        <ellipse cx="5" cy="5" rx="3" ry="3"
          stroke="red" stroke-width="2" fill="blue"/>
      </svg>
    `)
}

func TestRectInG(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "rect_in_g", `
        RRRRR____
        RRRRR____
        RRRRR____
        RRRRR____
        RRRRR____
        _________
        _________
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <g x="5" y="5">
          <rect width="5" height="5" fill="red" />
        </g>
      </svg>
    `)
}

func TestRectXYInG(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "rect_x_y_in_g", `
        _________
        _________
        __RRRRR__
        __RRRRR__
        __RRRRR__
        __RRRRR__
        __RRRRR__
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <g x="5" y="5">
          <rect x="2" y="2" width="5" height="5" fill="red" />
        </g>
      </svg>
    `)
}

func TestRectStrokeZero(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "rect_stroke_zero", `
        _________
        _________
        _________
        _________
        _________
        _________
        _________
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <rect x="2" y="2" width="5" height="5"
              stroke-width="0" stroke="red" fill="none" />
      </svg>
    `)
}

func TestRectWidthHeightZero(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "rect_fill", `
        _________
        _________
        _________
        _________
        _________
        _________
        _________
        _________
        _________
    `, `
        <style>
            @page { size: 9px }
            svg { display: block }
        </style>
        <svg width="0" height="0" xmlns="http://www.w3.org/2000/svg">
            <rect x="2" y="2" width="5" height="5" fill="red" />
        </svg>
    `)
}

// Test how the visibility is controlled with "visibility" and "display"
// attributes.

func TestVisibilityVisible(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "visibility_visible", `
        _________
        _________
        __RRRRR__
        __RRRRR__
        __RRRRR__
        __RRRRR__
        __RRRRR__
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <rect visibility="visible"
              x="2" y="2" width="5" height="5" fill="red" />
      </svg>
    `)
}

func TestVisibilityHidden(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "visibility_hidden", `
        _________
        _________
        _________
        _________
        _________
        _________
        _________
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <rect visibility="hidden"
              x="2" y="2" width="5" height="5" fill="red" />
      </svg>
    `)
}

func TestVisibilityInheritHidden(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "visibility_inherit_hidden", `
        _________
        _________
        _________
        _________
        _________
        _________
        _________
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <g visibility="hidden">
          <rect x="2" y="2" width="5" height="5" fill="red" />
        </g>
      </svg>
    `)
}

func TestVisibilityInheritVisible(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "visibility_inherit_visible", `
        _________
        _________
        __RRRRR__
        __RRRRR__
        __RRRRR__
        __RRRRR__
        __RRRRR__
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <g visibility="hidden">
          <rect visibility="visible"
                x="2" y="2" width="5" height="5" fill="red" />
        </g>
      </svg>
    `)
}

func TestDisplayInline(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "display_inline", `
        _________
        _________
        __RRRRR__
        __RRRRR__
        __RRRRR__
        __RRRRR__
        __RRRRR__
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <rect display="inline"
              x="2" y="2" width="5" height="5" fill="red" />
      </svg>
    `)
}

func TestDisplayNone(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "display_none", `
        _________
        _________
        _________
        _________
        _________
        _________
        _________
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <rect display="none"
              x="2" y="2" width="5" height="5" fill="red" />
      </svg>
    `)
}

func TestDisplayInheritNone(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "display_inherit_none", `
        _________
        _________
        _________
        _________
        _________
        _________
        _________
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <g display="none">
          <rect x="2" y="2" width="5" height="5" fill="red" />
        </g>
      </svg>
    `)
}

func TestDisplayInheritInline(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "display_inherit_inline", `
        _________
        _________
        _________
        _________
        _________
        _________
        _________
        _________
        _________
    `, `
      <style>
        @page { size: 9px }
        svg { display: block }
      </style>
      <svg width="9px" height="9px" xmlns="http://www.w3.org/2000/svg">
        <g display="none">
          <rect display="inline"
                x="2" y="2" width="5" height="5" fill="red" />
        </g>
      </svg>
    `)
}
