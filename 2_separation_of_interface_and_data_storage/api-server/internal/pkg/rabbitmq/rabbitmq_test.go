package rabbitmq

import (
	"encoding/json"
	"fmt"
	"testing"
)

const host = "amqp://root:123456@localhost:5672"

func TestPublish(t *testing.T) {
	// 需要创建两个exchange，分别是test和test2
	q, err := New(host)
	defer q.Close()
	q.Bind("test")

	q2, err := New(host)
	defer q2.Close()
	q2.Bind("test")

	q3, err := New(host)
	defer q3.Close()

	expect := "test"
	q3.Publish("test2", "any")
	q3.Publish("test", expect)

	fmt.Println("waiting for message")

	c := q.Consume()
	msg := <-c
	var actual interface{}
	err = json.Unmarshal(msg.Body, &actual)
	fmt.Println(msg)
	if err != nil {
		t.Error(err)
	}
	if actual != expect {
		t.Errorf("expected %s, actual %s", expect, actual)
	}
	if msg.ReplyTo != q3.Name {
		t.Error(msg)
	}

	c2 := q2.Consume()
	msg = <-c2
	err = json.Unmarshal(msg.Body, &actual)
	if err != nil {
		t.Error(err)
	}
	if actual != expect {
		t.Errorf("expected %s, actual %s", expect, actual)
	}
	if msg.ReplyTo != q3.Name {
		t.Error(msg)
	}
	q2.Send(msg.ReplyTo, "test3")
	c3 := q3.Consume()
	msg = <-c3
	if string(msg.Body) != `"test3"` {
		t.Error(string(msg.Body))
	}
}

func TestSend(t *testing.T) {
	q, err := New(host)
	defer q.Close()

	q2, err := New(host)
	defer q2.Close()

	expect := "test"
	expect2 := "test2"
	q2.Send(q.Name, expect)
	q2.Send(q2.Name, expect2)

	c := q.Consume()
	msg := <-c
	var actual interface{}
	err = json.Unmarshal(msg.Body, &actual)
	if err != nil {
		t.Error(err)
	}
	if actual != expect {
		t.Errorf("expected %s, actual %s", expect, actual)
	}

	c2 := q2.Consume()
	msg = <-c2
	err = json.Unmarshal(msg.Body, &actual)
	if err != nil {
		t.Error(err)
	}
	if actual != expect2 {
		t.Errorf("expected %s, actual %s", expect2, actual)
	}
}
