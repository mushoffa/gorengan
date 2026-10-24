package widget

import (
	"math"
)

func minOf(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	m := v[0]
	for _, x := range v {
		if x < m {
			m = x
		}
	}
	return m
}

func maxOf(v []float64) float64 {
	m := 0.0
	for _, x := range v {
		if x > m {
			m = x
		}
	}
	return m
}

func minf(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func maxf(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

// tidy removes floating point noise such as 0.6000000000000001.
func tidy(v float64) float64 {
	r := math.Round(v*1e9) / 1e9
	if r == 0 {
		return 0 // avoid "-0"
	}
	return r
}

// niceStep returns a round tick spacing (1, 2 or 5 x 10^n) giving about 5 ticks.
func niceStep(span float64) float64 {
	if span <= 0 {
		return 1
	}
	raw := span / 5
	pow := math.Pow(10, math.Floor(math.Log10(raw)))
	switch f := raw / pow; {
	case f <= 1:
		return pow
	case f <= 2:
		return 2 * pow
	case f <= 5:
		return 5 * pow
	default:
		return 10 * pow
	}
}
