package octopussy

import (
	"net/http"

	"github.com/streadway/amqp"
	"golang.org/x/net/websocket"
)

// Server is an AMQP-to-websockets bridge.
type Server struct {
	// Name of the exchange to use.
	Exchange string

	// Optional function that runs on error. Can be used for logging.
	OnError func(error)

	// Connection represents the underlying AMQP connection. It can be set
	// directly or constructed from a URL using the `Dial` call.
	Connection *amqp.Connection
}

// Dial sets up the AMQP connection if you don't want to provide it directly.
func (s *Server) Dial(url string) error {
	conn, err := amqp.Dial(url)
	if err == nil {
		s.Connection = conn
	}
	return err
}

// Handler returns an HTTP handler that can be directly connected to the router.
func (s *Server) Handler() http.Handler {
	return websocket.Handler(s.handle)
}

// Close closes the underlying AMQP connection.
func (s *Server) Close() error {
	return s.Connection.Close()
}

func (s *Server) handle(ws *websocket.Conn) {
	channel, err := s.Connection.Channel()
	if err != nil {
		s.catchError(err)
		return
	}
	requestHandler := &handler{
		channel:   channel,
		websocket: ws,
		exchange:  s.Exchange,
	}
	s.catchError(requestHandler.handle())
}

func (s *Server) catchError(err error) {
	if err == nil {
		return
	}
	if s.OnError == nil {
		return
	}
	s.OnError(err)
}
