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

func logError(err error) {
	log.Println(err)
}

func logRequest(r http.Request) {
	log.Printf("Connecting %s", r.RemoteAddr)
}

func createServer() (*octopussy.Server, error) {
	server := &octopussy.Server{
		Exchange: amqpExchange,
		OnError:  logError,
		OnConn:   logRequest,
		ConnectionFactory: octopussy.DialURLWithBackoff(
			amqpURL,
			amqpCARoot,
			customDNSResolvers,
		),
	}
	return server, server.SetupConnection()
}

func urlMatches(u *url.URL, pattern string) bool {
	pattern = strings.TrimSpace(pattern)
	hostname := strings.TrimSpace(strings.SplitN(u.Host, ":", 2)[0])
	match, err := regexp.MatchString(pattern, hostname)
	return (err == nil) && match
}

func checkOriginRegexp(pattern string) func(*http.Request) bool {
	if len(pattern) == 0 {
		log.Println("CORS is currently disabled")
	}
	return func(r *http.Request) bool {
		hostURL, err := url.Parse(r.Header.Get("Origin"))
		return err == nil && urlMatches(hostURL, pattern)
	}
}

func setupServer() (*octopussy.Server, error) {
	server, err := createServer()
	if err != nil {
		return nil, err
	}
	server.Upgrader = &websocket.Upgrader{
		CheckOrigin: checkOriginRegexp(originRegexp),
	}
	return server, nil
}

func serveHTTP() {
	parsedAddr := fmt.Sprintf("%s:%d", host, port)
	log.Println("Listening for WebSocket connections on " + parsedAddr)
	if development {
		log.Fatal(http.ListenAndServe(parsedAddr, nil))
		return
	}
	// Requires cert.pem and cert.key to be present. See cert_setup.sh
	log.Fatal(http.ListenAndServeTLS(parsedAddr, "cert.pem", "cert.key", nil))
}

func runServer(c *cli.Context) {
	server, err := setupServer()
	if err != nil {
		log.Println(err)
		return
	}
	defer server.Close()
	http.HandleFunc("/", server.Handler())
	serveHTTP()
}

func main() {
	app := cli.NewApp()
	app.Name = "AMQP-WebSocket push service"
	app.Usage = "provides real-time notifications from amqp topics"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		developmentFlag,
		hostFlag,
		portFlag,
		amqpURLFlag,
		amqpExchangeFlag,
		amqpURLFlag,
		customDNSResolversFlag,
		originRegexpFlag,
	}
	app.Action = runServer
	app.Run(os.Args)
}
