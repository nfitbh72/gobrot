package bigfloat

import (
	"fmt"
	"math/big"
)

type TBigFloats struct {
	Prec uint
}

func (m *TBigFloats) NewBigS(s string) *big.Float {
	bf := m.NewBig0()
	_, _ = fmt.Sscan(s, bf)
	return bf
}

func (m *TBigFloats) NewBig(f float64) *big.Float {
	return big.NewFloat(f).SetPrec(m.Prec)
}

func (m *TBigFloats) NewBig0() *big.Float {
	return m.NewBig(0.0)
}

func (m *TBigFloats) MulBig(n1 *big.Float, n2 *big.Float) *big.Float {
	return m.NewBig0().Mul(n1, n2)
}

func (m *TBigFloats) DivBig(n1 *big.Float, n2 *big.Float) *big.Float {
	return m.NewBig0().Quo(n1, n2)
}

func (m *TBigFloats) AddBig(n1 *big.Float, n2 *big.Float) *big.Float {
	return m.NewBig0().Add(n1, n2)
}

func (m *TBigFloats) SubBig(n1 *big.Float, n2 *big.Float) *big.Float {
	return m.NewBig0().Sub(n1, n2)
}
