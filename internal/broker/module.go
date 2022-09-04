package broker

import (
	"context"
	"sync"
	"therealbroker/pkg/broker"
	"time"
)

type Module struct {
	topics   map[string]*Topic
	closed   bool
	lock     sync.Mutex
	database Database
}

func NewModule() broker.Broker {

	return &Module{
		topics:   make(map[string]*Topic),
		closed:   false,
		database: NewDatabase("cassandra"),
	}
}

func (m *Module) getTopic(subject string) (*Topic, bool, error) {

	m.lock.Lock()
	defer m.lock.Unlock()
	if m.closed {
		return nil, true, broker.ErrUnavailable
	}

	topic, ok := m.topics[subject]
	if !ok {
		topic = NewTopic(subject, m.database) // pointer to database maybe?!
		m.topics[subject] = topic
		return topic, false, nil
	}
	return topic, true, nil
}

func (m *Module) Close() error {

	if m.closed {
		return broker.ErrUnavailable
	}

	m.closed = true
	return nil
}

func (m *Module) Publish(ctx context.Context, subject string, msg broker.Message) (int, error) {

	if m.closed {
		return -1, broker.ErrUnavailable
	}

	topic, _, err := m.getTopic(subject)
	if err != nil {
		return -1, err
	}

	createTime := time.Now()
	messageID := topic.Publish(msg, createTime)

	return messageID, nil
}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan broker.Message, error) {

	if m.closed {
		return nil, broker.ErrUnavailable
	}

	topic, _, err := m.getTopic(subject)
	if err != nil {
		return nil, err
	}

	channel := make(chan broker.Message, 100000) // TODO: make this a buffered channel
	newSubsriber := NewSubscriber(ctx, channel)
	topic.AddSubscriber(newSubsriber)
	return channel, nil
}

func (m *Module) Fetch(ctx context.Context, subject string, id int) (broker.Message, error) {

	if m.closed {
		return broker.Message{}, broker.ErrUnavailable
	}

	topic, exist, err := m.getTopic(subject)
	if err != nil {
		return broker.Message{}, err
	}

	if !exist {
		return broker.Message{}, broker.ErrInvalidID
	}

	message, err := topic.Fetch(id)
	if err != nil {
		return broker.Message{}, err
	}

	return *message, nil
}
