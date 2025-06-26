package mock

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// RoundTripperFunc a type representing mock transport function
type RoundTripperFunc func(*http.Request) (*http.Response, error)

// RoundTrip a wrapper function
func (fn RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

// AddMockTransport add mock transport to the http client and returns the mock object
func AddMockTransport(t *testing.T, client *http.Client) Transport {
	mockTrans := &mockTransport{
		recordedRequests: make([]*http.Request, 0),
	}
	client.Transport = RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
		mockTrans.mutex.Lock()
		mockTrans.recordedRequests = append(mockTrans.recordedRequests, request)

		// Process expectations under lock to prevent race conditions
		for _, exp := range mockTrans.expectations {
			if exp.met {
				continue
			}

			if strings.EqualFold(exp.request.URL.String(), request.URL.String()) && strings.EqualFold(exp.request.Method, request.Method) {
				exp.met = true
				mockTrans.mutex.Unlock()

				if len(exp.validator) == 0 {
					return exp.response, exp.error
				}

				for _, validator := range exp.validator {
					validator(t, request)
				}

				return exp.response, exp.error
			}
		}
		mockTrans.mutex.Unlock()

		assert.Failf(t, "Unexpected http request", "Request `%s %s` was not expected but client initiated", request.Method, request.URL)
		return nil, nil
	})
	return mockTrans
}

// Transport an interface for setting the mock expectations
type Transport interface {
	// Expect set http client expectation to make a request and return the response or error
	Expect(*http.Request) Expectation

	// ExpectWith expect request with method and url
	ExpectWith(method, url string) Expectation

	// AllExpectationMet returns TRUE when all expectation set on the mock is met, else returns FALSE
	AllExpectationMet() bool

	// RecordedRequests returns recorded requests to mock transport
	RecordedRequests() []*http.Request
}

// Expectation an interface exposing function for returning value for the given expectation
type Expectation interface {
	Return(*http.Response, error)
	ReturnWith(statusCode int, body string)
}

// RequestValidator a type representing a validator function
type RequestValidator func(t *testing.T, actual *http.Request) bool

type expectation struct {
	request   *http.Request
	response  *http.Response
	validator []RequestValidator
	error     error
	met       bool
}

func (e *expectation) Return(response *http.Response, err error) {
	e.response = response
	e.error = err
}

func (e *expectation) ReturnWith(statusCode int, body string) {
	e.response = &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

type mockTransport struct {
	expectations     []*expectation
	recordedRequests []*http.Request
	mutex            sync.RWMutex
}

func (m *mockTransport) RecordedRequests() []*http.Request {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return a copy to prevent external modifications
	result := make([]*http.Request, len(m.recordedRequests))
	copy(result, m.recordedRequests)
	return result
}

func (m *mockTransport) AllExpectationMet() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, exp := range m.expectations {
		if !exp.met {
			return false
		}
	}
	return true
}

func (m *mockTransport) ExpectWith(method, url string) Expectation {
	request, _ := http.NewRequestWithContext(context.TODO(), method, url, nil)
	return m.expect(request)
}

func (m *mockTransport) Expect(request *http.Request) Expectation {
	return m.expect(request, func(t *testing.T, actual *http.Request) bool {
		return assert.Equal(t, request, actual)
	})
}

func (m *mockTransport) expect(request *http.Request, validators ...RequestValidator) Expectation {
	expect := &expectation{
		request:   request,
		validator: validators,
	}

	m.mutex.Lock()
	m.expectations = append(m.expectations, expect)
	m.mutex.Unlock()

	return expect
}
