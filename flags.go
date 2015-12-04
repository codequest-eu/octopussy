package main

import (
	"time"

	"github.com/codegangsta/cli"
)

var (
	amqpURL, amqpExchange, host, originRegexp string
	port                                      int
	development                               bool
	killAfter                                 time.Duration
	hostFlag                                  = cli.StringFlag{
		Name:        "bind, b",
		Value:       "",
		Usage:       "host to bind the ws server",
		EnvVar:      "WS_BIND_HOST",
		Destination: &host,
	}
	portFlag = cli.IntFlag{
		Name:        "port, p",
		Value:       9090,
		Usage:       "port to listen on",
		EnvVar:      "WS_BIND_PORT",
		Destination: &port,
	}
	originRegexpFlag = cli.StringFlag{
		Name:        "allow-origin-regexp, r",
		Value:       "",
		Usage:       "a regexp string used to validate request origin header",
		EnvVar:      "WS_ALLOW_ORIGIN_REGEXP",
		Destination: &originRegexp,
	}
	amqpURLFlag = cli.StringFlag{
		Name:        "amqp-url, u",
		Usage:       "full amqp URL",
		EnvVar:      "AMQP_URL",
		Destination: &amqpURL,
	}
	amqpExchangeFlag = cli.StringFlag{
		Name:        "amqp-exchange, x",
		Value:       "",
		Usage:       "topic exchange name",
		EnvVar:      "AMQP_EXCHANGE",
		Destination: &amqpExchange,
	}
	developmentFlag = cli.BoolFlag{
		Name:        "development, d",
		Usage:       "run in development mode, without CORS and TLS",
		Destination: &development,
	}
	killAfterFlag = cli.DurationFlag{
		Name:        "kill-after, k",
		Value:       15 * time.Minute,
		Usage:       "time duration to wait for new connections",
		EnvVar:      "KILL_AFTER",
		Destination: &killAfter,
	}
)
