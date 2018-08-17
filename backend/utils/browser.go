package utils

import (
	"io"
	"net/http"
	"net/url"
)

const fakeUserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:61.0) Gecko/20100101 Firefox/61.0"

// NewBrowserRequest creates new http request looking like coming from Firefox
func NewBrowserRequest(method, requestURL string, optionalBody io.Reader) (*http.Request, error) {
	parsedURL, err := url.Parse(requestURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, requestURL, optionalBody)
	if err != nil {
		return nil, err
	}

	fakeOrigin := parsedURL.Scheme + "://" + parsedURL.Host + "/"

	req.Header.Add("User-Agent", fakeUserAgent)
	req.Header.Add("Accept-Language", "en-US,de-DE;q=0.5")
	req.Header.Add("Referer", fakeOrigin)
	req.Header.Add("Origin", fakeOrigin)

	return req, nil
}
