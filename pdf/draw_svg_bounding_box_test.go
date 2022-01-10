package pdf

// import (
// 	"testing"

// 	"github.com/benoitkugler/webrender/utils/testutils"
// )

// // Test how bounding boxes are defined for SVG tags.

// func TestBoundingBoxRect(t *testing.T) {
// 	// capt := testutils.CaptureLogs()
// 	// defer capt.AssertNoLogs(t)

// 	assertPixelsEqual(t, "boundingBoxRect", `
//         BBBBB
//         BBBBR
//         BBBRR
//         BBRRR
//         BRRRR
//     `, `
//       <style>
//         @page { size: 5px }
//         svg { display: block }
//       </style>
//       <svg width="5px" height="5px" xmlns="http://www.w3.org/2000/svg">
//         <defs>
//           <linearGradient id="grad" x1="0" y1="0" x2="1" y2="1"
//             gradientUnits="objectBoundingBox">
//             <stop stop-color="blue" offset="55%"></stop>
//             <stop stop-color="red" offset="55%"></stop>
//           </linearGradient>
//         </defs>
//         <rect x="0" y="0" width="5" height="5" fill="url(#grad)" />
//       </svg>
//     `)
// }

// func TestBoundingBoxCircle(t *testing.T) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

// 	assertPixelsEqual(t, "boundingBoxCircle", `
//         _________
//         _BBBBBB_
//         BBBBBBBR
//         BBBBBBRR
//         BBBBBRRR
//         BBBBRRRR
//         BBBRRRRR
//         BBRRRRRR
//         _RRRRRR_
//         _________
//     `, `
//       <style>
//         @page { size: 10px }
//         svg { display: block }
//       </style>
//       <svg width="10px" height="10px" xmlns="http://www.w3.org/2000/svg">
//         <defs>
//           <linearGradient id="grad" x1="0" y1="0" x2="1" y2="1"
//             gradientUnits="objectBoundingBox">
//             <stop stop-color="blue" offset="55%"></stop>
//             <stop stop-color="red" offset="55%"></stop>
//           </linearGradient>
//         </defs>
//         <circle cx="5" cy="5" r="4" fill="url(#grad)" />
//       </svg>
//     `)
// }

// func TestBoundingBoxEllipse(t *testing.T) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

// 	assertPixelsEqual(t, "boundingBoxEllipse", `
//         _________
//         _BBBBBB_
//         BBBBBBBR
//         BBBBBBRR
//         BBBBBRRR
//         BBBBRRRR
//         BBBRRRRR
//         BBRRRRRR
//         _RRRRRR_
//         _________
//     `, `
//       <style>
//         @page { size: 10px }
//         svg { display: block }
//       </style>
//       <svg width="10px" height="10px" xmlns="http://www.w3.org/2000/svg">
//         <defs>
//           <linearGradient id="grad" x1="0" y1="0" x2="1" y2="1"
//             gradientUnits="objectBoundingBox">
//             <stop stop-color="blue" offset="55%"></stop>
//             <stop stop-color="red" offset="55%"></stop>
//           </linearGradient>
//         </defs>
//         <ellipse cx="5" cy="5" rx="4" ry="4" fill="url(#grad)" />
//       </svg>
//     `)
// }

// func TestBoundingBoxLine(t *testing.T) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

// 	assertPixelsEqual(t, "boundingBoxLine", `
//         BB__
//         BBB_
//         BRR
//         _RRR
//         __RR
//     `, `
//       <style>
//         @page { size: 5px }
//         svg { display: block }
//       </style>
//       <svg width="5px" height="5px" xmlns="http://www.w3.org/2000/svg">
//         <defs>
//           <linearGradient id="grad" x1="0" y1="0" x2="1" y2="1"
//             gradientUnits="objectBoundingBox">
//             <stop stop-color="blue" offset="50%"></stop>
//             <stop stop-color="red" offset="50%"></stop>
//           </linearGradient>
//         </defs>
//         <line x1="0" y1="0" x2="5" y2="5"
//               stroke-width="1" stroke="url(#grad)" />
//       </svg>
//     `)
// }

// func TestBoundingBoxPolygon(t *testing.T) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

// 	assertPixelsEqual(t, "boundingBoxPolygon", `
//         BBBBB
//         BBBBR
//         BBBRR
//         BBRRR
//         BRRRR
//     `, `
//       <style>
//         @page { size: 5px }
//         svg { display: block }
//       </style>
//       <svg width="5px" height="5px" xmlns="http://www.w3.org/2000/svg">
//         <defs>
//           <linearGradient id="grad" x1="0" y1="0" x2="1" y2="1"
//             gradientUnits="objectBoundingBox">
//             <stop stop-color="blue" offset="55%"></stop>
//             <stop stop-color="red" offset="55%"></stop>
//           </linearGradient>
//         </defs>
//         <polygon points="0 0 0 5 5 5 5 0" fill="url(#grad)" />
//       </svg>
//     `)
// }

// func TestBoundingBoxPolyline(t *testing.T) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

// 	assertPixelsEqual(t, "boundingBoxPolyline", `
//         BBBBB
//         BBBBR
//         BBBRR
//         BBRRR
//         BRRRR
//     `, `
//       <style>
//         @page { size: 5px }
//         svg { display: block }
//       </style>
//       <svg width="5px" height="5px" xmlns="http://www.w3.org/2000/svg">
//         <defs>
//           <linearGradient id="grad" x1="0" y1="0" x2="1" y2="1"
//             gradientUnits="objectBoundingBox">
//             <stop stop-color="blue" offset="55%"></stop>
//             <stop stop-color="red" offset="55%"></stop>
//           </linearGradient>
//         </defs>
//         <polyline points="0 0 0 5 5 5 5 0" fill="url(#grad)" />
//       </svg>
//     `)
// }

