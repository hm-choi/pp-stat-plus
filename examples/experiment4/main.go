package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"time"

	"github.com/hm-choi/pp-stat-plus/config"
	"github.com/hm-choi/pp-stat-plus/engine"
	"github.com/hm-choi/pp-stat-plus/utils"
)

func main() {
	engine := engine.NewHEEngine(config.NewParameters(16, 11, 50, true))
	ageSlice, _ := utils.ReadCSV("../../examples/dataset/insurance.csv", 0)
	bmiSlice, _ := utils.ReadCSV("../../examples/dataset/insurance.csv", 2)
	smokerSlice, _ := utils.ReadCSV("../../examples/dataset/insurance.csv", 4)
	chargeSlice, _ := utils.ReadCSV("../../examples/dataset/insurance.csv", 6)

	for i := 0; i < len(chargeSlice); i++ {
		chargeSlice[i] = chargeSlice[i] / 1000.0
	}

	EVAL_NUM, B := 10, 20.0
	CG_ZSCORE_MRE0, CG_ZSCORE_MAE0, CG_ZSCORE_TIME0 := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	CG_KURT_MRE0, CG_KURT_MAE0, CG_KURT_TIME0 := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	CG_SKEW_MRE0, CG_SKEW_MAE0, CG_SKEW_TIME0 := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	AGE_CG_CORR_MRE0, AGE_CG_CORR_MAE0, AGE_CG_CORR_TIME0 := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	BMI_CG_CORR_MRE0, BMI_CG_CORR_MAE0, BMI_CG_CORR_TIME0 := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	SMOKER_CG_CORR_MRE0, SMOKER_CG_CORR_MAE0, SMOKER_CG_CORR_TIME0 := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)

	CG_ZSCORE_MRE1, CG_ZSCORE_MAE1, CG_ZSCORE_TIME1 := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	CG_KURT_MRE1, CG_KURT_MAE1, CG_KURT_TIME1 := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	CG_SKEW_MRE1, CG_SKEW_MAE1, CG_SKEW_TIME1 := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	AGE_CG_CORR_MRE1, AGE_CG_CORR_MAE1, AGE_CG_CORR_TIME1 := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	BMI_CG_CORR_MRE1, BMI_CG_CORR_MAE1, BMI_CG_CORR_TIME1 := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	SMOKER_CG_CORR_MRE1, SMOKER_CG_CORR_MAE1, SMOKER_CG_CORR_TIME1 := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	
	CG_ZSCORE_MRE_PPSTAT, CG_ZSCORE_MAE_PPSTAT, CG_ZSCORE_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	CG_KURT_MRE_PPSTAT, CG_KURT_MAE_PPSTAT, CG_KURT_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	CG_SKEW_MRE_PPSTAT, CG_SKEW_MAE_PPSTAT, CG_SKEW_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	AGE_CG_CORR_MRE_PPSTAT, AGE_CG_CORR_MAE_PPSTAT, AGE_CG_CORR_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	BMI_CG_CORR_MRE_PPSTAT, BMI_CG_CORR_MAE_PPSTAT, BMI_CG_CORR_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	SMOKER_CG_CORR_MRE_PPSTAT, SMOKER_CG_CORR_MAE_PPSTAT, SMOKER_CG_CORR_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)

	file, err := os.Create("output.txt")
	if err != nil {
		return
	}
	defer file.Close()
	level := 11
	age, _ := engine.Encrypt(ageSlice, level)
	bmi, _ := engine.Encrypt(bmiSlice, level)
	smoker, _ := engine.Encrypt(smokerSlice, level)
	charge, _ := engine.Encrypt(chargeSlice, level)

	_, _, skew_cg := utils.Skewness(chargeSlice)
	_, _, kurt_cg := utils.Kurtosis(chargeSlice)

	fmt.Println("==============================")

	for i := 0; i < int(EVAL_NUM); i++ {
		TIME := time.Now()
		zNormCharge, _ := engine.ZScoreNorm(charge, 100.0, false)
		CHARGE_ZNORM_TIME := time.Since(TIME)
		zSNcharge, _ := engine.Decrypt(zNormCharge)

		_, zScoreMreCharge := utils.CheckMRE(zSNcharge, zSNcharge, utils.ZScoreNorm(chargeSlice), len(chargeSlice))
		_, zScoreMaeCharge := utils.CheckMAE(zSNcharge, zSNcharge, utils.ZScoreNorm(chargeSlice), len(chargeSlice))

		CG_ZSCORE_MRE0[i] = zScoreMreCharge
		CG_ZSCORE_MAE0[i] = zScoreMaeCharge
		CG_ZSCORE_TIME0[i] = float64(CHARGE_ZNORM_TIME.Seconds())
		fmt.Println("Charge ZNorm <Basic>", zScoreMaeCharge, zScoreMreCharge, CHARGE_ZNORM_TIME)

		TIME = time.Now()
		zNormCharge, _ = engine.ZScoreNorm(charge, 100.0, true)
		CHARGE_ZNORM_TIME = time.Since(TIME)
		zSNcharge, _ = engine.Decrypt(zNormCharge)

		_, zScoreMreCharge = utils.CheckMRE(zSNcharge, zSNcharge, utils.ZScoreNorm(chargeSlice), len(chargeSlice))
		_, zScoreMaeCharge = utils.CheckMAE(zSNcharge, zSNcharge, utils.ZScoreNorm(chargeSlice), len(chargeSlice))

		CG_ZSCORE_MRE1[i] = zScoreMreCharge
		CG_ZSCORE_MAE1[i] = zScoreMaeCharge
		CG_ZSCORE_TIME1[i] = float64(CHARGE_ZNORM_TIME.Seconds())
		fmt.Println("Charge ZNorm <Fast>", zScoreMaeCharge, zScoreMreCharge, CHARGE_ZNORM_TIME)

		TIME = time.Now()
		zNormCharge_ppstat, _ := engine.ZScoreNorm_ppstat(charge, 100.0)
		CHARGE_ZNORM__PPSTAT_TIME := time.Since(TIME)
		zSNcharge_ppstat, _ := engine.Decrypt(zNormCharge_ppstat)

		_, zScoreMreCharge_ppstat := utils.CheckMRE(zSNcharge_ppstat, zSNcharge_ppstat, utils.ZScoreNorm(chargeSlice), len(chargeSlice))
		_, zScoreMaeCharge_ppstat := utils.CheckMAE(zSNcharge_ppstat, zSNcharge_ppstat, utils.ZScoreNorm(chargeSlice), len(chargeSlice))

		CG_ZSCORE_MRE_PPSTAT[i] = zScoreMreCharge_ppstat
		CG_ZSCORE_MAE_PPSTAT[i] = zScoreMaeCharge_ppstat
		CG_ZSCORE_TIME_PPSTAT[i] = float64(CHARGE_ZNORM__PPSTAT_TIME.Seconds())


		TIME = time.Now()
		skewCharge, _ := engine.Skewness(charge, B, false)
		CHARGE_SKEW_TIME := time.Since(TIME)
		skCharge, _ := engine.Decrypt(skewCharge)
		CG_SKEW_MRE0[i] = math.Abs(skCharge[0]-skew_cg) / math.Abs(skew_cg)
		CG_SKEW_MAE0[i] = math.Abs(skCharge[0] - skew_cg)
		CG_SKEW_TIME0[i] = float64(CHARGE_SKEW_TIME.Seconds())
		fmt.Println("Charge skewResult <Basic>", skCharge[0], math.Abs(skCharge[0]-skew_cg), math.Abs(skCharge[0]-skew_cg)/math.Abs(skew_cg), CHARGE_SKEW_TIME)

		TIME = time.Now()
		skewCharge, _ = engine.Skewness(charge, B, true)
		CHARGE_SKEW_TIME = time.Since(TIME)
		skCharge, _ = engine.Decrypt(skewCharge)
		CG_SKEW_MRE1[i] = math.Abs(skCharge[0]-skew_cg) / math.Abs(skew_cg)
		CG_SKEW_MAE1[i] = math.Abs(skCharge[0] - skew_cg)
		CG_SKEW_TIME1[i] = float64(CHARGE_SKEW_TIME.Seconds())
		fmt.Println("Charge skewResult <Fast>", skCharge[0], math.Abs(skCharge[0]-skew_cg), math.Abs(skCharge[0]-skew_cg)/math.Abs(skew_cg), CHARGE_SKEW_TIME)

		TIME = time.Now()
		skewCharge_ppstat, _ := engine.Skewness_ppstat(charge, B)
		CHARGE_SKEW_PPSTAT_TIME := time.Since(TIME)
		skCharge_ppstat, _ := engine.Decrypt(skewCharge_ppstat)
		CG_SKEW_MRE_PPSTAT[i] = math.Abs(skCharge_ppstat[0]-skew_cg) / math.Abs(skew_cg)
		CG_SKEW_MAE_PPSTAT[i] = math.Abs(skCharge_ppstat[0] - skew_cg)
		CG_SKEW_TIME_PPSTAT[i] = float64(CHARGE_SKEW_PPSTAT_TIME.Seconds())


		TIME = time.Now()
		kurtCharge, _ := engine.Kurtosis(charge, B, false)
		CHARGE_KURT_TIME := time.Since(TIME)
		ktCharge, _ := engine.Decrypt(kurtCharge)
		CG_KURT_MRE0[i] = math.Abs(ktCharge[0]-kurt_cg) / math.Abs(kurt_cg)
		CG_KURT_MAE0[i] = math.Abs(ktCharge[0] - kurt_cg)
		CG_KURT_TIME0[i] = float64(CHARGE_KURT_TIME.Seconds())
		fmt.Println("BCharge kurtResult <Basic>", ktCharge[0], math.Abs(ktCharge[0]-kurt_cg), math.Abs(ktCharge[0]-kurt_cg)/math.Abs(kurt_cg), CHARGE_KURT_TIME)

		TIME = time.Now()
		kurtCharge, _ = engine.Kurtosis(charge, B, true)
		CHARGE_KURT_TIME = time.Since(TIME)
		ktCharge, _ = engine.Decrypt(kurtCharge)
		CG_KURT_MRE1[i] = math.Abs(ktCharge[0]-kurt_cg) / math.Abs(kurt_cg)
		CG_KURT_MAE1[i] = math.Abs(ktCharge[0] - kurt_cg)
		CG_KURT_TIME1[i] = float64(CHARGE_KURT_TIME.Seconds())
		fmt.Println("BCharge kurtResult <Fast>", ktCharge[0], math.Abs(ktCharge[0]-kurt_cg), math.Abs(ktCharge[0]-kurt_cg)/math.Abs(kurt_cg), CHARGE_KURT_TIME)

		TIME = time.Now()
		kurtCharge_ppstat, _ := engine.Kurtosis(charge, B, false)
		CHARGE_KURT_PPSTAT_TIME := time.Since(TIME)
		ktCharge_ppstat, _ := engine.Decrypt(kurtCharge_ppstat)
		CG_KURT_MRE_PPSTAT[i] = math.Abs(ktCharge_ppstat[0]-kurt_cg) / math.Abs(kurt_cg)
		CG_KURT_MAE_PPSTAT[i] = math.Abs(ktCharge_ppstat[0] - kurt_cg)
		CG_KURT_TIME_PPSTAT[i] = float64(CHARGE_KURT_PPSTAT_TIME.Seconds())


		_, corrr1, _ := utils.Correlation(ageSlice, chargeSlice)
		TIME = time.Now()
		corr, _ := engine.PCorrCoeff(age, charge, B, false)
		AGE_CG_TIME := time.Since(TIME)
		corrResult, _ := engine.Decrypt(corr)
		AGE_CG_CORR_MAE0[i] = math.Abs(corrResult[0] - corrr1)
		AGE_CG_CORR_MRE0[i] = math.Abs(corrResult[0]-corrr1) / math.Abs(corrr1)
		AGE_CG_CORR_TIME0[i] = float64(AGE_CG_TIME.Seconds())
		fmt.Println("Correlation (BMI, CHARGE) <Basic>", corrResult[0], math.Abs(corrResult[0]-corrr1), math.Abs(corrResult[0]-corrr1)/corrr1)

		TIME = time.Now()
		corr, _ = engine.PCorrCoeff(age, charge, B, true)
		AGE_CG_TIME = time.Since(TIME)
		corrResult, _ = engine.Decrypt(corr)
		AGE_CG_CORR_MAE1[i] = math.Abs(corrResult[0] - corrr1)
		AGE_CG_CORR_MRE1[i] = math.Abs(corrResult[0]-corrr1) / math.Abs(corrr1)
		AGE_CG_CORR_TIME1[i] = float64(AGE_CG_TIME.Seconds())
		fmt.Println("Correlation (BMI, CHARGE) <Fast>", corrResult[0], math.Abs(corrResult[0]-corrr1), math.Abs(corrResult[0]-corrr1)/corrr1)

		TIME = time.Now()
		corr_ppstat, _ := engine.PCorrCoeff(age, charge, B, false)
		AGE_CG_PPSTAT_TIME := time.Since(TIME)
		corrResult_ppstat, _ := engine.Decrypt(corr_ppstat)
		AGE_CG_CORR_MAE_PPSTAT[i] = math.Abs(corrResult_ppstat[0] - corrr1)
		AGE_CG_CORR_MRE_PPSTAT[i] = math.Abs(corrResult_ppstat[0]-corrr1) / math.Abs(corrr1)
		AGE_CG_CORR_TIME_PPSTAT[i] = float64(AGE_CG_PPSTAT_TIME.Seconds())
		

		_, corrr2, _ := utils.Correlation(bmiSlice, chargeSlice)
		TIME = time.Now()
		corr2, _ := engine.PCorrCoeff(bmi, charge, B, false)
		BMI_CG_TIME := time.Since(TIME)
		corrResult2, _ := engine.Decrypt(corr2)
		BMI_CG_CORR_MAE0[i] = math.Abs(corrResult2[0] - corrr2)
		BMI_CG_CORR_MRE0[i] = math.Abs(corrResult2[0]-corrr2) / math.Abs(corrr2)
		BMI_CG_CORR_TIME0[i] = float64(BMI_CG_TIME.Seconds())
		fmt.Println("Correlation (BMI, CHARGE) <Basic>", corrResult2[0], math.Abs(corrResult2[0]-corrr2), math.Abs(corrResult2[0]-corrr2)/corrr2)

		TIME = time.Now()
		corr2, _ = engine.PCorrCoeff(bmi, charge, B, false)
		BMI_CG_TIME = time.Since(TIME)
		corrResult2, _ = engine.Decrypt(corr2)
		BMI_CG_CORR_MAE1[i] = math.Abs(corrResult2[0] - corrr2)
		BMI_CG_CORR_MRE1[i] = math.Abs(corrResult2[0]-corrr2) / math.Abs(corrr2)
		BMI_CG_CORR_TIME1[i] = float64(BMI_CG_TIME.Seconds())
		fmt.Println("Correlation (BMI, CHARGE) <Fast>", corrResult2[0], math.Abs(corrResult2[0]-corrr2), math.Abs(corrResult2[0]-corrr2)/corrr2)

		TIME = time.Now()
		corr2_ppstat, _ := engine.PCorrCoeff(bmi, charge, B, false)
		BMI_CG_PPSTAT_TIME := time.Since(TIME)
		corrResult2_ppstat, _ := engine.Decrypt(corr2_ppstat)
		BMI_CG_CORR_MAE_PPSTAT[i] = math.Abs(corrResult2_ppstat[0] - corrr2)
		BMI_CG_CORR_MRE_PPSTAT[i] = math.Abs(corrResult2_ppstat[0]-corrr2) / math.Abs(corrr2)
		BMI_CG_CORR_TIME_PPSTAT[i] = float64(BMI_CG_PPSTAT_TIME.Seconds())
		

		_, corrr3, _ := utils.Correlation(smokerSlice, chargeSlice)
		TIME = time.Now()
		corr3, _ := engine.PCorrCoeff(smoker, charge, B, false)
		SMOKER_CG_TIME := time.Since(TIME)
		corrResult3, _ := engine.Decrypt(corr3)
		SMOKER_CG_CORR_MAE0[i] = math.Abs(corrResult3[0] - corrr3)
		SMOKER_CG_CORR_MRE0[i] = math.Abs(corrResult3[0]-corrr3) / math.Abs(corrr3)
		SMOKER_CG_CORR_TIME0[i] = float64(SMOKER_CG_TIME.Seconds())
		fmt.Println("Correlation (BMI, CHARGE) <Basic>", corrResult3[0], math.Abs(corrResult3[0]-corrr3), math.Abs(corrResult3[0]-corrr3)/corrr3, SMOKER_CG_TIME)

		TIME = time.Now()
		corr3, _ = engine.PCorrCoeff(smoker, charge, B, false)
		SMOKER_CG_TIME = time.Since(TIME)
		corrResult3, _ = engine.Decrypt(corr3)
		SMOKER_CG_CORR_MAE1[i] = math.Abs(corrResult3[0] - corrr3)
		SMOKER_CG_CORR_MRE1[i] = math.Abs(corrResult3[0]-corrr3) / math.Abs(corrr3)
		SMOKER_CG_CORR_TIME1[i] = float64(SMOKER_CG_TIME.Seconds())
		fmt.Println("Correlation (BMI, CHARGE) <Fast>", corrResult3[0], math.Abs(corrResult3[0]-corrr3), math.Abs(corrResult3[0]-corrr3)/corrr3, SMOKER_CG_TIME)

		TIME = time.Now()
		corr3_ppstat, _ := engine.PCorrCoeff(smoker, charge, B, false)
		SMOKER_CG_PPSTAT_TIME := time.Since(TIME)
		corrResult3_ppstat, _ := engine.Decrypt(corr3_ppstat)
		SMOKER_CG_CORR_MAE_PPSTAT[i] = math.Abs(corrResult3_ppstat[0] - corrr3)
		SMOKER_CG_CORR_MRE_PPSTAT[i] = math.Abs(corrResult3_ppstat[0]-corrr3) / math.Abs(corrr3)
		SMOKER_CG_CORR_TIME_PPSTAT[i] = float64(SMOKER_CG_PPSTAT_TIME.Seconds())
	}
	result := ""
	result += fmt.Sprintf("[Ours-Basic][ZSCORE] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(CG_ZSCORE_MAE0), utils.StdDev(CG_ZSCORE_MAE0), utils.Mean(CG_ZSCORE_MRE0), utils.StdDev(CG_ZSCORE_MRE0), utils.Mean(CG_ZSCORE_TIME0), utils.StdDev(CG_ZSCORE_TIME0))
	result += fmt.Sprintf("[Ours-Basic][KURT] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(CG_KURT_MAE0), utils.StdDev(CG_KURT_MAE0), utils.Mean(CG_KURT_MRE0), utils.StdDev(CG_KURT_MRE0), utils.Mean(CG_KURT_TIME0), utils.StdDev(CG_KURT_TIME0))
	result += fmt.Sprintf("[Ours-Basic][SKEW] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(CG_SKEW_MAE0), utils.StdDev(CG_SKEW_MAE0), utils.Mean(CG_SKEW_MRE0), utils.StdDev(CG_SKEW_MRE0), utils.Mean(CG_SKEW_TIME0), utils.StdDev(CG_SKEW_TIME0))
	result += fmt.Sprintf("[Ours-Basic][AGE_CORR] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AGE_CG_CORR_MAE0), utils.StdDev(AGE_CG_CORR_MAE0), utils.Mean(AGE_CG_CORR_MRE0), utils.StdDev(AGE_CG_CORR_MRE0), utils.Mean(AGE_CG_CORR_TIME0), utils.StdDev(AGE_CG_CORR_TIME0))
	result += fmt.Sprintf("[Ours-Basic][BMI_CORR] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(BMI_CG_CORR_MAE0), utils.StdDev(BMI_CG_CORR_MAE0), utils.Mean(BMI_CG_CORR_MRE0), utils.StdDev(BMI_CG_CORR_MRE0), utils.Mean(BMI_CG_CORR_TIME0), utils.StdDev(BMI_CG_CORR_TIME0))
	result += fmt.Sprintf("[Ours-Basic][SMOKER_CORR] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(SMOKER_CG_CORR_MAE0), utils.StdDev(SMOKER_CG_CORR_MAE0), utils.Mean(SMOKER_CG_CORR_MRE0), utils.StdDev(SMOKER_CG_CORR_MRE0), utils.Mean(SMOKER_CG_CORR_TIME0), utils.StdDev(SMOKER_CG_CORR_TIME0))
	
	result += fmt.Sprintf("[Ours-Fast][ZSCORE] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(CG_ZSCORE_MAE1), utils.StdDev(CG_ZSCORE_MAE1), utils.Mean(CG_ZSCORE_MRE1), utils.StdDev(CG_ZSCORE_MRE1), utils.Mean(CG_ZSCORE_TIME1), utils.StdDev(CG_ZSCORE_TIME1))
	result += fmt.Sprintf("[Ours-Fast][KURT] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(CG_KURT_MAE1), utils.StdDev(CG_KURT_MAE1), utils.Mean(CG_KURT_MRE1), utils.StdDev(CG_KURT_MRE1), utils.Mean(CG_KURT_TIME1), utils.StdDev(CG_KURT_TIME1))
	result += fmt.Sprintf("[Ours-Fast][SKEW] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(CG_SKEW_MAE1), utils.StdDev(CG_SKEW_MAE1), utils.Mean(CG_SKEW_MRE1), utils.StdDev(CG_SKEW_MRE1), utils.Mean(CG_SKEW_TIME1), utils.StdDev(CG_SKEW_TIME1))
	result += fmt.Sprintf("[Ours-Fast][AGE_CORR] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AGE_CG_CORR_MAE1), utils.StdDev(AGE_CG_CORR_MAE1), utils.Mean(AGE_CG_CORR_MRE1), utils.StdDev(AGE_CG_CORR_MRE1), utils.Mean(AGE_CG_CORR_TIME1), utils.StdDev(AGE_CG_CORR_TIME1))
	result += fmt.Sprintf("[Ours-Fast][BMI_CORR] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(BMI_CG_CORR_MAE1), utils.StdDev(BMI_CG_CORR_MAE1), utils.Mean(BMI_CG_CORR_MRE1), utils.StdDev(BMI_CG_CORR_MRE1), utils.Mean(BMI_CG_CORR_TIME1), utils.StdDev(BMI_CG_CORR_TIME1))
	result += fmt.Sprintf("[Ours-Fast][SMOKER_CORR] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(SMOKER_CG_CORR_MAE1), utils.StdDev(SMOKER_CG_CORR_MAE1), utils.Mean(SMOKER_CG_CORR_MRE1), utils.StdDev(SMOKER_CG_CORR_MRE1), utils.Mean(SMOKER_CG_CORR_TIME1), utils.StdDev(SMOKER_CG_CORR_TIME1))
	
	result += fmt.Sprintf("[PPSTAT][ZSCORE] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(CG_ZSCORE_MAE_PPSTAT), utils.StdDev(CG_ZSCORE_MAE_PPSTAT), utils.Mean(CG_ZSCORE_MRE_PPSTAT), utils.StdDev(CG_ZSCORE_MRE_PPSTAT), utils.Mean(CG_ZSCORE_TIME_PPSTAT), utils.StdDev(CG_ZSCORE_TIME_PPSTAT))
	result += fmt.Sprintf("[PPSTAT][KURT] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(CG_KURT_MAE_PPSTAT), utils.StdDev(CG_KURT_MAE_PPSTAT), utils.Mean(CG_KURT_MRE_PPSTAT), utils.StdDev(CG_KURT_MRE_PPSTAT), utils.Mean(CG_KURT_TIME_PPSTAT), utils.StdDev(CG_KURT_TIME_PPSTAT))
	result += fmt.Sprintf("[PPSTAT][SKEW] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(CG_SKEW_MAE_PPSTAT), utils.StdDev(CG_SKEW_MAE_PPSTAT), utils.Mean(CG_SKEW_MRE_PPSTAT), utils.StdDev(CG_SKEW_MRE_PPSTAT), utils.Mean(CG_SKEW_TIME_PPSTAT), utils.StdDev(CG_SKEW_TIME_PPSTAT))
	result += fmt.Sprintf("[PPSTAT][AGE_CORR] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AGE_CG_CORR_MAE_PPSTAT), utils.StdDev(AGE_CG_CORR_MAE_PPSTAT), utils.Mean(AGE_CG_CORR_MRE_PPSTAT), utils.StdDev(AGE_CG_CORR_MRE_PPSTAT), utils.Mean(AGE_CG_CORR_TIME_PPSTAT), utils.StdDev(AGE_CG_CORR_TIME_PPSTAT))
	result += fmt.Sprintf("[PPSTAT][BMI_CORR] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(BMI_CG_CORR_MAE_PPSTAT), utils.StdDev(BMI_CG_CORR_MAE_PPSTAT), utils.Mean(BMI_CG_CORR_MRE_PPSTAT), utils.StdDev(BMI_CG_CORR_MRE_PPSTAT), utils.Mean(BMI_CG_CORR_TIME_PPSTAT), utils.StdDev(BMI_CG_CORR_TIME_PPSTAT))
	result += fmt.Sprintf("[PPSTAT][SMOKER_CORR] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(SMOKER_CG_CORR_MAE_PPSTAT), utils.StdDev(SMOKER_CG_CORR_MAE_PPSTAT), utils.Mean(SMOKER_CG_CORR_MRE_PPSTAT), utils.StdDev(SMOKER_CG_CORR_MRE_PPSTAT), utils.Mean(SMOKER_CG_CORR_TIME_PPSTAT), utils.StdDev(SMOKER_CG_CORR_TIME_PPSTAT))
	
	io.WriteString(file, result)
}
