package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

type CaseData struct {
    Case      int  	  `json:"case"`
    Degree    float64 `json:"degree"`
    Iteration int     `json:"iteration"`
    Time      float64 `json:"time"`
    MRE       float64 `json:"mre"`
}

type LevelData map[string]CaseData

func Get_deg_and_iter(level_int int, fast bool) (float64, int, int){
	cs, deg, iter := 0, 0.0, 0

	data, err := os.ReadFile("../../optimizer/result/lattigo_optimizer.json")
	if err != nil {
		log.Fatal(err)
	}

	var result map[string]LevelData
	if err := json.Unmarshal(data, &result); err != nil {
		log.Fatal(err)
	}

	level := strconv.Itoa(level_int + 1)
	if levelCases, ok := result[level]; ok {
		var chosen CaseData
		found := false

		if fast {
			if caseData, ok := levelCases["Fast"]; ok {
				chosen = caseData
				found = true
			}
		} else {
			if caseData, ok := levelCases["Basic"]; ok {
				chosen = caseData
				found = true
			}
		}

		if !found {
			for _, caseData := range levelCases {
				chosen = caseData
				break
			}
		}

		deg = chosen.Degree
		iter = chosen.Iteration
		cs = chosen.Case
	}
	return deg, iter, cs
}

func (e *HEEngine) ZScoreNorm(ct *HEData, B float64, fast bool) (*HEData, error) {
	const (
		chebyshevDegree = 2 // Degree for initial Chebyshev approximation
		newtonScale     = 2 // Scaling for Newton method
	)

	// Step 1: Compute mean μ
	mean, err := e.Mean(ct)
	if err != nil {
		return nil, fmt.Errorf("compute mean: %w", err)
	}

	// Step 2: Center data => X - μ
	centered, err := e.Sub(ct, mean)
	if err != nil {
		return nil, fmt.Errorf("center input: %w", err)
	}

	
	// Step 3: Compute inverse of standard deviation 1/σ using invSqrt
	invSigmaRefined, err := e.computeInvStd(ct, fast, newtonScale, B)
	if err != nil {
		return nil, fmt.Errorf("HENewtonInv: %w", err)
	}

	// Step 4: Extend 1/σ to all slots for element-wise multiplication
	invSigmaSlots, err := e.extendOneToMulty(invSigmaRefined, len(centered.Ciphertexts()), centered.Size())
	if err != nil {
		return nil, fmt.Errorf("extendOneToMulty: %w", err)
	}

	// Step 5: Final Z-score normalization: Z = (X - μ) × (1/σ)
	zscore, err := e.Mult(centered, invSigmaSlots)
	if err != nil {
		return nil, fmt.Errorf("final multiply: %w", err)
	}

	return zscore, nil
}

func (e *HEEngine) Kurtosis(ct *HEData, B float64, fast bool) (*HEData, error) {
	const (
		chebyshevDegree = 2
		newtonScale     = 2
		subtractBias    = 3.0 // for excess kurtosis
	)
	// Step 1: Compute mean μ
	mean, err := e.Mean(ct)
	if err != nil {
		return nil, fmt.Errorf("mean: %w", err)
	}

	// Step 2: Centered input: x = X - μ
	centered, err := e.Sub(ct, mean)
	if err != nil {
		return nil, fmt.Errorf("center: %w", err)
	}

	// Step 3: Compute x² and x⁴
	x2, err := e.Mult(centered, centered)
	if err != nil {
		return nil, fmt.Errorf("x^2: %w", err)
	}

	x4, err := e.Mult(x2, x2)
	if err != nil {
		return nil, fmt.Errorf("x^4: %w", err)
	}

	// Step 4: Compute E[x⁴] (numerator)
	numerator, err := e.Mean(x4)
	if err != nil {
		return nil, fmt.Errorf("mean of x^4: %w", err)
	}

	// Step 5: Compute inverse of standard deviation 1/σ using invSqrt
	invSigmaRefined, err := e.computeInvStd(ct, fast, newtonScale, B)
	if err != nil {
		return nil, fmt.Errorf("HENewtonInv: %w", err)
	}

	// Step ６: Compute 1/σ⁴ = (1/σ)² × (1/σ)²
	invSigma2, err := e.Mult(invSigmaRefined, invSigmaRefined)
	if err != nil {
		return nil, fmt.Errorf("inv σ²: %w", err)
	}

	invSigma4, err := e.Mult(invSigma2, invSigma2)
	if err != nil {
		return nil, fmt.Errorf("inv σ⁴: %w", err)
	}

	// Step 7: Apply inv(σ⁴) to numerator: E[x⁴] × (1/σ⁴)
	invSigma4Expanded, err := e.extendOneToMulty(invSigma4, len(numerator.Ciphertexts()), numerator.Size())
	if err != nil {
		return nil, fmt.Errorf("extendOneToMulty (inv σ⁴): %w", err)
	}

	if e.IsBTS {
		invSigma4Expanded, _ = e.DoBootstrap(invSigma4Expanded, 1)
	}

	kurtosis, err := e.Mult(numerator, invSigma4Expanded)
	if err != nil {
		return nil, fmt.Errorf("final multiply: %w", err)
	}

	// Step 8: Convert to excess kurtosis: K − 3
	kurtosis, err = e.SubConst(kurtosis, subtractBias)
	if err != nil {
		return nil, fmt.Errorf("subtract bias: %w", err)
	}

	return kurtosis, nil
}

