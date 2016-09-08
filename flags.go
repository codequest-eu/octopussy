package main

import "github.com/codegangsta/cli"

var (
	amqpURL            string
	amqpExchange       string
	amqpCARoot         string
	customDNSResolvers string
	host               string
	originRegexp       string
	port               int
	development        bool
	hostFlag           = cli.StringFlag{
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
	amqpCARootFlag = cli.StringFlag{
		Name:        "amqp-ca-root, c",
		Value:       "",
		Usage:       "AMQP CA root, if applicable",
		EnvVar:      "AMQP_CA_CERT",
		Destination: &amqpCARoot,
	}
	customDNSResolversFlag = cli.StringFlag{
		Name:        "custom-dns-resolvers",
		Value:       "",
		Usage:       "custom DNS resolvers, if applicable",
		EnvVar:      "CUSTOM_DNS_RESOLVERS",
		Destination: &customDNSResolvers,
	}
	developmentFlag = cli.BoolFlag{
		Name:        "development, d",
		Usage:       "run in development mode, without CORS and TLS",
		Destination: &development,
	}
)
