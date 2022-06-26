package jobs

import (
	"fmt"
	"sort"
	"strings"
)

const MatrixInstancePrefix = "PGQ_INSTANCE"

// MatrixArgValues is an array for all the values that one MatrixArg could have
type MatrixArgValues []string

// MatrixArgs is a map of Matrix Arguments.
// The keys are the names of the argument and
// - will be prefixed and uppercased when set for shell commands
// - will be changed into numbered args when set for sql commands
type MatrixArgs map[string]MatrixArgValues

// InstanceArguments are key=value pairs which is extrapolated from MatrixArgs
// One InstanceArguments is parsed to a step instance
type InstanceArguments map[string]string

/*
Example:
In a case where
  mas := MatrixArgs{"x": ["1, 2"], "y": ["3", "4"]}
`is := mas.Explode()` would result in:
  is := Instances{[
    map[string]{"x": "1", "y": "3"],
    map[string]{"x": "1", "y": "4"],
    map[string]{"x": "2", "y": "3"],
    map[string]{"x": "2", "y": "4"],
    ]
And the step commands would be run 4 times.
A shell step command would be run with the following instances:
- PGQ_INSTANCE_X=1 PGQ_INSTANCE_Y=3 {SCRIPT}
- PGQ_INSTANCE_X=1 PGQ_INSTANCE_Y=4 {SCRIPT}
- PGQ_INSTANCE_X=2 PGQ_INSTANCE_Y=3 {SCRIPT}
- PGQ_INSTANCE_X=2 PGQ_INSTANCE_Y=4 {SCRIPT}
A query step command (e.a. `select fn_myfunc($x, $y);`) would be run with the following instances:
- `select fn_myfunc($1, $2);` with args: [1, 3]
- `select fn_myfunc($1, $2);` with args: [1, 4]
- `select fn_myfunc($1, $2);` with args: [2, 3]
- `select fn_myfunc($1, $2);` with args: [2, 4]
*/

func (ias InstanceArguments) String() string {
	// loop over elements of slice
	var keyValues []string
	for key, value := range ias {
		if len(key) > 0 {
			keyValues = append(keyValues, fmt.Sprintf("'%s': '%s'",
				strings.Replace(key, "'", "''", -1),
				strings.Replace(value, "'", "''", -1),
			))
		}
	}
	if len(keyValues) == 0 {
		return "None"
	}
	sort.Strings(keyValues)
	return fmt.Sprintf("{ %s }", strings.Join(keyValues, ", "))
}

func (ias InstanceArguments) Clone() (newMia InstanceArguments) {
	newMia = make(InstanceArguments)
	for key, value := range ias {
		newMia[key] = value
	}
	return newMia
}

func (ias InstanceArguments) AsEnv() []string {
	var env []string
	for key, value := range ias {
		env = append(env, fmt.Sprintf("%s_%s=%s", MatrixInstancePrefix, strings.ToUpper(key), value))
	}
	return env
}

// ParseQuery can take a query with named arguments and convert it into a query with numbered arguments.
// Inspired by https://github.com/jackc/pgx/issues/387#issuecomment-798348824
func (ias InstanceArguments) ParseQuery(query string) (parsedQuery string, args []interface{}) {
	var i = 1
	parsedQuery = query
	// Loop the named args and replace with placeholders
	for argName, argValue := range ias {
		parsedQuery = strings.Replace(parsedQuery, fmt.Sprintf("${%s}", argName), argValue, -1)
		if strings.Contains(parsedQuery, ":"+argName) {
			parsedQuery = strings.ReplaceAll(parsedQuery, ":"+argName, fmt.Sprint(`$`, i))
			args = append(args, argValue)
			i++
		}
	}
	// Return
	// - the query with replaced placeholders and
	// - an array of arguments as a []interface{} which can be directly parsed to .Query()
	return parsedQuery, args
}

func (mavs MatrixArgValues) Explode(key string, collected []InstanceArguments) (exploded []InstanceArguments) {
	for _, value := range mavs {
		for _, mia := range collected {
			mia = mia.Clone()
			mia[key] = value
			exploded = append(exploded, mia)
		}
	}
	return exploded
}

func (mas MatrixArgs) Instances() (ias []InstanceArguments) {
	for arg, values := range mas {
		if len(ias) == 0 {
			for _, value := range values {
				ias = append(ias, InstanceArguments{arg: value})
			}
		} else {
			ias = values.Explode(arg, ias)
		}
	}
	if len(ias) == 0 {
		ias = append(ias, InstanceArguments{"": ""})
	}
	return ias
}