func (e *HEEngine) Skewness(ct *HEData, B float64, fast bool) (*HEData, error) {
	const (
		chebyshevDegree = 2
		newtonScale     = 2
	)
	// Step 1: Compute mean μ
	mean, err := e.Mean(ct)
	if err != nil {
		return nil, fmt.Errorf("compute mean: %w", err)
	}

	// Step 2: Centered data x = X - μ
	centered, err := e.Sub(ct, mean)
	if err != nil {
		return nil, fmt.Errorf("center data: %w", err)
	}

	// Step 3: Compute x² and x³
	x2, err := e.Mult(centered, centered)
	if err != nil {
		return nil, fmt.Errorf("compute x^2: %w", err)
	}

	x3, err := e.Mult(centered, x2)
	if err != nil {
		return nil, fmt.Errorf("compute x^3: %w", err)
	}

	// Step 4: Compute E[x³]
	numerator, err := e.Mean(x3)
	if err != nil {
		return nil, fmt.Errorf("mean of x^3: %w", err)
	}

	// Step 5: Compute inverse of standard deviation 1/σ using invSqrt
	invSigmaRefined, err := e.computeInvStd(ct, fast, newtonScale, B)
	if err != nil {
		return nil, fmt.Errorf("HENewtonInv: %w", err)
	}

	// Step 6: Compute (1/σ)³ = (1/σ) × (1/σ)²
	invSigma2, err := e.Mult(invSigmaRefined, invSigmaRefined)
	if err != nil {
		return nil, fmt.Errorf("inv σ²: %w", err)
	}

	invSigma3, err := e.Mult(invSigmaRefined, invSigma2)
	if err != nil {
		return nil, fmt.Errorf("inv σ³: %w", err)
	}

	// Step 7: Final result: E[x³] × (1/σ³)
	invSigma3Expanded, err := e.extendOneToMulty(invSigma3, len(numerator.Ciphertexts()), numerator.Size())
	if err != nil {
		return nil, fmt.Errorf("extendOneToMulty: %w", err)
	}

	if e.IsBTS {
		invSigma3Expanded, _ = e.DoBootstrap(invSigma3Expanded, 1)
	}

	skewness, err := e.Mult(numerator, invSigma3Expanded)
	if err != nil {
		return nil, fmt.Errorf("final multiply: %w", err)
	}

	return skewness, nil
}

