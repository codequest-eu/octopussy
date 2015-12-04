package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/codegangsta/cli"
	octopussy "github.com/codequest-eu/octopussy/lib"
	"github.com/gorilla/websocket"
)

func startServer(amqpURL string, amqpExchange string) (*octopussy.Server, error) {
	server := &octopussy.Server{
		Exchange: amqpExchange,
		OnError: func(e error) {
			log.Println(e.Error())
		},
		OnConn: func(r http.Request) {
			log.Println("Connecting " + r.RemoteAddr)
		},
		ConnectionFactory: octopussy.DialURL(amqpURL),
	}
	if err := server.SetupConnection(); err != nil {
		return nil, err
	}
	return server, nil
}

func checkOriginRegexp(pattern string) func(*http.Request) bool {
	if len(pattern) == 0 {
		log.Println("CORS is currently disabled")
	}
	return func(r *http.Request) bool {
		origin := r.Header["Origin"]
		if len(origin) == 0 {
			return false
		}
		hostURL, err := url.Parse(origin[0])
		if err != nil {
			return false
		}
		pattern := strings.TrimSpace(pattern)
		hostname := strings.TrimSpace(strings.SplitN(hostURL.Host, ":", 2)[0])
		match, err := regexp.MatchString(pattern, hostname)
		return (err == nil) && match
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "AMQP-WebSocket push service"
	app.Usage = "provides real-time notifications from amqp topics"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{developmentFlag, hostFlag, portFlag, amqpURLFlag, amqpExchangeFlag, originRegexpFlag}
	app.Action = func(c *cli.Context) {
		server, err := startServer(amqpURL, amqpExchange)
		if err != nil {
			log.Println("AMQP Unreachable")
			return
		}
		defer server.Close()
		server.Upgrader = &websocket.Upgrader{
			CheckOrigin: checkOriginRegexp(originRegexp),
		}
		http.HandleFunc("/", server.Handler())
		parsedAddr := fmt.Sprintf("%s:%d", host, port)
		log.Println("Listening for WebSocket connections on " + parsedAddr)
		if development {
			log.Fatal(http.ListenAndServe(parsedAddr, nil))
			return
		}
		// Requires cert.pem and cert.key to be present. See cert_setup.sh
		log.Fatal(http.ListenAndServeTLS(parsedAddr, "cert.pem", "cert.key", nil))
	}
	app.Run(os.Args)
}
