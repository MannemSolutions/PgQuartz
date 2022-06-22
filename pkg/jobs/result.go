package jobs

import (
	"fmt"
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
	lines = strings.TrimSuffix(lines, "\n")
	for _, line := range strings.Split(lines, "\n") {
		result = append(result, ResultLine(line))
	}
	return result
}

func (r Result) String() string {
	var lines []string
	for _, line := range r {
		lines = append(lines, fmt.Sprintf("'%s'", strings.Replace(string(line), "'", "''", -1)))
	}
	return fmt.Sprintf("[ %s ]", strings.Join(lines, ", "))
}

func (r Result) Contains(part string) bool {
	for _, l := range r {
		if l.Contains(part) {
			if debug() {
				log.Debugf("%s contains %s", r.String(), part)
			}
			log.Debug()
			return true
		}
	}
	if debug() {
		log.Debugf("%s does not contain %s", r.String(), part)
	}
	return false
}

func (r Result) ContainsLine(line string) bool {
	for _, l := range r {
		if string(l) == line {
			if debug() {
				log.Debugf("%s contains line %s", r.String(), line)
			}
			return true
		}
	}
	if debug() {
		log.Debugf("%s does not contain line %s", r.String(), line)
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
				if debug() {
					log.Debugf("%s contains regexp %s", r.String(), exp)
				}
				return true
			}
		}
	}
	if debug() {
		log.Debugf("%s does not contain regexp %s", r.String(), exp)
	}
	return false
}

func (r ResultLine) Contains(part string) bool {
	return strings.Contains(string(r), part)
}
