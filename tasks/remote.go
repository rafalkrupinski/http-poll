package tasks

import (
	"errors"
	"github.com/dghubble/sling"
	"net/http"
)

func (rs *RemoteSpecification) request(method string) (*sling.Sling, *http.Request, error) {
	client := sling.New().Client(rs.Client)

	client, err := SlingMethod(client, method, rs.Address)
	if err != nil {
		return nil, nil, err
	}

	req, err := client.Request()
	return client, req, err
}

func (rs *RemoteSpecification) GetObject(target interface{}) (*http.Response, error) {
	client, req, err := rs.request(http.MethodGet)
	if req != nil {
		return nil, err
	}

	return client.Do(req, target, nil)
}

func SlingMethod(s *sling.Sling, method, address string) (*sling.Sling, error) {
	switch method {
	case http.MethodGet:
		return s.Get(address), nil
	case http.MethodPost:
		return s.Post(address), nil
	default:
		return nil, errors.New("Unexpected method " + method)
	}
}
