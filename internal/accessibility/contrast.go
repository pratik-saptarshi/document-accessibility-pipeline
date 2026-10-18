// Package accessibility provides small, pure accessibility calculations.
package accessibility

import "math"

// RGB is an opaque sRGB color. Callers must supply the rendered color values;
// this type does not model transparency or infer foreground/background roles.
type RGB struct {
	R uint8
	G uint8
	B uint8
}

// RelativeLuminance returns the WCAG relative luminance of an sRGB color.
// It ranges from 0 (black) to 1 (white).
func RelativeLuminance(c RGB) float64 {
	r := linearizeSRGB(float64(c.R) / 255)
	g := linearizeSRGB(float64(c.G) / 255)
	b := linearizeSRGB(float64(c.B) / 255)
	return 0.2126*r + 0.7152*g + 0.0722*b
}

// ContrastRatio returns the WCAG contrast ratio for two rendered sRGB colors.
// It ranges from 1:1 to 21:1. Applying a ratio to a WCAG criterion requires
// separate context, including the criterion, text size, and its exceptions;
// this function does not classify text or establish conformance.
func ContrastRatio(a, b RGB) float64 {
	y1, y2 := RelativeLuminance(a), RelativeLuminance(b)
	if y1 < y2 {
		y1, y2 = y2, y1
	}
	return (y1 + 0.05) / (y2 + 0.05)
}

func linearizeSRGB(channel float64) float64 {
	if channel <= 0.04045 {
		return channel / 12.92
	}
	return math.Pow((channel+0.055)/1.055, 2.4)
}
