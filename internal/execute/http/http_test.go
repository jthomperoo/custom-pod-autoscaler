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

package http_test

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	gohttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/config"
	"github.com/jthomperoo/custom-pod-autoscaler/v2/internal/execute/http"
)

const (
	validCACert = `-----BEGIN CERTIFICATE-----
MIIDtTCCAp2gAwIBAgIUNTpMyBZ6+m5tPnJv4MlRhOUm4skwDQYJKoZIhvcNAQEL
BQAwajELMAkGA1UEBhMCR0IxDzANBgNVBAgMBkFudHJpbTEQMA4GA1UEBwwHQmVs
ZmFzdDEWMBQGA1UECgwNUGxvcGNpdHkgSW5jLjEgMB4GA1UEAwwXY3VzdG9tcG9k
YXV0b3NjYWxlci5jb20wHhcNMjIwMzI3MTk0NjEyWhcNMzIwMzI0MTk0NjEyWjBq
MQswCQYDVQQGEwJHQjEPMA0GA1UECAwGQW50cmltMRAwDgYDVQQHDAdCZWxmYXN0
MRYwFAYDVQQKDA1QbG9wY2l0eSBJbmMuMSAwHgYDVQQDDBdjdXN0b21wb2RhdXRv
c2NhbGVyLmNvbTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMljqzjS
HDU+OxlYW8P8KJjCabE1K9TgCFh/r5RW1IyN7U5k3GxIP6IcKxFrZGvGGuWWv8uu
g85FuvyeOkaDUWesJ1MdtqaoZUz88gbw4kLqx+cYMk2AjNUsUfIO1M7DtHWtLXg2
KSrycygUibCGsPro/xBkjka0NDWy5CBvJYODwWSb48h01iOXYgFzBE/EOeLA5GB3
HUh6Jyo8RUI1b+FQ48aX6Z0vmL1khgJj/F3OUXoWNpj7fPs6zhJfDIIXKzMgAG64
Mw0gBqgxeq1IdVvV2N7oLSaa3dvGZxPN5AhugbNV5NN91S1TgMEDkx1nDqPou/FI
3eRxtziytsW59h0CAwEAAaNTMFEwHQYDVR0OBBYEFDefDEDbDwZYgWepTj8hqzf/
FGMOMB8GA1UdIwQYMBaAFDefDEDbDwZYgWepTj8hqzf/FGMOMA8GA1UdEwEB/wQF
MAMBAf8wDQYJKoZIhvcNAQELBQADggEBALonKAB8GD0mQGkEdKo01U1YBiBRhZNH
kfHZNtLwP7YruUc5UVKoB7aO7b20UOSeRHb/O4MlwhztTFIfIUUPDEw6cifKt+fN
2O2QHqlV1aZZRojA1EWaa36JGY4WB8nMt5CO+lUFB9EwAa4+9s5679QHu21CbW3m
xVBWRp3C/dFpNpPQ69yEW2aJrlU3d87szMtI7UeJrmH0w2RjIgodTtjcAb/+9TRc
3P7Jpc1fGsQXRSTFfPy/LsdaGW5lwazqxaXpjYUJe2FCbpE37oqJw8dQDykDcbyU
Ivi3KBFKDE3BJWHqC6h8dXlLau0pWXWBwvVfiAAAiGCU7pDnmQJWAw8=
-----END CERTIFICATE-----`
	validClientCert = `-----BEGIN CERTIFICATE-----
MIIDUTCCAjkCAWQwDQYJKoZIhvcNAQELBQAwajELMAkGA1UEBhMCR0IxDzANBgNV
BAgMBkFudHJpbTEQMA4GA1UEBwwHQmVsZmFzdDEWMBQGA1UECgwNUGxvcGNpdHkg
SW5jLjEgMB4GA1UEAwwXY3VzdG9tcG9kYXV0b3NjYWxlci5jb20wHhcNMjIwMzI3
MTk0NzIyWhcNMjYwMzI2MTk0NzIyWjBzMQswCQYDVQQGEwJHQjEPMA0GA1UECAwG
QW50cmltMRAwDgYDVQQHDAdCZWxmYXN0MR8wHQYDVQQKDBZDaHVja2xlcGxvcCBJ
bmR1c3RyaWVzMSAwHgYDVQQDDBdjdXN0b21wb2RhdXRvc2NhbGVyLmNvbTCCASIw
DQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAL7Zm6DFRnEPbxtnZh12bbYo9Wvo
r+IGm7gjx6UAB57YCKunodEk5Rd1kezNbY5agkQl6Rjp6XW+ys18CV9ZXCJN+AHo
v8b4rx+yyzFDdR8yF4XGQEvP84FkAr5WpXblJC9VxSxcOEcUFQMvp1yHUWdylaDg
PqzIRdyIF0s4gKDl7fCSoWghvpOV+iRfK5kheMsgI+zalFmztYmm/p9KRisXErKr
mC8RAAP326zwk9VICgX4+4fVP0qseQJsnNZt3ctV3RCJNY/4f0V4CwIme715FuDC
loy3TeYeYcFdAFw+6v0X9cuFasa8076cQULli66QvqJXtQa5eg51NWDRet8CAwEA
ATANBgkqhkiG9w0BAQsFAAOCAQEANA9doRKfWCc215FaSkwmkiiPNI+fMmEJ2coy
xvyTUjCIdUMMAcBlwIRTYyCm0qiZt6mhdE5RMT8NkF3Dx3TWVU/5fkaokgMBs3mN
FvxXEJLptwafAzfbAzIsGYStB/n9FF356yKB9uF5DVQrGfsvCJCxXVXsNAATVIrO
j24NdssUSiMLrXx3+aWptcZS+8KM6oTRFLcoThilw9LWqg6UzyJNEa5pYYrj464M
74e/3TkGNjQrPWefopgoFH5TOh4VzjKyYPs85x0r2+L0yHGMAsbjtiGnKy40fhsE
nY0nV1d2wY03OX1Y53UuUNYAS7Z42SZPozxsnqYLntRW3RyF/g==
-----END CERTIFICATE-----`
	validClientKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAvtmboMVGcQ9vG2dmHXZttij1a+iv4gabuCPHpQAHntgIq6eh
0STlF3WR7M1tjlqCRCXpGOnpdb7KzXwJX1lcIk34Aei/xvivH7LLMUN1HzIXhcZA
S8/zgWQCvlalduUkL1XFLFw4RxQVAy+nXIdRZ3KVoOA+rMhF3IgXSziAoOXt8JKh
aCG+k5X6JF8rmSF4yyAj7NqUWbO1iab+n0pGKxcSsquYLxEAA/fbrPCT1UgKBfj7
h9U/Sqx5Amyc1m3dy1XdEIk1j/h/RXgLAiZ7vXkW4MKWjLdN5h5hwV0AXD7q/Rf1
y4VqxrzTvpxBQuWLrpC+ole1Brl6DnU1YNF63wIDAQABAoIBAQCtMOqzBfM7zIbG
dpnLaNp88URuLZXW5qpPe2DEUneX5XJQ2+nT0sh29oF7RJ0EDwyh7UlQC42KBZ8i
xEn5Fi2vD9RdXysR1EGP4X/Vb+MMcTe5dUSJx+exZuG6ewTjFWQ19H7fF70We5np
70fQhxgPaYNxn64gvAnx7Co/X8ISv6QwldaHLnmYUgws+4jbRhedGadMK2uk2Q6m
QtkbdRXR23a6bP2eMvIlyQjZ/ymdLnGpTVVqt/8bUkmZZcapW1KnsupRNZDKd5Mv
AW8Cb5aARIOmg8oIY+TGsvTdrzI6EZhOcGupM47Ek9xMXReUXWZm1mbF/uVmy1GE
WsRIadUxAoGBAOVGn2Qjs88mTK3n54VA6oLc5KnaJMWzb+g/d7d7yxqLGShhyLi7
Nyh6z822Z2f2TSEZm5tpmtruy8PjbDP5pTXh0EdGrlVoFTbixlEJmikivdCG0EdZ
nDyAoTNZarw9Nv4RbpKQevF+CHmsd4zeHJzxa4vvfnvSfw4g4d8nCteJAoGBANUY
ZOyfdd+pFerebcR0iLdN0PuBTUy42TrG6udugXWN98HjiVxzCPPySt7l1IFyYMNU
QNVtfSjTn0rxqTR8Kmf6TXEYvCeSOnD2GBC05GNgEfIc17Ws7DRVU6u5R3TSJkZT
B85LC0QTaHkdIHBtqkS8prDzknBgNDRgOnE/lz0nAoGBAMA5X7UzgbNxZuR/A8ri
zDr/O+9z51Shxncvjw2UioosiOEkaocG343euY69GSE+jRftQlsgRSa9ArWvXK6O
5YaxVlAL4GnWo8Kqip6ysD9A02ebP9AyPx0ysvQ8SZKcuBh3QP88hvclNbHdeTXv
momylvpxxtfFeaS8yOKw9xQRAoGAPgeE0VA1mq54f87Rev9FEL6pF5zy6GNicHaY
yFdlfdeeiCT4xh2CPKiJ3MpgCnJF8nXjDX16kljPpJwl2e5O1ceJpWNC2e357wnj
xXEyji7X6nc032B/vVgdK/6Z60qE87GVsxorJFzV63NsDu4NQ0b66sVsGiQW7iwY
twCAwL8CgYBkLwaYD6EXtWud3IOyFooLWfubvo9tQBFNQ5P+kvwt0pNlUAA/Yar3
vZMN3TPRTuI/cWf+QQufW1zPJLIQLOPJ0wfYNGC/eEX5DzKSU6TYiYcumnmE4xYB
ecUYC36yD3rjMBkxDmINgyGLJvO3CEFNvCItjDUcNv2rsvWJifermA==
-----END RSA PRIVATE KEY-----`
)

func stringPtr(s string) *string {
	return &s
}

type testHTTPClient struct {
	RoundTripReactor func(req *gohttp.Request) (*gohttp.Response, error)
}

func (f *testHTTPClient) RoundTrip(req *gohttp.Request) (*gohttp.Response, error) {
	return f.RoundTripReactor(req)
}

type testReader struct {
	ReadReactor  func(p []byte) (n int, err error)
	CloseReactor func() error
}

func (f *testReader) Read(p []byte) (n int, err error) {
	return f.ReadReactor(p)
}

func (f *testReader) Close() error {
	return f.CloseReactor()
}

func TestExecute_ExecuteWithValue(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})
	var tests = []struct {
		description string
		expected    string
		expectedErr error
		method      *config.Method
		value       string
		execute     http.Execute
	}{
		{
			"Fail, missing HTTP method configuration",
			"",
			errors.New(`missing required 'http' configuration on method`),
			&config.Method{
				Type: "http",
			},
			"test",
			http.Execute{},
		},
		{
			"Fail, invalid HTTP method",
			"",
			errors.New(`net/http: invalid method "*?"`),
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method: "*?",
					URL:    "https://custompodautoscaler.com",
				},
			},
			"test",
			http.Execute{},
		},
		{
			"Fail, unknown parameter mode",
			"",
			errors.New(`unknown parameter mode 'unknown'`),
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "unknown",
				},
			},
			"test",
			http.Execute{},
		},
		{
			"Fail, CA cert provided, fail to read",
			"",
			errors.New(`Fail to read file`),
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
					CACert:        stringPtr("/test.pem"),
				},
			},
			"test",
			http.Execute{
				ReadFile: func(filename string) ([]byte, error) {
					return nil, errors.New("Fail to read file")
				},
			},
		},
		{
			"Fail, CA cert provided, invalid PEM",
			"",
			errors.New(`failed to populate CA root pool for HTTP execute`),
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
					CACert:        stringPtr("/test.pem"),
				},
			},
			"test",
			http.Execute{
				ReadFile: func(filename string) ([]byte, error) {
					return []byte("invalid"), nil
				},
			},
		},
		{
			"Fail, client cert and key provided, fail to read cert",
			"",
			errors.New(`Fail to read cert file`),
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
					ClientCert:    stringPtr("/client.cert"),
					ClientKey:     stringPtr("/client.key"),
				},
			},
			"test",
			http.Execute{
				ReadFile: func(filename string) ([]byte, error) {
					if filename == "/client.cert" {
						return nil, errors.New("Fail to read cert file")
					}
					return []byte("key"), nil
				},
			},
		},
		{
			"Fail, client cert and key provided, fail to read key",
			"",
			errors.New(`Fail to read key file`),
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
					ClientCert:    stringPtr("/client.cert"),
					ClientKey:     stringPtr("/client.key"),
				},
			},
			"test",
			http.Execute{
				ReadFile: func(filename string) ([]byte, error) {
					if filename == "/client.key" {
						return nil, errors.New("Fail to read key file")
					}
					return []byte("cert"), nil
				},
			},
		},
		{
			"Fail, client cert and key provided, invalid cert and key",
			"",
			errors.New(`tls: failed to find any PEM data in certificate input`),
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
					ClientCert:    stringPtr("/client.cert"),
					ClientKey:     stringPtr("/client.key"),
				},
			},
			"test",
			http.Execute{
				ReadFile: func(filename string) ([]byte, error) {
					return []byte("invalid"), nil
				},
			},
		},
		{
			"Fail, fail to generate Go HTTP client",
			"",
			errors.New(`fail to generate Go HTTP client`),
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
				},
			},
			"test",
			http.Execute{
				ClientGenerator: func(tlsConfig *tls.Config) (*gohttp.Client, error) {
					return nil, errors.New("fail to generate Go HTTP client")
				},
			},
		},
		{
			"Fail, request fail",
			"",
			errors.New(`Get "https://custompodautoscaler.com?value=test": Test network error!`),
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
				},
			},
			"test",
			http.Execute{
				ClientGenerator: func(tlsConfig *tls.Config) (*gohttp.Client, error) {
					return &gohttp.Client{
						Transport: &testHTTPClient{
							func(req *gohttp.Request) (*gohttp.Response, error) {
								return nil, errors.New("Test network error!")
							},
						},
					}, nil
				},
			},
		},
		{
			"Fail, timeout",
			"",
			errors.New(`Get "https://custompodautoscaler.com?value=test": context deadline exceeded`),
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
				},
				Timeout: 5,
			},
			"test",
			http.Execute{
				ClientGenerator: func(tlsConfig *tls.Config) (*gohttp.Client, error) {
					return func() *gohttp.Client {
						testserver := httptest.NewServer(gohttp.HandlerFunc(func(rw gohttp.ResponseWriter, req *gohttp.Request) {
							time.Sleep(10 * time.Millisecond)
							// Send response to be tested
							rw.Write([]byte(`OK`))
						}))
						defer testserver.Close()

						return testserver.Client()
					}(), nil
				},
			},
		},
		{
			"Fail, invalid response body",
			"",
			errors.New(`fail to read body!`),
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
				},
			},
			"test",
			http.Execute{
				ClientGenerator: func(tlsConfig *tls.Config) (*gohttp.Client, error) {
					return &gohttp.Client{
						Transport: &testHTTPClient{
							func(req *gohttp.Request) (*gohttp.Response, error) {
								resp := &gohttp.Response{
									Body: &testReader{
										ReadReactor: func(p []byte) (n int, err error) {
											return 0, errors.New("fail to read body!")
										},
									},
									Header: gohttp.Header{},
								}
								resp.Header.Set("Content-Length", "1")
								return resp, nil
							},
						},
					}, nil
				},
			},
		},
		{
			"Fail, bad response code",
			"",
			errors.New(`HTTP request failed, status: [400], response: 'bad request!'`),
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
					SuccessCodes: []int{
						200,
						202,
					},
				},
			},
			"test",
			http.Execute{
				ClientGenerator: func(tlsConfig *tls.Config) (*gohttp.Client, error) {
					return &gohttp.Client{
						Transport: &testHTTPClient{
							func(req *gohttp.Request) (*gohttp.Response, error) {
								return &gohttp.Response{
									Body:       io.NopCloser(strings.NewReader("bad request!")),
									Header:     gohttp.Header{},
									StatusCode: 400,
								}, nil
							},
						},
					}, nil
				},
			},
		},
		{
			"Success, POST, body parameter, 3 headers",
			"Success!",
			nil,
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method:        "POST",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "body",
					Headers: map[string]string{
						"a": "testa",
						"b": "testb",
						"c": "testc",
					},
					SuccessCodes: []int{
						200,
						202,
					},
				},
			},
			"test",
			http.Execute{
				ClientGenerator: func(tlsConfig *tls.Config) (*gohttp.Client, error) {
					return &gohttp.Client{
						Transport: &testHTTPClient{
							func(req *gohttp.Request) (*gohttp.Response, error) {

								if !cmp.Equal(req.Method, "POST") {
									return nil, fmt.Errorf("Invalid method, expected 'POST', got '%s'", req.Method)
								}

								// Read the request body
								body, err := io.ReadAll(req.Body)
								if err != nil {
									return nil, err
								}

								if !cmp.Equal(req.Header.Get("a"), "testa") {
									return nil, fmt.Errorf("Missing header 'a'")
								}
								if !cmp.Equal(req.Header.Get("b"), "testb") {
									return nil, fmt.Errorf("Missing header 'a'")
								}
								if !cmp.Equal(req.Header.Get("c"), "testc") {
									return nil, fmt.Errorf("Missing header 'a'")
								}
								if !cmp.Equal(string(body), "test") {
									return nil, fmt.Errorf("Invalid body, expected 'test', got '%s'", body)
								}

								return &gohttp.Response{
									Body:       io.NopCloser(strings.NewReader("Success!")),
									Header:     gohttp.Header{},
									StatusCode: 200,
								}, nil
							},
						},
					}, nil
				},
			},
		},
		{
			"Success, GET, query parameter, 1 header",
			"Success!",
			nil,
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
					Headers: map[string]string{
						"a": "testa",
					},
					SuccessCodes: []int{
						200,
						202,
					},
				},
			},
			"test",
			http.Execute{
				ClientGenerator: func(tlsConfig *tls.Config) (*gohttp.Client, error) {
					return &gohttp.Client{
						Transport: &testHTTPClient{
							func(req *gohttp.Request) (*gohttp.Response, error) {

								if !cmp.Equal(req.Method, "GET") {
									return nil, fmt.Errorf("Invalid method, expected 'GET', got '%s'", req.Method)
								}

								query := req.URL.Query()

								if !cmp.Equal(query.Get("value"), "test") {
									return nil, fmt.Errorf("Invalid query param, expected 'test', got '%s'", query.Get("value"))
								}

								if !cmp.Equal(req.Header.Get("a"), "testa") {
									return nil, fmt.Errorf("Missing header 'a'")
								}

								return &gohttp.Response{
									Body:       io.NopCloser(strings.NewReader("Success!")),
									Header:     gohttp.Header{},
									StatusCode: 200,
								}, nil
							},
						},
					}, nil
				},
			},
		},
		{
			"Success, CA cert provided",
			"Success!",
			nil,
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
					CACert:        stringPtr("/ca.pem"),
					Headers: map[string]string{
						"a": "testa",
					},
					SuccessCodes: []int{
						200,
						202,
					},
				},
			},
			"test",
			http.Execute{
				ReadFile: func(filename string) ([]byte, error) {
					switch filename {
					case "/ca.pem":
						return []byte(validCACert), nil
					default:
						return nil, errors.New("unknown file name")
					}
				},
				ClientGenerator: func(tlsConfig *tls.Config) (*gohttp.Client, error) {
					return &gohttp.Client{
						Transport: &testHTTPClient{
							func(req *gohttp.Request) (*gohttp.Response, error) {

								if !cmp.Equal(req.Method, "GET") {
									return nil, fmt.Errorf("Invalid method, expected 'GET', got '%s'", req.Method)
								}

								query := req.URL.Query()

								if !cmp.Equal(query.Get("value"), "test") {
									return nil, fmt.Errorf("Invalid query param, expected 'test', got '%s'", query.Get("value"))
								}

								if !cmp.Equal(req.Header.Get("a"), "testa") {
									return nil, fmt.Errorf("Missing header 'a'")
								}

								return &gohttp.Response{
									Body:       io.NopCloser(strings.NewReader("Success!")),
									Header:     gohttp.Header{},
									StatusCode: 200,
								}, nil
							},
						},
					}, nil
				},
			},
		},
		{
			"Success, client cert, and client key provided",
			"Success!",
			nil,
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
					ClientCert:    stringPtr("/client.cert"),
					ClientKey:     stringPtr("/client.key"),
					Headers: map[string]string{
						"a": "testa",
					},
					SuccessCodes: []int{
						200,
						202,
					},
				},
			},
			"test",
			http.Execute{
				ReadFile: func(filename string) ([]byte, error) {
					switch filename {
					case "/client.cert":
						return []byte(validClientCert), nil
					case "/client.key":
						return []byte(validClientKey), nil
					default:
						return nil, errors.New("unknown file name")
					}
				},
				ClientGenerator: func(tlsConfig *tls.Config) (*gohttp.Client, error) {
					return &gohttp.Client{
						Transport: &testHTTPClient{
							func(req *gohttp.Request) (*gohttp.Response, error) {

								if !cmp.Equal(req.Method, "GET") {
									return nil, fmt.Errorf("Invalid method, expected 'GET', got '%s'", req.Method)
								}

								query := req.URL.Query()

								if !cmp.Equal(query.Get("value"), "test") {
									return nil, fmt.Errorf("Invalid query param, expected 'test', got '%s'", query.Get("value"))
								}

								if !cmp.Equal(req.Header.Get("a"), "testa") {
									return nil, fmt.Errorf("Missing header 'a'")
								}

								return &gohttp.Response{
									Body:       io.NopCloser(strings.NewReader("Success!")),
									Header:     gohttp.Header{},
									StatusCode: 200,
								}, nil
							},
						},
					}, nil
				},
			},
		},
		{
			"Success, CA cert, client cert, and client key provided",
			"Success!",
			nil,
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
					CACert:        stringPtr("/ca.pem"),
					ClientCert:    stringPtr("/client.cert"),
					ClientKey:     stringPtr("/client.key"),
					Headers: map[string]string{
						"a": "testa",
					},
					SuccessCodes: []int{
						200,
						202,
					},
				},
			},
			"test",
			http.Execute{
				ReadFile: func(filename string) ([]byte, error) {
					switch filename {
					case "/ca.pem":
						return []byte(validCACert), nil
					case "/client.cert":
						return []byte(validClientCert), nil
					case "/client.key":
						return []byte(validClientKey), nil
					default:
						return nil, errors.New("unknown file name")
					}
				},
				ClientGenerator: func(tlsConfig *tls.Config) (*gohttp.Client, error) {
					return &gohttp.Client{
						Transport: &testHTTPClient{
							func(req *gohttp.Request) (*gohttp.Response, error) {

								if !cmp.Equal(req.Method, "GET") {
									return nil, fmt.Errorf("Invalid method, expected 'GET', got '%s'", req.Method)
								}

								query := req.URL.Query()

								if !cmp.Equal(query.Get("value"), "test") {
									return nil, fmt.Errorf("Invalid query param, expected 'test', got '%s'", query.Get("value"))
								}

								if !cmp.Equal(req.Header.Get("a"), "testa") {
									return nil, fmt.Errorf("Missing header 'a'")
								}

								return &gohttp.Response{
									Body:       io.NopCloser(strings.NewReader("Success!")),
									Header:     gohttp.Header{},
									StatusCode: 200,
								}, nil
							},
						},
					}, nil
				},
			},
		},
		{
			"Success, GET, query parameter, 0 headers",
			"Success!",
			nil,
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method:        "GET",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "query",
					SuccessCodes: []int{
						200,
						202,
					},
				},
			},
			"test",
			http.Execute{
				ClientGenerator: func(tlsConfig *tls.Config) (*gohttp.Client, error) {
					return &gohttp.Client{
						Transport: &testHTTPClient{
							func(req *gohttp.Request) (*gohttp.Response, error) {

								if !cmp.Equal(req.Method, "GET") {
									return nil, fmt.Errorf("Invalid method, expected 'GET', got '%s'", req.Method)
								}

								query := req.URL.Query()

								if !cmp.Equal(query.Get("value"), "test") {
									return nil, fmt.Errorf("Invalid query param, expected 'test', got '%s'", query.Get("value"))
								}

								return &gohttp.Response{
									Body:       io.NopCloser(strings.NewReader("Success!")),
									Header:     gohttp.Header{},
									StatusCode: 200,
								}, nil
							},
						},
					}, nil
				},
			},
		},
		{
			"Success, PUT, body parameter, 0 headers",
			"Success!",
			nil,
			&config.Method{
				Type: "http",
				HTTP: &config.HTTP{
					Method:        "PUT",
					URL:           "https://custompodautoscaler.com",
					ParameterMode: "body",
					SuccessCodes: []int{
						200,
						202,
					},
				},
			},
			"test",
			http.Execute{
				ClientGenerator: func(tlsConfig *tls.Config) (*gohttp.Client, error) {
					return &gohttp.Client{
						Transport: &testHTTPClient{
							func(req *gohttp.Request) (*gohttp.Response, error) {

								if !cmp.Equal(req.Method, "PUT") {
									return nil, fmt.Errorf("Invalid method, expected 'PUT', got '%s'", req.Method)
								}

								// Read the request body
								body, err := io.ReadAll(req.Body)
								if err != nil {
									return nil, err
								}

								if !cmp.Equal(string(body), "test") {
									return nil, fmt.Errorf("Invalid body, expected 'test', got '%s'", body)
								}

								return &gohttp.Response{
									Body:       io.NopCloser(strings.NewReader("Success!")),
									Header:     gohttp.Header{},
									StatusCode: 200,
								}, nil
							},
						},
					}, nil
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := test.execute.ExecuteWithValue(test.method, test.value)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, result) {
				t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}

func TestExecute_GetType(t *testing.T) {
	var tests = []struct {
		description string
		expected    string
		execute     http.Execute
	}{
		{
			"Return type",
			"http",
			http.Execute{},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := test.execute.GetType()
			if !cmp.Equal(test.expected, result) {
				t.Errorf("metrics mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}
