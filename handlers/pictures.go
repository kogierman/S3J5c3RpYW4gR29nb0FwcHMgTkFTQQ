package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"example.com/nasa-apod-fetcher/nasa"
)

const timeFormat = "2006-01-02" // YYYY-MM-DD

type ImageGetter interface {
	GetImages(string, string) ([]string, error)
}

type Pictures struct {
	img ImageGetter
}

type response struct {
	URLs  []string `json:"urls,omitempty"`
	Error string   `json:"error,omitempty"`
}

func NewPictures(img ImageGetter) *Pictures {
	return &Pictures{img}
}

func sendResponse(status int, resp response, rw http.ResponseWriter) {
	respBytes, err := json.Marshal(resp)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Internal server error"))
		return
	}
	rw.WriteHeader(status)
	rw.Write(respBytes)
}

func (p *Pictures) Handle(rw http.ResponseWriter, req *http.Request) {
	var resp response
	var start, end string
	params := req.URL.Query()
	if startDate, ok := params["start_date"]; ok && len(startDate) == 1 {
		start = startDate[0]
	}
	if endDate, ok := params["end_date"]; ok && len(endDate) == 1 {
		end = endDate[0]
	}
	if err := p.validate(start, end); err != nil {
		resp.Error = err.Error()
		sendResponse(http.StatusUnprocessableEntity, resp, rw)
		return
	}

	urls, err := p.img.GetImages(start, end)
	if errors.Is(err, nasa.ErrRateLimitExceeded) {
		resp.Error = "rate limit exceeded"
		sendResponse(http.StatusTooManyRequests, resp, rw)
		return
	}
	if err != nil {
		resp.Error = err.Error()
		sendResponse(http.StatusInternalServerError, resp, rw)
		return
	}
	resp.URLs = urls

	sendResponse(200, resp, rw)
}

func (p *Pictures) validate(startDate, endDate string) error {
	var start, end time.Time
	var err error
	now := time.Now()
	if startDate != "" {
		start, err = time.Parse(timeFormat, startDate)
		if err != nil {
			return errors.New("`start_date` should be in the format YYYY-MM-DD")
		}
	} else {
		start = now
	}

	if endDate != "" {
		if startDate == "" {
			return errors.New("`end_date` cannot be used without `start_date`")
		}
		end, err = time.Parse(timeFormat, endDate)
		if err != nil {
			return errors.New("`end_date` should be in the format YYYY-MM-DD")
		}
	} else {
		end = now
	}

	if start.After(now) || end.After(now) {
		return errors.New("`start_date` and `end_date` cannot be future dates")
	}

	if end.Before(start) {
		return errors.New("`start_date` must be before `end_date`")
	}

	return nil
}
