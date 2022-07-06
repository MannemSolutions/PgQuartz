package jobs

import (
	"fmt"
	"strings"

	"github.com/mannemsolutions/PgQuartz/pkg/pg"
)

type Connections map[string]pg.Conn

func (cs Connections) Execute(connName string, role string, query string, batchMode bool, args InstanceArguments) (result Result, err error) {
	var response pg.Result
	var c pg.Conn
	var exists bool
	if c, exists = cs[connName]; !exists {
		return nil, fmt.Errorf("connection %s does not exist", connName)
	} else if err = c.VerifyRole(role); err != nil {
		log.Infof("skipping command %s (%s): %s", query, args.String(), err.Error())
		return result, err
	}

	if batchMode {
		for _, qry := range strings.Split(query, ";") {
			numberedArgsQuery, numberedArgs := args.ParseQuery(qry)
			if response, err = c.GetAll(numberedArgsQuery, numberedArgs...); err != nil {
				return nil, err
			} else {
				result = result.Append(NewResult(response.AsStringArray()))
			}
		}
		return result, nil
	} else {
		numberedArgsQuery, numberedArgs := args.ParseQuery(query)
		if response, err = c.GetAll(numberedArgsQuery, numberedArgs...); err != nil {
			return nil, err
		} else {
			return NewResult(response.AsStringArray()), nil
		}
	}
}
