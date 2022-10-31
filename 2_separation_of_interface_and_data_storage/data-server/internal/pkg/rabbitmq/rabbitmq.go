package rabbitmq

import (
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	channel  *amqp.Channel
	conn     *amqp.Connection
	Name     string
	exchange string
}

// 普通New和反复创建直到成功的New
// 普通的New
func New(s string) (*RabbitMQ, error) {
	conn, e := amqp.Dial(s)
	if e != nil {
		return nil, e
	}

	ch, e := conn.Channel()
	if e != nil {
		return nil, e
	}

	q, e := ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if e != nil {
		return nil, e
	}

	mq := new(RabbitMQ)
	mq.channel = ch
	mq.conn = conn
	mq.Name = q.Name
	return mq, nil
}

// 反复创建直到成功的New
func RetryNew(s string) *RabbitMQ {
	for {
		mq, e := New(s)
		if e == nil {
			return mq
		}
		time.Sleep(5 * time.Second)
	}
}

func (q *RabbitMQ) GetNotifyChan() (chan *amqp.Error, chan string) {
	notifyCloseChan := q.channel.NotifyClose(make(chan *amqp.Error, 1))
	notifyCancelChan := q.channel.NotifyCancel(make(chan string, 1))
	return notifyCloseChan, notifyCancelChan
}

func (q *RabbitMQ) Bind(exchange string) {
	e := q.channel.ExchangeDeclare(
		exchange, // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if e != nil {
		log.Println(e)
		return
	}
	e = q.channel.QueueBind(
		q.Name,   // queue name
		"",       // routing key
		exchange, // exchange
		false,
		nil)
	if e != nil {
		log.Println(e)
	}
	q.exchange = exchange
}

func (q *RabbitMQ) Send(queue string, body interface{}) {
	str, e := json.Marshal(body)
	if e != nil {
		log.Println(e)
		return
	}
	e = q.channel.Publish("",
		queue,
		false,
		false,
		amqp.Publishing{
			ReplyTo: q.Name,
			Body:    []byte(str),
		})
	if e != nil {
		log.Println(e)
	}
}

func (q *RabbitMQ) Publish(exchange string, body interface{}) {
	e := q.channel.ExchangeDeclare(
		exchange, // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if e != nil {
		log.Println(e)
		return
	}
	str, e := json.Marshal(body)
	if e != nil {
		log.Println(e)
		return
	}
	e = q.channel.Publish(exchange,
		"",
		false,
		false,
		amqp.Publishing{
			ReplyTo: q.Name,
			Body:    []byte(str),
		})
	if e != nil {
		log.Println(e)
	}
}

func (q *RabbitMQ) Consume() <-chan amqp.Delivery {
	c, e := q.channel.Consume(q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if e != nil {
		log.Println(e)
		return nil
	}
	return c
}

func (q *RabbitMQ) Close() {
	q.channel.Close()
	q.conn.Close()
}
