package octopussy

import (
	"encoding/json"
	"io"

	"github.com/streadway/amqp"
)

type handler struct {
	channel   *amqp.Channel
	websocket io.ReadWriter
	exchange  string
	queue     *amqp.Queue
	topics    []string
}

func (h *handler) handle() error {
	if err := h.setUp(); err != nil {
		return err
	}
	return h.consume()
}

func (h *handler) setUp() error {
	if err := h.getTopics(); err != nil {
		return err
	}
	if err := h.declareExchange(); err != nil {
		return err
	}
	if err := h.declareQueue(); err != nil {
		return err
	}
	return h.subscribeToTopics()
}

func (h *handler) getTopics() error {
	return json.NewDecoder(h.websocket).Decode(&h.topics)
}

func (h *handler) declareExchange() error {
	return h.channel.ExchangeDeclare(
		h.exchange, // name
		"topic",    // type
		true,       // durable
		false,      // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
}

func (h *handler) declareQueue() error {
	queue, err := h.channel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err == nil {
		h.queue = &queue
	}
	return err
}

func (h *handler) subscribeToTopics() error {
	for _, topic := range h.topics {
		if err := h.subscribeToTopic(topic); err != nil {
			return err
		}
	}
	return nil
}

func (h *handler) subscribeToTopic(topic string) error {
	return h.channel.QueueBind(
		h.queue.Name, // queue name
		topic,        // routing key
		h.exchange,   // exchange
		false,        // no-wait
		nil,          // arguments
	)
}

func (h *handler) consume() error {
	messages, err := h.channel.Consume(
		h.queue.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return err
	}
	for message := range messages {
		if _, err := h.websocket.Write(message.Body); err != nil {
			return err
		}
	}
	return nil
}
