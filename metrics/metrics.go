package metrics

import (
	"baies/bn"

	"github.com/JohnPierman/bngo/models"
)

type Result struct {
	Total     int
	TP, FP, FN, TN int
	Accuracy  float64
	Precision float64
	Recall    float64
	F1        float64
}

func LOOCV(edges [][2]string, samples []map[string]int, alpha float64) (Result, error) {
	var res Result
	res.Total = len(samples)

	for i := 0; i < len(samples); i++ {
		var train []map[string]int
		for j := 0; j < len(samples); j++ {
			if j != i {
				train = append(train, samples[j])
			}
		}

		bnLoo, err := models.NewBayesianNetwork(edges)
		if err != nil {
			return res, err
		}
		if err := bn.Fit(bnLoo, train, alpha, edges); err != nil {
			return res, err
		}

		testSample := samples[i]
		ev := map[string]int{
			"pge1":           testSample["pge1"],
			"ik_duration":    testSample["ik_duration"],
			"clamp_duration": testSample["clamp_duration"],
			"ncfdna_t4":      testSample["ncfdna_t4"],
			"mcfdna_t4":      testSample["mcfdna_t4"],
		}

		values, err := bn.Predict(bnLoo, ev, "prolonged_ventilation")
		if err != nil {
			return res, err
		}

		predicted := 0
		if values[1] >= 0.5 {
			predicted = 1
		}
		actual := testSample["prolonged_ventilation"]

		if predicted == 1 && actual == 1 {
			res.TP++
		}
		if predicted == 1 && actual == 0 {
			res.FP++
		}
		if predicted == 0 && actual == 1 {
			res.FN++
		}
		if predicted == 0 && actual == 0 {
			res.TN++
		}
	}

	correct := res.TP + res.TN
	res.Accuracy = float64(correct) / float64(res.Total)
	if res.TP+res.FP > 0 {
		res.Precision = float64(res.TP) / float64(res.TP+res.FP)
	}
	if res.TP+res.FN > 0 {
		res.Recall = float64(res.TP) / float64(res.TP+res.FN)
	}
	if res.Precision+res.Recall > 0 {
		res.F1 = 2 * res.Precision * res.Recall / (res.Precision + res.Recall)
	}

	return res, nil
}