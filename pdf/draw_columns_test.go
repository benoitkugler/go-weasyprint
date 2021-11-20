package pdf

import (
	"testing"

	"github.com/benoitkugler/go-weasyprint/utils/testutils"
)

// Test how columns are drawn.

func TestColumnRule_1(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "solid", `
        a_r_a
        a_r_a
        _____
    `, `
      <style>
        img { display: inline-block; width: 1px; height: 1px }
        div { columns: 2; column-rule-style: solid;
              column-rule-width: 1px; column-gap: 3px;
              column-rule-color: red }
        body { margin: 0; font-size: 0; background: white}
        @page { margin: 0; size: 5px 3px }
      </style>
      <div>
        <img src="../resources_test/blue.jpg">
        <img src="../resources_test/blue.jpg">
        <img src="../resources_test/blue.jpg">
        <img src="../resources_test/blue.jpg">
      </div>`)
}

func TestColumnRule_2(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "dotted", `
        a_r_a
        a___a
        a_r_a
    `, `
      <style>
        img { display: inline-block; width: 1px; height: 1px }
        div { columns: 2; column-rule-style: dotted;
              column-rule-width: 1px; column-gap: 3px;
              column-rule-color: red }
        body { margin: 0; font-size: 0; background: white}
        @page { margin: 0; size: 5px 3px }
      </style>
      <div>
        <img src="../resources_test/blue.jpg">
        <img src="../resources_test/blue.jpg">
        <img src="../resources_test/blue.jpg">
        <img src="../resources_test/blue.jpg">
        <img src="../resources_test/blue.jpg">
        <img src="../resources_test/blue.jpg">
      </div>`)
}
