package aiven

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type (
	Service struct {
		CloudName  string         `json:"cloud_name"`
		CreateTime string         `json:"create_time"`
		UpdateTime string         `json:"update_time"`
		GroupList  []string       `json:"group_list"`
		NodeCount  int            `json:"node_count"`
		Plan       string         `json:"plan"`
		Name       string         `json:"service_name"`
		Type       string         `json:"service_type"`
		Uri        string         `json:"service_uri"`
		State      string         `json:"state"`
		Metadata   interface{}    `json:"metadata"`
		Users      []*ServiceUser `json:"users"`
	}

	ServicesHandler struct {
		client *Client
	}

	CreateServiceRequest struct {
		Cloud       string `json:"cloud,omitempty"`
		GroupName   string `json:"group_name,omitempty"`
		Plan        string `json:"plan,omitempty"`
		ServiceName string `json:"service_name"`
		ServiceType string `json:"service_type"`
	}

	UpdateServiceRequest struct {
		Cloud     string `json:"cloud,omitempty"`
		GroupName string `json:"group_name,omitempty"`
		Plan      string `json:"plan,omitempty"`
		Powered   bool   `json:"powered"` // TODO: figure out if we can overwrite the default?
	}

	ServiceResponse struct {
		APIResponse
		Service *Service `json:"service"`
	}

	ServiceListResponse struct {
		APIResponse
		Services []*Service `json:"services"`
	}
)

// Hostname parses the hostname out of the Service URI.
func (s *Service) Hostname() (string, error) {
	hn, _, err := getHostPort(s.Uri)
	return hn, err
}

// Port parses the port out of the service URI.
func (s *Service) Port() (string, error) {
	_, port, err := getHostPort(s.Uri)
	return port, err
}

func getHostPort(uri string) (string, string, error) {
	hostUrl, err := url.Parse(uri)
	if err != nil {
		return "", "", err
	}

	if hostUrl.Host == "" {
		return hostUrl.Scheme, hostUrl.Opaque, nil
	}

	sp := strings.Split(hostUrl.Host, ":")
	if len(sp) != 2 {
		return "", "", ErrInvalidHost
	}

	return sp[0], sp[1], nil
}

func (h *ServicesHandler) Create(project string, req CreateServiceRequest) (*Service, error) {
	rsp, err := h.client.doPostRequest(fmt.Sprintf("/project/%s/service", project), req)
	if err != nil {
		return nil, err
	}

	return parseServiceResponse(rsp)
}

func (h *ServicesHandler) Get(project, service string) (*Service, error) {
	rsp, err := h.client.doGetRequest(fmt.Sprintf("/project/%s/service/%s", project, service), nil)
	if err != nil {
		return nil, err
	}

	return parseServiceResponse(rsp)
}

func (h *ServicesHandler) Update(project, service string, req UpdateServiceRequest) (*Service, error) {
	rsp, err := h.client.doPutRequest(fmt.Sprintf("/project/%s/service/%s", project, service), req)
	if err != nil {
		return nil, err
	}

	return parseServiceResponse(rsp)
}

func (h *ServicesHandler) Delete(project, service string) error {
	bts, err := h.client.doDeleteRequest(fmt.Sprintf("/project/%s/service/%s", project, service), nil)
	if err != nil {
		return err
	}

	return handleDeleteResponse(bts)
}

func (h *ServicesHandler) List(project string) ([]*Service, error) {
	rsp, err := h.client.doGetRequest(fmt.Sprintf("/project/%s/service", project), nil)
	if err != nil {
		return nil, err
	}

	var response *ServiceListResponse
	if err := json.Unmarshal(rsp, &response); err != nil {
		return nil, err
	}

	if len(response.Errors) != 0 {
		return nil, errors.New(response.Message)
	}

	return response.Services, nil
}

func parseServiceResponse(rsp []byte) (*Service, error) {
	var response *ServiceResponse
	if err := json.Unmarshal(rsp, &response); err != nil {
		return nil, err
	}

	if len(response.Errors) != 0 {
		return nil, errors.New(response.Message)
	}

	return response.Service, nil
}
