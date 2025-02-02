package sendgrid

import (
	"bytes"
	"compress/gzip"
	"context"

	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type iSendClient interface {
	Send(email *mail.SGMailV3) (*rest.Response, error)
	SendWithHeaders(email *mail.SGMailV3, headers map[string]string) (*rest.Response, error)
	SendWithContext(ctx context.Context, email *mail.SGMailV3, headers map[string]string) (*rest.Response, error)
}

type requestPreparer interface {
	prepareRequest(email *mail.SGMailV3, headers map[string]string) (rest.Request, error)
}

type Client struct {
	iSendClient
	requestPreparer
}

type ApiKeyClient struct {
	Client
	apiKey string
}

// NewApiKeySendClient constructs a new Twilio SendGrid client given an API key
func NewApiKeySendClient(key string) *ApiKeyClient {
	return &ApiKeyClient{apiKey: key}
}

// Send sends an email through Twilio SendGrid
func (cl *Client) Send(email *mail.SGMailV3) (*rest.Response, error) {
	return cl.SendWithContext(context.Background(), email, nil)
}

// SendWithHeaders sends an email through Twilio SendGrid with request headers
func (cl *Client) SendWithHeaders(email *mail.SGMailV3, headers map[string]string) (*rest.Response, error) {
	return cl.SendWithContext(context.Background(), email, headers)
}

// SendWithContext sends an email through Twilio SendGrid with context.Context.
func (cl *Client) SendWithContext(ctx context.Context, email *mail.SGMailV3, headers map[string]string) (*rest.Response, error) {
	request, err := cl.prepareRequest(email, headers)
	if err != nil {
		return nil, err
	}
	return MakeRequestWithContext(ctx, request)
}

func (cl *ApiKeyClient) prepareRequest(email *mail.SGMailV3, headers map[string]string) (rest.Request, error) {
	var request rest.Request
	request = GetRequest(cl.apiKey, "/v3/mail/send", "")
	request.Method = "POST"

	for k, v := range headers {
		request.Headers[k] = v
	}

	request.Body = mail.GetRequestBody(email)
	// when Content-Encoding header is set to "gzip"
	// mail body is compressed using gzip according to
	// https://docs.sendgrid.com/api-reference/mail-send/mail-send#mail-body-compression
	if request.Headers["Content-Encoding"] == "gzip" {
		var gzipped bytes.Buffer
		gz := gzip.NewWriter(&gzipped)
		if _, err := gz.Write(request.Body); err != nil {
			return request, err
		}
		if err := gz.Flush(); err != nil {
			return request, err
		}
		if err := gz.Close(); err != nil {
			return request, err
		}
		request.Body = gzipped.Bytes()
	}

	return request, nil
}

// sendGetSentReqBodyHeaders sends an email and returns response along with request body
func (cl *ApiKeyClient) sendGetSentReqBodyHeaders(email *mail.SGMailV3, headers map[string]string) (*rest.Response, []byte, error) {
	return cl.sendWithContextGetSentReqBodyHeaders(context.Background(), email, headers)
}

// sendWithContextGetSentReqBodyHeaders sends an email with context and returns response along with request body
func (cl *ApiKeyClient) sendWithContextGetSentReqBodyHeaders(ctx context.Context, email *mail.SGMailV3, headers map[string]string) (*rest.Response, []byte, error) {
	request, err := cl.prepareRequest(email, headers)
	if err != nil {
		return nil, nil, err
	}
	response, err := MakeRequestWithContext(ctx, request)
	return response, request.Body, err
}
