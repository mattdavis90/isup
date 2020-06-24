package config

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"time"
)

type Client struct {
	CA      *string
	Cert    *string
	Key     *string
	Timeout *time.Duration
	client  *http.Client
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{},
	}
}

func (c *Client) Check() error {
	if c.Timeout == nil {
		timeout := 5 * time.Second
		c.Timeout = &timeout
	}

	tlsConfig := &tls.Config{}

	if c.Cert != nil && c.Key != nil {
		cert, err := tls.LoadX509KeyPair(*c.Cert, *c.Key)
		if err != nil {
			return err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if c.CA != nil {
		ca, err := ioutil.ReadFile(*c.CA)
		if err != nil {
			return err
		}
		tlsConfig.RootCAs = x509.NewCertPool()
		tlsConfig.RootCAs.AppendCertsFromPEM(ca)
	}

	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	c.client = &http.Client{
		Transport: transport,
		Timeout:   *c.Timeout,
	}

	return nil
}

func (c *Client) HttpClient() http.Client {
	return *c.client
}
