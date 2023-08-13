/*
Copyright 2021 The Custom Pod Autoscaler Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package http handles interactions over HTTP
package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	gohttp "net/http"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/config"
)

// Type http represents an HTTP request
const Type = "http"

const (
	// BodyParameterMode is a configuration flag specifying the value passed via
	// HTTP should be through the HTTP body parameter
	BodyParameterMode = "body"
	// QueryParameterMode is a configuration flag specifying the value passed
	// via HTTP should be through the HTTP query parameter
	QueryParameterMode = "query"
)

const (
	// QueryParameterKey is the key of the query parameter passed if the query
	// parameter mode is used, in the form https://example.com?value="DATA"
	QueryParameterKey = "value"
)

// DefaultExecute creates an HTTP executer with a default Go HTTP client generator and readfile utility
func DefaultExecute() *Execute {
	return &Execute{
		ClientGenerator: defaultClientGenerator,
		ReadFile:        os.ReadFile,
	}
}

func defaultClientGenerator(tlsConfig *tls.Config) (*gohttp.Client, error) {
	if tlsConfig != nil {
		return &gohttp.Client{
			Transport: &gohttp.Transport{
				TLSClientConfig: tlsConfig,
			},
		}, nil
	}
	return gohttp.DefaultClient, nil
}

// Execute represents a way to execute HTTP requests with values as parameters.
type Execute struct {
	ClientGenerator func(tlsConfig *tls.Config) (*gohttp.Client, error)
	ReadFile        func(filename string) ([]byte, error)
}

// ExecuteWithValue executes an HTTP request with the value provided as
// parameter, configurable to be either in the body or query string
func (e *Execute) ExecuteWithValue(method *config.Method, value string) (string, error) {
	if method.HTTP == nil {
		return "", fmt.Errorf("missing required 'http' configuration on method")
	}

	glog.V(4).Infof("Making HTTP request, method: '%s', URL: '%s'", method.HTTP.Method, method.HTTP.URL)

	// Set up request using method and URL provided
	req, err := gohttp.NewRequest(method.HTTP.Method, method.HTTP.URL, nil)
	if err != nil {
		return "", err
	}

	// Set parameter value, based on configuration option
	switch method.HTTP.ParameterMode {
	case BodyParameterMode:
		// Set body parameter
		req.Body = io.NopCloser(strings.NewReader(value))
	case QueryParameterMode:
		// Set query parameter
		query := req.URL.Query()
		query.Add(QueryParameterKey, value)
		req.URL.RawQuery = query.Encode()
	default:
		return "", fmt.Errorf("unknown parameter mode '%s'", method.HTTP.ParameterMode)
	}

	tlsConfig, err := e.getTLSConfig(method.HTTP)
	if err != nil {
		return "", err
	}

	httpClient, err := e.ClientGenerator(tlsConfig)
	if err != nil {
		return "", err
	}

	// Add headers
	for key, val := range method.HTTP.Headers {
		req.Header.Add(key, val)
	}

	// Set up a context to provide an HTTP request timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(method.Timeout)*time.Millisecond)
	defer cancel()

	// Make request
	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", err
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Check for a successful response code
	success := false
	for _, successCode := range method.HTTP.SuccessCodes {
		if resp.StatusCode == successCode {
			success = true
			break
		}
	}

	if !success {
		return "", fmt.Errorf("HTTP request failed, status: [%d], response: '%s'", resp.StatusCode, string(body))
	}

	return string(body), nil
}

// GetType returns the http executer type
func (e *Execute) GetType() string {
	return Type
}

func (e *Execute) getTLSConfig(config *config.HTTP) (*tls.Config, error) {
	if config.CACert == nil && config.ClientCert == nil && config.ClientKey == nil {
		return nil, nil
	}

	var caCertPool *x509.CertPool
	if config.CACert != nil {
		caCert, err := e.ReadFile(*config.CACert)
		if err != nil {
			return nil, err
		}
		caCertPool = x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			return nil, errors.New("failed to populate CA root pool for HTTP execute")
		}
	}

	var cert *tls.Certificate
	if config.ClientCert != nil && config.ClientKey != nil {
		certPEMBlock, err := e.ReadFile(*config.ClientCert)
		if err != nil {
			return nil, err
		}
		keyPEMBlock, err := e.ReadFile(*config.ClientKey)
		if err != nil {
			return nil, err
		}
		loadedCert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
		if err != nil {
			return nil, err
		}
		cert = &loadedCert
	}

	tlsConfig := &tls.Config{}
	if caCertPool != nil {
		tlsConfig.RootCAs = caCertPool
	}

	if cert != nil {
		tlsConfig.Certificates = []tls.Certificate{*cert}
	}

	return tlsConfig, nil
}
