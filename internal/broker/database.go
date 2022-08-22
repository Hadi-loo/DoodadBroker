package broker

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"therealbroker/pkg/broker"
	"time"

	_ "github.com/lib/pq"
)

type Database interface {
	AddMessage(subject string, id int, message *broker.Message, createTime time.Time) error
	FetchMessage(subject string, id int) (*broker.Message, time.Time, error)
	RemoveMessage(subject string, id int) error
}

type InMemoryBlock struct {
	subject    string
	id         int
	createTime time.Time
	message    *broker.Message
}

type InMemory struct {
	messages map[string][]InMemoryBlock
	lock     sync.Mutex
}

type PostgreSQL struct {
	client  *sql.DB
	queries []string
	ticker  *time.Ticker
	lock    sync.Mutex
}

const (
	psql_host     = "localhost"
	psql_port     = 5432
	psql_user     = "postgres"
	psql_password = "password"
	psql_dbname   = "BrokerDB"
)

func NewDatabase(DBType string) Database {

	switch DBType {
	case "inmemory":
		return newInMemoryDatabase()
	case "postgres":
		return newPostgresDatabase()
	default:
		return nil
	}
}

// ================================================================================================

func newPostgresDatabase() *PostgreSQL {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		psql_host, psql_port, psql_user, psql_password, psql_dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS messages (
		id 					int,
		subject 			varchar(255),
		createTime			timestamp,
		body				varchar,
		expirationSeconds	float8,
		PRIMARY KEY 		(subject, id)
	);`
	_, err = db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}

	// is this part necessary?
	db.SetMaxOpenConns(50)
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(time.Millisecond * 100)
	fmt.Println("Successfully connected to PostgreSQL")
	psql := &PostgreSQL{
		client: db,
		ticker: ticker,
	}
	go psql.WriteMessages()
	return psql
}

func (db *PostgreSQL) WriteMessages() {
	for {
		select {
		case <-db.ticker.C:
			if len(db.queries) > 0 {
				db.lock.Lock()
				query := fmt.Sprintf("INSERT INTO messages (id, subject, createTime, body, expirationSeconds) VALUES %s",
					strings.Join(db.queries, ","))
				db.queries = []string{}
				db.lock.Unlock()
				_, err := db.client.Exec(query)
				if err != nil {
					log.Fatal(err)
				}
			}
		default:
			if len(db.queries) > 100 {
				db.lock.Lock()
				query := fmt.Sprintf("INSERT INTO messages (id, subject, createTime, body, expirationSeconds) VALUES %s",
					strings.Join(db.queries, ","))
				db.queries = []string{}
				db.lock.Unlock()
				_, err := db.client.Exec(query)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

func (db *PostgreSQL) AddMessage(subject string, id int, message *broker.Message, createTime time.Time) error {

	// without batching
	// query := `
	// INSERT INTO messages (id, subject, createTime, body, expirationSeconds)
	// VALUES ($1, $2, $3, $4, $5)`

	// createTimeString := createTime.Format("2006-01-02 15:04:05")
	// _, err := db.client.Exec(query, id, subject, createTimeString, message.Body, message.Expiration.Seconds())
	// if err != nil {
	// 	return err
	// }
	// return nil

	// with batching, this is still slower than the in-memory implementation but much faster than the without batching implementation
	query := fmt.Sprintf(`
	(%d, '%s', '%s', '%s', %f)`,
		id, subject, createTime.Format("2006-01-02 15:04:05"), message.Body, message.Expiration.Seconds())

	db.lock.Lock()
	db.queries = append(db.queries, query)
	db.lock.Unlock()

	return nil
}

func (db *PostgreSQL) FetchMessage(subject string, id int) (*broker.Message, time.Time, error) {

	query := `
	SELECT body, createTime, expirationSeconds FROM messages WHERE subject = $1 AND id = $2`
	row := db.client.QueryRow(query, subject, id)
	if row.Err() != nil {
		return &broker.Message{}, time.Time{}, row.Err()
	}

	var (
		body               string
		createTime         time.Time
		expirationDuration float64
	)

	err := row.Scan(&body, &createTime, &expirationDuration)
	if err != nil {
		return &broker.Message{}, time.Time{}, err
	}

	return &broker.Message{
		Body:       body,
		Expiration: time.Duration(expirationDuration) * time.Second,
	}, createTime, nil
}

func (db *PostgreSQL) RemoveMessage(subject string, id int) error {

	query := `
	DELETE FROM messages WHERE subject = $1 AND id = $2`
	_, err := db.client.Exec(query, subject, id)
	if err != nil {
		return err
	}
	return nil
}

// ================================================================================================

func newInMemoryDatabase() *InMemory {

	return &InMemory{
		messages: make(map[string][]InMemoryBlock),
	}
}

func (db *InMemory) AddMessage(subject string, id int, message *broker.Message, createTime time.Time) error {

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

func (db *InMemory) RemoveMessage(subject string, id int) error {
	// TODO: implement this. although it's not really necessary yet
	return nil
}
