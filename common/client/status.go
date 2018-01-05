package client

import (
	"fmt"
	"net/http"
)

// GlobalStatus sends a GET request to the daemon. Returns the status details
// of the different components running in the daemon.
func (cli Client) GlobalStatus() (string, error) {
	serverResp, err := cli.R().Get("/healthz")
	if err != nil {
		return "", fmt.Errorf("error while connecting to daemon: %s", err)
	}

	if serverResp.StatusCode() != http.StatusOK {
		return "", processErrorBody(serverResp.Body(), nil)
	}
	// TODO (awander): add StatusResponse
	// var resp endpoint.StatusResponse
	// if err := json.Unmarshal(serverResp.Body(), &resp); err != nil {
	// 	return nil, err
	// }
	// return &resp, nil

	return "OK", nil
}
