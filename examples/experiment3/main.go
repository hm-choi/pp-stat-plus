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
	EVAL_NUM, B := 10, 50.0
	ageSlice, _ := utils.ReadCSV("../../examples/dataset/adult_dataset.csv", 0)
	hpwSlice, _ := utils.ReadCSV("../../examples/dataset/adult_dataset.csv", 12)
	eduSlice, _ := utils.ReadCSV("../../examples/dataset/adult_dataset.csv", 4)

	age, _ := engine.Encrypt(ageSlice, 11)
	hpw, _ := engine.Encrypt(hpwSlice, 11)
	edu, _ := engine.Encrypt(eduSlice, 11)

	_, _, skew_age := utils.Skewness(ageSlice)
	_, _, kurt_age := utils.Kurtosis(ageSlice)

	_, _, skew_hpw := utils.Skewness(hpwSlice)
	_, _, kurt_hpw := utils.Kurtosis(hpwSlice)

	_, _, skew_edu := utils.Skewness(eduSlice)
	_, _, kurt_edu := utils.Kurtosis(eduSlice)

	fmt.Println("utils.Mean(ageSlice), utils.Mean(hpwSlice), utils.Mean(eduSlice), ", utils.Mean(ageSlice), utils.Mean(hpwSlice), utils.Mean(eduSlice))

	AgeNorm_MRE, AgeNorm_MAE, AgeNorm_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	AgeSkew_MRE, AgeSkew_MAE, AgeSkew_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	AgeKurt_MRE, AgeKurt_MAE, AgeKurt_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)

	HPWNorm_MRE, HPWNorm_MAE, HPWNorm_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	HPWSkew_MRE, HPWSkew_MAE, HPWSkew_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	HPWKurt_MRE, HPWKurt_MAE, HPWKurt_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)

	EduNorm_MRE, EduNorm_MAE, EduNorm_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	EduSkew_MRE, EduSkew_MAE, EduSkew_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	EduKurt_MRE, EduKurt_MAE, EduKurt_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)

	AgeHPW_CORR_MRE, AgeHPW_CORR_MAE, AgeHPW_CORR_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	AGE_EDU_CORR_MRE, AGE_EDU_CORR_MAE, AGE_EDU_CORR_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)


	AgeNorm_MRE_PPSTAT, AgeNorm_MAE_PPSTAT, AgeNorm_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	AgeSkew_MRE_PPSTAT, AgeSkew_MAE_PPSTAT, AgeSkew_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	AgeKurt_MRE_PPSTAT, AgeKurt_MAE_PPSTAT, AgeKurt_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)

	HPWNorm_MRE_PPSTAT, HPWNorm_MAE_PPSTAT, HPWNorm_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	HPWSkew_MRE_PPSTAT, HPWSkew_MAE_PPSTAT, HPWSkew_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	HPWKurt_MRE_PPSTAT, HPWKurt_MAE_PPSTAT, HPWKurt_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)

	EduNorm_MRE_PPSTAT, EduNorm_MAE_PPSTAT, EduNorm_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	EduSkew_MRE_PPSTAT, EduSkew_MAE_PPSTAT, EduSkew_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	EduKurt_MRE_PPSTAT, EduKurt_MAE_PPSTAT, EduKurt_TIME_PPSTAT := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)

	AgeHPW_CORR_MRE_PPSTAT, AgeHPW_CORR_MAE_PPSTAT, AgeHPW_CORR_TIME_PPSTAT 	:= make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	AGE_EDU_CORR_MRE_PPSTAT, AGE_EDU_CORR_MAE_PPSTAT, AGE_EDU_CORR_TIME_PPSTAT 	:= make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)


	file, err := os.Create("output.txt")
	if err != nil {
		return
	}
	defer file.Close()

	for i := 0; i < int(EVAL_NUM); i++ {
		fmt.Println("[============ Age Test ============]")

		TIME := time.Now()
		zScoreNorm1, _ := engine.ZScoreNorm(age, B, false)
		AGE_ZNORM_TIME := time.Since(TIME)
		zSNAge, _ := engine.Decrypt(zScoreNorm1)
		fmt.Println("Age ZNorm", zSNAge[0:1], utils.ZScoreNorm(ageSlice)[:1], AGE_ZNORM_TIME)

		TIME = time.Now()
		zScoreNorm_ppstat1, _ := engine.ZScoreNorm_ppstat(age, B)
		AGE_ZNORM_PPSTAT_TIME := time.Since(TIME)
		zSNAge_ppstat, _ := engine.Decrypt(zScoreNorm_ppstat1)


		TIME = time.Now()
		skew1, _ := engine.Skewness(age, B, false)
		AGE_SKEW_TIME := time.Since(TIME)
		skewAge, _ := engine.Decrypt(skew1)
		AgeSkew_MRE[i] = math.Abs(skewAge[0]-skew_age) / math.Abs(skew_age)
		AgeSkew_MAE[i] = math.Abs(skewAge[0] - skew_age)
		AgeSkew_TIME[i] = float64(AGE_SKEW_TIME.Seconds())
		fmt.Println("Age skewResult", skewAge[0], math.Abs(skewAge[0]-skew_age), math.Abs(skewAge[0]-skew_age)/math.Abs(skew_age), AGE_SKEW_TIME)

		TIME = time.Now()
		skew_ppstat1, _ := engine.Skewness_ppstat(age, B)
		AGE_SKEW_PPSTAT_TIME := time.Since(TIME)
		skewAge_ppstat, _ := engine.Decrypt(skew_ppstat1)
		AgeSkew_MRE_PPSTAT[i] = math.Abs(skewAge_ppstat[0]-skew_age) / math.Abs(skew_age)
		AgeSkew_MAE_PPSTAT[i] = math.Abs(skewAge_ppstat[0] - skew_age)
		AgeSkew_TIME_PPSTAT[i] = float64(AGE_SKEW_PPSTAT_TIME.Seconds())


		TIME = time.Now()
		kurt1, _ := engine.Kurtosis(age, B, false)
		AGE_KURT_TIME := time.Since(TIME)
		kurtAge, _ := engine.Decrypt(kurt1)
		AgeKurt_MRE[i] = math.Abs(kurtAge[0]-kurt_age) / math.Abs(kurt_age)
		AgeKurt_MAE[i] = math.Abs(kurtAge[0] - kurt_age)
		AgeKurt_TIME[i] = float64(AGE_KURT_TIME.Seconds())
		fmt.Println("Age kurtResult", kurtAge[0], math.Abs(kurtAge[0]-kurt_age), math.Abs(kurtAge[0]-kurt_age)/math.Abs(kurt_age), AGE_KURT_TIME)

		TIME = time.Now()
		kurt_ppstat1, _ := engine.Kurtosis_ppstat(age, B)
		AGE_KURT_PPSTAT_TIME := time.Since(TIME)
		kurtAge_ppstat, _ := engine.Decrypt(kurt_ppstat1)
		AgeKurt_MRE_PPSTAT[i] = math.Abs(kurtAge_ppstat[0]-kurt_age) / math.Abs(kurt_age)
		AgeKurt_MAE_PPSTAT[i] = math.Abs(kurtAge_ppstat[0] - kurt_age)
		AgeKurt_TIME_PPSTAT[i] = float64(AGE_KURT_PPSTAT_TIME.Seconds())


		fmt.Println("[============ HPW Test ============]")

		TIME = time.Now()
		zScoreNorm2, _ := engine.ZScoreNorm(hpw, 100.0, false)
		HPW_ZNORM_TIME := time.Since(TIME)
		zSNHpw, _ := engine.Decrypt(zScoreNorm2)
		fmt.Println("HPW ZNorm", zSNHpw[0:1], utils.ZScoreNorm(zSNHpw)[:1], HPW_ZNORM_TIME)

		TIME = time.Now()
		zScoreNorm_ppstat2, _ := engine.ZScoreNorm_ppstat(hpw, 100.0)
		HPW_ZNORM_PPSTAT_TIME := time.Since(TIME)
		zSNHpw_ppstat, _ := engine.Decrypt(zScoreNorm_ppstat2)


		TIME = time.Now()
		skew2, _ := engine.Skewness(hpw, B, false)
		HPW_SKEW_TIME := time.Since(TIME)
		skewHpw, _ := engine.Decrypt(skew2)
		HPWSkew_MRE[i] = math.Abs(skewHpw[0]-skew_hpw) / math.Abs(skew_hpw)
		HPWSkew_MAE[i] = math.Abs(skewHpw[0] - skew_hpw)
		HPWSkew_TIME[i] = float64(HPW_SKEW_TIME.Seconds())
		fmt.Println("HPW skewResult", skewHpw[0], math.Abs(skewHpw[0]-skew_hpw), math.Abs(skewHpw[0]-skew_hpw)/math.Abs(skew_hpw), HPW_SKEW_TIME)

		TIME = time.Now()
		skew_ppstat2, _ := engine.Skewness(hpw, B, false)
		HPW_SKEW_PPSTAT_TIME := time.Since(TIME)
		skewHpw_ppstat, _ := engine.Decrypt(skew_ppstat2)
		HPWSkew_MRE_PPSTAT[i] = math.Abs(skewHpw_ppstat[0]-skew_hpw) / math.Abs(skew_hpw)
		HPWSkew_MAE_PPSTAT[i] = math.Abs(skewHpw_ppstat[0] - skew_hpw)
		HPWSkew_TIME_PPSTAT[i] = float64(HPW_SKEW_PPSTAT_TIME.Seconds())


		TIME = time.Now()
		kurt2, _ := engine.Kurtosis(hpw, B, false)
		HPW_KURT_TIME := time.Since(TIME)
		kurtHpw, _ := engine.Decrypt(kurt2)
		HPWKurt_MRE[i] = math.Abs(kurtHpw[0]-kurt_hpw) / math.Abs(kurt_hpw)
		HPWKurt_MAE[i] = math.Abs(kurtHpw[0] - kurt_hpw)
		HPWKurt_TIME[i] = float64(HPW_KURT_TIME.Seconds())
		fmt.Println("HPW kurtResult", kurtHpw[0], math.Abs(kurtHpw[0]-kurt_hpw), math.Abs(kurtHpw[0]-kurt_hpw)/math.Abs(kurt_hpw), HPW_KURT_TIME)

		TIME = time.Now()
		kurt_ppstat2, _ := engine.Kurtosis_ppstat(hpw, B)
		HPW_KURT_PPSTAT_TIME := time.Since(TIME)
		kurtHpw_ppstat, _ := engine.Decrypt(kurt_ppstat2)
		HPWKurt_MRE_PPSTAT[i] = math.Abs(kurtHpw_ppstat[0]-kurt_hpw) / math.Abs(kurt_hpw)
		HPWKurt_MAE_PPSTAT[i] = math.Abs(kurtHpw_ppstat[0] - kurt_hpw)
		HPWKurt_TIME_PPSTAT[i] = float64(HPW_KURT_PPSTAT_TIME.Seconds())


		fmt.Println("[============ Edu Test ============]")

		TIME = time.Now()
		zScoreNorm3, _ := engine.ZScoreNorm(edu, 100.0, false)
		EDU_ZNORM_TIME := time.Since(TIME)
		zSNEdu, _ := engine.Decrypt(zScoreNorm3)
		fmt.Println("Edu ZNorm", zSNEdu[0:1], utils.ZScoreNorm(zSNEdu)[:1], EDU_ZNORM_TIME)

		TIME = time.Now()
		zScoreNorm_ppstat3, _ := engine.ZScoreNorm_ppstat(edu, 100.0)
		EDU_ZNORM_PPSTAT_TIME := time.Since(TIME)
		zSNEdu_ppstat, _ := engine.Decrypt(zScoreNorm_ppstat3)


		TIME = time.Now()
		skew3, _ := engine.Skewness(edu, B, false)
		EDU_SKEW_TIME := time.Since(TIME)
		skewEdu, _ := engine.Decrypt(skew3)
		EduSkew_MRE[i] = math.Abs(skewEdu[0]-skew_edu) / math.Abs(skew_edu)
		EduSkew_MAE[i] = math.Abs(skewEdu[0] - skew_edu)
		EduSkew_TIME[i] = float64(EDU_SKEW_TIME.Seconds())
		fmt.Println("Edu skewResult", skewEdu[0], math.Abs(skewEdu[0]-skew_edu), math.Abs(skewEdu[0]-skew_edu)/math.Abs(skew_edu), EDU_SKEW_TIME)

		TIME = time.Now()
		skew_ppstat3, _ := engine.Skewness_ppstat(edu, B)
		EDU_SKEW_PPSTAT_TIME := time.Since(TIME)
		skewEdu_ppstat, _ := engine.Decrypt(skew_ppstat3)
		EduSkew_MRE_PPSTAT[i] = math.Abs(skewEdu_ppstat[0]-skew_edu) / math.Abs(skew_edu)
		EduSkew_MAE_PPSTAT[i] = math.Abs(skewEdu_ppstat[0] - skew_edu)
		EduSkew_TIME_PPSTAT[i] = float64(EDU_SKEW_PPSTAT_TIME.Seconds())


		TIME = time.Now()
		kurt3, _ := engine.Kurtosis(edu, B, false)
		EDU_KURT_TIME := time.Since(TIME)
		kurtEdu, _ := engine.Decrypt(kurt3)
		EduKurt_MRE[i] = math.Abs(kurtEdu[0]-kurt_edu) / math.Abs(kurt_edu)
		EduKurt_MAE[i] = math.Abs(kurtEdu[0] - kurt_edu)
		EduKurt_TIME[i] = float64(EDU_KURT_TIME.Seconds())
		fmt.Println("Edu kurtResult", kurtEdu[0], math.Abs(kurtEdu[0]-kurt_edu), math.Abs(kurtEdu[0]-kurt_edu)/math.Abs(kurt_edu), EDU_KURT_TIME)

		TIME = time.Now()
		kurt_ppstat3, _ := engine.Kurtosis_ppstat(edu, B)
		EDU_KURT_PPSTAT_TIME := time.Since(TIME)
		kurtEdu_ppstat, _ := engine.Decrypt(kurt_ppstat3)
		EduKurt_MRE_PPSTAT[i] = math.Abs(kurtEdu_ppstat[0]-kurt_edu) / math.Abs(kurt_edu)
		EduKurt_MAE_PPSTAT[i] = math.Abs(kurtEdu_ppstat[0] - kurt_edu)
		EduKurt_TIME_PPSTAT[i] = float64(EDU_KURT_PPSTAT_TIME.Seconds())


		fmt.Println("[============ Corr Test ============]")
		_, corrr, _ := utils.Correlation(ageSlice, hpwSlice)
		TIME = time.Now()
		corr, _ := engine.PCorrCoeff(age, hpw, B, false)
		CORR_TIME := time.Since(TIME)
		corrtResult, _ := engine.Decrypt(corr)
		AgeHPW_CORR_MRE[i] = math.Abs(corrtResult[0]-corrr) / math.Abs(corrr)
		AgeHPW_CORR_MAE[i] = math.Abs(corrtResult[0] - corrr)
		AgeHPW_CORR_TIME[i] = float64(CORR_TIME.Seconds())
		fmt.Println("corrtResult(AGE vs HPW)", corrtResult[0], corrr, math.Abs(corrtResult[0]-corrr), math.Abs(corrtResult[0]-corrr)/math.Abs(corrr), CORR_TIME)

		TIME = time.Now()
		corr_ppstat, _ := engine.PCorrCoeff_ppstat(age, hpw, B)
		CORR_PPSTAT_TIME := time.Since(TIME)
		corrtResult_ppstat, _ := engine.Decrypt(corr_ppstat)
		AgeHPW_CORR_MRE_PPSTAT[i] = math.Abs(corrtResult_ppstat[0]-corrr) / math.Abs(corrr)
		AgeHPW_CORR_MAE_PPSTAT[i] = math.Abs(corrtResult_ppstat[0] - corrr)
		AgeHPW_CORR_TIME_PPSTAT[i] = float64(CORR_PPSTAT_TIME.Seconds())

		_, corrr, _ = utils.Correlation(ageSlice, eduSlice)
		TIME = time.Now()
		corr, _ = engine.PCorrCoeff(age, edu, B, false)
		CORR_TIME = time.Since(TIME)
		corrtResult, _ = engine.Decrypt(corr)
		AGE_EDU_CORR_MRE[i] = math.Abs(corrtResult[0]-corrr) / math.Abs(corrr)
		AGE_EDU_CORR_MAE[i] = math.Abs(corrtResult[0] - corrr)
		AGE_EDU_CORR_TIME[i] = float64(CORR_TIME.Seconds())
		fmt.Println("corrtResult(AGE vs EDU)", corrtResult[0], corrr, math.Abs(corrtResult[0]-corrr), math.Abs(corrtResult[0]-corrr)/math.Abs(corrr), CORR_TIME)

		TIME = time.Now()
		corr_ppstat, _ = engine.PCorrCoeff_ppstat(age, edu, B)
		CORR_PPSTAT_TIME = time.Since(TIME)
		corrtResult_ppstat, _ = engine.Decrypt(corr_ppstat)
		AGE_EDU_CORR_MRE_PPSTAT[i] = math.Abs(corrtResult_ppstat[0]-corrr) / math.Abs(corrr)
		AGE_EDU_CORR_MAE_PPSTAT[i] = math.Abs(corrtResult_ppstat[0] - corrr)
		AGE_EDU_CORR_TIME_PPSTAT[i] = float64(CORR_PPSTAT_TIME.Seconds())

		ageZNorm := utils.ZScoreNorm(ageSlice)
		hpwZNorm := utils.ZScoreNorm(hpwSlice)
		eduZNorm := utils.ZScoreNorm(eduSlice)
		_, zScoreMreAge := utils.CheckMRE(zSNAge, zSNAge, ageZNorm, len(ageSlice))
		_, zScoreMaeAge := utils.CheckMAE(zSNAge, zSNAge, ageZNorm, len(ageSlice))
		_, zScoreMreAge_ppstat := utils.CheckMRE(zSNAge_ppstat, zSNAge_ppstat, ageZNorm, len(ageSlice))
		_, zScoreMaeAge_ppstat := utils.CheckMAE(zSNAge_ppstat, zSNAge_ppstat, ageZNorm, len(ageSlice))

		_, zScoreMreHpw := utils.CheckMRE(zSNHpw, zSNHpw, hpwZNorm, len(hpwSlice))
		_, zScoreMaeHpw := utils.CheckMAE(zSNHpw, zSNHpw, hpwZNorm, len(hpwSlice))
		_, zScoreMreHpw_ppstat := utils.CheckMRE(zSNHpw_ppstat, zSNHpw_ppstat, hpwZNorm, len(hpwSlice))
		_, zScoreMaeHpw_ppstat := utils.CheckMAE(zSNHpw_ppstat, zSNHpw_ppstat, hpwZNorm, len(hpwSlice))

		_, zScoreMreEdu := utils.CheckMRE(zSNEdu, zSNEdu, eduZNorm, len(eduSlice))
		_, zScoreMaeEdu := utils.CheckMAE(zSNEdu, zSNEdu, eduZNorm, len(eduSlice))
		_, zScoreMreEdu_ppstat := utils.CheckMRE(zSNEdu_ppstat, zSNEdu_ppstat, eduZNorm, len(eduSlice))
		_, zScoreMaeEdu_ppstat := utils.CheckMAE(zSNEdu_ppstat, zSNEdu_ppstat, eduZNorm, len(eduSlice))

		AgeNorm_MRE[i] = zScoreMreAge
		AgeNorm_MAE[i] = zScoreMaeAge
		AgeNorm_TIME[i] = float64(AGE_ZNORM_TIME.Seconds())
		HPWNorm_MRE[i] = zScoreMreHpw
		HPWNorm_MAE[i] = zScoreMaeHpw
		HPWNorm_TIME[i] = float64(HPW_ZNORM_TIME.Seconds())
		EduNorm_MRE[i] = zScoreMreEdu
		EduNorm_MAE[i] = zScoreMaeEdu
		EduNorm_TIME[i] = float64(EDU_ZNORM_TIME.Seconds())
		fmt.Println("ZNorm Age (MRE, MAE)", zScoreMreAge, zScoreMaeAge, AGE_ZNORM_TIME)
		fmt.Println("ZNorm Hpw (MRE, MAE)", zScoreMreHpw, zScoreMaeHpw, HPW_ZNORM_TIME)
		fmt.Println("ZNorm Edu (MRE, MAE)", zScoreMreEdu, zScoreMaeEdu, EDU_ZNORM_TIME)

		AgeNorm_MRE_PPSTAT[i] = zScoreMreAge_ppstat
		AgeNorm_MAE_PPSTAT[i] = zScoreMaeAge_ppstat
		AgeNorm_TIME_PPSTAT[i] = float64(AGE_ZNORM_PPSTAT_TIME.Seconds())
		HPWNorm_MRE_PPSTAT[i] = zScoreMreHpw_ppstat
		HPWNorm_MAE_PPSTAT[i] = zScoreMaeHpw_ppstat
		HPWNorm_TIME_PPSTAT[i] = float64(HPW_ZNORM_PPSTAT_TIME.Seconds())
		EduNorm_MRE_PPSTAT[i] = zScoreMreEdu_ppstat
		EduNorm_MAE_PPSTAT[i] = zScoreMaeEdu_ppstat
		EduNorm_TIME_PPSTAT[i] = float64(EDU_ZNORM_PPSTAT_TIME.Seconds())
	}

	fmt.Println("[============ END ============]")
	result := fmt.Sprintf("[Ours][AGE_ZSCORE] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AgeNorm_MAE), utils.StdDev(AgeNorm_MAE), utils.Mean(AgeNorm_MRE), utils.StdDev(AgeNorm_MRE), utils.Mean(AgeNorm_TIME), utils.StdDev(AgeNorm_TIME))
	result += fmt.Sprintf("[Ours][AGE_SKEW] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AgeSkew_MAE), utils.StdDev(AgeSkew_MAE), utils.Mean(AgeSkew_MRE), utils.StdDev(AgeSkew_MRE), utils.Mean(AgeSkew_TIME), utils.StdDev(AgeSkew_TIME))
	result += fmt.Sprintf("[Ours][AGE_KURT] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AgeKurt_MAE), utils.StdDev(AgeKurt_MAE), utils.Mean(AgeKurt_MRE), utils.StdDev(AgeKurt_MRE), utils.Mean(AgeKurt_TIME), utils.StdDev(AgeKurt_TIME))
	result += fmt.Sprintf("[Ours][HPW_ZSCORE] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(HPWNorm_MAE), utils.StdDev(HPWNorm_MAE), utils.Mean(HPWNorm_MRE), utils.StdDev(HPWNorm_MRE), utils.Mean(HPWNorm_TIME), utils.StdDev(HPWNorm_TIME))
	result += fmt.Sprintf("[Ours][HPW_SKEW] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(HPWSkew_MAE), utils.StdDev(HPWSkew_MAE), utils.Mean(HPWSkew_MRE), utils.StdDev(HPWSkew_MRE), utils.Mean(HPWSkew_TIME), utils.StdDev(HPWSkew_TIME))
	result += fmt.Sprintf("[Ours][HPW_KURT] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(HPWKurt_MAE), utils.StdDev(HPWKurt_MAE), utils.Mean(HPWKurt_MRE), utils.StdDev(HPWKurt_MRE), utils.Mean(HPWKurt_TIME), utils.StdDev(HPWKurt_TIME))
	result += fmt.Sprintf("[Ours][EDU_ZSCORE] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(EduNorm_MAE), utils.StdDev(EduNorm_MAE), utils.Mean(EduNorm_MRE), utils.StdDev(EduNorm_MRE), utils.Mean(EduNorm_TIME), utils.StdDev(EduNorm_TIME))
	result += fmt.Sprintf("[Ours][EDU_SKEW] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(EduSkew_MAE), utils.StdDev(EduSkew_MAE), utils.Mean(EduSkew_MRE), utils.StdDev(EduSkew_MRE), utils.Mean(EduSkew_TIME), utils.StdDev(EduSkew_TIME))
	result += fmt.Sprintf("[Ours][EDU_KURT] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(EduKurt_MAE), utils.StdDev(EduKurt_MAE), utils.Mean(EduKurt_MRE), utils.StdDev(EduKurt_MRE), utils.Mean(EduKurt_TIME), utils.StdDev(EduKurt_TIME))
	result += fmt.Sprintf("[Ours][AGEHPW_CORR] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AgeHPW_CORR_MAE), utils.StdDev(AgeHPW_CORR_MAE), utils.Mean(AgeHPW_CORR_MRE), utils.StdDev(AgeHPW_CORR_MRE), utils.Mean(AgeHPW_CORR_TIME), utils.StdDev(AgeHPW_CORR_TIME))
	result += fmt.Sprintf("[Ours][AGEEDU_CORR] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AGE_EDU_CORR_MAE), utils.StdDev(AGE_EDU_CORR_MAE), utils.Mean(AGE_EDU_CORR_MRE), utils.StdDev(AGE_EDU_CORR_MRE), utils.Mean(AGE_EDU_CORR_TIME), utils.StdDev(AGE_EDU_CORR_TIME))
	io.WriteString(file, result)

	result_ppstat := fmt.Sprintf("[PP-STAT][AGE_ZSCORE] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AgeNorm_MAE_PPSTAT), utils.StdDev(AgeNorm_MAE_PPSTAT), utils.Mean(AgeNorm_MRE_PPSTAT), utils.StdDev(AgeNorm_MRE_PPSTAT), utils.Mean(AgeNorm_TIME_PPSTAT), utils.StdDev(AgeNorm_TIME_PPSTAT))
	result_ppstat += fmt.Sprintf("[PP-STAT][AGE_SKEW] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AgeSkew_MAE_PPSTAT), utils.StdDev(AgeSkew_MAE_PPSTAT), utils.Mean(AgeSkew_MRE_PPSTAT), utils.StdDev(AgeSkew_MRE_PPSTAT), utils.Mean(AgeSkew_TIME_PPSTAT), utils.StdDev(AgeSkew_TIME_PPSTAT))
	result_ppstat += fmt.Sprintf("[PP-STAT][AGE_KURT] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AgeKurt_MAE_PPSTAT), utils.StdDev(AgeKurt_MAE_PPSTAT), utils.Mean(AgeKurt_MRE_PPSTAT), utils.StdDev(AgeKurt_MRE_PPSTAT), utils.Mean(AgeKurt_TIME_PPSTAT), utils.StdDev(AgeKurt_TIME_PPSTAT))
	result_ppstat += fmt.Sprintf("[PP-STAT][HPW_ZSCORE] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(HPWNorm_MAE_PPSTAT), utils.StdDev(HPWNorm_MAE_PPSTAT), utils.Mean(HPWNorm_MRE_PPSTAT), utils.StdDev(HPWNorm_MRE_PPSTAT), utils.Mean(HPWNorm_TIME_PPSTAT), utils.StdDev(HPWNorm_TIME_PPSTAT))
	result_ppstat += fmt.Sprintf("[PP-STAT][HPW_KURT] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(HPWKurt_MAE_PPSTAT), utils.StdDev(HPWKurt_MAE_PPSTAT), utils.Mean(HPWKurt_MRE_PPSTAT), utils.StdDev(HPWKurt_MRE_PPSTAT), utils.Mean(HPWKurt_TIME_PPSTAT), utils.StdDev(HPWKurt_TIME_PPSTAT))
	result_ppstat += fmt.Sprintf("[PP-STAT][EDU_ZSCORE] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(EduNorm_MAE_PPSTAT), utils.StdDev(EduNorm_MAE_PPSTAT), utils.Mean(EduNorm_MRE_PPSTAT), utils.StdDev(EduNorm_MRE_PPSTAT), utils.Mean(EduNorm_TIME_PPSTAT), utils.StdDev(EduNorm_TIME_PPSTAT))
	result_ppstat += fmt.Sprintf("[PP-STAT][EDU_SKEW] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(EduSkew_MAE_PPSTAT), utils.StdDev(EduSkew_MAE_PPSTAT), utils.Mean(EduSkew_MRE_PPSTAT), utils.StdDev(EduSkew_MRE_PPSTAT), utils.Mean(EduSkew_TIME_PPSTAT), utils.StdDev(EduSkew_TIME_PPSTAT))
	result_ppstat += fmt.Sprintf("[PP-STAT][EDU_KURT] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(EduKurt_MAE_PPSTAT), utils.StdDev(EduKurt_MAE_PPSTAT), utils.Mean(EduKurt_MRE_PPSTAT), utils.StdDev(EduKurt_MRE_PPSTAT), utils.Mean(EduKurt_TIME_PPSTAT), utils.StdDev(EduKurt_TIME_PPSTAT))
	result_ppstat += fmt.Sprintf("[PP-STAT][AGEHPW_CORR] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AgeHPW_CORR_MAE_PPSTAT), utils.StdDev(AgeHPW_CORR_MAE_PPSTAT), utils.Mean(AgeHPW_CORR_MRE_PPSTAT), utils.StdDev(AgeHPW_CORR_MRE_PPSTAT), utils.Mean(AgeHPW_CORR_TIME_PPSTAT), utils.StdDev(AgeHPW_CORR_TIME_PPSTAT))
	result_ppstat += fmt.Sprintf("[PP-STAT][AGEEDU_CORR] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n\n", utils.Mean(AGE_EDU_CORR_MAE_PPSTAT), utils.StdDev(AGE_EDU_CORR_MAE_PPSTAT), utils.Mean(AGE_EDU_CORR_MRE_PPSTAT), utils.StdDev(AGE_EDU_CORR_MRE_PPSTAT), utils.Mean(AGE_EDU_CORR_TIME_PPSTAT), utils.StdDev(AGE_EDU_CORR_TIME_PPSTAT))
	io.WriteString(file, result_ppstat)
}
