package dao

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/otokaze/go-kit/progressbar"
)

var (
	ErrCanceled = errors.New("Canceled")
)

type dao struct {
	bar     *progressbar.Bar
	httpCli *http.Client
	token   *token
}

func New() *dao {
	return &dao{
		bar: progressbar.New(nil),
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
