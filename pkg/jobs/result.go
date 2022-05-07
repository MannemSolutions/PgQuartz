package jobs

import (
	"regexp"
	"strings"
)

type ResultLine string

type Result []ResultLine

func NewResult(lines []string) (result Result) {
	for _, line := range lines {
		result = append(result, ResultLine(line))
	}
	return result
}

func NewResultFromString(lines string) (result Result) {
	return NewResult(strings.Split(lines, "\n"))
}

func (r Result) Contains(part string) bool {
	for _, l := range r {
		if l.Contains(part) {
			return true
		}
	}
	return false
}

func (r Result) ContainsLine(line string) bool {
	for _, l := range r {
		if string(l) == line {
			return true
		}
	}
	return false
}

func (r Result) RegExpContains(exp string) bool {
	if re, err := regexp.Compile(exp); err != nil {
		log.Errorf("Could not compile Regular Expression %s: %e", exp, err)
		return false
	} else {
		for _, l := range r {
			if re.Match([]byte(l)) {
				return true
			}
		}
	}
	return false
}

func (r ResultLine) Contains(part string) bool {
	return strings.Contains(string(r), part)
}
