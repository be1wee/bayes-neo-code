package output

import (
	"fmt"
	"strings"

	"github.com/JohnPierman/bngo/factors"
)

func label(val int, variable string) string {
	switch variable {
	case "pge1":
		if val == 0 {
			return "нет"
		}
		return "да"
	case "ik_duration":
		if val == 0 {
			return "<120"
		}
		return ">=120"
	case "clamp_duration":
		if val == 0 {
			return "<30"
		}
		return ">=30"
	case "ncfdna_t4", "mcfdna_t4":
		switch val {
		case 0:
			return "низкий"
		case 1:
			return "средний"
		case 2:
			return "высокий"
		}
	case "prolonged_ventilation":
		if val == 0 {
			return "<24ч"
		}
		return ">=24ч"
	}
	return "?"
}

func runes(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

func pad(s string, w int) string {
	n := runes(s)
	if n >= w {
		return s
	}
	return s + strings.Repeat(" ", w-n)
}

func printRow(cells []string, widths []int) {
	parts := make([]string, len(cells))
	for i, c := range cells {
		parts[i] = pad(c, widths[i])
	}
	fmt.Println(strings.Join(parts, "  "))
}

func PrintCPD(cpd *factors.TabularCPD) {
	variable := cpd.Variable
	parents := cpd.Evidence
	values := cpd.Values
	variableCard := cpd.VariableCard

	fmt.Printf("\n%s\n", variable)

	if len(parents) == 0 {
		headers := []string{variable, "P"}
		widths := []int{runes(variable), 8}
		printRow(headers, widths)
		for i := 0; i < variableCard; i++ {
			printRow([]string{label(i, variable), fmt.Sprintf("%.4f", values[0][i])}, widths)
		}
		return
	}

	parentCards := make([]int, len(parents))
	for i, p := range parents {
		parentCards[i] = cpd.EvidenceCard[p]
	}

	headers := make([]string, 0)
	widths := make([]int, 0)
	for _, p := range parents {
		headers = append(headers, p)
		widths = append(widths, runes(p))
	}
	for k := 0; k < variableCard; k++ {
		h := "P(" + label(k, variable) + ")"
		headers = append(headers, h)
		widths = append(widths, runes(h))
	}

	printRow(headers, widths)

	idx := 0
	var generate func(pos int, current []int)
	generate = func(pos int, current []int) {
		if pos == len(parents) {
			cells := make([]string, 0)
			for i, v := range current {
				cells = append(cells, label(v, parents[i]))
			}
			for k := 0; k < variableCard; k++ {
				cells = append(cells, fmt.Sprintf("%.4f", values[idx][k]))
			}
			printRow(cells, widths)
			idx++
			return
		}
		for v := 0; v < parentCards[pos]; v++ {
			generate(pos+1, append(current, v))
		}
	}
	generate(0, []int{})
}