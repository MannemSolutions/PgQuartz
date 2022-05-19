package jobs

import (
	"fmt"

	"github.com/mannemsolutions/PgQuartz/pkg/pg"
)

type Connections map[string]pg.Conn

func (cs Connections) Execute(connName string, query string, args InstanceArguments) (result Result, err error) {
	var response pg.Result
	numberedArgsQuery, numberedArgs := args.ParseQuery(query)
	if c, exists := cs[connName]; !exists {
		return nil, fmt.Errorf("connection %s does not exist", connName)
	} else if response, err = c.GetAll(numberedArgsQuery, numberedArgs...); err != nil {
		return nil, err
	} else {
		return NewResult(response.AsStringArray()), nil
	}
}
