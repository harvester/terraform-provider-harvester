package util

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	maxHTTPRedirect = 5
)

func DoPost(url, data, cacert string, insecure bool, headers map[string]string) (*http.Response, error) {
	if url == "" {
		return nil, fmt.Errorf("doing post: URL is nil")
	}

	jsonBytes := []byte(data)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}, //nolint:gosec
		Proxy:           http.ProxyFromEnvironment,
	}

	if cacert != "" {
		// Get the SystemCertPool, continue with an empty pool on error
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		// Append our cert to the system pool
		if ok := rootCAs.AppendCertsFromPEM([]byte(cacert)); !ok {
			log.Println("No certs appended, using system certs only")
		}
		transport.TLSClientConfig.RootCAs = rootCAs
	}

	client.Transport = transport

	return client.Do(req) //nolint:gosec
}

func DoGet(url, username, password, token, cacert string, insecure bool) (*http.Response, error) {
	start := time.Now()

	if url == "" {
		return nil, fmt.Errorf("doing get: URL is nil")
	}
	log.Println("Getting from ", url)

	client := &http.Client{
		Timeout: time.Duration(60 * time.Second),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxHTTPRedirect {
				return fmt.Errorf("stopped after %d redirects", maxHTTPRedirect)
			}
			if len(token) > 0 {
				req.Header.Add("Authorization", "Bearer "+token)
			} else if len(username) > 0 && len(password) > 0 {
				s := username + ":" + password
				req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s)))
			}
			return nil
		},
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}, //nolint:gosec
		Proxy:           http.ProxyFromEnvironment,
	}

	if cacert != "" {
		// Get the SystemCertPool, continue with an empty pool on error
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		// Append our cert to the system pool
		if ok := rootCAs.AppendCertsFromPEM([]byte(cacert)); !ok {
			log.Println("No certs appended, using system certs only")
		}
		transport.TLSClientConfig.RootCAs = rootCAs
	}
	client.Transport = transport

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("doing get: %v", err)
	}
	if len(token) > 0 {
		req.Header.Add("Authorization", "Bearer "+token)
	} else if len(username) > 0 && len(password) > 0 {
		s := username + ":" + password
		req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s)))
	}
	// Timings recorded as part of internal metrics
	log.Println("Time to get req: ", float64((time.Since(start))/time.Millisecond), " ms")

	return client.Do(req) //nolint:gosec
}

// DoPostWithTransport sends a context-aware POST request using the given transport.
// Use this for requests that need a pre-configured transport (e.g. from rest.TransportFor)
// which handles TLS, proxy, and authentication settings.
func DoPostWithTransport(ctx context.Context, url string, body io.Reader, contentType string, transport http.RoundTripper) (*http.Response, error) {
	if url == "" {
		return nil, fmt.Errorf("doing post: URL is nil")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	client := &http.Client{Transport: transport}
	return client.Do(req) //nolint:gosec // G107: URL comes from trusted provider configuration
}