// // TODO :pytest.mark.xfail
// // func TestBoundingBoxText(t *testing.T) {
// // 	capt := testutils.CaptureLogs()
// // 	defer capt.AssertNoLogs(t)

// //     assertPixelsEqual(t, "boundingBoxText", 2, 2, `
// //         BB
// //         BR
// //     `, `
// //       <style>
// //         @font-face { src: url(weasyprint.otf); font-family: weasyprint }
// //         @page { size: 2px }
// //         svg { display: block }
// //       </style>
// //       <svg width="2px" height="2px" xmlns="http://www.w3.org/2000/svg">
// //         <defs>
// //           <linearGradient id="grad" x1="0" y1="0" x2="1" y2="1"
// //             gradientUnits="objectBoundingBox">
// //             <stop stop-color="blue" offset="55%"></stop>
// //             <stop stop-color="red" offset="55%"></stop>
// //           </linearGradient>
// //         </defs>
// //         <text x="0" y="1" font-family="weasyprint" font-size="2"
// //               fill="url(#grad)">
// //           A
// //         </text>
// //       </svg>
// //     `)

// func TestBoundingBoxPathHv(t *testing.T) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

// 	assertPixelsEqual(t, "boundingBoxPathHv", `
//         BBBBB
//         BBBBR
//         BBBRR
//         BBRRR
//         BRRRR
//     `, `
//       <style>
//         @page { size: 5px }
//         svg { display: block }
//       </style>
//       <svg width="5px" height="5px" xmlns="http://www.w3.org/2000/svg">
//         <defs>
//           <linearGradient id="grad" x1="0" y1="0" x2="1" y2="1"
//             gradientUnits="objectBoundingBox">
//             <stop stop-color="blue" offset="55%"></stop>
//             <stop stop-color="red" offset="55%"></stop>
//           </linearGradient>
//         </defs>
//         <path d="m 5 0 v 5 h -5 V 0 H 5 z" fill="url(#grad)" />
//       </svg>
//     `)
// }

// func TestBoundingBoxPathL(t *testing.T) {
// 	capt := testutils.CaptureLogs()
// 	defer capt.AssertNoLogs(t)

// 	assertPixelsEqual(t, "boundingBoxPathL", `
//         BBBBB
//         BBBBR
//         BBBRR
//         BBRRR
//         BRRRR
//     `, `
//       <style>
//         @page { size: 5px }
//         svg { display: block }
//       </style>
//       <svg width="5px" height="5px" xmlns="http://www.w3.org/2000/svg">
//         <defs>
//           <linearGradient id="grad" x1="0" y1="0" x2="1" y2="1"
//             gradientUnits="objectBoundingBox">
//             <stop stop-color="blue" offset="55%"></stop>
//             <stop stop-color="red" offset="55%"></stop>
//           </linearGradient>
//         </defs>
//         <path d="M 5 0 l 0 5 l -5 0 L 0 0 z" fill="url(#grad)" />
//       </svg>
//     `)
// }

// // // TODO: pytest.mark.xfail
// // func TestBoundingBoxPathCt *testing.T():
// // capt := testutils.CaptureLogs()
// // defer capt.AssertNoLogs(t)

// //     assertPixelsEqual(t, "boundingBoxPathC",  `
// //         BBB_
// //         BBR_
// //         ____
// //         BBB_
// //         BBR_
// //     `, `
// //       <style>
// //         @page { size: 5px }
// //         svg { display: block }
// //       </style>
// //       <svg width="5px" height="5px" xmlns="http://www.w3.org/2000/svg">
// //         <defs>
// //           <linearGradient id="grad" x1="0" y1="0" x2="1" y2="1"
// //             gradientUnits="objectBoundingBox">
// //             <stop stop-color="blue" offset="55%"></stop>
// //             <stop stop-color="red" offset="55%"></stop>
// //           </linearGradient>
// //         </defs>
// //         <g fill="none" stroke="url(#grad)" stroke-width="2">
// //           <path d="M 0 1 C 0 1 1 1 3 1" />
// //           <path d="M 0 4 c 0 0 1 0 3 0" />
// //         </g>
// //       </svg>
// //     `)
// // }

// // TODO: pytest.mark.xfail
// // func TestBoundingBoxPathS(t *testing.T) {
// // 	capt := testutils.CaptureLogs()
// // 	defer capt.AssertNoLogs(t)

// //     assertPixelsEqual(t, "boundingBoxPathS",  `
// //         BBB_
// //         BBR_
// //         ____
// //         BBB_
// //         BBR_
// //     `, `
// //       <style>
// //         @page { size: 5px }
// //         svg { display: block }
// //       </style>
// //       <svg width="5px" height="5px" xmlns="http://www.w3.org/2000/svg">
// //         <defs>
// //           <linearGradient id="grad" x1="0" y1="0" x2="1" y2="1"
// //             gradientUnits="objectBoundingBox">
// //             <stop stop-color="blue" offset="55%"></stop>
// //             <stop stop-color="red" offset="55%"></stop>
// //           </linearGradient>
// //         </defs>
// //         <g fill="none" stroke="url(#grad)" stroke-width="2">
// //           <path d="M 0 1 S 1 1 3 1" />
// //           <path d="M 0 4 s 1 0 3 0" />
// //         </g>
// //       </svg>
// //     `)
