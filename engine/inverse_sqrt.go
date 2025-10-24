package engine

import (
	"fmt"
	"math"
	"math/big"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

// GetChebyshevPoly returns the Chebyshev polynomial approximation of f the
// in the interval [-K, K] for the given degree.
func GetChebyshevPoly(K float64, degree int, f64 func(x float64) (y float64)) bignum.Polynomial {

	FBig := func(x *big.Float) (y *big.Float) {
		xF64, _ := x.Float64()
		return new(big.Float).SetPrec(x.Prec()).SetFloat64(f64(xF64))
	}

	var prec uint = 128

	interval := bignum.Interval{
		A:     *bignum.NewFloat(-K, prec),
		B:     *bignum.NewFloat(K, prec),
		Nodes: degree,
	}
	// Returns the polynomial.
	return bignum.ChebyshevApproximation(FBig, interval)
}

func (e *HEEngine) ChebyshevInvSqrt(ct *HEData, mode int, B float64) (*HEData, error) {
	cpData := ct.CopyData()
	d := 9.0

	var F func(float64) float64

	switch mode {
	case 1:
		cpData, _ = e.MultConst(cpData, 2.0/B)
		F = func(x float64) (y float64) {
			if x > -1.0 {
				return 1 / math.Sqrt(B/2) / (math.Sqrt(x + 1.0))
			} else {
				return 0
			}
		}
	case 2:
		F = func(x float64) (y float64) {
			if x > -1.0 {
				return 1 / math.Sqrt(B) / (math.Sqrt(x + 1.0))
			} else {
				return 0
			}
		}
	case 0:
		F = func(x float64) float64 {
			if x > -1.0 {
				return 1 / math.Sqrt(x+1.0)
			}
			return 0
		}
	default:
		return nil, fmt.Errorf("invalid InvSqrt mode: %d", mode)

	}

	scaled_ct, err := e.SubConst(cpData, 1)
	if err != nil {
		return nil, err
	}
	gcbsp := GetChebyshevPoly(1.0, int(math.Pow(2, float64(d))-2), F)
	poly := polynomial.NewPolynomial(gcbsp)
	polyEval := polynomial.NewEvaluator(e.params, e.Evaluator())

	scaledCtxts := scaled_ct.Ciphertexts()
	invCtxts := []*rlwe.Ciphertext{}
	targetScale := e.params.DefaultScale().Div(rlwe.NewScale(2))
	for i := 0; i < len(scaledCtxts); i++ {
		p2, _ := polyEval.Evaluate(scaledCtxts[i], poly, targetScale)
		p2.Scale = p2.Scale.Mul(rlwe.NewScale(2))
		conj, _ := e.evaluator.ConjugateNew(p2)
		e.evaluator.Add(p2, conj, p2)
		p2.Scale = scaledCtxts[i].Scale
		invCtxts = append(invCtxts, p2)
	}
	return NewHEData(invCtxts, ct.Size(), invCtxts[0].Level(), ct.Scale()), nil
}

func (e *HEEngine) ChebyshevInvSqrt_deg(ct *HEData, mode int, B float64, deg float64) (*HEData, error) {
	
	cpData := ct.CopyData()
	if cpData.Level() - int(deg) < 0 {
		if e.IsBTS {
			cpData, _ = e.DoBootstrap(cpData, e.params.MaxLevel())
		}
	}
	
	d := deg

	var F func(float64) float64

	switch mode {
	case 1:
		F = func(x float64) (y float64) {
			if x > -1.0 {
				return 1 / math.Sqrt(B/2) / (math.Sqrt(x + 1.0))
			} else {
				return 0
			}
		}
	case 2:
		F = func(x float64) (y float64) {
			if x > -1.0 {
				return 1 / math.Sqrt(B) / (math.Sqrt(x + 1.0))
			} else {
				return 0
			}
		}
	case 0:
		F = func(x float64) float64 {
			if x > -1.0 {
				return 1 / math.Sqrt(x+1.0)
			}
			return 0
		}
	default:
		return nil, fmt.Errorf("invalid InvSqrt mode: %d", mode)

	}

	scaled_ct, err := e.SubConst(cpData, 1)
	if err != nil {
		return nil, err
	}
	gcbsp := GetChebyshevPoly(1.0, int(math.Pow(2, float64(d))-2), F)
	poly := polynomial.NewPolynomial(gcbsp)
	polyEval := polynomial.NewEvaluator(e.params, e.Evaluator())

	scaledCtxts := scaled_ct.Ciphertexts()
	invCtxts := []*rlwe.Ciphertext{}
	targetScale := e.params.DefaultScale().Div(rlwe.NewScale(2))
	for i := 0; i < len(scaledCtxts); i++ {
		p2, _ := polyEval.Evaluate(scaledCtxts[i], poly, targetScale)
		p2.Scale = p2.Scale.Mul(rlwe.NewScale(2))
		conj, _ := e.evaluator.ConjugateNew(p2)
		e.evaluator.Add(p2, conj, p2)
		p2.Scale = scaledCtxts[i].Scale
		invCtxts = append(invCtxts, p2)
	}
	return NewHEData(invCtxts, ct.Size(), invCtxts[0].Level(), ct.Scale()), nil
}

func (e *HEEngine) HENewtonInv(ct, init *HEData, B float64, iter, mode int) (*HEData, error) {
	N := 1.0
	x, y := ct.CopyData(), init.CopyData()
	switch mode {
	case 1:
		x, _ = e.MultConst(x, B)
	case 2:
		N = 2
	case 3:
		N = 2
		x, _ = e.MultConst(x, B/N)
	}

	if e.IsBTS {
		x, _ = e.DoBootstrap(x, 2)
	}

	for _ = range iter {
		if e.IsBTS {
			y, _ = e.DoBootstrap(y, 2)
		}

		tmp_a, _ := e.MultConst(y, float64((N+1))/float64(N))
		tmp_b, _ := e.Mult(x, y)

		if N == 2.0 {
			y, _ = e.Mult(y, y)
		}
		tmp_b, _ = e.Mult(tmp_b, y)
		y, _ = e.Sub(tmp_a, tmp_b)
	}
	return y, nil
}

func (e *HEEngine) CryptoInvSqrt(ct *HEData, scaled_ct *HEData, B float64, deg float64, iter int, cheb_mode int, nt_mode int) (*HEData, error) {
	
	y, err := e.ChebyshevInvSqrt_deg(scaled_ct, cheb_mode, B, deg)
	if err != nil {
		return y, err
	}

	return e.HENewtonInv(ct, y, B, iter, nt_mode)
}