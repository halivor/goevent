package exists

import (
	mw "github.com/halivor/goevent/middleware"
)

type exists struct {
	*mw.MwTmpl
}

func init() {
	mw.Register(mw.T_TRANSFER, New)
}

func New() mw.Mwer {
	return &exists{
		MwTmpl: mw.NewTmpl(),
	}
}

func (t *exists) Produce(id mw.QId, message interface{}) interface{} {
	if cs, ok := t.Cs[id]; ok {
		for _, c := range cs {
			if i := c.Consume(message); i != nil {
				return i
			}
		}
	}
	return nil
}
