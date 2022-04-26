package jobs

import (
	"fmt"

	"github.com/mannemsolutions/PgQuartz/pkg/pg"
)

type Connections map[string]pg.Conn

func (cs Connections) Execute(connName string, sql string, expected string) (err error) {
	var response string
	if c, exists := cs[connName]; !exists {
		return fmt.Errorf("connection %s does not exist", connName)
	} else if expected == "" {
		return c.Exec(sql)
	} else if response, err = c.GetOneField(sql); err != nil {
		return err
	} else if response != expected {
		return fmt.Errorf("query %s did no respond as expected (expected: %s, response: %s)", sql, expected,
			response)
	}
	return nil
}
