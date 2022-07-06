package pg

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/jackc/pgx/v4"
)

var (
	UnexpctedRole = fmt.Errorf("we are connected to a database with another role then wished for")
)

type Conn struct {
	Type       string `yaml:"type"`
	ConnParams Dsn    `yaml:"conn_params"`
	Role       string `yaml:"role"`
	conn       *pgx.Conn
}

func NewConn(connParams Dsn) (c *Conn) {
	return &Conn{
		ConnParams: connParams,
	}
}

func (c *Conn) DbName() (dbName string) {
	value, ok := c.ConnParams["dbname"]
	if ok {
		return value
	}
	value = os.Getenv("PGDATABASE")
	if value != "" {
		return value
	}
	return c.UserName()
}

func (c *Conn) UserName() (userName string) {
	value, ok := c.ConnParams["user"]
	if ok {
		return value
	}
	value = os.Getenv("PGUSER")
	if value != "" {
		return value
	}
	currentUser, err := user.Current()
	if err != nil {
		log.Panic("cannot determine current user")

	}
	return currentUser.Username
}

// connectStringValue uses proper quoting for connect string values
func connectStringValue(objectName string) (escaped string) {
	return fmt.Sprintf("'%s'", strings.Replace(objectName, "'", "\\'", -1))
}

func (c *Conn) DSN() (dsn string) {
	var pairs []string
	for key, value := range c.ConnParams {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, connectStringValue(value)))
	}
	return strings.Join(pairs[:], " ")
}

func (c *Conn) Connect() (err error) {
	if c.conn != nil {
		if c.conn.IsClosed() {
			c.conn = nil
		} else {
			return nil
		}
	}
	c.conn, err = pgx.Connect(ctx, c.DSN())
	if err != nil {
		c.conn = nil
		return err
	}
	return nil
}

func (c *Conn) CheckExists(query string, args ...interface{}) (exists bool, err error) {
	err = c.Connect()
	if err != nil {
		return false, err
	}
	var answer string
	err = c.conn.QueryRow(ctx, query, args...).Scan(&answer)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err == nil {
		return true, nil
	}
	return false, err
}

func (c *Conn) Exec(query string, args ...interface{}) (err error) {
	err = c.Connect()
	if err != nil {
		return err
	}
	_, err = c.conn.Exec(ctx, query, args...)
	return err
}

func (c *Conn) GetOneField(query string, args ...interface{}) (answer string, err error) {
	err = c.Connect()
	if err != nil {
		return "", err
	}

	err = c.conn.QueryRow(ctx, query, args...).Scan(&answer)
	if err != nil {
		return "", fmt.Errorf("runQueryGetOneField (%s) failed: %v\n", query, err)
	}
	return answer, nil
}

func (c *Conn) GetAll(query string, args ...interface{}) (answer Result, err error) {
	err = c.Connect()
	if err != nil {
		return answer, err
	}
	var cursor pgx.Rows
	if cursor, err = c.conn.Query(ctx, query, args...); err != nil {
		return answer, err
	} else {
		for _, header := range cursor.FieldDescriptions() {
			answer.header = append(answer.header, string(header.Name))
		}
		for cursor.Next() {
			var row []string
			var cols []interface{}
			if cols, err = cursor.Values(); err != nil {
				return answer, err
			} else {
				for _, col := range cols {
					row = append(row, fmt.Sprint(col))
				}
			}
			answer.rows = append(answer.rows, row)
		}
	}
	return answer, nil
}

func (c *Conn) VerifyRole() error {
	if c.Role == "all" {
		return nil
	}
	if _, ok := ValidRoles[c.Role]; !ok {
		return fmt.Errorf("invalid role was specified for conn %s", c.ConnParams.String(true))
	}
	if role, err := c.GetOneField("select case pg_is_in_recovery() when true then 'standby' else 'primary' end"); err != nil {
		return err
	} else if role != c.Role {
		log.Debugf("actual role %s != expected role %s", role, c.Role)
		return UnexpctedRole
	}
	log.Debugf("actual role is as expected %s", c.Role)
	return nil
}
