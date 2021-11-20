package pdf

import (
	"testing"

	"github.com/benoitkugler/go-weasyprint/utils/testutils"
)

// Test how gradients are drawn.

func TestLinearGradients_1(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient_1", `
        _____
        _____
        _____
        BBBBB
        BBBBB
        RRRRR
        RRRRR
        RRRRR
        RRRRR
    `, `<style>@page { size: 5px 9px; background: linear-gradient(
      white, white 3px, blue 0, blue 5px, red 0, red
    )`)
}

func TestLinearGradients_2(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient_2", `
        _____
        _____
        _____
        BBBBB
        BBBBB
        RRRRR
        RRRRR
        RRRRR
        RRRRR
    `, `<style>@page { size: 5px 9px; background: linear-gradient(
      white 3px, blue 0, blue 5px, red 0
    )`)
}

func TestLinearGradients_3(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient_3", `
        ___BBrrrr
        ___BBrrrr
        ___BBrrrr
        ___BBrrrr
        ___BBrrrr
    `, `<style>@page { size: 9px 5px; background: linear-gradient(
      to right, white 3px, blue 0, blue 5px, red 0
    )`)
}

func TestLinearGradients_4(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient_4", `
        BBBBBBrrrr
        BBBBBBrrrr
        BBBBBBrrrr
        BBBBBBrrrr
        BBBBBBrrrr
    `, `<style>@page { size: 10px 5px; background: linear-gradient(
      to right, blue 5px, blue 6px, red 6px, red 9px
    )`)
}

func TestLinearGradients_5(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient_5", `
        rBrrrBrrrB
        rBrrrBrrrB
        rBrrrBrrrB
        rBrrrBrrrB
        rBrrrBrrrB
    `, `<style>@page { size: 10px 5px; background: repeating-linear-gradient(
      to right, blue 50%, blue 60%, red 60%, red 90%
    )`)
}

func TestLinearGradients_6(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient_6", `
        BBBrrrrrr
        BBBrrrrrr
        BBBrrrrrr
        BBBrrrrrr
        BBBrrrrrr
    `, `<style>@page { size: 9px 5px; background: linear-gradient(
      to right, blue 3px, blue 3px, red 3px, red 3px
    )`)
}

func TestLinearGradients_7(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient_7", `
        hhhhhhhhh
        hhhhhhhhh
        hhhhhhhhh
        hhhhhhhhh
        hhhhhhhhh
    `, `<style>@page { size: 9px 5px; background: repeating-linear-gradient(
      to right, black 3px, black 3px, #800080 3px, #800080 3px
    )`)
}

func TestLinearGradients_8(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient_8", `
        BBBBBBBBB
        BBBBBBBBB
        BBBBBBBBB
        BBBBBBBBB
        BBBBBBBBB
    `, `<style>@page { size: 9px 5px; background: repeating-linear-gradient(
      to right, blue 3px
    )`)
}

func TestLinearGradients_9(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient_9", `
        BBBBBBBBB
        BBBBBBBBB
        BBBBBBBBB
        BBBBBBBBB
        BBBBBBBBB
    `, `<style>@page { size: 9px 5px; background: repeating-linear-gradient(
      45deg, blue 3px
    )`)
}

func TestLinearGradients_10(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient_10", `
        BBBBBBBBB
        BBBBBBBBB
        BBBBBBBBB
        BBBBBBBBB
        BBBBBBBBB
    `, `<style>@page { size: 9px 5px; background: linear-gradient(
      45deg, blue 3px, red 3px, red 3px, blue 3px
    )`)
}

func TestLinearGradients_11(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient_11", `
        BBBrBBBBB
        BBBrBBBBB
        BBBrBBBBB
        BBBrBBBBB
        BBBrBBBBB
    `, `<style>@page { size: 9px 5px; background: linear-gradient(
      to right, blue 3px, red 3px, red 4px, blue 4px
    )`)
}

func TestLinearGradients_12(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient_12", `
        BBBBBBBBB
        BBBBBBBBB
        BBBBBBBBB
        BBBBBBBBB
        BBBBBBBBB
    `, `<style>@page { size: 9px 5px; background: repeating-linear-gradient(
      to right, red 3px, blue 3px, blue 4px, red 4px
    )`)
}

func TestLinearGradients_13(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "linear_gradient_13", `
        _____
        _____
        _____
        SSSSS
        SSSSS
        RRRRR
        RRRRR
        RRRRR
        RRRRR
    `, `<style>@page { size: 5px 9px; background: linear-gradient(
      white, white 3px, rgba(255, 0, 0, 0.751) 0, rgba(255, 0, 0, 0.751) 5px,
      red 0, red
    )`)
}

func TestRadialGradients_1(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "radial_gradient_1", `
        BBBBBB
        BBBBBB
        BBBBBB
        BBBBBB
        BBBBBB
        BBBBBB
    `, `<style>@page { size: 6px; background:
      radial-gradient(red -30%, blue -10%)`)
}

func TestRadialGradients_2(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "radial_gradient_2", `
        RRRRRR
        RRRRRR
        RRRRRR
        RRRRRR
        RRRRRR
        RRRRRR
    `, `<style>@page { size: 6px; background:
      radial-gradient(red 110%, blue 130%)`)
}

func TestRadialGradients_3(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "radial_gradient_3", `
        BzzzzzzzzB
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzRRzzzz
        zzzzRRzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        BzzzzzzzzB
    `, `<style>@page { size: 10px 16px; background:
      radial-gradient(red 20%, blue 80%)`)
}

func TestRadialGradients_4(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "radial_gradient_4", `
        BzzzzzzzzB
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzRRzzzz
        zzzzRRzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        BzzzzzzzzB
    `, `<style>@page { size: 10px 16px; background:
      radial-gradient(red 50%, blue 50%)`)
}

func TestRadialGradients_5(t *testing.T) {
	capt := testutils.CaptureLogs()
	defer capt.AssertNoLogs(t)

	assertPixelsEqual(t, "radial_gradient_5", `
        SzzzzzzzzS
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzRRzzzz
        zzzzRRzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        zzzzzzzzzz
        SzzzzzzzzS
    `, `<style>@page { size: 10px 16px; background:
      radial-gradient(red 50%, rgba(255, 0, 0, 0.751) 50%)`)
}