func (e *HEEngine) PCorrCoeff(ct1, ct2 *HEData, B float64, fast bool) (*HEData, error) {
	const (
		chebyshevDegree = 2
		newtonScale     = 2
	)

	// Step 1: Compute means
	meanX, err := e.Mean(ct1)
	if err != nil {
		return nil, fmt.Errorf("mean of ct1: %w", err)
	}
	meanY, err := e.Mean(ct2)
	if err != nil {
		return nil, fmt.Errorf("mean of ct2: %w", err)
	}

	// Step 2: Centered data
	xCentered, err := e.Sub(ct1, meanX)
	if err != nil {
		return nil, fmt.Errorf("center ct1: %w", err)
	}
	yCentered, err := e.Sub(ct2, meanY)
	if err != nil {
		return nil, fmt.Errorf("center ct2: %w", err)
	}

	// Step 3: Numerator = E[(X - μx)(Y - μy)]
	mulXY, err := e.Mult(xCentered, yCentered)
	if err != nil {
		return nil, fmt.Errorf("x·y: %w", err)
	}
	numerator, err := e.Mean(mulXY)
	if err != nil {
		return nil, fmt.Errorf("mean of x·y: %w", err)
	}

	// Step 4: Compute inverse std for X
	invStdX, err := e.computeInvStd(ct1, fast, newtonScale, B)
	if err != nil {
		return nil, fmt.Errorf("computeInvStd (X): %w", err)
	}

	// Step 5: Compute inverse std for Y
	invStdY, err := e.computeInvStd(ct2, fast, newtonScale, B)
	if err != nil {
		return nil, fmt.Errorf("computeInvStd (Y): %w", err)
	}

	// Step 6: Compute PCC = numerator × (1/σx) × (1/σy)
	denominator, err := e.Mult(invStdX, invStdY)
	if err != nil {
		return nil, fmt.Errorf("σx·σy inverse: %w", err)
	}

	denominatorExpanded, err := e.extendOneToMulty(denominator, len(numerator.Ciphertexts()), numerator.Size())
	if err != nil {
		return nil, fmt.Errorf("extendOneToMulty: %w", err)
	}

	pcc, err := e.Mult(numerator, denominatorExpanded)
	if err != nil {
		return nil, fmt.Errorf("final multiply: %w", err)
	}

	return pcc, nil
}

func (e *HEEngine) ZScoreNorm_ppstat(ct *HEData, B float64) (*HEData, error) {
	const (
		chebyshevDegree = 2 // Degree for initial Chebyshev approximation
		newtonIter      = 5 // Iteration count for Newton refinement
		newtonScale     = 2 // Scaling for Newton method
		bootstrapDepth  = 3 // Depth used when bootstrapping initial guess
	)

	// Step 1: Compute mean μ
	mean, err := e.Mean(ct)
	if err != nil {
		return nil, fmt.Errorf("compute mean: %w", err)
	}

	// Step 2: Approximate variance Var(X)
	denom := float64(ct.Size()) * B
	varianceApprox, err := varianceWithCustomDenom(e, ct, denom, denom*B)
	if err != nil {
		return nil, fmt.Errorf("compute variance (approx): %w", err)
	}

	// Step 3: Center data => X - μ
	centered, err := e.Sub(ct, mean)
	if err != nil {
		return nil, fmt.Errorf("center input: %w", err)
	}

	// Select representative ciphertext for approximation
	varApproxCtxt, err := e.selectOneCtxt(varianceApprox)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt (variance approx): %w", err)
	}

	// Step 4: Initial guess for 1/σ using Chebyshev

	// Optional: Bootstrap the initial guess for higher precision
	if e.IsBTS {
		varApproxCtxt, err = e.DoBootstrap(varApproxCtxt, 9)
		if err != nil {
			return nil, fmt.Errorf("bootstrap (invSqrt init): %w", err)
		}
	}

	invSigmaInit, err := e.ChebyshevInvSqrt(varApproxCtxt, chebyshevDegree, B*B)
	if err != nil {
		return nil, fmt.Errorf("ChebyshevInvSqrt: %w", err)
	}

	// Optional: Bootstrap the initial guess for higher precision
	if e.IsBTS {
		invSigmaInit, err = e.DoBootstrap(invSigmaInit, bootstrapDepth)
		if err != nil {
			return nil, fmt.Errorf("bootstrap (invSqrt init): %w", err)
		}
	}

	// Step 5: Refine variance and compute Newton-based 1/σ
	denom = float64(ct.Size()) * math.Sqrt(2)
	varianceRefined, err := varianceWithCustomDenom(e, ct, denom, denom * math.Sqrt(2))
	if err != nil {
		return nil, fmt.Errorf("compute refined variance: %w", err)
	}
	varRefinedCtxt, err := e.selectOneCtxt(varianceRefined)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt (variance refined): %w", err)
	}

	invSigmaRefined, err := e.HENewtonInv(varRefinedCtxt, invSigmaInit, B, newtonIter, newtonScale)
	if err != nil {
		return nil, fmt.Errorf("HENewtonInv: %w", err)
	}

	// Step 6: Extend 1/σ to all slots for element-wise multiplication
	invSigmaSlots, err := e.extendOneToMulty(invSigmaRefined, len(centered.Ciphertexts()), centered.Size())
	if err != nil {
		return nil, fmt.Errorf("extendOneToMulty: %w", err)
	}

	// Step 7: Final Z-score normalization: Z = (X - μ) × (1/σ)
	zscore, err := e.Mult(centered, invSigmaSlots)
	if err != nil {
		return nil, fmt.Errorf("final multiply: %w", err)
	}

	return zscore, nil
}

