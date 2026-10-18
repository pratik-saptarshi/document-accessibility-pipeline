package accessibility

// Tests for WCAG 2.2 contrast evaluation.
//
// Expected values below were computed INDEPENDENTLY of this package's
// implementation, using a separate Python script implementing the WCAG 2.2
// relative-luminance and contrast-ratio definitions:
//
//	black on white            = 21.0
//	#767676 on white          = 4.542225  (passes AA normal text, 4.5:1)
//	#777777 on white          = 4.478089  (fails AA normal, passes AA large 3:1)
//	#595959 on white          = 7.004729  (passes AAA normal text, 7:1)
//	#5a5a5a on white          = 6.896926  (fails AAA normal, passes AAA large 4.5:1)
//	#808080 on white          = 3.949440  (fails AA normal, passes AA large)
//	#005a9c on white          = 7.136726  (colored pair, passes AAA normal)
//	#b30000 on white          = 7.199816  (colored pair, passes AAA normal)
//	RelativeLuminance(white)   = 1.0
//	RelativeLuminance(black)   = 0.0
//	RelativeLuminance(#808080) = 0.2158605

import (
	"image/color"
	"math"
	"strconv"
	"testing"
)

func mustHex(t *testing.T, s string) color.Color {
	t.Helper()
	if len(s) != 6 {
		t.Fatalf("mustHex: want 6 hex digits, got %q", s)
	}
	var v [3]uint8
	for i := 0; i < 3; i++ {
		n, err := strconv.ParseUint(s[2*i:2*i+2], 16, 8)
		if err != nil {
			t.Fatalf("mustHex(%q): %v", s, err)
		}
		v[i] = uint8(n)
	}
	return color.NRGBA{R: v[0], G: v[1], B: v[2], A: 0xff}
}

func TestRelativeLuminanceExtremes(t *testing.T) {
	if got := RelativeLuminance(color.White); got != 1 {
		t.Errorf("RelativeLuminance(white) = %v, want 1", got)
	}
	if got := RelativeLuminance(color.Black); got != 0 {
		t.Errorf("RelativeLuminance(black) = %v, want 0", got)
	}
}

func TestRelativeLuminanceMidGray(t *testing.T) {
	// Independent reference: 0.2158605 for #808080.
	if got := RelativeLuminance(mustHex(t, "808080")); math.Abs(got-0.2158605) > 1e-6 {
		t.Errorf("RelativeLuminance(#808080) = %v, want 0.2158605", got)
	}
}

func TestContrastRatioBlackOnWhiteIs21(t *testing.T) {
	if got := ContrastRatio(color.Black, color.White); got != 21 {
		t.Errorf("ContrastRatio(black, white) = %v, want 21", got)
	}
}

func TestContrastRatioIdenticalColorsIsOne(t *testing.T) {
	c := mustHex(t, "005a9c")
	if got := ContrastRatio(c, c); got != 1 {
		t.Errorf("ContrastRatio(same, same) = %v, want 1", got)
	}
}

func TestContrastRatioIsSymmetric(t *testing.T) {
	fg, bg := mustHex(t, "005a9c"), mustHex(t, "ffffff")
	if a, b := ContrastRatio(fg, bg), ContrastRatio(bg, fg); a != b {
		t.Errorf("ContrastRatio not symmetric: fg/bg = %v, bg/fg = %v", a, b)
	}
}

func TestContrastRatioReferencePairs(t *testing.T) {
	white := mustHex(t, "ffffff")
	cases := []struct {
		name string
		fg   string
		want float64
	}{
		{"black", "000000", 21.0},
		{"near-AA-pass", "767676", 4.542225},
		{"near-AA-fail", "777777", 4.478089},
		{"near-AAA-pass", "595959", 7.004729},
		{"near-AAA-fail", "5a5a5a", 6.896926},
		{"mid-gray", "808080", 3.949440},
		{"blue", "005a9c", 7.136726},
		{"red", "b30000", 7.199816},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ContrastRatio(mustHex(t, tc.fg), white)
			if math.Abs(got-tc.want) > 0.002 {
				t.Errorf("ContrastRatio(#%s, white) = %v, want %v", tc.fg, got, tc.want)
			}
		})
	}
}

func TestMeetsThresholdExactBoundary(t *testing.T) {
	// The comparison itself must be inclusive: exactly 4.5:1 passes,
	// anything below fails. Tested at the unit level so color
	// quantization cannot blur the boundary.
	if !meetsThreshold(4.5, ThresholdAANormal) {
		t.Error("meetsThreshold(4.5, 4.5) = false, want true")
	}
	if meetsThreshold(4.4999999, ThresholdAANormal) {
		t.Error("meetsThreshold(4.4999999, 4.5) = true, want false")
	}
}

