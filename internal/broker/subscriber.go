package broker

import (
	"context"
	"sync"
	"therealbroker/pkg/broker"
)

type Subscriber struct {
	queue      chan *broker.Message
	OutChannel chan broker.Message
	lock       sync.Mutex
	ctx        context.Context
}

func NewSubscriber(ctx context.Context, channel chan broker.Message) *Subscriber {

	return &Subscriber{
		queue:      make(chan *broker.Message, 1000),
		OutChannel: channel,
		ctx:        ctx,
	}
}

func (s *Subscriber) AddMessageToQueue(msg *broker.Message) {

	s.lock.Lock()
	defer s.lock.Unlock()
	s.queue <- msg
	// TODO: don't know what to do here yet
}

func (s *Subscriber) Listen() {

	for {
		select {
		case <-s.ctx.Done():
			return

		case msg := <-s.queue:
			s.OutChannel <- *msg
			// TODO: don't know what to do here yet
		}
	}
}
