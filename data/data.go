package data

import (
	"sort"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

func discPGE1(v string) (int, bool) {
	v = strings.TrimSpace(v)
	if v == "0" {
		return 0, true
	}
	if v == "1" || v == "1*" {
		return 1, true
	}
	return 0, false
}

func discIK(v string) (int, bool) {
	x, err := strconv.ParseFloat(strings.ReplaceAll(v, ",", "."), 64)
	if err != nil {
		return 0, false
	}
	if x < 120 {
		return 0, true
	}
	return 1, true
}

func discClamp(v string) (int, bool) {
	x, err := strconv.ParseFloat(strings.ReplaceAll(v, ",", "."), 64)
	if err != nil {
		return 0, false
	}
	if x < 30 {
		return 0, true
	}
	return 1, true
}

func discLevel(v string, q1, q3 float64) (int, bool) {
	x, err := strconv.ParseFloat(strings.ReplaceAll(v, ",", "."), 64)
	if err != nil {
		return 0, false
	}
	if x < q1 {
		return 0, true
	}
	if x < q3 {
		return 1, true
	}
	return 2, true
}

func discIVL(v string) (int, bool) {
	x, err := strconv.ParseFloat(strings.ReplaceAll(v, ",", "."), 64)
	if err != nil {
		return 0, false
	}
	if x >= 24 {
		return 1, true
	}
	return 0, true
}

func quantile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := p * float64(len(sorted)-1)
	lo := int(idx)
	hi := lo + 1
	if hi >= len(sorted) {
		return sorted[lo]
	}
	frac := idx - float64(lo)
	return sorted[lo]*(1-frac) + sorted[hi]*frac
}

type rawRow struct {
	pge1, ik, clamp, ncf, mcf, ivl string
}

func LoadSamples(path, sheet string) ([]map[string]int, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, err
	}

	var raw []rawRow

	for i, row := range rows {
		if i < 2 {
			continue
		}
		get := func(col string) string {
			c, _ := excelize.ColumnNameToNumber(col)
			if c-1 < len(row) {
				return row[c-1]
			}
			return ""
		}
		raw = append(raw, rawRow{
			pge1:  get("I"),
			ik:    get("Z"),
			clamp: get("AA"),
			ncf:   get("DA"),
			mcf:   get("DE"),
			ivl:   get("BE"),
		})
	}

	var ncfVals, mcfVals []float64
	for _, r := range raw {
		if x, err := strconv.ParseFloat(strings.ReplaceAll(r.ncf, ",", "."), 64); err == nil {
			ncfVals = append(ncfVals, x)
		}
		if x, err := strconv.ParseFloat(strings.ReplaceAll(r.mcf, ",", "."), 64); err == nil {
			mcfVals = append(mcfVals, x)
		}
	}
	sort.Float64s(ncfVals)
	sort.Float64s(mcfVals)

	q1n := quantile(ncfVals, 0.25)
	q3n := quantile(ncfVals, 0.75)
	q1m := quantile(mcfVals, 0.25)
	q3m := quantile(mcfVals, 0.75)

	var samples []map[string]int
	for _, r := range raw {
		p, ok1 := discPGE1(r.pge1)
		c, ok2 := discIK(r.ik)
		cl, ok3 := discClamp(r.clamp)
		n, ok4 := discLevel(r.ncf, q1n, q3n)
		m, ok5 := discLevel(r.mcf, q1m, q3m)
		iv, ok6 := discIVL(r.ivl)
		if !(ok1 && ok2 && ok3 && ok4 && ok5 && ok6) {
			continue
		}
		samples = append(samples, map[string]int{
			"pge1":                  p,
			"ik_duration":           c,
			"clamp_duration":        cl,
			"ncfdna_t4":             n,
			"mcfdna_t4":             m,
			"prolonged_ventilation": iv,
		})
	}

	return samples, nil
}