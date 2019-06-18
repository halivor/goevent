package middleware

import ()

type MwTmpl struct {
	Id  QId
	QId map[string]QId
	Cs  map[QId]map[Consumer]struct{}
}

func NewTmpl() *MwTmpl {
	return &MwTmpl{
		Id:  10000,
		QId: make(map[string]QId),
		Cs:  make(map[QId]map[Consumer]struct{}),
	}
}

func (t *MwTmpl) Bind(q string, a Action, c interface{}) QId {
	id, ok := t.QId[q]
	if !ok {
		t.Id++
		id = t.Id
		t.QId[q] = id
		t.Cs[id] = make(map[Consumer]struct{}, 8)
	}
	if cc, ok := c.(Consumer); ok && a == A_CONSUME {
		t.Cs[id][cc] = struct{}{}
	}
	return id
}

func (t *MwTmpl) Produce(id QId, message interface{}) interface{} {
	if cs, ok := t.Cs[id]; ok {
		for c, _ := range cs {
			c.Consume(message)
		}
	}
	return nil
}

func (t *MwTmpl) GetQId(q string) QId {
	if id, ok := t.QId[q]; ok {
		return id
	}
	return -1
}

func (t *MwTmpl) Release(q string, c interface{}) {
	if id, ok := t.QId[q]; ok {
		if cs, ok := t.Cs[id]; ok {
			if cc, ok := c.(Consumer); ok {
				delete(cs, cc)
			}
			if len(cs) == 0 {
				delete(t.Cs, id)
				// 不能删除, 存在producer, 最后一个consumer被删除后,
				// 后续加入的consumer会导致Qid重新分配，之前的producer生产出的数据会无人消费
				// delete(t.QId, q)
			}
		}
	}
}
