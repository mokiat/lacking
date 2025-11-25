package game

import (
	"github.com/mokiat/lacking/util/async"
)

type resourceHandle struct {
	resourceLoader ResourceLoader[any]
	promise        async.Promise[any]
	refCount       int
}