func (e *HEEngine) Kurtosis_ppstat(ct *HEData, B float64) (*HEData, error) {
	const (
		chebyshevDegree = 2
		newtonIter      = 5
		newtonScale     = 2
		bootstrapDepth  = 3
		subtractBias    = 3.0 // for excess kurtosis
	)

	// Step 1: Compute mean μ
	mean, err := e.Mean(ct)
	if err != nil {
		return nil, fmt.Errorf("mean: %w", err)
	}

	// Step 2: Centered input: x = X - μ
	centered, err := e.Sub(ct, mean)
	if err != nil {
		return nil, fmt.Errorf("center: %w", err)
	}

	// Step 3: Compute x² and x⁴
	x2, err := e.Mult(centered, centered)
	if err != nil {
		return nil, fmt.Errorf("x^2: %w", err)
	}

	x4, err := e.Mult(x2, x2)
	if err != nil {
		return nil, fmt.Errorf("x^4: %w", err)
	}

	// Step 4: Compute E[x⁴] (numerator)
	numerator, err := e.Mean(x4)
	if err != nil {
		return nil, fmt.Errorf("mean of x^4: %w", err)
	}

	// Step 5: Approximate variance Var(X)
	denom := float64(ct.Size()) * B
	varianceApprox, err := varianceWithCustomDenom(e, ct, denom, denom*B)
	if err != nil {
		return nil, fmt.Errorf("variance (approx): %w", err)
	}

	varApproxCtxt, err := e.selectOneCtxt(varianceApprox)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt (variance approx): %w", err)
	}

	// Step 6: Initial approximation of 1/σ using Chebyshev

	// Optional: Bootstrap the initial guess for higher precision
	if e.IsBTS {
		varApproxCtxt, err = e.DoBootstrap(varApproxCtxt, 9)
		if err != nil {
			return nil, fmt.Errorf("bootstrap (invSqrt init): %w", err)
		}
	}
	invSigmaInit, err := e.ChebyshevInvSqrt(varApproxCtxt, chebyshevDegree, B*B)
	if err != nil {
		return nil, fmt.Errorf("ChebyshevInvSqrt: %w", err)
	}

	if e.IsBTS {
		invSigmaInit, err = e.DoBootstrap(invSigmaInit, bootstrapDepth)
		if err != nil {
			return nil, fmt.Errorf("bootstrap (Chebyshev init): %w", err)
		}
	}

	// Step 7: Refine 1/σ using Newton method
	denom = float64(ct.Size()) * math.Sqrt(2)
	varianceRefined, err := varianceWithCustomDenom(e, ct, denom, denom * math.Sqrt(2))
	if err != nil {
		return nil, fmt.Errorf("compute refined variance: %w", err)
	}
	varRefinedCtxt, err := e.selectOneCtxt(varianceRefined)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt (variance refined): %w", err)
	}

	invSigma, err := e.HENewtonInv(varRefinedCtxt, invSigmaInit, B, newtonIter, newtonScale)
	if err != nil {
		return nil, fmt.Errorf("HENewtonInv: %w", err)
	}

	// Step 8: Compute 1/σ⁴ = (1/σ)² × (1/σ)²
	invSigma2, err := e.Mult(invSigma, invSigma)
	if err != nil {
		return nil, fmt.Errorf("inv σ²: %w", err)
	}

	invSigma4, err := e.Mult(invSigma2, invSigma2)
	if err != nil {
		return nil, fmt.Errorf("inv σ⁴: %w", err)
	}

	// Step 9: Apply inv(σ⁴) to numerator: E[x⁴] × (1/σ⁴)
	invSigma4Expanded, err := e.extendOneToMulty(invSigma4, len(numerator.Ciphertexts()), numerator.Size())
	if err != nil {
		return nil, fmt.Errorf("extendOneToMulty (inv σ⁴): %w", err)
	}

	kurtosis, err := e.Mult(numerator, invSigma4Expanded)
	if err != nil {
		return nil, fmt.Errorf("final multiply: %w", err)
	}

	// Step 10: Convert to excess kurtosis: K − 3
	kurtosis, err = e.SubConst(kurtosis, subtractBias)
	if err != nil {
		return nil, fmt.Errorf("subtract bias: %w", err)
	}

	return kurtosis, nil
}

