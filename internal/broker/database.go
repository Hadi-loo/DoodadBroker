package broker

import (
	"sync"
	"therealbroker/pkg/broker"
	"time"
)

type Database interface {
	AddMessage(subject string, id int, message *broker.Message, createTime time.Time) error
	FetchMessage(subject string, id int) (*broker.Message, time.Time, error)
	RemoveMessage(id int) error
}

type InMemoryBlock struct {
	subject    string
	id         int
	createTime time.Time
	message    *broker.Message
}

type InMemory struct {
	// messages []InMemoryBlock
	messages map[string][]InMemoryBlock
	lock     sync.Mutex
}

func NewDatabase() Database {
	return &InMemory{
		// messages: make([]InMemoryBlock, 0),
		messages: make(map[string][]InMemoryBlock),
	}
}

func (db *InMemory) AddMessage(subject string, id int, message *broker.Message, createTime time.Time) error {

	// var lock = &sync.RWMutex{}
	// lock.Lock()
	// defer lock.Unlock()
	// db.messages = append(db.messages, InMemoryBlock{
	// 	subject:    subject,
	// 	id:         id,
	// 	createTime: createTime,
	// 	message:    message,
	// })

	db.lock.Lock()
	defer db.lock.Unlock()

	_, ok := db.messages[subject]
	if !ok {
		db.messages[subject] = make([]InMemoryBlock, 0)
	}
	db.messages[subject] = append(db.messages[subject], InMemoryBlock{
		subject:    subject,
		id:         id,
		createTime: createTime,
		message:    message,
	})

	return nil
}

func (db *InMemory) FetchMessage(subject string, id int) (*broker.Message, time.Time, error) {

	// can be done concurrently which idk is a good idea or not
	// for _, msg := range db.messages {
	// 	if msg.subject == subject && msg.id == id {
	// 		return msg.message, msg.createTime, nil
	// 	}
	// }
	// return &broker.Message{}, time.Time{}, broker.ErrInvalidID

	db.lock.Lock()
	defer db.lock.Unlock()

	topic, ok := db.messages[subject]
	if !ok {
		return &broker.Message{}, time.Time{}, broker.ErrInvalidID
	}
	for _, msg := range topic {
		if msg.id == id {
			return msg.message, msg.createTime, nil
		}
	}
	return &broker.Message{}, time.Time{}, broker.ErrInvalidID
}

func (db *InMemory) RemoveMessage(id int) error {
	// TODO: implement this. although it's not really necessary yet
	return nil
}
