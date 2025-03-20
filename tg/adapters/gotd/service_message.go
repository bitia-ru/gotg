package gotd

import (
	"github.com/bitia-ru/gotg/tg"
	"time"
)

type ServiceMessageData struct {
	id        int64
	action    tg.ServiceMessageAction
	timestamp time.Time

	where tg.Peer
}

type ServiceMessage struct {
	ServiceMessageData
}

func (m *ServiceMessage) ID() int64 {
	return m.id
}

func (m *ServiceMessage) Where() tg.Peer {
	return m.where
}

func (m *ServiceMessage) CreatedAt() time.Time {
	return m.timestamp
}

func (m *ServiceMessage) Action() tg.ServiceMessageAction {
	return m.action
}

type ServiceMessageWithSubject struct {
	ServiceMessage

	subject tg.Peer
}

func (m *ServiceMessageWithSubject) Subject() tg.Peer {
	return m.subject
}
