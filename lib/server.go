package octopussy

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

// Server is an AMQP-to-websockets bridge.
type Server struct {
	// Name of the exchange to use.
	Exchange string

	// Optional function that runs on error. Can be used for logging.
	OnError func(error)

	// Optional function called with every incoming connection.
	OnConn func(http.Request)

	// Connection represents the underlying AMQP connection. It can be set
	// directly or constructed from a URL using the `Dial` call.
	Connection *amqp.Connection

	// websocket.Upgrader used for setting up the WS connection
	Upgrader *websocket.Upgrader
}

var DefaultUpgrader = &websocket.Upgrader{}

// Dial sets up the AMQP connection if you don't want to provide it directly.
func (s *Server) Dial(url string) error {
	conn, err := amqp.Dial(url)
	if err == nil {
		s.Connection = conn
		var errChan chan *amqp.Error
		conn.NotifyClose(errChan)
		go func() {
			log.Fatal(<-errChan)
		}()
	}
	return err
}

// Handler returns an HTTP handler that can be directly connected to the router.
func (s *Server) Handler() func(http.ResponseWriter, *http.Request) {
	if s.Upgrader == nil {
		s.Upgrader = DefaultUpgrader
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := s.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			s.catchError(err)
			return
		}
		channel, err := s.Connection.Channel()
		if err != nil {
			s.catchError(err)
			return
		}
		if s.OnConn != nil {
			s.OnConn(*r)
		}
		s.catchError(newHandler(channel, ws, s.Exchange).handle())
	}
}

// Close closes the underlying AMQP connection.
func (s *Server) Close() error {
	return s.Connection.Close()
}

func (s *Server) catchError(err error) {
	if err == nil || s.OnError == nil {
		return
	}
	s.OnError(err)
}
