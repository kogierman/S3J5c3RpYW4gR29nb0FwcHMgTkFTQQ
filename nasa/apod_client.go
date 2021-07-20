package nasa

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const APODHost = "https://api.nasa.gov"

var ErrRateLimitExceeded error = errors.New("rate limit exceeded")
var ErrNASAReqFailed error = errors.New("request to NASA API failed")
var ErrNASAParsingError error = errors.New("couldn't parse response from NASA API")

type APOD struct {
	apiKey         string
	concurrentReqs int
	httpClient     *http.Client

	concurrencyLimiter chan struct{}
}

func NewAPOD(k string, c int, h *http.Client) *APOD {
	if h == nil {
		h = &http.Client{
			Timeout: 10 * time.Second,
		}
	}
	cl := make(chan struct{}, c)
	for i := 0; i < c; i++ {
		cl <- struct{}{}
	}
	return &APOD{k, c, h, cl}
}

func (a *APOD) GetImages(start, end string) ([]string, error) {
	var req *http.Request
	query := url.Values{}
	query.Add("api_key", a.apiKey)
	if start != "" {
		query.Add("start_date", start)
	}
	if end != "" {
		query.Add("end_date", end)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/planetary/apod?%s", APODHost, query.Encode()), nil)
	if err != nil {
		return nil, err
	}

	<-a.concurrencyLimiter
	resp, err := a.httpClient.Do(req)
	if err != nil {
		a.concurrencyLimiter <- struct{}{}
		return nil, err
	}
	a.concurrencyLimiter <- struct{}{}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrRateLimitExceeded
	}

	if resp.StatusCode >= http.StatusBadRequest {
		log.Printf("Error [%d]: %s", resp.StatusCode, body)
		return nil, ErrNASAReqFailed
	}

	return a.getURLsFromResponse(body, start == "" && end == "")
}

func (a *APOD) getURLsFromResponse(body []byte, singleImage bool) ([]string, error) {
	type image struct {
		URL string `json:"url"`
	}

	var parsed []image
	if singleImage {
		var img image
		err := json.Unmarshal(body, &img)
		if err != nil {
			return nil, ErrNASAParsingError
		}
		parsed = append(parsed, img)
	} else {
		err := json.Unmarshal(body, &parsed)
		if err != nil {
			return nil, ErrNASAParsingError
		}
	}

	urls := make([]string, 0, len(parsed))
	for _, img := range parsed {
		urls = append(urls, img.URL)
	}
	return urls, nil
}
