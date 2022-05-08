package pg

import "strings"

type Row []string
type Result struct {
	header Row
	rows   []Row
}

func (r Result) AsMapArray() (arraysOfMaps []map[string]string) {
	for _, row := range r.rows {
		m := make(map[string]string)
		for i, c := range row {
			m[r.header[i]] = c
		}
		arraysOfMaps = append(arraysOfMaps, m)
	}
	return arraysOfMaps
}

func (r Result) AsStringArray(params ...string) (arraysOfStrings []string) {
	var delimiter = "\t"
	if len(params) > 0 {
		delimiter = params[0]
	}
	for _, row := range r.rows {
		arraysOfStrings = append(arraysOfStrings, strings.Join(row, delimiter))
	}
	return arraysOfStrings
}