func (e *HEEngine) Skewness_ppstat(ct *HEData, B float64) (*HEData, error) {
	const (
		chebyshevDegree = 2
		newtonIter      = 5
		newtonScale     = 2
		bootstrapDepth  = 3
	)

	// Step 1: Compute mean μ
	mean, err := e.Mean(ct)
	if err != nil {
		return nil, fmt.Errorf("compute mean: %w", err)
	}

	// Step 2: Centered data x = X - μ
	centered, err := e.Sub(ct, mean)
	if err != nil {
		return nil, fmt.Errorf("center data: %w", err)
	}

	// Step 3: Compute x² and x³
	x2, err := e.Mult(centered, centered)
	if err != nil {
		return nil, fmt.Errorf("compute x^2: %w", err)
	}

	x3, err := e.Mult(centered, x2)
	if err != nil {
		return nil, fmt.Errorf("compute x^3: %w", err)
	}

	// Step 4: Compute E[x³]
	numerator, err := e.Mean(x3)
	if err != nil {
		return nil, fmt.Errorf("mean of x^3: %w", err)
	}

	// Step 5: Approximate variance σ²
	denom := float64(ct.Size()) * B
	varianceApprox, err := varianceWithCustomDenom(e, ct, denom, denom*B)
	if err != nil {
		return nil, fmt.Errorf("variance (approx): %w", err)
	}

	varApproxCtxt, err := e.selectOneCtxt(varianceApprox)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt (variance approx): %w", err)
	}

	// Step 6: Initial approximation of 1/σ using Chebyshev
	
	// Optional: Bootstrap the initial guess for higher precision
	if e.IsBTS {
		varApproxCtxt, err = e.DoBootstrap(varApproxCtxt, 9)
		if err != nil {
			return nil, fmt.Errorf("bootstrap (invSqrt init): %w", err)
		}
	}
	invSigmaInit, err := e.ChebyshevInvSqrt(varApproxCtxt, chebyshevDegree, B*B)
	if err != nil {
		return nil, fmt.Errorf("ChebyshevInvSqrt: %w", err)
	}

	if e.IsBTS {
		invSigmaInit, err = e.DoBootstrap(invSigmaInit, bootstrapDepth)
		if err != nil {
			return nil, fmt.Errorf("bootstrap (Chebyshev init): %w", err)
		}
	}

	// Step 7: Refine inverse std dev using Newton
	denom = float64(ct.Size()) * math.Sqrt(2)
	varianceRefined, err := varianceWithCustomDenom(e, ct, denom, denom * math.Sqrt(2))
	if err != nil {
		return nil, fmt.Errorf("compute refined variance: %w", err)
	}
	varRefinedCtxt, err := e.selectOneCtxt(varianceRefined)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt (variance refined): %w", err)
	}

	invSigma, err := e.HENewtonInv(varRefinedCtxt, invSigmaInit, B, newtonIter, newtonScale)
	if err != nil {
		return nil, fmt.Errorf("HENewtonInv: %w", err)
	}

	// Step 8: Compute (1/σ)³ = (1/σ) × (1/σ)²
	invSigma2, err := e.Mult(invSigma, invSigma)
	if err != nil {
		return nil, fmt.Errorf("inv σ²: %w", err)
	}

	invSigma3, err := e.Mult(invSigma, invSigma2)
	if err != nil {
		return nil, fmt.Errorf("inv σ³: %w", err)
	}

	// Step 9: Final result: E[x³] × (1/σ³)
	invSigma3Expanded, err := e.extendOneToMulty(invSigma3, len(numerator.Ciphertexts()), numerator.Size())
	if err != nil {
		return nil, fmt.Errorf("extendOneToMulty: %w", err)
	}

	skewness, err := e.Mult(numerator, invSigma3Expanded)
	if err != nil {
		return nil, fmt.Errorf("final multiply: %w", err)
	}

	return skewness, nil
}

