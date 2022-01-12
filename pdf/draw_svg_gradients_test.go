package pdf

import (
	"testing"

	"github.com/benoitkugler/webrender/utils/testutils"
)

// Test how SVG simple gradients are drawn.

func TestLinearGradient(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient", `
        BBBBBBBBBB
        BBBBBBBBBB
        BBBBBBBBBB
        BBBBBBBBBB
        BBBBBBBBBB
        RRRRRRRRRR
        RRRRRRRRRR
        RRRRRRRRRR
        RRRRRRRRRR
        RRRRRRRRRR
    `, `
      <style>
        @page { size: 10px }
        svg { display: block }
      </style>
      <svg width="10px" height="10px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <linearGradient id="grad" x1="0" y1="0" x2="0" y2="1"
            gradientUnits="objectBoundingBox">
            <stop stop-color="blue" offset="50%"></stop>
            <stop stop-color="red" offset="50%"></stop>
          </linearGradient>
        </defs>
        <rect x="0" y="0" width="10" height="10" fill="url(#grad)" />
      </svg>
    `)
}

func TestLinearGradientUserspace(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient_userspace", `
        BBBBBBBBBB
        BBBBBBBBBB
        BBBBBBBBBB
        BBBBBBBBBB
        BBBBBBBBBB
        RRRRRRRRRR
        RRRRRRRRRR
        RRRRRRRRRR
        RRRRRRRRRR
        RRRRRRRRRR
    `, `
      <style>
        @page { size: 10px }
        svg { display: block }
      </style>
      <svg width="10px" height="10px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <linearGradient id="grad" x1="0" y1="0" x2="0" y2="10"
            gradientUnits="userSpaceOnUse">
            <stop stop-color="blue" offset="50%"></stop>
            <stop stop-color="red" offset="50%"></stop>
          </linearGradient>
        </defs>
        <rect x="0" y="0" width="10" height="10" fill="url(#grad)" />
      </svg>
    `)
}

func TestLinearGradientMulticolor(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient_multicolor", `
        BBBBBBBBBB
        BBBBBBBBBB
        RRRRRRRRRR
        RRRRRRRRRR
        GGGGGGGGGG
        GGGGGGGGGG
        vvvvvvvvvv
        vvvvvvvvvv
    `, `
      <style>
        @page { size: 10px 8px }
        svg { display: block }
      </style>
      <svg width="10px" height="8px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <linearGradient id="grad" x1="0" y1="0" x2="0" y2="1"
            gradientUnits="objectBoundingBox">
            <stop stop-color="blue" offset="25%"></stop>
            <stop stop-color="red" offset="25%"></stop>
            <stop stop-color="red" offset="50%"></stop>
            <stop stop-color="lime" offset="50%"></stop>
            <stop stop-color="lime" offset="75%"></stop>
            <stop stop-color="rgb(128,0,128)" offset="75%"></stop>
          </linearGradient>
        </defs>
        <rect x="0" y="0" width="10" height="8" fill="url(#grad)" />
      </svg>
    `)
}

func TestLinearGradientMulticolorUserspace(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient_multicolor_userspace", `
        BBBBBBBBBB
        BBBBBBBBBB
        RRRRRRRRRR
        RRRRRRRRRR
        GGGGGGGGGG
        GGGGGGGGGG
        vvvvvvvvvv
        vvvvvvvvvv
    `, `
      <style>
        @page { size: 10px 8px }
        svg { display: block }
      </style>
      <svg width="10px" height="8px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <linearGradient id="grad" x1="0" y1="0" x2="0" y2="8"
            gradientUnits="userSpaceOnUse">
            <stop stop-color="blue" offset="25%"></stop>
            <stop stop-color="red" offset="25%"></stop>
            <stop stop-color="red" offset="50%"></stop>
            <stop stop-color="lime" offset="50%"></stop>
            <stop stop-color="lime" offset="75%"></stop>
            <stop stop-color="rgb(128,0,128)" offset="75%"></stop>
          </linearGradient>
        </defs>
        <rect x="0" y="0" width="10" height="8" fill="url(#grad)" />
      </svg>
    `)
}

// TODO: pytest.mark.xfail
// func TestLinearGradientTransform(t *testing.T) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

// 	assertPixelsEqual(t, "linear_gradient_transform", `
//         BBBBBBBBBB
//         RRRRRRRRRR
//         GGGGGGGGGG
//         vvvvvvvvvv
//         vvvvvvvvvv
//         vvvvvvvvvv
//         vvvvvvvvvv
//         vvvvvvvvvv
//     `, `
//       <style>
//         @page { size: 10px 8px}
//         svg { display: block }
//       </style>
//       <svg width="10px" height="8px" xmlns="http://www.w3.org/2000/svg">
//         <defs>
//           <linearGradient id="grad" x1="0" y1="0" x2="0" y2="1"
//             gradientUnits="objectBoundingBox" gradientTransform="scale(0.5)">
//             <stop stop-color="blue" offset="25%"></stop>
//             <stop stop-color="red" offset="25%"></stop>
//             <stop stop-color="red" offset="50%"></stop>
//             <stop stop-color="lime" offset="50%"></stop>
//             <stop stop-color="lime" offset="75%"></stop>
//             <stop stop-color="rgb(128,0,128)" offset="75%"></stop>
//           </linearGradient>
//         </defs>
//         <rect x="0" y="0" width="10" height="8" fill="url(#grad)" />
//       </svg>
//     `)
// }

func TestLinearGradientRepeat(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient_repeat", `
        BBBBBBBBBB
        BBBBBBBBBB
        RRRRRRRRRR
        RRRRRRRRRR
        GGGGGGGGGG
        GGGGGGGGGG
        vvvvvvvvvv
        vvvvvvvvvv
        BBBBBBBBBB
        BBBBBBBBBB
        RRRRRRRRRR
        RRRRRRRRRR
        GGGGGGGGGG
        GGGGGGGGGG
        vvvvvvvvvv
        vvvvvvvvvv
    `, `
      <style>
        @page { size: 10px 16px }
        svg { display: block }
      </style>
      <svg width="11px" height="16px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <linearGradient id="grad" x1="0" y1="0" x2="0" y2="0.5"
            gradientUnits="objectBoundingBox" spreadMethod="repeat">
            <stop stop-color="blue" offset="25%"></stop>
            <stop stop-color="red" offset="25%"></stop>
            <stop stop-color="red" offset="50%"></stop>
            <stop stop-color="lime" offset="50%"></stop>
            <stop stop-color="lime" offset="75%"></stop>
            <stop stop-color="rgb(128,0,128)" offset="75%"></stop>
          </linearGradient>
        </defs>
        <rect x="0" y="0" width="11" height="16" fill="url(#grad)" />
      </svg>
    `)
}

// TODO: pytest.mark.xfail
// func TestLinearGradientRepeatLong(t *testing.T) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

// 	assertPixelsEqual(t, "linear_gradient_repeat_long", `
//         BBBBBBBBBB
//         RRRRRRRRRR
//         GGGGGGGGGG
//         vvvvvvvvvv
//         BBBBBBBBBB
//         RRRRRRRRRR
//         GGGGGGGGGG
//         vvvvvvvvvv
//         BBBBBBBBBB
//         RRRRRRRRRR
//         GGGGGGGGGG
//         vvvvvvvvvv
//         BBBBBBBBBB
//         RRRRRRRRRR
//         GGGGGGGGGG
//         vvvvvvvvvv
//     `, `
//       <style>
//         @page { size: 10px 16px }
//         svg { display: block }
//       </style>
//       <svg width="11px" height="16px" xmlns="http://www.w3.org/2000/svg">
//         <defs>
//           <linearGradient id="grad" x1="0" y1="0" x2="0" y2="0.25"
//             gradientUnits="objectBoundingBox" spreadMethod="repeat">
//             <stop stop-color="blue" offset="25%"></stop>
//             <stop stop-color="red" offset="25%"></stop>
//             <stop stop-color="red" offset="50%"></stop>
//             <stop stop-color="lime" offset="50%"></stop>
//             <stop stop-color="lime" offset="75%"></stop>
//             <stop stop-color="rgb(128,0,128)" offset="75%"></stop>
//           </linearGradient>
//         </defs>
//         <rect x="0" y="0" width="11" height="16" fill="url(#grad)" />
//       </svg>
//     `)
// }

func TestLinearGradientReflect(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient_reflect", `
        BBBBBBBBBB
        BBBBBBBBBB
        RRRRRRRRRR
        RRRRRRRRRR
        GGGGGGGGGG
        GGGGGGGGGG
        vvvvvvvvvv
        vvvvvvvvvv
        vvvvvvvvvv
        vvvvvvvvvv
        GGGGGGGGGG
        GGGGGGGGGG
        RRRRRRRRRR
        RRRRRRRRRR
        BBBBBBBBBB
        BBBBBBBBBB
    `, `
      <style>
        @page { size: 10px 16px }
        svg { display: block }
      </style>
      <svg width="11px" height="16px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <linearGradient id="grad" x1="0" y1="0" x2="0" y2="0.5"
            gradientUnits="objectBoundingBox" spreadMethod="reflect">
            <stop stop-color="blue" offset="25%"></stop>
            <stop stop-color="red" offset="25%"></stop>
            <stop stop-color="red" offset="50%"></stop>
            <stop stop-color="lime" offset="50%"></stop>
            <stop stop-color="lime" offset="75%"></stop>
            <stop stop-color="rgb(128,0,128)" offset="75%"></stop>
          </linearGradient>
        </defs>
        <rect x="0" y="0" width="11" height="16" fill="url(#grad)" />
      </svg>
    `)
}

func TestRadialGradient(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "radial_gradient", `
        rrrrrrrrrr
        rrrrrrrrrr
        rrrrBBrrrr
        rrrBBBBrrr
        rrBBBBBBrr
        rrBBBBBBrr
        rrrBBBBrrr
        rrrrBBrrrr
        rrrrrrrrrr
        rrrrrrrrrr
    `, `
      <style>
        @page { size: 10px }
        svg { display: block }
      </style>
      <svg width="10px" height="10px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <radialGradient id="grad" cx="0.5" cy="0.5" r="0.5"
            fx="0.5" fy="0.5" fr="0.2"
            gradientUnits="objectBoundingBox">
            <stop stop-color="blue" offset="25%"></stop>
            <stop stop-color="red" offset="25%"></stop>
          </radialGradient>
        </defs>
        <rect x="0" y="0" width="10" height="10" fill="url(#grad)" />
      </svg>
    `)
}

func TestRadialGradientUserspace(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "radial_gradient_userspace", `
        rrrrrrrrrr
        rrrrrrrrrr
        rrrrBBrrrr
        rrrBBBBrrr
        rrBBBBBBrr
        rrBBBBBBrr
        rrrBBBBrrr
        rrrrBBrrrr
        rrrrrrrrrr
        rrrrrrrrrr
    `, `
      <style>
        @page { size: 10px }
        svg { display: block }
      </style>
      <svg width="10px" height="10px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <radialGradient id="grad" cx="5" cy="5" r="5" fx="5" fy="5" fr="2"
            gradientUnits="userSpaceOnUse">
            <stop stop-color="blue" offset="25%"></stop>
            <stop stop-color="red" offset="25%"></stop>
          </radialGradient>
        </defs>
        <rect x="0" y="0" width="10" height="10" fill="url(#grad)" />
      </svg>
    `)
}

func TestRadialGradientMulticolor(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "radial_gradient_multicolor", `
        rrrrrrrrrr
        rrrGGGGrrr
        rrGGBBGGrr
        rGGBBBBGGr
        rGBBBBBBGr
        rGBBBBBBGr
        rGGBBBBGGr
        rrGGBBGGrr
        rrrGGGGrrr
        rrrrrrrrrr
    `, `
      <style>
        @page { size: 10px }
        svg { display: block }
      </style>
      <svg width="10px" height="10px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <radialGradient id="grad" cx="0.5" cy="0.5" r="0.5"
            fx="0.5" fy="0.5" fr="0.2"
            gradientUnits="objectBoundingBox">
            <stop stop-color="blue" offset="33%"></stop>
            <stop stop-color="lime" offset="33%"></stop>
            <stop stop-color="lime" offset="66%"></stop>
            <stop stop-color="red" offset="66%"></stop>
          </radialGradient>
        </defs>
        <rect x="0" y="0" width="10" height="10" fill="url(#grad)" />
      </svg>
    `)
}

func TestRadialGradientMulticolorUserspace(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "radial_gradient_multicolor_userspace", `
        rrrrrrrrrr
        rrrGGGGrrr
        rrGGBBGGrr
        rGGBBBBGGr
        rGBBBBBBGr
        rGBBBBBBGr
        rGGBBBBGGr
        rrGGBBGGrr
        rrrGGGGrrr
        rrrrrrrrrr
    `, `
      <style>
        @page { size: 10px }
        svg { display: block }
      </style>
      <svg width="10px" height="10px" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <radialGradient id="grad" cx="5" cy="5" r="5"
            fx="5" fy="5" fr="2"
            gradientUnits="userSpaceOnUse">
            <stop stop-color="blue" offset="33%"></stop>
            <stop stop-color="lime" offset="33%"></stop>
            <stop stop-color="lime" offset="66%"></stop>
            <stop stop-color="red" offset="66%"></stop>
          </radialGradient>
        </defs>
        <rect x="0" y="0" width="10" height="10" fill="url(#grad)" />
      </svg>
    `)
}

// TODO: pytest.mark.xfail
// func TestRadialGradientRepeat(t *testing.T) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

// 	assertPixelsEqual(t, "radial_gradient_repeat", `
//         GBrrrrrrBG
//         BrrGGGGrrB
//         rrGGBBGGrr
//         rGGBBBBGGr
//         rGBBBBBBGr
//         rGBBBBBBGr
//         rGGBBBBGGr
//         rrGGBBGGrr
//         BrrGGGGrrB
//         GBrrrrrrBG
//     `, `
//       <style>
//         @page { size: 10px }
//         svg { display: block }
//       </style>
//       <svg width="10px" height="10px" xmlns="http://www.w3.org/2000/svg">
//         <defs>
//           <radialGradient id="grad" cx="0.5" cy="0.5" r="0.5"
//             fx="0.5" fy="0.5" fr="0.2"
//             gradientUnits="objectBoundingBox" spreadMethod="repeat">
//             <stop stop-color="blue" offset="33%"></stop>
//             <stop stop-color="lime" offset="33%"></stop>
//             <stop stop-color="lime" offset="66%"></stop>
//             <stop stop-color="red" offset="66%"></stop>
//           </radialGradient>
//         </defs>
//         <rect x="0" y="0" width="10" height="10" fill="url(#grad)" />
//       </svg>
//     `)
// }

// TODO: pytest.mark.xfail
// func TestRadialGradientReflect(t *testing.T) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

// 	assertPixelsEqual(t, "radial_gradient_reflect", `
//         BGrrrrrrGB
//         GrrGGGGrrG
//         rrGGBBGGrr
//         rGGBBBBGGr
//         rGBBBBBBGr
//         rGBBBBBBGr
//         rGGBBBBGGr
//         rrGGBBGGrr
//         GrrGGGGrrG
//         BGrrrrrrGB
//     `, `
//       <style>
//         @page { size: 10px }
//         svg { display: block }
//       </style>
//       <svg width="10px" height="10px" xmlns="http://www.w3.org/2000/svg">
//         <defs>
//           <radialGradient id="grad" cx="0.5" cy="0.5" r="0.5"
//             fx="0.5" fy="0.5" fr="0.2"
//             gradientUnits="objectBoundingBox" spreadMethod="reflect">
//             <stop stop-color="blue" offset="33%"></stop>
//             <stop stop-color="lime" offset="33%"></stop>
//             <stop stop-color="lime" offset="66%"></stop>
//             <stop stop-color="red" offset="66%"></stop>
//           </radialGradient>
//         </defs>
//         <rect x="0" y="0" width="10" height="10" fill="url(#grad)" />
//       </svg>
//     `)
// }
