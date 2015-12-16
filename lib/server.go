package octopussy

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

// ConnectionFactory is a function capable of creating an AMQP connection.
type ConnectionFactory func() (*amqp.Connection, error)

// Server is an AMQP-to-websockets bridge.
type Server struct {
	// Name of the exchange to use.
	Exchange string

	// Optional function that runs on error. Can be used for logging.
	OnError func(error)

	// Optional function called with every incoming connection.
	OnConn func(http.Request)

	// ConnectionFactory is a function capable of creating the underlying
	// AMQP connection.
	ConnectionFactory ConnectionFactory

	// websocket.Upgrader used for setting up the WS connection
	Upgrader *websocket.Upgrader

	// amqp.Connection
	conn *amqp.Connection
}

var DefaultUpgrader = &websocket.Upgrader{}

func DialURL(url string) ConnectionFactory {
	return func() (*amqp.Connection, error) {
		return amqp.Dial(url)
	}
}

// Handler returns an HTTP handler that can be directly connected to the router.
func (s *Server) Handler() func(http.ResponseWriter, *http.Request) {
	if s.Upgrader == nil {
		s.Upgrader = DefaultUpgrader
	}
	return s.handlerFunc
}

func (s *Server) runConnCallback(r *http.Request) {
	if s.OnConn != nil {
		s.OnConn(*r)
	}
}

func (s *Server) handlerFunc(w http.ResponseWriter, r *http.Request) {
	ws, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.catchError(err)
		return
	}
	channel, err := s.getChannel()
	if err != nil {
		log.Fatalf("Cannot establish AMQP channel: %v", err)
	}
	s.runConnCallback(r)
	s.catchError(newHandler(channel, ws, s.Exchange).handle())
}

func (s *Server) getChannel() (*amqp.Channel, error) {
	return s.getChannelWithRetry(false)
}

func (s *Server) getChannelWithRetry(isRetry bool) (*amqp.Channel, error) {
	channel, err := s.conn.Channel()
	if err == nil || isRetry {
		return channel, err
	}
	if err := s.SetupConnection(); err != nil {
		return nil, err
	}
	return s.getChannelWithRetry(true)
}

func (s *Server) SetupConnection() error {
	conn, err := s.ConnectionFactory()
	if err == nil {
		s.conn = conn
	}
	return err
}

// Close closes the underlying AMQP connection.
func (s *Server) Close() error {
	return s.conn.Close()
}

func (s *Server) catchError(err error) {
	if err == nil || s.OnError == nil {
		return
	}
	s.OnError(err)
}
