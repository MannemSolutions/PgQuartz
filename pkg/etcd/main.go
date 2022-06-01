package etcd

import (
	"context"
)

var ctx = context.Background()

func InitContext(c context.Context) {
	ctx = c
}
