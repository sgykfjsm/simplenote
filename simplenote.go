package simplenote

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	API_BASE_URL     = "https://simple-note.appspot.com"
	AUTH_URL         = API_BASE_URL + "/api/login"
	DATA_URL         = API_BASE_URL + "/api2/data"
	INDEX_URL        = API_BASE_URL + "/api2/index"
	TAG_URL          = API_BASE_URL + "/api2/tag"
	MIN_INDEX_LENGTH = 1
	MAX_INDEX_LENGTH = 100
)

type Error interface {
	error
	Status() int
}

type ErrorResponse struct {
	Code int
	Err  error
}

func (e *ErrorResponse) Error() string {
	return e.Err.Error()
}

func (e *ErrorResponse) Status() int {
	return e.Code
}

type Client struct {
	// Email is required with every request
	Email string
	// Token is passed from simplenote API
	Token []byte
	// Debug output the log if true
	Debug bool
}

type Data struct {
	// Response From /api2/data
	Modifydate string   `json:"modifydate"` // Modifydate is last modified date, in seconds since epoch and set by client
	Tags       []string `json:"tags"`       // Tags is set by client, some set by server
	Deleted    int      `json:"deleted"`    // Deleted is integer pseudo-boolean, in trash or not and set by client
	Createdate string   `json:"createdate"` // Createdate is note created date, in seconds since epoch and set by client
	Systemtags []string `json:"systemtags"` // Systemtags is set by client, some set by server
	Content    string   `json:"content"`    // Content is data content and set by set by client when creating, otherwise may be set by server
	Version    int      `json:"version"`    // Version is integer, number set by server, track note changes
	Syncnum    int      `json:"syncnum"`    // Syncnum is integer, number set by server, track note content changes
	Key        string   `json:"key"`        // Key is note identifier and set by server
	ShareKey   string   `json:"sharekey"`   // ShareKey is shared note identifier and set by server
	PublishKey string   `json:"publishkey"` // PublishKey is published note identifier and set by server
	Minversion int      `json:"minversion"` // Minversion is integer, number set by server, miniumum version available for note
	// Response From /api2/index
	Count int        `json:"count"`
	Data  []struct { // This is intended to make struct simplify.
		Modifydate string      `json:"modifydate"`
		Tags       []string    `json:"tags"`
		Deleted    int         `json:"deleted"`
		Createdate string      `json:"createdate"`
		Systemtags []string    `json:"systemtags"`
		Content    string      `json:"content"`
		Version    int         `json:"version"`
		Syncnum    int         `json:"syncnum"`
		Key        string      `json:"key"`
		Sharekey   string      `json:"sharekey"`
		Publishkey string      `json:"publishkey"`
		Minversion int         `json:"minversion"`
		Count      int         `json:"count"`
		Data       interface{} `json:"data"`
		Time       string      `json:"time"`
		Mark       string      `json:"mark"`
	} `json:"data"`
	Time string `json:"time"`
	Mark string `json:"mark"`
}

func auth(email, password string) ([]byte, *ErrorResponse) {
	query := fmt.Sprintf("email=%s&password=%s", email, password)
	encodeQuery := base64.StdEncoding.EncodeToString([]byte(query))
	res, err := http.Post(AUTH_URL, "text/plain", strings.NewReader(encodeQuery))
	if err != nil || res.StatusCode != http.StatusOK {
		return nil, &ErrorResponse{
			Code: res.StatusCode,
			Err:  errors.New(http.StatusText(res.StatusCode)),
		}
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, &ErrorResponse{
			Code: http.StatusInternalServerError,
			Err:  errors.New(http.StatusText(http.StatusInternalServerError)),
		}
	}

	return body, nil
}

func (s *Client) buildNoteData(q url.Values) ([]byte, error) {
	var data Data

	if _, ok := q["content"]; ok {
		data.Content = q.Get("content")
	}

	if _, ok := q["tags"]; ok {
		data.Tags = q["tags"]
	}

	if _, ok := q["key"]; ok {
		data.Createdate = fmt.Sprintf("%f", UnixNowSecondFloat64())
	} else {
		data.Modifydate = fmt.Sprintf("%f", UnixNowSecondFloat64())
	}

	if _, ok := q["deleted"]; ok {
		data.Deleted = 1
	}

	return json.Marshal(data)
}

