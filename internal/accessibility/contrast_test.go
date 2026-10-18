package accessibility

import (
	"math"
	"testing"
)

func TestRelativeLuminanceAndContrastRatio(t *testing.T) {
	cases := []struct {
		name  string
		color RGB
		want  float64
	}{
		{name: "black", color: RGB{}, want: 0},
		{name: "white", color: RGB{R: 255, G: 255, B: 255}, want: 1},
		{name: "red", color: RGB{R: 255}, want: 0.2126},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := RelativeLuminance(tc.color); math.Abs(got-tc.want) > 0.0001 {
				t.Errorf("RelativeLuminance(%+v) = %.7f, want %.7f", tc.color, got, tc.want)
			}
		})
	}

	black := RGB{}
	white := RGB{R: 255, G: 255, B: 255}
	if got := ContrastRatio(black, white); got != 21 {
		t.Errorf("ContrastRatio(black, white) = %v, want 21", got)
	}
	if got := ContrastRatio(white, black); got != 21 {
		t.Errorf("ContrastRatio(white, black) = %v, want 21", got)
	}
}

func TestWCAG22ContrastThresholdContexts(t *testing.T) {
	white := RGB{R: 255, G: 255, B: 255}
	mediumGray := RGB{R: 118, G: 118, B: 118} // 4.54:1 against white.
	ratio := ContrastRatio(mediumGray, white)

	// These are explicit WCAG 2.2 threshold contexts. The calculation does not
	// infer text size, weight, content type, exceptions, or conformance.
	contexts := []struct {
		name      string
		threshold float64
		wantPass  bool
	}{
		{name: "SC 1.4.3 normal text AA", threshold: 4.5, wantPass: true},
		{name: "SC 1.4.3 large text AA", threshold: 3, wantPass: true},
		{name: "SC 1.4.6 normal text AAA", threshold: 7, wantPass: false},
		{name: "SC 1.4.6 large text AAA", threshold: 4.5, wantPass: true},
		{name: "SC 1.4.11 non-text contrast", threshold: 3, wantPass: true},
	}
	for _, tc := range contexts {
		t.Run(tc.name, func(t *testing.T) {
			got := ratio >= tc.threshold
			if got != tc.wantPass {
				t.Errorf("contrast ratio %.3f against threshold %.1f = %v, want %v", ratio, tc.threshold, got, tc.wantPass)
			}
		})
	}
}
