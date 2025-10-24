package optimizer

import (
	"log"
	"math"
	"slices"
	"time"

	"github.com/hm-choi/pp-stat-plus/engine"
	"github.com/hm-choi/pp-stat-plus/utils"
)

type Dtuple struct {
	D float64
	C, I int
	M, T float64
}

type Rtuple struct {
	D float64
	C, I int
	M, T float64
}

func GetOptIter(e *engine.HEEngine, ct *engine.HEData, scaled_ct *engine.HEData, ans []float64, deg, B float64, i_max int, delta float64) (int, float64, float64) {

	M, T := make([]float64, i_max), make([]float64, i_max)

	start := time.Now()

	y, _ := ChebyshevInvSqrt_deg_log(e, scaled_ct, 1, B, deg)

	N := 2
	x:= ct.CopyData()

	for i := range i_max {
		
		if e.IsBTS {
			y, _ = e.DoBootstrap(y, 4)
		}	

		tmp_a_c, _ := e.MultConst(y, float64((N+1))/float64(N))
		tmp_b_c, _ := e.Mult(x, y)

		y, _ = e.Mult(y, y)

		tmp_b_c, _ = e.Mult(tmp_b_c, y)
		y, _ = e.Sub(tmp_a_c, tmp_b_c)
		
		elapsed := time.Since(start).Seconds()

		after_iter_c, _ := e.Decrypt(y)
		
		log.Println("Iter", i+1)
		log.Println("Level x -", x.Level(), "y -", y.Level())
		log.Println("Time", elapsed)

		_, MRE  := utils.CheckMRE(after_iter_c, after_iter_c, ans, ct.Size())
		log.Println("MRE", MRE)

		M[i] = MRE
		T[i] = elapsed

		log.Println()
	}

	MREmin := slices.Min(M)
	ell := math.Floor(math.Log10(MREmin))
	alpha := MREmin / math.Pow(10, ell)
	MREdelta := (math.Floor(alpha) + delta) * math.Pow(10, ell)
	I := -1
	for i, v := range M {
		if v <= MREdelta {
			I = i
			break
		}
	}

	return I+1, M[I], T[I]
}


func Optimizing(e *engine.HEEngine, d_min, d_max float64, i_max int, START, MIDDLE, STOP float64, N int, theta, delta float64) (map[int][]Rtuple) {
	
	B := STOP
	
	// Generate input test values and ground truth (1/sqrt(x))
	test1 := utils.Linspace(START, MIDDLE, N/4)
	test2 := utils.Linspace(MIDDLE, STOP, N/4)
	test := append(test1, test2...)

	invS := make([]float64, N/2)
	for i, v := range test {
		invS[i] = 1.0 / math.Sqrt(v)
	}

	D := make(map[int][]Dtuple)
	R := make(map[int][]Rtuple)

	
	l_bts := 0
	l_afterBTS := e.Params().MaxLevel()

	for l := e.Params().MaxLevel(); l >= l_bts+1; l-- {

		log.Println("*******************************************************************")
		log.Println("Level", l)
		log.Println("*******************************************************************")
		log.Println()

		for d_e := d_min; d_e <= d_max; d_e++ {

			ct_base, _   := e.Encrypt(test, l)

			scaled_ct, _ := e.MultConst(ct_base, 2.0/B)
			ct, _ 		 := e.MultConst(ct_base, 1.0/2)

			log.Println("===================================================================")
			log.Println("Degree", math.Pow(2, d_e)-2)
			log.Println("===================================================================")
			log.Println()

			if ct_base.Level() >= l_bts + 3 {

				log.Println("-------------------------------------------------------------------")
				log.Println("No Pre-BTS")
				log.Println("-------------------------------------------------------------------")

				i, m, t := GetOptIter(e, ct, scaled_ct, invS, d_e, B, i_max, delta)
				D[l] = append(D[l], Dtuple{d_e, 0, i, m, t})
			}
			if ct_base.Level() <= l_afterBTS -2 {

				log.Println("-------------------------------------------------------------------")
				log.Println("Pre-BTS")
				log.Println("-------------------------------------------------------------------")

				start := time.Now()
				ct_base, _ = e.MultConst(ct_base, 1.0/B)
				ct_base, _ = e.DoBootstrap(ct_base, e.Params().Parameters.MaxLevel())
				ct_base, _ = e.MultConst(ct_base, B)
		
				scaled_ct, _ = e.MultConst(ct_base, 2.0/B)
				ct, _ 		 = e.MultConst(ct_base, 1.0/2)
				elapsed := time.Since(start).Seconds()

				i, m, t := GetOptIter(e, ct, scaled_ct, invS, d_e, B, i_max, delta)
				D[l] = append(D[l], Dtuple{d_e, 1, i, m, t+elapsed})
			}

		}
	}

	for L, tuples := range D {
		if len(tuples) == 0 {
			continue
		}

		Mmin := tuples[0].M
		for _, t := range tuples[1:] {
			if t.M < Mmin {
				Mmin = t.M
			}
		}

		ell  := math.Floor(math.Log10(Mmin))
		alpha := Mmin / math.Pow(10, ell)
		Mtheta := (math.Floor(alpha) + theta) * math.Pow(10, ell)

		var u1 Dtuple
		found := false
		for _, t := range tuples {
			if t.M <= Mtheta {
				if !found || t.T < u1.T {
					u1 = t
					found = true
				}
			}
		}

		u2 := tuples[0]
		for _, t := range tuples[1:] {
			if t.T < u2.T {
				u2 = t
			}
		}

		R[L] = []Rtuple{
			{D: u1.D, C: u1.C, I: u1.I, M:u1.M, T:u1.T},
			{D: u2.D, C: u2.C, I: u2.I, M:u2.M, T:u2.T},
		}
	}

	return R
}