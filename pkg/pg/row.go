package pg

import (
	"fmt"
	"strings"
)

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
	var delimiter = ", "
	if len(params) > 0 {
		delimiter = params[0]
	}
	for _, row := range r.rows {
		var cols []string
		for i, col := range row {
			cols = append(cols, fmt.Sprintf("{%s}={%s}", r.header[i],
				strings.Replace(col, "'", "''", -1)))
		}
		arraysOfStrings = append(arraysOfStrings, strings.Join(cols, delimiter))
	}
	return arraysOfStrings
}
