package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hm-choi/pp-stat-plus/config"
	"github.com/hm-choi/pp-stat-plus/engine"
	"github.com/hm-choi/pp-stat-plus/optimizer"
)

type CaseInfo struct {
	Case      int  `json:"case"`
	Degree    int     `json:"degree"`
	Iteration int     `json:"iteration"`
	Time      float64 `json:"time"`
	Mre       float64 `json:"mre"`
}

func main() {
	
	f, _ := os.Create("output.log")
	mw := io.MultiWriter(os.Stdout, f)
    log.SetOutput(mw)

	const (
		DATA_SIZE = 32768 // Slot size
		START     = 0.001
		MIDDLE	  = 1
		STOP      = 100.0
	)

	engine := engine.NewHEEngine(config.NewParameters(16, 11, 50, true))

	d_min, d_max := 4.0, 9.0
	i_max := 15

	R := optimizer.Optimizing(engine, d_min, d_max, i_max, START, MIDDLE, STOP, DATA_SIZE*2, 1.0, 1.0)

	fmt.Println(R)

	jsonData := make(map[string]map[string]CaseInfo)

	for level, tuples := range R {
		levelKey := fmt.Sprintf("%d", level)
		jsonData[levelKey] = make(map[string]CaseInfo)

		if len(tuples) < 1 {
			continue
		}

		u1 := tuples[0]
		u2 := tuples[1]

		if u1.T <= u2.T {
			u2 = optimizer.Rtuple{}
		} else if u1.M >= u2.M {
			u1 = optimizer.Rtuple{}
		}

		// u1
		if u1 != (optimizer.Rtuple{}) {
			caseName := "Basic"
			jsonData[levelKey][caseName] = CaseInfo{
				Case:      u1.C,
				Degree:    int(u1.D),
				Iteration: u1.I,
				Time:      u1.T,
				Mre:       u1.M,
			}
		}

		// u2
		if u2 != (optimizer.Rtuple{}) {
			caseName := "Fast"
			jsonData[levelKey][caseName] = CaseInfo{
				Case:      u2.C,
				Degree:    int(u2.D),
				Iteration: u2.I,
				Time:      u2.T,
				Mre:       u2.M,
			}
		}
	}

	out, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))

	if err := os.WriteFile("lattigo_optimizer.json", out, 0644); err != nil {
		panic(err)
	}
}
