// Package accessibility provides WCAG 2.2 contrast evaluation for generated
// and remediated documents.
//
// The ratios and threshold predicates here are ADVISORY measurements, not
// conformance claims: they do not establish WCAG conformance of any
// document, and they MUST NOT be wired into CI quality gates or acceptance
// criteria as pass/fail verdicts. The repository's accepted conformance
// target remains PDF/UA-2 machine validation plus human semantic review (see
// docs/decisions/0001-pdf-ua-2-validation-target.md). Per owner direction
// (2026-10-18, recorded in docs/decisions/0014-profile-policy-decisions.md),
// APCA is not used: contrast is evaluated against the WCAG 2.2 contrast
// ratio definition.
//
// Large-text status is never inferred from the two colors alone. Callers
// pass largeText explicitly; per WCAG 2.2, "large" means at least 18pt
// (24px) regular or 14pt (18.66px) bold, a property of the rendered text,
// not of its colors.
package accessibility

import (
	"image/color"
	"math"
)

// WCAG 2.2 contrast thresholds (SC 1.4.3 Contrast (Minimum) for AA,
// SC 1.4.6 Contrast (Enhanced) for AAA).
const (
	// ThresholdAANormal is the AA minimum contrast ratio for normal text.
	ThresholdAANormal = 4.5
	// ThresholdAALarge is the AA minimum contrast ratio for large text.
	ThresholdAALarge = 3.0
	// ThresholdAAANormal is the AAA enhanced contrast ratio for normal text.
	ThresholdAAANormal = 7.0
	// ThresholdAAALarge is the AAA enhanced contrast ratio for large text.
	ThresholdAAALarge = 4.5
)

// RelativeLuminance returns the WCAG 2.2 relative luminance of c, in the
// range [0, 1].
//
// The sRGB component values are linearized per the WCAG 2.2 definition:
//
//	if CsRGB <= 0.04045 then C = CsRGB/12.92
//	else C = ((CsRGB+0.055)/1.055)^2.4
//
// and combined as L = 0.2126*R + 0.7152*G + 0.0722*B.
//
// Colors are evaluated from their straight source sRGB channels as opaque
// values; alpha is ignored. This is an advisory source-channel measurement,
// not the contrast of a color composited over a particular background or the
// rendered appearance of a partially transparent color. The
// computation uses 16-bit channel components (via NRGBA64) so that
// 16-bit inputs are not truncated to 8 bits before linearization.
//
// Unsupported: alpha=0 (fully transparent) inputs are out of scope for the
// current backlog. The behavior for such inputs is type-dependent and not
// covered by the contrast guarantees in this package.
func RelativeLuminance(c color.Color) float64 {
	var n color.NRGBA64
	switch c := c.(type) {
	case color.NRGBA:
		// NRGBA stores straight (non-premultiplied) 8-bit channels. Preserve
		// those channels directly: converting through RGBA would multiply by
		// alpha and quantize low-alpha colors on the way back to straight RGB.
		n = color.NRGBA64{
			R: uint16(c.R) * 0x101,
			G: uint16(c.G) * 0x101,
			B: uint16(c.B) * 0x101,
			A: uint16(c.A) * 0x101,
		}
	case color.NRGBA64:
		// NRGBA64 already stores straight 16-bit channels.
		n = c
	default:
		// Other color.Color implementations expose their appearance through
		// RGBA(), whose channels are premultiplied by alpha. Convert those
		// according to the standard color model; only concrete NRGBA types can
		// promise that their original straight channels should be retained.
		n = color.NRGBA64Model.Convert(c).(color.NRGBA64)
	}
	return 0.2126*linearize(float64(n.R)/65535) +
		0.7152*linearize(float64(n.G)/65535) +
		0.0722*linearize(float64(n.B)/65535)
}

// linearize applies the WCAG 2.2 sRGB transfer function to one component
// in the range [0, 1].
func linearize(cs float64) float64 {
	if cs <= 0.04045 {
		return cs / 12.92
	}
	return math.Pow((cs+0.055)/1.055, 2.4)
}

// ContrastRatio returns the WCAG 2.2 contrast ratio of fg (text) over bg
// (background):
//
//	ratio = (L1 + 0.05) / (L2 + 0.05)
//
// where L1 is the lighter and L2 the darker of the two relative
// luminances. The result ranges from 1 (identical colors) to 21 (black on
// white).
//
// This replaces the earlier APCA-based Contrast function: APCA is not
// used anywhere in this package per owner direction (2026-10-18). The
// name differs deliberately — an APCA lightness contrast (Lc) is not a
// ratio, so reusing the old name would mislead.
func ContrastRatio(fg, bg color.Color) float64 {
	l1 := RelativeLuminance(fg)
	l2 := RelativeLuminance(bg)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

// meetsThreshold reports whether ratio meets threshold. The comparison is
// inclusive: a ratio of exactly the threshold passes.
func meetsThreshold(ratio, threshold float64) bool {
	return ratio >= threshold
}

// MeetsAA reports whether the fg/bg pair meets the WCAG 2.2 AA minimum
// contrast (SC 1.4.3): 4.5:1 for normal text, 3:1 for large text.
//
// The largeText argument selects the threshold explicitly; it is never
// inferred from the colors. The result is an advisory label input, not a
// conformance verdict or a CI gate.
func MeetsAA(fg, bg color.Color, largeText bool) bool {
	threshold := ThresholdAANormal
	if largeText {
		threshold = ThresholdAALarge
	}
	return meetsThreshold(ContrastRatio(fg, bg), threshold)
}

// MeetsAAA reports whether the fg/bg pair meets the WCAG 2.2 AAA enhanced
// contrast (SC 1.4.6): 7:1 for normal text, 4.5:1 for large text.
//
// As with MeetsAA, largeText is an explicit caller input and the result is
// advisory only.
func MeetsAAA(fg, bg color.Color, largeText bool) bool {
	threshold := ThresholdAAANormal
	if largeText {
		threshold = ThresholdAAALarge
	}
	return meetsThreshold(ContrastRatio(fg, bg), threshold)
}
