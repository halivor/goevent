package check

import (
	mw "github.com/halivor/goevent/middleware"
)

type check struct {
	*mw.MwTmpl
}

func init() {
	mw.Register(mw.T_CHECK, New)
}

func New() mw.Mwer {
	return &check{
		MwTmpl: mw.NewTmpl(),
	}
}