func (e *HEEngine) PCorrCoeff_ppstat(ct1, ct2 *HEData, B float64) (*HEData, error) {
	const (
		chebyshevDegree = 2
		newtonIter      = 5
		newtonScale     = 2
		bootstrapDepth  = 3
	)

	// Step 1: Compute means
	meanX, err := e.Mean(ct1)
	if err != nil {
		return nil, fmt.Errorf("mean of ct1: %w", err)
	}
	meanY, err := e.Mean(ct2)
	if err != nil {
		return nil, fmt.Errorf("mean of ct2: %w", err)
	}

	// Step 2: Centered data
	xCentered, err := e.Sub(ct1, meanX)
	if err != nil {
		return nil, fmt.Errorf("center ct1: %w", err)
	}
	yCentered, err := e.Sub(ct2, meanY)
	if err != nil {
		return nil, fmt.Errorf("center ct2: %w", err)
	}

	// Step 3: Numerator = E[(X - μx)(Y - μy)]
	mulXY, err := e.Mult(xCentered, yCentered)
	if err != nil {
		return nil, fmt.Errorf("x·y: %w", err)
	}
	numerator, err := e.Mean(mulXY)
	if err != nil {
		return nil, fmt.Errorf("mean of x·y: %w", err)
	}

	// Step 4: Compute inverse std for X
	invStdX, err := e.computeInvStd_ppstat(ct1, chebyshevDegree, newtonIter, newtonScale, bootstrapDepth, B)
	if err != nil {
		return nil, fmt.Errorf("computeInvStd (X): %w", err)
	}

	// Step 5: Compute inverse std for Y
	invStdY, err := e.computeInvStd_ppstat(ct2, chebyshevDegree, newtonIter, newtonScale, bootstrapDepth, B)
	if err != nil {
		return nil, fmt.Errorf("computeInvStd (Y): %w", err)
	}

	// Step 6: Compute PCC = numerator × (1/σx) × (1/σy)
	denominator, err := e.Mult(invStdX, invStdY)
	if err != nil {
		return nil, fmt.Errorf("σx·σy inverse: %w", err)
	}

	denominatorExpanded, err := e.extendOneToMulty(denominator, len(numerator.Ciphertexts()), numerator.Size())
	if err != nil {
		return nil, fmt.Errorf("extendOneToMulty: %w", err)
	}

	pcc, err := e.Mult(numerator, denominatorExpanded)
	if err != nil {
		return nil, fmt.Errorf("final multiply: %w", err)
	}

	return pcc, nil
}


func (e *HEEngine) computeInvStd_ppstat(ct *HEData, chebDeg, newtonIter, newtonScale, bootstrapDepth int, B float64) (*HEData, error) {
	denom := float64(ct.Size()) * B

	// Approximate variance
	varianceApprox, err := varianceWithCustomDenom(e, ct, denom, denom*B)
	if err != nil {
		return nil, fmt.Errorf("variance (approx): %w", err)
	}
	varApproxCtxt, err := e.selectOneCtxt(varianceApprox)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt (approx): %w", err)
	}

	// Initial guess for 1/σ
	
	// Optional: Bootstrap the initial guess for higher precision
	if e.IsBTS {
		varApproxCtxt, err = e.DoBootstrap(varApproxCtxt, 9)
		if err != nil {
			return nil, fmt.Errorf("bootstrap (invSqrt init): %w", err)
		}
	}
	invSigmaInit, err := e.ChebyshevInvSqrt(varApproxCtxt, chebDeg, B*B)
	if err != nil {
		return nil, fmt.Errorf("ChebyshevInvSqrt: %w", err)
	}
	if e.IsBTS {
		invSigmaInit, err = e.DoBootstrap(invSigmaInit, bootstrapDepth)
		if err != nil {
			return nil, fmt.Errorf("bootstrap: %w", err)
		}
	}

	// Refined variance
	denom = float64(ct.Size()) * math.Sqrt(2)
	varianceRefined, err := varianceWithCustomDenom(e, ct, denom, denom * math.Sqrt(2))
	if err != nil {
		return nil, fmt.Errorf("compute refined variance: %w", err)
	}
	varRefinedCtxt, err := e.selectOneCtxt(varianceRefined)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt (refined): %w", err)
	}

	// Newton refinement
	invStd, err := e.HENewtonInv(varRefinedCtxt, invSigmaInit, B, newtonIter, newtonScale)
	if err != nil {
		return nil, fmt.Errorf("HENewtonInv: %w", err)
	}

	return invStd, nil
}

