package layers

import "math"

// ApplyLogCurve returns a copy of the layers with the log curve applied to each material height.
//
// Layer values are depths where 1.0 means "no material" (see GetLayers) and 1-v is the height of the
// material. The height h becomes log(1+k*h)/log(1+k), which keeps 0 and 1 fixed but spreads out small
// heights. Those belong to pixels that escaped after many iterations, which sit almost flat on the base
// in a linear mapping, so the structure near the set boundary gets visible relief. k <= 0 means linear:
// the layers are returned as they are. Values at or above 1.0 are left alone so empty stays empty.
func ApplyLogCurve(layers [][][]float64, k float64) [][][]float64 {
	if k <= 0 {
		return layers
	}
	scale := math.Log1p(k)
	out := make([][][]float64, len(layers))
	for n, layer := range layers {
		out[n] = make([][]float64, len(layer))
		for i, column := range layer {
			out[n][i] = make([]float64, len(column))
			for j, v := range column {
				if v < 1.0 {
					h := math.Log1p(k*(1.0-v)) / scale
					v = 1.0 - h
				}
				out[n][i][j] = v
			}
		}
	}
	return out
}
