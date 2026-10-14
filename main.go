package main

import (
	"fmt"
	"log"

	"baies/bn"
	"baies/data"
	"baies/metrics"
	"baies/output"
)

func main() {
	samples, err := data.LoadSamples("База Коарктации ДНК.xlsx", "Лист1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Пациентов после фильтрации: %d\n", len(samples))

	edges := [][2]string{
		{"pge1", "ncfdna_t4"},
		{"pge1", "mcfdna_t4"},
		{"ik_duration", "ncfdna_t4"},
		{"ik_duration", "mcfdna_t4"},
		{"clamp_duration", "mcfdna_t4"},
		{"ncfdna_t4", "prolonged_ventilation"},
		{"mcfdna_t4", "prolonged_ventilation"},
	}

	network, err := bn.BuildNetwork(edges)
	if err != nil {
		log.Fatal(err)
	}

	if err := bn.Fit(network, samples, 1.0, edges); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Сеть обучена (со сглаживанием a = 1.0).")

	for _, node := range network.Nodes() {
		cpd, err := network.GetCPD(node)
		if err != nil {
			log.Fatal(err)
		}
		output.PrintCPD(cpd)
	}

	evidence := map[string]int{
		"pge1":           1,
		"ik_duration":    1,
		"clamp_duration": 1,
		"ncfdna_t4":      2,
		"mcfdna_t4":      2,
	}
	values, err := bn.Predict(network, evidence, "prolonged_ventilation")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nПрогноз:\n")
	fmt.Printf("P(ИВЛ < 24 ч)  = %.4f\n", values[0])
	fmt.Printf("P(ИВЛ >= 24 ч) = %.4f\n", values[1])

	res, err := metrics.LOOCV(edges, samples, 1.0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nLOOCV\n")
	fmt.Printf("Всего пациентов: %d\n", res.Total)
	fmt.Printf("TP=%d, FP=%d, FN=%d, TN=%d\n", res.TP, res.FP, res.FN, res.TN)
	fmt.Printf("Accuracy = %.4f\n", res.Accuracy)
	fmt.Printf("Precision = %.4f\n", res.Precision)
	fmt.Printf("Recall = %.4f\n", res.Recall)
	fmt.Printf("F1 = %.4f\n", res.F1)
}