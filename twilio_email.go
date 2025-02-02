package sendgrid

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"

	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailClient struct {
	Client
	emailOptions TwilioEmailOptions
}

// TwilioEmailOptions for GetTwilioEmailRequest
type TwilioEmailOptions struct {
	Username string
	Password string
	Endpoint string
	Host     string
}

// NewTwilioEmailSendClient constructs a new Twilio Email client given a username and password
func NewTwilioEmailSendClient(username, password string) *EmailClient {
	return &EmailClient{emailOptions: TwilioEmailOptions{Username: username, Password: password}}
}

// GetTwilioEmailRequest create Request
// @return [Request] a default request object
func GetTwilioEmailRequest(twilioEmailOptions TwilioEmailOptions) rest.Request {
	credentials := twilioEmailOptions.Username + ":" + twilioEmailOptions.Password
	encodedCreds := base64.StdEncoding.EncodeToString([]byte(credentials))

	options := options{
		Auth:     "Basic " + encodedCreds,
		Endpoint: "/v3/mail/send",
		Host:     twilioEmailOptions.Host,
	}

	if options.Host == "" {
		options.Host = "https://email.twilio.com"
	}

	return requestNew(options)
}

// prepareRequest prepares the email request with the given headers
func (cl *EmailClient) prepareRequest(email *mail.SGMailV3, headers map[string]string) (rest.Request, error) {
	var request rest.Request
	request = GetTwilioEmailRequest(cl.emailOptions)
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
