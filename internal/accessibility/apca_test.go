package accessibility

import (
	"image/color"
	"math"
	"strconv"
	"strings"
	"testing"
)

// Reference pairs below are the published APCA-W3 keystone checks
// ("TESTING YOUR IMPLEMENTATION") from the canonical APCA documentation at
// https://git.apcacontrast.com/documentation/README.html — "keystone checks
// with no rounding. The first color is TEXT and the second color is
// BACKGROUND", valid for APCA W3 to 0.1.9 with the G-series constants.
func TestContrastAPCAW3KeystoneValues(t *testing.T) {
	cases := []struct {
		name string
		fg   color.Color // text
		bg   color.Color // background
		want float64
	}{
		{"dark gray on white", mustHex("#888888"), mustHex("#ffffff"), 63.056469930209424},
		{"white on dark gray", mustHex("#ffffff"), mustHex("#888888"), -68.54146436644962},
		{"black on mid gray", mustHex("#000000"), mustHex("#aaaaaa"), 58.146262578561334},
		{"mid gray on black", mustHex("#aaaaaa"), mustHex("#000000"), -56.24113336839742},
		{"dark slate on pale", mustHex("#112233"), mustHex("#ddeeff"), 91.66830811481631},
		{"pale on dark slate", mustHex("#ddeeff"), mustHex("#112233"), -93.06770049484275},
		{"dark slate on gray", mustHex("#112233"), mustHex("#444444"), 8.32326136957393},
		{"gray on dark slate", mustHex("#444444"), mustHex("#112233"), -7.526878460278154},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Contrast(tc.fg, tc.bg)
			if math.Abs(got-tc.want) > 1e-6 {
				t.Errorf("Contrast() = %v, want %v", got, tc.want)
			}
		})
	}
}

// The extremes of the Lc scale for APCA-W3 (same canonical source): black
// text on a white background is the maximum normal-polarity value (~Lc 106)
// and white text on a black background is the maximum reverse-polarity value
// (~Lc -107). Polarity is carried by the sign: positive means dark text on a
// light background, negative means light text on a dark background.
func TestContrastPolarityExtremes(t *testing.T) {
	if got := Contrast(color.Black, color.White); math.Abs(got-106) > 1 {
		t.Errorf("Contrast(black, white) = %v, want approximately 106", got)
	}
	if got := Contrast(color.White, color.Black); math.Abs(got+107) > 1 {
		t.Errorf("Contrast(white, black) = %v, want approximately -107", got)
	}
}

// Pairs below the low-clip threshold report Lc 0 ("contrast too low"): no
// readable difference is predicted, and nothing near a threshold boundary is
// implied.
func TestContrastIdenticalColorsReportsZero(t *testing.T) {
	for _, c := range []color.Color{color.Black, color.White, mustHex("#808080")} {
		if got := Contrast(c, c); got != 0 {
			t.Errorf("Contrast(same color) = %v, want 0", got)
		}
	}
}

func mustHex(s string) color.Color {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		panic("test helper expects #rrggbb")
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		panic(err)
	}
	return color.NRGBA{
		R: uint8(v >> 16),
		G: uint8(v >> 8),
		B: uint8(v),
		A: 0xff,
	}
}
