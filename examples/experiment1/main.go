package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"time"

	"github.com/hm-choi/pp-stat-plus/config"
	"github.com/hm-choi/pp-stat-plus/engine"
	invsqrt "github.com/hm-choi/pp-stat-plus/examples/inv_sqrt"
	"github.com/hm-choi/pp-stat-plus/utils"
)

func main() {
	
	f, _ := os.Create("output.log")
	mw := io.MultiWriter(os.Stdout, f)
    log.SetOutput(mw)

	file, err := os.Create("output.txt")
	if err != nil {
		fmt.Println("Failed to create file:", err)
		return
	}
	defer file.Close()

	e := engine.NewHEEngine(config.NewParameters(16, 11, 50, true))

	const (
		DATA_SIZE = 32768 // Slot size
		B         = 100.0
		START     = 0.001
		MIDDLE    = 1.0
		STOP      = B
		EVAL_NUM  = 10
	)

	// Generate input test values and ground truth (1/sqrt(x))
	test1 := utils.Linspace(START, MIDDLE, DATA_SIZE/2)
	test2 := utils.Linspace(MIDDLE, STOP, DATA_SIZE/2)
	test := append(test1, test2...)

	invS := make([]float64, DATA_SIZE)
	for i, v := range test {
		invS[i] = 1.0 / math.Sqrt(v)
	}

	scaling_depth := 1

	l_bts := 0
	for level := e.Params().MaxLevel(); level >= l_bts+1; level-- {

		HSTAT_MRE, HSTAT_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
		PSTAT_MRE, PSTAT_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
		INVSQRT_BASIC_MRE, INVSQRT_BASIC_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
		INVSQRT_FAST_MRE, INVSQRT_FAST_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)

		log.Println("*******************************************************************")
		log.Println("Level", level)
		log.Println("*******************************************************************")

		for i := 0; i < EVAL_NUM; i++ {
				
			ct_base, _   := e.Encrypt(test, level)

			scaled_ct, _ := e.MultConst(ct_base, 2.0/B)
			ct, _ 		 := e.MultConst(ct_base, 1.0/2)

			// HEaaN-Stat method
			if level > 1 {
				log.Println("-------------------------------------------------------------------")
				log.Println("HEaaN-Stat")
				log.Println("-------------------------------------------------------------------")

				start := time.Now()
				heStat, _ := invsqrt.HEStat(e, ct, 21, B)
				elapsed := time.Since(start)
				hsResult, _ := e.Decrypt(heStat)
				_, hsMRE := utils.CheckMRE(hsResult, hsResult, invS, ct.Size())
				HSTAT_MRE[i], HSTAT_TIME[i] = hsMRE, elapsed.Seconds()
				log.Println("HEaaN-Stat", hsResult[:3], hsMRE, elapsed)
			}

			// PP-Stat method
			log.Println("-------------------------------------------------------------------")
			log.Println("PP-Stat")
			log.Println("-------------------------------------------------------------------")

			start := time.Now()
			ppStat, _ := e.CryptoInvSqrt(ct, scaled_ct, B, 9.0, 6, 1, 2)
			elapsed := time.Since(start)
			psResult, _ := e.Decrypt(ppStat)
			_, psMRE := utils.CheckMRE(psResult, psResult, invS, ct.Size())
			PSTAT_MRE[i], PSTAT_TIME[i] = psMRE, elapsed.Seconds()
			log.Println("CryptoInvSqrt", psResult[:3], psMRE, elapsed)


			// Proposed method
			log.Println("-------------------------------------------------------------------")
			log.Println("Proposed method")
			log.Println("-------------------------------------------------------------------")			

			log.Println("<Basic>")

			deg, iter, cs := engine.Get_deg_and_iter(ct_base.Level() - scaling_depth, false)

			start = time.Now()
			if cs == 1 {
				ct_temp, _ := e.MultConst(ct_base, 1.0/B)
				ct_temp, _ = e.DoBootstrap(ct_temp, e.Params().Parameters.MaxLevel())
				ct_temp, _ = e.MultConst(ct_temp, B)

				scaled_ct, _ = e.MultConst(ct_temp, 2.0/B)
				ct, _ 		 = e.MultConst(ct_temp, 1.0/2)
			}
			cryptoInvSqrt, _ := e.CryptoInvSqrt(ct, scaled_ct, B, deg, iter, 1, 2)
			elapsed = time.Since(start)
			cisResult, _ := e.Decrypt(cryptoInvSqrt)
			_, cisMRE := utils.CheckMRE(cisResult, cisResult, invS, ct.Size())
			INVSQRT_BASIC_MRE[i], INVSQRT_BASIC_TIME[i] = cisMRE, elapsed.Seconds()
			log.Println("CryptoInvSqrt-BASIC", cisResult[:3], cisMRE, elapsed)

			log.Println("<Fast>")

			deg, iter, cs = engine.Get_deg_and_iter(ct_base.Level() - scaling_depth, true)

			start = time.Now()
			if cs == 1 {
				ct_temp, _ := e.MultConst(ct_base, 1.0/B)
				ct_temp, _ = e.DoBootstrap(ct_temp, e.Params().Parameters.MaxLevel())
				ct_temp, _ = e.MultConst(ct_temp, B)

				scaled_ct, _ = e.MultConst(ct_temp, 2.0/B)
				ct, _ 		 = e.MultConst(ct_temp, 1.0/2)
			}
			cryptoInvSqrt, _ = e.CryptoInvSqrt(ct, scaled_ct, B, deg, iter, 1, 2)
			elapsed = time.Since(start)
			cisResult, _ = e.Decrypt(cryptoInvSqrt)
			_, cisMRE = utils.CheckMRE(cisResult, cisResult, invS, ct.Size())
			INVSQRT_FAST_MRE[i], INVSQRT_FAST_TIME[i] = cisMRE, elapsed.Seconds()
			log.Println("CryptoInvSqrt-FAST", cisResult[:3], cisMRE, elapsed)
			log.Println()
		}
			
		result := fmt.Sprintf("*Level %d*\n", level)
		result += fmt.Sprintf("[HSTAT] MRE %e (%e), TIME %.3f (%.3f)\n",
		utils.Mean(HSTAT_MRE), utils.StdDev(HSTAT_MRE),
		utils.Mean(HSTAT_TIME), utils.StdDev(HSTAT_TIME),
		)
		result += fmt.Sprintf("[PSTAT] MRE %e (%e), TIME %.3f (%.3f)\n",
			utils.Mean(PSTAT_MRE), utils.StdDev(PSTAT_MRE),
			utils.Mean(PSTAT_TIME), utils.StdDev(PSTAT_TIME),
		)
		result += fmt.Sprintf("[OURS-BASIC] MRE %e (%e), TIME %.3f (%.3f)\n",
			utils.Mean(INVSQRT_BASIC_MRE), utils.StdDev(INVSQRT_BASIC_MRE),
			utils.Mean(INVSQRT_BASIC_TIME), utils.StdDev(INVSQRT_BASIC_TIME),
		)
		result += fmt.Sprintf("[OURS-FAST] MRE %e (%e), TIME %.3f (%.3f)\n",
			utils.Mean(INVSQRT_FAST_MRE), utils.StdDev(INVSQRT_FAST_MRE),
			utils.Mean(INVSQRT_FAST_TIME), utils.StdDev(INVSQRT_FAST_TIME),
		)

		log.Println(result)
		io.WriteString(file, result)
	}

}
