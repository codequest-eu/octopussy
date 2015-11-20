package main

import (
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/codegangsta/cli"
	octopussy "github.com/codequest-eu/octopussy/lib"
	"github.com/gorilla/websocket"
)

func startServer(amqpUrl string, amqpExchange string) (*octopussy.Server, error) {
	server := &octopussy.Server{
		Exchange: amqpExchange,
		OnError: func(e error) {
			log.Println(e.Error())
		},
		OnConn: func(r http.Request) {
			log.Println("Connecting " + r.RemoteAddr)
		},
	}
	if err := server.Dial(amqpUrl); err != nil {
		return nil, err
	}
	return server, nil
}

func checkOriginRegexp(pattern string) func(*http.Request) bool {
	if len(pattern) == 0 {
		log.Println("CORS is currently disabled")
	}
	return func(r *http.Request) bool {
		origin := (r.Header["Origin"])
		if len(origin) == 0 {
			return true
		}
		match, err := regexp.Match(pattern, []byte(origin[0]))
		if err != nil {
			return false
		}
		return match
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "AMQP-WebSocket push service"
	app.Usage = "provides real-time notifications from amqp topics"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{developmentFlag, hostFlag, portFlag, amqpUrlFlag, amqpExchangeFlag, originRegexpFlag}
	app.Action = func(c *cli.Context) {
		server, err := startServer(amqpUrl, amqpExchange)
		if err != nil {
			log.Println("AMQP Unreachable")
			return
		}
		server.Upgrader = &websocket.Upgrader{
			CheckOrigin: checkOriginRegexp(originRegexp),
		}
		http.HandleFunc("/", server.Handler())
		parsedAddr := host + ":" + strconv.Itoa(port)
		log.Println("Listening for WebSocket connections on " + parsedAddr)
		if development {
			log.Fatal(http.ListenAndServe(parsedAddr, nil))
		} else {
			// Requires cert.pem and cert.key to be present. See cert_setup.sh
			log.Fatal(http.ListenAndServeTLS(parsedAddr, "cert.pem", "cert.key", nil))
		}
	}
	app.Run(os.Args)
}
