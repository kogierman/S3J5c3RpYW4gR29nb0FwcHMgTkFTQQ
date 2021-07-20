package nasa_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"example.com/nasa-apod-fetcher/nasa"
)

// TEST HELPERS

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(fn),
	}
}

func NewMockedCall(t *testing.T, status int, body string, validator func(*http.Request) error) roundTripFunc {
	return func(req *http.Request) *http.Response {
		if validator != nil {
			if e := validator(req); e != nil {
				t.Error(e)
			}
		}
		return &http.Response{
			StatusCode: status,
			Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
		}
	}
}

func NewQueryParamsValidator(expectedValues map[string]interface{}) func(*http.Request) error {
	return func(req *http.Request) error {
		params := req.URL.Query()
		for key, value := range expectedValues {
			switch value {
			case nil:
				if _, exist := params[key]; exist {
					return fmt.Errorf("%s queryparam shouldn't exist", key)
				}
			default:
				if params[key][0] != value {
					return fmt.Errorf("Invalid value for %s queryparam: %s [expected %s]", key, params[key][0], value)
				}
			}
		}
		return nil
	}
}

func TestRateLimitError(t *testing.T) {
	testClient := NewTestClient(NewMockedCall(t, http.StatusTooManyRequests, "rate limit exceeded", nil))
	apod := nasa.NewAPOD("key", 10, testClient)

	_, err := apod.GetImages("", "")
	if !errors.Is(err, nasa.ErrRateLimitExceeded) {
		t.Error("Invalid error type received")
	}
}

func TestApodRequestError(t *testing.T) {
	testClient := NewTestClient(NewMockedCall(t, http.StatusInternalServerError, "ISE", nil))
	apod := nasa.NewAPOD("key", 10, testClient)

	_, err := apod.GetImages("", "")
	if !errors.Is(err, nasa.ErrNASAReqFailed) {
		t.Error("Invalid error type received")
	}
}

func TestParsingError(t *testing.T) {
	testClient := NewTestClient(NewMockedCall(t, http.StatusOK, "{invalidJSON}}", nil))
	apod := nasa.NewAPOD("key", 10, testClient)

	_, err := apod.GetImages("", "")
	if !errors.Is(err, nasa.ErrNASAParsingError) {
		t.Error("Invalid error type received")
	}
}

func TestQueryParamsAll(t *testing.T) {
	expectedParams := map[string]interface{}{
		"start_date": "2021-01-01",
		"end_date":   "2021-03-01",
		"api_key":    "key",
	}
	testClient := NewTestClient(NewMockedCall(t, http.StatusOK, "{}", NewQueryParamsValidator(expectedParams)))
	apod := nasa.NewAPOD("key", 10, testClient)

	apod.GetImages("2021-01-01", "2021-03-01")
}

func TestQueryParamsOnlyStart(t *testing.T) {
	expectedParams := map[string]interface{}{
		"start_date": "2021-01-01",
		"end_date":   nil,
	}
	testClient := NewTestClient(NewMockedCall(t, http.StatusOK, "{}", NewQueryParamsValidator(expectedParams)))
	apod := nasa.NewAPOD("key", 10, testClient)

	apod.GetImages("2021-01-01", "")
}

func TestQueryParamsOnlyEnd(t *testing.T) {
	expectedParams := map[string]interface{}{
		"start_date": nil,
		"end_date":   "2021-01-01",
	}
	testClient := NewTestClient(NewMockedCall(t, http.StatusOK, "{}", NewQueryParamsValidator(expectedParams)))
	apod := nasa.NewAPOD("key", 10, testClient)

	apod.GetImages("", "2021-01-01")
}

func TestQueryParamsNone(t *testing.T) {
	expectedParams := map[string]interface{}{
		"start_date": nil,
		"end_date":   nil,
	}
	testClient := NewTestClient(NewMockedCall(t, http.StatusOK, "{}", NewQueryParamsValidator(expectedParams)))
	apod := nasa.NewAPOD("key", 10, testClient)

	apod.GetImages("", "")
}

func TestResponseSingleImage(t *testing.T) {
	testClient := NewTestClient(NewMockedCall(t, http.StatusOK, `{"url": "xyz"}`, nil))
	apod := nasa.NewAPOD("key", 10, testClient)

	imgs, _ := apod.GetImages("", "")
	if len(imgs) != 1 || imgs[0] != "xyz" {
		t.Errorf("Incorrectly parsed response: %v", imgs)
	}
}

func TestResponseMultipleImages(t *testing.T) {
	testClient := NewTestClient(NewMockedCall(t, http.StatusOK, `[{"url": "xyz"},{"url": "foo"}]`, nil))
	apod := nasa.NewAPOD("key", 10, testClient)

	imgs, _ := apod.GetImages("mock", "mock")
	if len(imgs) != 2 || imgs[0] != "xyz" || imgs[1] != "foo" {
		t.Errorf("Incorrectly parsed response: %v", imgs)
	}
}
