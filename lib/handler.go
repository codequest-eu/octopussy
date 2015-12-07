package octopussy

import (
	"encoding/json"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

const (
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

type handler struct {
	channel   *amqp.Channel
	messages  <-chan amqp.Delivery
	websocket *websocket.Conn
	ticker    *time.Ticker
	exchange  string
	queue     *amqp.Queue
	topics    []string
}

type step func() error

type stepper struct {
	err error
}

func (s *stepper) perform(steps ...step) {
	for _, st := range steps {
		if s.err == nil {
			s.err = st()
		}
	}
}

func newHandler(ch *amqp.Channel, ws *websocket.Conn, ex string) *handler {
	return &handler{
		channel:   ch,
		websocket: ws,
		exchange:  ex,
	}
}

func (h *handler) handle() error {
	if err := h.setUp(); err != nil {
		return err
	}
	return h.consume()
}

func (h *handler) setUp() error {
	s := &stepper{}
	s.perform(h.getTopics, h.declareExchange, h.declareQueue, h.subscribeToTopics)
	if s.err != nil {
		return s.err
	}
	h.setUpTicker()
	return h.createChannel()
}

func (h *handler) getTopics() error {
	_, r, err := h.websocket.NextReader()
	if err != nil {
		return err
	}
	if err := json.NewDecoder(r).Decode(&h.topics); err != nil {
		return err
	}
	return nil
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

func (h *handler) createChannel() error {
	msgChan, err := h.channel.Consume(
		h.queue.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // arguments
	)
	if err == nil {
		h.messages = msgChan
	}
	return err
}

func pongHandler(ws *websocket.Conn) func(_ string) error {
	return func(_ string) error {
		ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	}
}

func (h *handler) setUpTicker() {
	h.ticker = time.NewTicker(pingPeriod)
	h.websocket.SetReadDeadline(time.Now().Add(pongWait))
	h.websocket.SetPongHandler(pongHandler(h.websocket))
}

func (h *handler) consume() error {
	defer h.Close()
	for {
		if err := h.consumeOne(); err != nil {
			return handlePipeError(err)
		}
	}
	return nil
}

func (h *handler) consumeOne() error {
	select {
	case msg := <-h.messages:
		return h.websocket.WriteMessage(websocket.TextMessage, msg.Body)
	case <-h.ticker.C:
		break
	}
	ttl := time.Now().Add(5 * time.Second)
	return h.websocket.WriteControl(websocket.PingMessage, []byte{}, ttl)
}

func handlePipeError(err error) error {
	if err == syscall.EPIPE {
		return nil
	}
	return err
}

func (h *handler) Close() {
	h.channel.Close()
	h.websocket.Close()
	h.ticker.Stop()
}
