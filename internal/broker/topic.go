package broker

import (
	"fmt"
	"sync"
	"therealbroker/pkg/broker"
	"time"

	"github.com/go-redis/redis"
)

type IDGenerator struct {
	// lastID int
	// lock   sync.Mutex
	client *redis.Client
}

func NewIDGenerator() *IDGenerator {

	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	return &IDGenerator{client: client}
}

func (idGen *IDGenerator) GetNewID(subject string) int {
	// idGen.lock.Lock()
	// defer idGen.lock.Unlock()
	// idGen.lastID++
	// return idGen.lastID
	res, _ := idGen.client.Incr(subject).Result()

	fmt.Println(res)

	return int(res)
}

type Topic struct {
	subject      string
	subscribers  []*Subscriber
	id_generator *IDGenerator
	database     Database
	lock         sync.Mutex
}

func NewTopic(subject string, database Database) *Topic {

	return &Topic{
		subject:      subject,
		subscribers:  make([]*Subscriber, 0),
		id_generator: NewIDGenerator(),
		database:     database,
		// TODO
	}
}

func (topic *Topic) AddMessage(msg *broker.Message, createTime time.Time) int {
	messageID := topic.id_generator.GetNewID(topic.subject)
	err := topic.database.AddMessage(topic.subject, messageID, msg, createTime)
	if err != nil {
		fmt.Println(err)
		return -1 // probably should use log or something
	}

	// TODO: implement expiration of messages
	return messageID
}

func (topic *Topic) Publish(msg broker.Message, createTime time.Time) int {
	messageID := topic.AddMessage(&msg, createTime)
	for _, subscriber := range topic.subscribers {
		subscriber.AddMessageToQueue(&msg)
	}
	return messageID
}

func (topic *Topic) AddSubscriber(newSubscriber *Subscriber) {
	topic.lock.Lock()
	defer topic.lock.Unlock()
	topic.subscribers = append(topic.subscribers, newSubscriber)
}

func IsExpired(message *broker.Message, createTime time.Time) bool {
	return (time.Since(createTime) > message.Expiration)
}

func (topic *Topic) Fetch(messageID int) (*broker.Message, error) {

	message, createTime, err := topic.database.FetchMessage(topic.subject, messageID)
	if err != nil {
		return &broker.Message{}, broker.ErrInvalidID
	}

	if IsExpired(message, createTime) {
		return &broker.Message{}, broker.ErrExpiredID
	}

	return message, nil
}
