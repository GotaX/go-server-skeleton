package client

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"

	"github.com/GotaX/go-server-skeleton/pkg/errors"
)

const DefaultEndpoint = "http://localhost:3500/v1.0"

type DaprHttpClient interface {
	Invoke(ctx context.Context, service, method string, req, resp interface{}) error
}

func NewDaprHttpClient() DaprHttpClient {
	return &daprHttpClient{
		Client:   resty.New(),
		Endpoint: DefaultEndpoint,
	}
}

type daprHttpClient struct {
	Client   *resty.Client
	Endpoint string
}

func (c *daprHttpClient) Invoke(ctx context.Context, service, method string, req, resp interface{}) error {
	addr := fmt.Sprintf("%s/invoke/%s/method/%s", c.Endpoint, service, method)

	response, err := c.Client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&resp).
		Post(addr)

	if err != nil {
		return err
	}
	if response.IsError() {
		return errors.FromHttp(response.RawResponse)
	}
	return nil
}
