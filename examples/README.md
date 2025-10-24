# Experimental Evaluation

There are four example codes and two datasets for the performance evaluation of PP-STAT with HE-DAP.

## Usage of Experiments
To run the n-th experiment, 
(1) Change the directory into *experiment{n}*  \
(2) Run the following command:
```bash
    go run main.go
```

## Dataset
We use two real-world datasets: the UCI Adult Income dataset [1] and the Medical Cost dataset [2].
These datasets are provided in the following CSV files:
```
    examples\dataset\adult_dataset.csv
    examples\dataset\insurance.csv
```

## Experiment 1: Performance of Inverse Square Root
We evaluate the accuracy and efficiency of HE-DAP and compare them to those of the existing SotA methods, HEaaN-STAT [3] and
PP-STAT [4]. l denotes the ciphertext level, d denotes the detree of the Chebyshev polynomials, c denotes the Pre-BTS predictor, and i denotes the number of iterations.

### Table. Inverse square root computation on Lattigo.

| Method                    | l  | d    | c | i  | MRE               | Time(s)  |
|---------------------------|----|------|---|----|-------------------|----------|
| HEaaN-STAT                | 11 | -    | - | 21 | 5.28 × 10⁻³       | 232.0    |
| PP-STAT                   | 11 | 2⁹−2 | - |  6 | 5.38 × 10⁻⁷       | 98.2     |
| PP-STAT w/ **HE-DAP**     | 11 | 2⁷−2 | 0 |  5 | 6.81 × 10⁻⁹       | 50.2     |
| HEaaN-STAT                | 7  | -    | - | 21 | 5.31 × 10⁻³       | 451.0    |
| PP-STAT                   | 7  | 2⁹−2 | - |  6 | 5.05 × 10⁻⁴       | 142.0    |
| PP-STAT w/ **HE-DAP (A)** | 7  | 2⁶−2 | 0 |  8 | 6.81 × 10⁻⁹       | 137.0    |
| PP-STAT w/ **HE-DAP (S)** | 7  | 2⁶−2 | 1 |  3 | 2.78 × 10⁻⁴       | 92.0     |


## Experiment 2: Performance of Statistical Operations
We evaluate HE-DAP on the statistical operations in PP-STAT [3] to verify the effect of parameter optimization. These operations include Z-score normalization (ZNorm), skewness (Skew), kurtosis (Kurt), and the Pearson correlation coefficient (PCC). B denotes the scaling constant, and the results for ciphertext level 11 are shown.

### Table. Performance of statistical measures on Lattigo

| Measure      | Method                | B   | MRE              | Time(s)  |
|--------------|-----------------------|-----|------------------|-----------|
| Z-Score      | PP-STAT               | 100 | 4.687 × 10⁻⁸     | 145.64    |
|              | PP-STAT w/ **HE-DAP** | 100 | 3.290 × 10⁻⁸     | 96.25     |
| Skewness     | PP-STAT               | 20  | 1.506 × 10⁻⁶     | 157.83    |
|              | PP-STAT w/ **HE-DAP** | 20  | 1.451 × 10⁻⁶     | 109.81    |
| Kurtosis     | PP-STAT               | 20  | 2.276 × 10⁻⁷     | 155.71    |
|              | PP-STAT w/ **HE-DAP** | 20  | 1.299 × 10⁻⁷     | 107.08    |
| Correlation  | PP-STAT               | 20  | 1.181 × 10⁻⁷     | 290.97    |
|              | PP-STAT w/ **HE-DAP** | 20  | 7.410 × 10⁻⁸     | 197.12    |


## Experiment 3: Evaluation on Real-world Datasets
We evaluate the performance of HE-DAP using four statistical measures—ZNorm, Skew, Kurt, and PCC—on the same real-world dataset.

### Expeirment 3.1 
### Table. Performance of statistical measures over the *Adult* dataset (with fixed scaling factor 𝐵 = 50). Runtime reduction (R) is computed as (1 − (𝑏)/(𝑎)) × 100%. Kurtosis is reported as excess kurtosis (i.e., normal kurtosis minus 3).

