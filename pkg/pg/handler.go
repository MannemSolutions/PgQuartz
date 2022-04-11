package pg

type Handler struct {
	conn *Conn
}

func NewPgHandler(connParams Dsn) (ph *Handler) {
	ph = &Handler{
		conn: NewConn(connParams),
	}
	return ph
}
