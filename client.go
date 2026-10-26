package textbelt

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const defaultURL = "https://textbelt.com/text"

type Client struct {
	url      string
	apiKey   string
	testMode bool
	sender   *string
}

func NewClient(apiKey string, opts ...Option) *Client {
	o := options{
		url:      defaultURL,
		testMode: false,
	}
	for _, opt := range opts {
		opt(&o)
	}

	return &Client{
		apiKey:   apiKey,
		url:      o.url,
		testMode: o.testMode,
		sender:   o.sender,
	}
}

type options struct {
	url      string
	testMode bool
	sender   *string
}

// Option is used to customize
type Option func(*options)

// WithURL sets the TextBelt API endpoint to POST to.
func WithURL(url string) Option {
	return func(o *options) {
		o.url = url
	}
}

// WithTestMode specifies whether or not test mode is enabled.
func WithTestMode(mode bool) Option {
	return func(o *options) {
		o.testMode = mode
	}
}

// WithSender specifies the organization/message sender.
func WithSender(sender string) Option {
	return func(o *options) {
		o.sender = &sender
	}
}

type Response struct {
	Success        bool    `json:"success"`
	Error          *string `json:"error,omitempty"`
	QuotaRemaining int     `json:"quotaRemaining"`
	TextID         *string `json:"textId,omitempty"`
}

func (c *Client) Send(number, message string) (Response, error) {
	key := c.apiKey
	if c.testMode {
		key += "_test"
	}
	values := url.Values{
		"phone":   {number},
		"message": {message},
		"key":     {key},
	}
	if c.sender != nil && *c.sender != "" {
		values.Add("sender", *c.sender)
	}

	resp, err := http.PostForm(c.url, values)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var body Response
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Response{}, fmt.Errorf("reading response body: %w", err)
	}

	if !body.Success {
		if body.Error != nil {
			return body, errors.New(*body.Error)
		}
		return body, errors.New("received an unsuccessful message without an error message")
	}

	return body, nil
}
