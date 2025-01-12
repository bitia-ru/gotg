package gotd

import gotdTg "github.com/gotd/td/tg"

type updateContext struct {
	entities gotdTg.Entities
	update   any
}
