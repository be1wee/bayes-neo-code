package bn

import (
	"github.com/JohnPierman/bngo/factors"
	"github.com/JohnPierman/bngo/inference"
	"github.com/JohnPierman/bngo/models"
)

func BuildNetwork(edges [][2]string) (*models.BayesianNetwork, error) {
	return models.NewBayesianNetwork(edges)
}

func getParents(edges [][2]string, node string) []string {
	var parents []string
	for _, e := range edges {
		if e[1] == node {
			parents = append(parents, e[0])
		}
	}
	return parents
}

func getCardinality(samples []map[string]int, variable string) int {
	max := -1
	for _, s := range samples {
		if v, ok := s[variable]; ok {
			if v > max {
				max = v
			}
		}
	}
	return max + 1
}

func makeKey(parents []string, sample map[string]int) string {
	key := ""
	for _, p := range parents {
		key += string(rune(sample[p] + '0'))
	}
	return key
}

func Fit(bn *models.BayesianNetwork, samples []map[string]int, alpha float64, edges [][2]string) error {
	for _, node := range bn.Nodes() {
		parents := getParents(edges, node)
		nodeCard := getCardinality(samples, node)

		type keyCount struct {
			counts []int
		}
		table := make(map[string]*keyCount)

		for _, s := range samples {
			key := makeKey(parents, s)
			if _, ok := table[key]; !ok {
				table[key] = &keyCount{counts: make([]int, nodeCard)}
			}
			table[key].counts[s[node]]++
		}

		var values [][]float64
		if len(parents) == 0 {
			key := ""
			kc, ok := table[key]
			if !ok {
				kc = &keyCount{counts: make([]int, nodeCard)}
			}
			row := make([]float64, nodeCard)
			sum := 0.0
			for i := 0; i < nodeCard; i++ {
				row[i] = float64(kc.counts[i]) + alpha
				sum += row[i]
			}
			for i := range row {
				row[i] /= sum
			}
			values = append(values, row)
		} else {
			parentCards := make([]int, len(parents))
			for i, p := range parents {
				parentCards[i] = getCardinality(samples, p)
			}

			var generate func(idx int, current []int)
			generate = func(idx int, current []int) {
				if idx == len(parents) {
					key := ""
					for _, v := range current {
						key += string(rune(v + '0'))
					}
					kc, ok := table[key]
					if !ok {
						kc = &keyCount{counts: make([]int, nodeCard)}
					}
					row := make([]float64, nodeCard)
					sum := 0.0
					for i := 0; i < nodeCard; i++ {
						row[i] = float64(kc.counts[i]) + alpha
						sum += row[i]
					}
					for i := range row {
						row[i] /= sum
					}
					values = append(values, row)
					return
				}
				for v := 0; v < parentCards[idx]; v++ {
					generate(idx+1, append(current, v))
				}
			}
			generate(0, []int{})
		}

		evidenceCard := make(map[string]int)
		for _, p := range parents {
			evidenceCard[p] = getCardinality(samples, p)
		}

		newCPD, err := factors.NewTabularCPD(node, nodeCard, values, parents, evidenceCard)
		if err != nil {
			return err
		}
		if err := bn.AddCPD(newCPD); err != nil {
			return err
		}
	}
	return nil
}

func Predict(bn *models.BayesianNetwork, evidence map[string]int, target string) ([]float64, error) {
	ve, err := inference.NewVariableElimination(bn)
	if err != nil {
		return nil, err
	}
	result, err := ve.Query([]string{target}, evidence)
	if err != nil {
		return nil, err
	}
	return result.Values, nil
}