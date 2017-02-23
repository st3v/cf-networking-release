package client

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/lager"

	"lib/json_client"
	"silk-controller/models"
)

type ControllerClient interface {
	List() ([]models.Lease, error)
	Acquire() (*models.AcquireLeaseResponse, error)
	Renew(request models.Lease) (*models.AcquireLeaseResponse, error)
	Delete(leaseID string) error
}

type apiClient struct {
	json_client.JsonClient
	localVtepIP string
}

func New(logger lager.Logger, baseURL string, localVtepIP string) ControllerClient {
	return &apiClient{
		localVtepIP: localVtepIP,
		JsonClient:  json_client.New(logger, http.DefaultClient, baseURL),
	}
}

func (c *apiClient) List() ([]models.Lease, error) {
	var leases []models.Lease
	err := c.Do("GET", "/leases", nil, &leases, "")
	return leases, err
}

func (c *apiClient) Acquire() (*models.AcquireLeaseResponse, error) {
	var resp models.AcquireLeaseResponse
	request := models.NewLeaseRequest{
		VtepIP: c.localVtepIP,
	}
	err := c.Do("POST", "/leases", request, &resp, "")
	return &resp, err
}

func (c *apiClient) Renew(request models.Lease) (*models.AcquireLeaseResponse, error) {
	var resp models.AcquireLeaseResponse
	route := fmt.Sprintf("/leases/%s", request.ID)
	err := c.Do("PUT", route, request, &resp, "")
	return &resp, err
}

func (c *apiClient) Delete(leaseID string) error {
	route := fmt.Sprintf("/leases/%s", leaseID)
	err := c.Do("DELETE", route, nil, nil, "")
	return err
}