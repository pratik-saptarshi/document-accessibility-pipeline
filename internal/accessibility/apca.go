// Package accessibility provides bounded, advisory-only perceptual contrast
// evaluation for generated and remediated documents.
//
// # DRAFT-STAGE / ADVISORY STATUS — READ FIRST
//
// This package implements APCA (Accessible Perceptual Contrast Algorithm),
// the contrast method proposed for WCAG 3.0. APCA is a WORKING DRAFT: it is
// not a W3C Recommendation, it is not part of WCAG 2.x, and its constants
// may change before WCAG 3.0 is finalized.
//
// Values produced here are ADVISORY DESIGN GUIDANCE ONLY. They MUST NOT be
// used as pass/fail conformance gates, they MUST NOT be wired into CI
// quality gates or acceptance criteria, and they do not establish WCAG
// conformance of any kind. The repository's accepted conformance target
// remains PDF/UA-2 machine validation plus human semantic review (see
// docs/decisions/0001-pdf-ua-2-validation-target.md).
package accessibility

import (
	"image/color"
	"math"
)

// APCA-W3 (0.1.9, 0.0.98G-4g) constants for sRGB input, per the canonical
// reference "APCA 0.1.9 4g constants and math" at
// https://git.apcacontrast.com/documentation/README.html.
const (
	apcaMainTRC = 2.4 // sRGB linearization exponent

	apcaSRco = 0.2126729 // sRGB luminance coefficients
	apcaSGco = 0.7151522
	apcaSBco = 0.0721750

	apcaNormBG  = 0.56 // normal polarity (dark text, light bg) exponents
	apcaNormTXT = 0.57
	apcaRevBG   = 0.65 // reverse polarity (light text, dark bg) exponents
	apcaRevTXT  = 0.62

	apcaBlkThrs = 0.022 // soft black clamp threshold
	apcaBlkClmp = 1.414 // soft black clamp exponent

	apcaScaleBoW = 1.14  // normal polarity scaler
	apcaScaleWoB = 1.14  // reverse polarity scaler
	apcaLoBoWOff = 0.027 // normal polarity low offset
	apcaLoWoBOff = 0.027 // reverse polarity low offset
	apcaLoClip   = 0.1   // "contrast too low" threshold on scaled SAPC
)

// Contrast returns the APCA lightness contrast (Lc) of fg (text) over bg
// (background) using the APCA-W3 sRGB model.
//
// ADVISORY / DRAFT STATUS: APCA is a WCAG 3.0 working-draft method, not a
// conformance standard. The returned Lc is advisory design guidance only: do
// NOT use it as a pass/fail gate. Positive Lc means dark text on a light
// background; negative Lc means light text on a dark background. Lc 0 means
// the pair is below the low-clip threshold ("contrast too low").
//
// Colors are evaluated as opaque sRGB values; alpha is ignored. Colors
// outside the sRGB gamut are clamped by the standard library conversion.
func Contrast(fg, bg color.Color) float64 {
	yTxt := softClamp(luminance(fg))
	yBg := softClamp(luminance(bg))

	var sapc, offset float64
	if yBg > yTxt {
		// Normal polarity: dark text on a light background.
		sapc = (math.Pow(yBg, apcaNormBG) - math.Pow(yTxt, apcaNormTXT)) * apcaScaleBoW
		offset = apcaLoBoWOff
	} else {
		// Reverse polarity: light text on a dark background.
		sapc = (math.Pow(yBg, apcaRevBG) - math.Pow(yTxt, apcaRevTXT)) * apcaScaleWoB
		offset = apcaLoWoBOff
	}

	if math.Abs(sapc) < apcaLoClip {
		return 0
	}
	if sapc > 0 {
		return (sapc - offset) * 100
	}
	return (sapc + offset) * 100
}

// luminance linearizes an sRGB color to relative luminance Y:
//
//	Y = (R/255)^2.4 * 0.2126729 + (G/255)^2.4 * 0.7151522 + (B/255)^2.4 * 0.0721750
func luminance(c color.Color) float64 {
	n, ok := c.(color.NRGBA)
	if !ok {
		n = color.NRGBAModel.Convert(c).(color.NRGBA)
	}
	return math.Pow(float64(n.R)/255, apcaMainTRC)*apcaSRco +
		math.Pow(float64(n.G)/255, apcaMainTRC)*apcaSGco +
		math.Pow(float64(n.B)/255, apcaMainTRC)*apcaSBco
}

// softClamp raises near-black luminance so perceptual contrast does not
// collapse at the bottom of the scale:
//
//	clampedY = (0.022 - Y)^1.414 + Y
//
// applied independently to each color whose Y is at or below 0.022.
func softClamp(y float64) float64 {
	if y <= apcaBlkThrs {
		return math.Pow(apcaBlkThrs-y, apcaBlkClmp) + y
	}
	return y
}
