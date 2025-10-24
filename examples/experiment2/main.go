package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"time"

	"github.com/hm-choi/pp-stat-plus/config"
	"github.com/hm-choi/pp-stat-plus/engine"
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

	// Benchmark settings
	DATA_SIZE := 1000000 // Datasize
	RANGE, RANGE2 := 20.0, 100.0
	EVAL_NUM := 10

	
	fasts := []bool{false, true}
	l_bts := 0
	for level := e.Params().MaxLevel(); level >= l_bts+1; level-- {

		log.Println("*******************************************************************")
		log.Println("Level", level)
		log.Println("*******************************************************************")
		log.Println()

		// Generate random plaintext data
		values1 := make([]float64, DATA_SIZE)
		values2 := make([]float64, DATA_SIZE)
		values3 := make([]float64, DATA_SIZE)
		for j := 0; j < DATA_SIZE; j++ {
			values1[j] = RANGE * rand.Float64()
			values2[j] = RANGE * rand.Float64()
			values3[j] = RANGE2 * rand.Float64()
		}

		// Encrypt inputs
		ctxt1, _ := e.Encrypt(values1, level)
		ctxt2, _ := e.Encrypt(values2, level)
		ctxt3, _ := e.Encrypt(values3, level)

		for _, fast := range fasts {

			// Result arrays (MRE and TIME only)
			ZSCORE_MRE, ZSCORE_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
			KURT_MRE, KURT_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
			SKEW_MRE, SKEW_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
			CORR_MRE, CORR_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)

			
			ZSCORE_MRE_PPSTAT, ZSCORE_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
			KURT_MRE_PPSTAT, KURT_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
			SKEW_MRE_PPSTAT, SKEW_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
			CORR_MRE_PPSTAT, CORR_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)

			for i := 0; i < EVAL_NUM; i++ {

				// [Z-Score Normalization]
				log.Println("[Z-Score Normalization]")
				start := time.Now()
				zNormReal := utils.ZScoreNorm(values3)
				zNorm, _ := e.ZScoreNorm(ctxt3, RANGE2, fast)
				duration := time.Since(start)
				zNormResult, _ := e.Decrypt(zNorm)
				_, mre := utils.CheckMRE(zNormResult, zNormResult, zNormReal, len(zNormReal))
				log.Println("[Ours]Z-Score MRE:", mre, duration)
				ZSCORE_MRE[i] = mre
				ZSCORE_TIME[i] = duration.Seconds()

				start = time.Now()
				zNormReal = utils.ZScoreNorm(values3)
				zNorm, _ = e.ZScoreNorm_ppstat(ctxt3, RANGE2)
				duration = time.Since(start)
				zNormResult, _ = e.Decrypt(zNorm)
				_, mre = utils.CheckMRE(zNormResult, zNormResult, zNormReal, len(zNormReal))
				log.Println("[PP-STAT]Z-Score MRE:", mre, duration)
				ZSCORE_MRE_PPSTAT[i] = mre
				ZSCORE_TIME_PPSTAT[i] = duration.Seconds()

				// [Skewness]
				log.Println("[Skewness]")
				_, _, skewReal := utils.Skewness(values1)
				start = time.Now()
				skew, _ := e.Skewness(ctxt1, RANGE, fast)
				duration = time.Since(start)
				skewResult, _ := e.Decrypt(skew)
				mre = math.Abs(skewResult[0]-skewReal) / math.Abs(skewReal)
				log.Println("[Ours]Skewness MRE:", mre, duration)
				SKEW_MRE[i] = mre
				SKEW_TIME[i] = duration.Seconds()

				_, _, skewReal = utils.Skewness(values1)
				start = time.Now()
				skew, _ = e.Skewness_ppstat(ctxt1, RANGE)
				duration = time.Since(start)
				skewResult, _ = e.Decrypt(skew)
				mre = math.Abs(skewResult[0]-skewReal) / math.Abs(skewReal)
				log.Println("[PP-STAT]Skewness MRE:", mre, duration)
				SKEW_MRE_PPSTAT[i] = mre
				SKEW_TIME_PPSTAT[i] = duration.Seconds()

				// [Kurtosis]
				log.Println("[Kurtosis]")
				_, _, kurtReal := utils.Kurtosis(values1)
				start = time.Now()
				kurt, _ := e.Kurtosis(ctxt1, RANGE, fast)
				duration = time.Since(start)
				kurtResult, _ := e.Decrypt(kurt)
				mre = math.Abs(kurtResult[0]-kurtReal) / math.Abs(kurtReal)
				log.Println("[Ours]Kurtosis MRE:", mre, duration)
				KURT_MRE[i] = mre
				KURT_TIME[i] = duration.Seconds()

				_, _, kurtReal = utils.Kurtosis(values1)
				start = time.Now()
				kurt, _ = e.Kurtosis_ppstat(ctxt1, RANGE)
				duration = time.Since(start)
				kurtResult, _ = e.Decrypt(kurt)
				mre = math.Abs(kurtResult[0]-kurtReal) / math.Abs(kurtReal)
				log.Println("[PP-STAT]Kurtosis MRE:", mre, duration)
				KURT_MRE_PPSTAT[i] = mre
				KURT_TIME_PPSTAT[i] = duration.Seconds()

				// [Correlation]
				log.Println("[Correlation]")
				_, corrReal, _ := utils.Correlation(values1, values2)
				start = time.Now()
				corr, _ := e.PCorrCoeff(ctxt1, ctxt2, RANGE, fast)
				duration = time.Since(start)
				corrResult, _ := e.Decrypt(corr)
				mre = math.Abs(corrResult[0]-corrReal) / math.Abs(corrReal)
				log.Println("[Ours]Correlation MRE:", mre, duration)
				CORR_MRE[i] = mre
				CORR_TIME[i] = duration.Seconds()

				_, corrReal, _ = utils.Correlation(values1, values2)
				start = time.Now()
				corr, _ = e.PCorrCoeff_ppstat(ctxt1, ctxt2, RANGE)
				duration = time.Since(start)
				corrResult, _ = e.Decrypt(corr)
				mre = math.Abs(corrResult[0]-corrReal) / math.Abs(corrReal)
				log.Println("[PP-STAT]Correlation MRE:", mre, duration)
				CORR_MRE_PPSTAT[i] = mre
				CORR_TIME_PPSTAT[i] = duration.Seconds()
			}

			// Write summary results to output file
			result := fmt.Sprintf("*Level %d, MRE constraint %t*\n", level, !fast)
			result += fmt.Sprintf("[Ours][ZSCORE] MRE %e (%e), TIME %f (%f)\n", utils.Mean(ZSCORE_MRE), utils.StdDev(ZSCORE_MRE), utils.Mean(ZSCORE_TIME), utils.StdDev(ZSCORE_TIME))
			result += fmt.Sprintf("[Ours][SKEWNESS] MRE %e (%e), TIME %f (%f)\n", utils.Mean(SKEW_MRE), utils.StdDev(SKEW_MRE), utils.Mean(SKEW_TIME), utils.StdDev(SKEW_TIME))
			result += fmt.Sprintf("[Ours][KURTOSIS] MRE %e (%e), TIME %f (%f)\n", utils.Mean(KURT_MRE), utils.StdDev(KURT_MRE), utils.Mean(KURT_TIME), utils.StdDev(KURT_TIME))
			result += fmt.Sprintf("[Ours][CORREL] MRE %e (%e), TIME %f (%f)\n", utils.Mean(CORR_MRE), utils.StdDev(CORR_MRE), utils.Mean(CORR_TIME), utils.StdDev(CORR_TIME))

			result_ppstat := fmt.Sprintf("[PP-STAT][ZSCORE] MRE %e (%e), TIME %f (%f)\n", utils.Mean(ZSCORE_MRE_PPSTAT), utils.StdDev(ZSCORE_MRE_PPSTAT), utils.Mean(ZSCORE_TIME_PPSTAT), utils.StdDev(ZSCORE_TIME_PPSTAT))
			result_ppstat += fmt.Sprintf("[PP-STAT][SKEWNESS] MRE %e (%e), TIME %f (%f)\n", utils.Mean(SKEW_MRE_PPSTAT), utils.StdDev(SKEW_MRE_PPSTAT), utils.Mean(SKEW_TIME_PPSTAT), utils.StdDev(SKEW_TIME_PPSTAT))
			result_ppstat += fmt.Sprintf("[PP-STAT][KURTOSIS] MRE %e (%e), TIME %f (%f)\n", utils.Mean(KURT_MRE_PPSTAT), utils.StdDev(KURT_MRE_PPSTAT), utils.Mean(KURT_TIME_PPSTAT), utils.StdDev(KURT_TIME_PPSTAT))
			result_ppstat += fmt.Sprintf("[PP-STAT][CORREL] MRE %e (%e), TIME %f (%f)\n", utils.Mean(CORR_MRE_PPSTAT), utils.StdDev(CORR_MRE_PPSTAT), utils.Mean(CORR_TIME_PPSTAT), utils.StdDev(CORR_TIME_PPSTAT))
			io.WriteString(file, result)
			io.WriteString(file, result_ppstat)
		}
	}
}