func (e *HEEngine) computeInvStd(ct *HEData, fast bool, newtonScale int, B float64) (*HEData, error) {
	
	deg, iter, cs := Get_deg_and_iter(ct.Level() - 2, fast)

	denom := float64(ct.Size()) * B
	varianceApprox, err := varianceWithCustomDenom(e, ct, denom, denom*B)
	if err != nil {
		return nil, fmt.Errorf("variance (approx): %w", err)
	}
	varApproxCtxt, err := e.selectOneCtxt(varianceApprox)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt (approx): %w", err)
	}
	
	if cs == 1 {
		if e.IsBTS {
			varApproxCtxt, err = e.DoBootstrap(varApproxCtxt, e.params.MaxLevel())
			if err != nil {
				return nil, fmt.Errorf("bootstrap (invSqrt init): %w", err)
			}
		}

		varRefinedCtxt, err := e.MultConst(varApproxCtxt, (B*B)/2)
		if err != nil {
			return nil, fmt.Errorf("variance (refined): %w", err)
		}

		// Newton refinement
		invStd, err := e.CryptoInvSqrt(varRefinedCtxt, varApproxCtxt, B*B, deg, iter-1, 2, newtonScale)
		if err != nil {
			return nil, fmt.Errorf("HENewtonInv: %w", err)
		}

		return invStd, nil

	} else {

		denom := float64(ct.Size()) * math.Sqrt(2)
		varianceRefined, err := varianceWithCustomDenom(e, ct, denom, denom * math.Sqrt(2))
		if err != nil {
			return nil, fmt.Errorf("variance (refined): %w", err)
		}
		varRefinedCtxt, err := e.selectOneCtxt(varianceRefined)
		if err != nil {
			return nil, fmt.Errorf("selectOneCtxt (refined): %w", err)
		}

		// Newton refinement
		invStd, err := e.CryptoInvSqrt(varRefinedCtxt, varApproxCtxt, B*B, deg, iter-1, 2, newtonScale)
		if err != nil {
			return nil, fmt.Errorf("HENewtonInv: %w", err)
		}

		return invStd, nil
	}
}

func varianceWithCustomDenom(e *HEEngine, ct *HEData, xDenom, xSquareDenom float64) (*HEData, error) {
	// Step 1: Compute E[X]
	meanXScaled, err := e.MultConst(ct, 1.0/xDenom)
	if err != nil {
		return nil, fmt.Errorf("MultConst(1/xDenom): %w", err)
	}

	meanX, err := e.Sum(meanXScaled)
	if err != nil {
		return nil, fmt.Errorf("Sum(E[x]): %w", err)
	}

	// Step 2: Compute (E[X])^2
	squaredMeanX, err := e.Mult(meanX, meanX)
	if err != nil {
		return nil, fmt.Errorf("square of E[x]: %w", err)
	}

	// Step 3: Compute X^2
	ctSquared, err := e.Mult(ct, ct)
	if err != nil {
		return nil, fmt.Errorf("X^2: %w", err)
	}

	// Step 4: Compute E[X^2]
	meanXSquaredScaled, err := e.MultConst(ctSquared, 1.0/xSquareDenom)
	if err != nil {
		return nil, fmt.Errorf("MultConst(1/xSquareDenom): %w", err)
	}

	meanXSquared, err := e.Sum(meanXSquaredScaled)
	if err != nil {
		return nil, fmt.Errorf("Sum(E[x^2]): %w", err)
	}

	// Step 5: Return variance = E[X^2] - (E[X])^2
	variance, err := e.Sub(meanXSquared, squaredMeanX)
	if err != nil {
		return nil, fmt.Errorf("E[x^2] - E[x]^2: %w", err)
	}

	return variance, nil
}

func (e *HEEngine) selectOneCtxt(ct *HEData) (*HEData, error) {
	size := ct.Size()
	if size > e.params.MaxSlots() {
		size = e.params.MaxSlots()
	}
	ctxt := make([]*rlwe.Ciphertext, 1)
	ctxt[0] = ct.Ciphertexts()[0].CopyNew()
	return NewHEData(ctxt, size, ct.Level(), ct.Scale()), nil
}

func (e *HEEngine) extendOneToMulty(ct *HEData, num, size int) (*HEData, error) {
	ctxts := make([]*rlwe.Ciphertext, num)
	for i := 0; i < num; i++ {
		ctxts[i] = ct.Ciphertexts()[0].CopyNew()
	}
	return NewHEData(ctxts, size, ctxts[0].Level(), ct.Scale()), nil
}
