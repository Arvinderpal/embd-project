package client

import (
	"fmt"
	"net/http"
)

func (cli Client) StartControllers(conf []byte) error {

	logger.Debugf("POST /controller")

	serverResp, err := cli.R().SetBody(conf).Post("/controller")
	if err != nil {
		return fmt.Errorf("error while connecting to daemon: %s", err)
	}

	if serverResp.StatusCode() != http.StatusOK &&
		serverResp.StatusCode() != http.StatusAccepted {
		return processErrorBody(serverResp.Body(), nil)
	}
	return nil
}

func (cli Client) StopController(machineID, controllerID string) error {

	logger.Debugf("DELETE /controller/%q/%q/%q", machineID, controllerID)

	serverResp, err := cli.R().Delete("/controller/" + machineID + "/" + controllerID)
	if err != nil {
		return fmt.Errorf("error while connecting to daemon: %s", err)
	}

	if serverResp.StatusCode() != http.StatusOK &&
		serverResp.StatusCode() != http.StatusAccepted {
		return processErrorBody(serverResp.Body(), nil)
	}
	return nil
}
