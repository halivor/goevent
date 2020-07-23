package service

type EventType int32

type Value interface {
	Event() EventType
	Data() []byte
	String() string
}

const (
	EVENT_ADD EventType = 1 << iota
	EVENT_MOD
	EVENT_DEL
)

var etm = map[EventType]string{
	EVENT_ADD: "event add",
	EVENT_MOD: "event mod",
	EVENT_DEL: "event del",
}

func (et EventType) String() string {
	return etm[et]
}