func (s *Client) request(method, requestURL string, q url.Values) (Data, *ErrorResponse) {
	if s.Debug {
		log.Printf("Method: %s", method)
		log.Printf("URL: %s", requestURL)
		log.Printf("Values: %s", q.Encode())
	}

	data := Data{}

	r, _ := url.Parse(requestURL)
	var res *http.Response
	var err error
	switch method {
	case http.MethodPost:
		var b []byte
		if b, err = s.buildNoteData(q); err != nil {
			return data, &ErrorResponse{
				Code: http.StatusInternalServerError,
				Err:  errors.New(http.StatusText(http.StatusInternalServerError)),
			}
		}
		q := url.Values{}
		q.Set("email", s.Email)
		q.Set("auth", string(s.Token))
		r.RawQuery = q.Encode()
		res, err = http.Post(r.String(), "application/json", bytes.NewReader(b))
	case http.MethodDelete:
		q.Set("email", s.Email)
		q.Set("auth", string(s.Token))
		r.RawQuery = q.Encode()
		req, err := http.NewRequest(http.MethodDelete, r.String(), nil)
		if err != nil {
			return data, &ErrorResponse{
				Code: http.StatusInternalServerError,
				Err:  errors.New(http.StatusText(http.StatusInternalServerError)),
			}
		}
		res, err = http.DefaultClient.Do(req)
	default: // http.MethodGet
		q.Set("email", s.Email)
		q.Set("auth", string(s.Token))
		r.RawQuery = q.Encode()
		res, err = http.Get(r.String())
	}

	if err != nil {
		return data, &ErrorResponse{
			Code: res.StatusCode,
			Err:  err,
		}
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return data, &ErrorResponse{
			Code: res.StatusCode,
			Err:  errors.New(res.Status),
		}
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return data, &ErrorResponse{
			Code: http.StatusInternalServerError,
			Err:  errors.New(http.StatusText(http.StatusInternalServerError)),
		}
	}

	err = json.Unmarshal(body, &data)
	return data, nil
}

func (s *Client) Index(length int, since, mark string) (Data, *ErrorResponse) {
	q := url.Values{}
	if length > MAX_INDEX_LENGTH {
		length = MAX_INDEX_LENGTH
	} else if length < MIN_INDEX_LENGTH {
		length = MIN_INDEX_LENGTH
	}
	q.Set("length", fmt.Sprintf("%v", length))

	if since != "" {
		q.Set("since", since)
	}

	if mark != "" {
		q.Set("mark", mark)
	}

	return s.request(http.MethodGet, INDEX_URL, q)
}

func (s *Client) Get(key string) (Data, *ErrorResponse) {
	return s.request(http.MethodGet, DATA_URL+"/"+key, url.Values{})
}

func (s *Client) Add(content string, tags []string) (Data, *ErrorResponse) {
	q := url.Values{}
	q.Set("content", content)
	for _, tag := range tags {
		q.Add("tags", tag)
	}

	return s.request(http.MethodPost, DATA_URL, q)
}

func (s *Client) Update(key, content string, is_delete bool) (Data, *ErrorResponse) {
	q := url.Values{}
	q.Set("content", content)
	if is_delete {
		q.Set("deleted", "1")
	}
	return s.request(http.MethodPost, DATA_URL+"/"+key, q)
}

func (s *Client) Delete(key string) *ErrorResponse {
	_, err := s.request(http.MethodDelete, DATA_URL+"/"+key, url.Values{})
	return err
}

func New(email, password string, debug bool) (*Client, *ErrorResponse) {
	token, err := auth(email, password)
	if err != nil {
		return nil, err
	}

	return &Client{
		Email: email,
		Token: token,
		Debug: debug,
	}, nil
}

func UnixNowSecondFloat64() float64 {
	return float64(time.Now().UnixNano()) / float64(time.Second)
}