func TestMeetsAANormalTextBoundary(t *testing.T) {
	white := mustHex(t, "ffffff")
	// #767676 -> 4.542225:1, just above the 4.5:1 AA normal-text line.
	if !MeetsAA(mustHex(t, "767676"), white, false) {
		t.Error("MeetsAA(#767676, white, normal) = false, want true (4.54 >= 4.5)")
	}
	// #777777 -> 4.478089:1, just below it.
	if MeetsAA(mustHex(t, "777777"), white, false) {
		t.Error("MeetsAA(#777777, white, normal) = true, want false (4.48 < 4.5)")
	}
}

func TestMeetsAALargeTextBoundary(t *testing.T) {
	white := mustHex(t, "ffffff")
	// #777777 -> 4.478089:1 clears the 3:1 large-text line.
	if !MeetsAA(mustHex(t, "777777"), white, true) {
		t.Error("MeetsAA(#777777, white, large) = false, want true (4.48 >= 3.0)")
	}
	// #808080 -> 3.949440:1 clears large text but not normal text.
	if !MeetsAA(mustHex(t, "808080"), white, true) {
		t.Error("MeetsAA(#808080, white, large) = false, want true (3.95 >= 3.0)")
	}
	if MeetsAA(mustHex(t, "808080"), white, false) {
		t.Error("MeetsAA(#808080, white, normal) = true, want false (3.95 < 4.5)")
	}
}

func TestMeetsAAANormalVsLarge(t *testing.T) {
	white := mustHex(t, "ffffff")
	// #595959 -> 7.004729:1 clears the 7:1 AAA normal-text line.
	if !MeetsAAA(mustHex(t, "595959"), white, false) {
		t.Error("MeetsAAA(#595959, white, normal) = false, want true (7.00 >= 7.0)")
	}
	// #5a5a5a -> 6.896926:1 misses AAA normal but clears AAA large (4.5:1).
	if MeetsAAA(mustHex(t, "5a5a5a"), white, false) {
		t.Error("MeetsAAA(#5a5a5a, white, normal) = true, want false (6.90 < 7.0)")
	}
	if !MeetsAAA(mustHex(t, "5a5a5a"), white, true) {
		t.Error("MeetsAAA(#5a5a5a, white, large) = false, want true (6.90 >= 4.5)")
	}
}

func TestLargeTextIsNeverInferredFromColors(t *testing.T) {
	// The largeText flag is the ONLY thing that selects the 3:1 line.
	// Identical colors must not change the outcome by themselves:
	// #777777 on white is 4.48:1, which passes AA only as large text.
	white := mustHex(t, "ffffff")
	fg := mustHex(t, "777777")
	if MeetsAA(fg, white, false) {
		t.Error("MeetsAA normal-text must not pass 4.48:1")
	}
	if !MeetsAA(fg, white, true) {
		t.Error("MeetsAA large-text must pass 4.48:1")
	}
}

func TestColorTypesAgree(t *testing.T) {
	// color.RGBA, color.NRGBA, and color.Gray for the same sRGB gray
	// must produce the same ratio; alpha handling must not leak in.
	gray := uint8(0x77)
	a := ContrastRatio(color.RGBA{R: gray, G: gray, B: gray, A: 0xff}, color.White)
	b := ContrastRatio(color.NRGBA{R: gray, G: gray, B: gray, A: 0xff}, color.White)
	c := ContrastRatio(color.Gray{Y: gray}, color.White)
	if math.Abs(a-b) > 1e-9 || math.Abs(a-c) > 1e-9 {
		t.Errorf("color type disagreement: RGBA=%v NRGBA=%v Gray=%v", a, b, c)
	}
}

// TestRelativeLuminancePreserves16BitPrecision requires luminance to be
// computed from 16-bit channel components, not truncated to 8-bit.
// NRGBA64{30464,30464,30464} against white has a WCAG ratio of ~4.5076
// (passes AA normal text); 8-bit quantization to #777777 gives ~4.4781
// (fails). The API accepts any color.Color with no documented 8-bit
// restriction (Codex P2 on PR #154).
func TestRelativeLuminancePreserves16BitPrecision(t *testing.T) {
	c := color.NRGBA64{R: 30464, G: 30464, B: 30464, A: 65535}
	ratio := ContrastRatio(c, color.White)
	if ratio < ThresholdAANormal {
		t.Fatalf("16-bit gray NRGBA64{30464} vs white: ratio %v must meet AA normal %v (16-bit precision must be preserved)", ratio, ThresholdAANormal)
	}
	if math.Abs(ratio-4.5076) > 0.01 {
		t.Fatalf("16-bit gray NRGBA64{30464} vs white: ratio %v, want ~4.5076", ratio)
	}
}
