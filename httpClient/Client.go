package httpClient

import "C"
import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"io"
	"net/http"
	"time"
)

type Config struct {
	RetryTimes    int
	RetryInterval int
	Header        http.Header
	Log           aws.Logger
}
type Client struct {
	http.Client
	Config
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	for k, v := range c.Header {
		req.Header[k] = v
	}
	for i := 0; i < c.RetryTimes; i++ {
		resp, err := c.Client.Do(req)
		if err != nil {
			fmt.Println(err)
			//c.Log.Log(err.Error())
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		} else {
			body, _ := io.ReadAll(resp.Body)
			fmt.Println(string(body))
			fmt.Println(resp.StatusCode)
			//c.Log.Log(resp.Request.URL, resp.Status)
		}
		time.Sleep(time.Duration(c.RetryInterval) * time.Second)
	}
	return nil, errors.New("http request failed")
}

func New(config *Config) *Client {
	if config == nil {
		config = &Config{
			RetryTimes:    5,
			RetryInterval: 2,
			Header:        http.Header{},
			Log:           nil,
		}
	}
	if config.RetryTimes == 0 {
		config.RetryTimes = 5
	}
	if config.RetryInterval == 0 {
		config.RetryInterval = 2
	}
	return &Client{
		Client: http.Client{},
		Config: *config,
	}

}
