package octopussy

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"time"

	"github.com/codequest-eu/dnsdialer"
	"github.com/streadway/amqp"
)

func DialURL(url, rootCACert, dnsResolvers string) ConnectionFactory {
	return func() (*amqp.Connection, error) {
		config, err := buildAMQPConfig(rootCACert, dnsResolvers)
		if err != nil {
			return nil, err
		}
		return amqp.DialConfig(url, config)
	}
}

func DialURLWithBackoff(url, rootCACert, dnsResolvers string) ConnectionFactory {
	return func() (conn *amqp.Connection, err error) {
		config, err := buildAMQPConfig(rootCACert, dnsResolvers)
		if err != nil {
			return nil, err
		}
		for i := 1; ; i++ {
			if conn, err = amqp.DialConfig(url, config); err == nil || i == 5 {
				return
			}
			log.Printf("AMQP connection failed, attempt %d", i)
			time.Sleep(time.Duration(i*3) * time.Second)
		}
		return
	}
}

func buildAMQPConfig(rootCACert, dnsResolvers string) (amqp.Config, error) {
	ret := amqp.Config{Heartbeat: 10 * time.Second}
	if dnsResolvers != "" {
		resolvers := dnsdialer.ParseResolvers(dnsResolvers)
		if len(resolvers) == 0 {
			return ret, errors.New("No valid DNS resolvers")
		}
		ret.Dial = dnsdialer.NewDialer(resolvers).Dial
	}
	if rootCACert == "" {
		return ret, nil
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM([]byte(rootCACert)) {
		return ret, errors.New("Root certificate not added")
	}
	ret.TLSClientConfig = &tls.Config{RootCAs: roots}
	return ret, nil
}
