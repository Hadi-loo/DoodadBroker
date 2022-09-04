package broker

import (
	"context"
	"sync"
	"therealbroker/pkg/broker"
)

type Subscriber struct {
	OutChannel chan broker.Message
	lock       sync.Mutex
	ctx        context.Context
}

func NewSubscriber(ctx context.Context, channel chan broker.Message) *Subscriber {

	return &Subscriber{
		OutChannel: channel,
		ctx:        ctx,
	}
}

