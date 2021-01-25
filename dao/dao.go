package dao

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

type dao struct {
	httpCli *http.Client
	token   *token
}

func New() *dao {
	return &dao{
		httpCli: &http.Client{
			Transport: &http.Transport{
				Proxy: nil,
				Dial: (&net.Dialer{
					Timeout:   60 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: 60 * time.Second,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
				MaxIdleConnsPerHost: 100,
				DisableCompression:  true,
			},
		},
		token: &token{},
	}
}
