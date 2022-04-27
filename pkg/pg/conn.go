package pg

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/jackc/pgx/v4"
)

type Conn struct {
	Type       string `yaml:"type"`
	ConnParams Dsn    `yaml:"conn_params"`
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
	c.conn, err = pgx.Connect(context.Background(), c.DSN())
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
	err = c.conn.QueryRow(context.Background(), query, args...).Scan(&answer)
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
	_, err = c.conn.Exec(context.Background(), query, args...)
	return err
}

func (c *Conn) GetOneField(query string, args ...interface{}) (answer string, err error) {
	err = c.Connect()
	if err != nil {
		return "", err
	}

	err = c.conn.QueryRow(context.Background(), query, args...).Scan(&answer)
	if err != nil {
		return "", fmt.Errorf("runQueryGetOneField (%s) failed: %v\n", query, err)
	}
	return answer, nil
}
