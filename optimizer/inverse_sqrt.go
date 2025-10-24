package optimizer

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/hm-choi/pp-stat-plus/engine"
	"github.com/hm-choi/pp-stat-plus/utils"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)


func ChebyshevInvSqrt_deg_log(e *engine.HEEngine, ct *engine.HEData, mode int, B float64, deg float64) (*engine.HEData, error) {
	
	cpData := ct.CopyData()
	if cpData.Level() - int(deg) < 0 {
		if e.IsBTS {
			cpData, _ = e.DoBootstrap(cpData, e.Params().MaxLevel())
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
	gcbsp := engine.GetChebyshevPoly(1.0, int(math.Pow(2, float64(d))-2), F)
	poly := polynomial.NewPolynomial(gcbsp)
	polyEval := polynomial.NewEvaluator(e.Params(), e.Evaluator())

	scaledCtxts := scaled_ct.Ciphertexts()
	invCtxts := []*rlwe.Ciphertext{}
	targetScale := e.Params().DefaultScale().Div(rlwe.NewScale(2))
	for i := 0; i < len(scaledCtxts); i++ {
		p2, _ := polyEval.Evaluate(scaledCtxts[i], poly, targetScale)
		p2.Scale = p2.Scale.Mul(rlwe.NewScale(2))
		conj, _ := e.Evaluator().ConjugateNew(p2)
		e.Evaluator().Add(p2, conj, p2)
		p2.Scale = scaledCtxts[i].Scale
		invCtxts = append(invCtxts, p2)
	}
	return engine.NewHEData(invCtxts, ct.Size(), invCtxts[0].Level(), ct.Scale()), nil
}

func HENewtonInv_log(e *engine.HEEngine, ct, init *engine.HEData, B float64, iter, mode int, inv_ans []float64, start time.Time) (error) {

	N := 1.0
	x, y := ct.CopyData(), init.CopyData()

	switch mode {
	case 1:
		x, _ = e.MultConst(x, B)
	case 2:
		N = 2
		// x, _ = e.MultConst(x, 1.0/N)
	case 3:
		N = 2
		x, _ = e.MultConst(x, B/N)
	}

	for i := range iter {
		
		if e.IsBTS {
			y, _ = e.DoBootstrap(y, 2)
		}	

		tmp_a_c, _ := e.MultConst(y, float64((N+1))/float64(N))
		tmp_b_c, _ := e.Mult(x, y)

		if N == 2.0 {
			y, _ = e.Mult(y, y)
		}

		tmp_b_c, _ = e.Mult(tmp_b_c, y)
		y, _ = e.Sub(tmp_a_c, tmp_b_c)
		
		elapsed := time.Since(start).Seconds()

		after_iter_c, _ := e.Decrypt(y)
		
		log.Println("Iter", i+1)
		log.Println("Level x -", x.Level(), "y -", y.Level())
		log.Println("Time", elapsed)

		_, MRE  := utils.CheckMRE(after_iter_c, after_iter_c, inv_ans, ct.Size())
		log.Println("MRE(CT)", MRE)

		log.Println()
	}

	return nil
}

func CryptoInvSqrt_log(e *engine.HEEngine, ct *engine.HEData, scaled_ct *engine.HEData, B float64, deg float64, i_max int, inv_ans []float64, start time.Time) (error) {

	if scaled_ct.Level() - int(deg) < 0 {
		if e.IsBTS {
			scaled_ct, _ = e.DoBootstrap(scaled_ct, e.Params().MaxLevel())
		}
	}

	y, _ := ChebyshevInvSqrt_deg_log(e, scaled_ct, 1, B, deg)


	if e.IsBTS {
		y, _ = e.DoBootstrap(y, 2)
	}

	return HENewtonInv_log(e, ct, y, B, i_max, 2, inv_ans, start)
}