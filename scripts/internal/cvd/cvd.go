package cvd

import (
	"math"

	"github.com/sunagawasei/dotfiles/scripts/internal/colorutil"
)

type Type struct {
	Name   string
	Matrix [3][3]float64
}

var Types = []Type{
	{
		Name: "protanopia",
		Matrix: [3][3]float64{
			{0.152286, 1.052583, -0.204868},
			{0.114503, 0.786281, 0.099216},
			{-0.003882, -0.048116, 1.051998},
		},
	},
	{
		Name: "deuteranopia",
		Matrix: [3][3]float64{
			{0.367322, 0.860646, -0.227968},
			{0.280085, 0.672501, 0.047413},
			{-0.011820, 0.042940, 0.968881},
		},
	},
	{
		Name: "tritanopia",
		Matrix: [3][3]float64{
			{1.255528, -0.076749, -0.178779},
			{-0.078411, 0.930809, 0.147602},
			{0.004733, 0.691367, 0.303900},
		},
	},
}

// Apply simulates one Machado CVD matrix and returns CIELAB coordinates.
func Apply(hex string, matrix [3][3]float64) ([3]float64, error) {
	r, g, b, err := colorutil.StrictRGB01(hex)
	if err != nil {
		return [3]float64{}, err
	}
	rl := colorutil.SRGBToLinear(r)
	gl := colorutil.SRGBToLinear(g)
	bl := colorutil.SRGBToLinear(b)
	rl2 := matrix[0][0]*rl + matrix[0][1]*gl + matrix[0][2]*bl
	gl2 := matrix[1][0]*rl + matrix[1][1]*gl + matrix[1][2]*bl
	bl2 := matrix[2][0]*rl + matrix[2][1]*gl + matrix[2][2]*bl
	rl2 = colorutil.Clamp01(rl2)
	gl2 = colorutil.Clamp01(gl2)
	bl2 = colorutil.Clamp01(bl2)
	return LinearRGBToLab(rl2, gl2, bl2), nil
}

// LinearRGBToLab converts linear sRGB with a D65 white point to CIELAB.
func LinearRGBToLab(r, g, b float64) [3]float64 {
	x := 0.4124564*r + 0.3575761*g + 0.1804375*b
	y := 0.2126729*r + 0.7151522*g + 0.0721750*b
	z := 0.0193339*r + 0.1191920*g + 0.9503041*b
	xn, yn, zn := 0.95047, 1.0, 1.08883
	f := func(value float64) float64 {
		d := 6.0 / 29.0
		if value > d*d*d {
			return math.Cbrt(value)
		}
		return value/(3*d*d) + 4.0/29.0
	}
	fx, fy, fz := f(x/xn), f(y/yn), f(z/zn)
	l := 116*fy - 16
	a := 500 * (fx - fy)
	bb := 200 * (fy - fz)
	return [3]float64{l, a, bb}
}

func DeltaE76(first, second [3]float64) float64 {
	dl := first[0] - second[0]
	da := first[1] - second[1]
	db := first[2] - second[2]
	return math.Sqrt(dl*dl + da*da + db*db)
}