| Operation | Feature(s)  | PP-STAT MRE | (a) PP-STAT Runtime(s) | PP-STAT w/ **HE-DAP** MRE | (b) PP-STAT w/ **HE-DAP** Runtime(s) | R (%) |
|-----------|-------------|-------------|-------------------------|---------------------------|---------------------------------------|-------|
| ZNorm | AGE | 2.82 × 10⁻⁸ | 110.83 | 2.12 × 10⁻⁸ | 62.89 | **43.27** |
|       | EDU | 5.21 × 10⁻⁸ | 109.86 | 5.30 × 10⁻⁸ | 61.02 | **44.49** |
|       | HPW | 5.93 × 10⁻⁸ | 109.47 | 5.23 × 10⁻⁸ | 61.09 | **44.19** |
| Skew  | AGE | 5.97 × 10⁻⁸ | 112.86 | 5.92 × 10⁻⁸ | 64.25 | **43.09** |
|       | EDU | 8.63 × 10⁻⁸ | 112.91 | 8.88 × 10⁻⁸ | 63.49 | **43.77** |
|       | HPW | 1.04 × 10⁻⁷ | 112.43 | 4.63 × 10⁻⁸ | 63.48 | **43.54** |
| Kurt  | AGE | 5.21 × 10⁻⁶ | 113.00 | 1.20 × 10⁻⁶ | 63.75 | **43.60** |
|       | EDU | 6.41 × 10⁻⁷ | 112.87 | 6.41 × 10⁻⁷ | 64.16 | **43.19** |
|       | HPW | 3.54 × 10⁻⁷ | 112.79 | 6.41 × 10⁻⁷ | 63.09 | **44.05** |
| PCC   | AGE vs HPW | 2.50 × 10⁻⁸ | 223.32 | 3.39 × 10⁻⁸ | 124.86 | **44.07** |
|       | AGE vs EDU | 2.65 × 10⁻⁸ | 223.41 | 4.65 × 10⁻⁸ | 124.49 | **44.26** |

*Abbreviations: AGE, EDU = education-num, HPW = hours-per-week*


### Experiment 3.2
### Table. Performance of statistical measures over the *Insurance* dataset. The scaling factor 𝐵 is set to 100 for Z-score normalization and 20 for all other evaluations. Runtime reduction (R) is computed as (1 − (𝑏)/(𝑎)) × 100%. Kurtosis is reported as excess kurtosis (i.e., normal kurtosis minus 3).


| Operation  | Feature(s)       | PP-STAT MRE | (a) PP-STAT Runtime(s) | PP-STAT w/ **HE-DAP** MRE | (b) PP-STAT w/ **HE-DAP** Runtime(s) | R(%)  |
|------------|------------------|-------------|----------------|-------------|----------------|-------|
| ZNorm      | Charges          | 1.88 × 10⁻⁸ | 110.29         | 1.22 × 10⁻⁸ | 61.07          | 44.64 |
| Skew       | Charges          | 5.92 × 10⁻⁸ | 112.34         | 3.74 × 10⁻⁸ | 63.41          | 43.56 |
| Kurt       | Charges          | 2.96 × 10⁻⁷ | 112.32         | 1.07 × 10⁻⁷ | 63.56          | 43.43 |
| PCC        | AGE vs Charges   | 1.84 × 10⁻⁸ | 222.26         | 2.97 × 10⁻⁸ | 124.69         | 43.91 |
|            | BMI vs Charges   | 3.29 × 10⁻⁸ | 222.64         | 4.01 × 10⁻⁸ | 124.16         | 44.22 |
|            | Smoker vs Charges| 1.22 × 10⁻⁸ | 222.74         | 2.89 × 10⁻⁸ | 123.57         | 44.31 |

*Abbreviations: AGE, Body Mass Index: BMI*

[1] Barry Becker and Ronny Kohavi. 1996. Adult. UCI Machine Learning Repository.
DOI: https://doi.org/10.24432/C5XW20. \
[2] Nahida Akter and Ashadun Nobi. 2018. Investigation of the financial stability of S&P 500 using realized volatility and stock returns distribution. Journal of Risk and Financial Management 11, 2 (2018), 22. \
[3] Younho Lee, Jinyeong Seo, Yujin Name, Jiseok Chae, and Jung Hee Cheon. 2023. HEaaN-STAT: a privacy-preserving statistical analysis toolkit for large-scale numerical, ordinal, and categorical data. IEEE Transactions on Dependable and Secure Computing (2023). \
[4] Hyunmin Choi. 2025. PP-STAT: An Efficient Privacy-Preserving Statistical Analysis Framework using Homomorphic Encryption. arXiv preprint arXiv:2508.12093(2025). To appear in CIKM2025.
