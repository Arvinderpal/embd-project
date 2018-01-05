package client

import (
	"fmt"
	"net/http"
)

func (cli Client) StartDrivers(conf []byte) error {

	logger.Debugf("POST /driver")

	serverResp, err := cli.R().SetBody(conf).Post("/driver")
	if err != nil {
		return fmt.Errorf("error while connecting to daemon: %s", err)
	}

	if serverResp.StatusCode() != http.StatusOK &&
		serverResp.StatusCode() != http.StatusAccepted {
		return processErrorBody(serverResp.Body(), nil)
	}
	return nil
}

func (cli Client) StopDriver(machineID, driverType, driverID string) error {

	logger.Debugf("DELETE /driver/%q/%q/%q", machineID, driverType, driverID)

	serverResp, err := cli.R().Delete("/driver/" + machineID + "/" + driverType + "/" + driverID)
	if err != nil {
		return fmt.Errorf("error while connecting to daemon: %s", err)
	}

	if serverResp.StatusCode() != http.StatusOK &&
		serverResp.StatusCode() != http.StatusAccepted {
		return processErrorBody(serverResp.Body(), nil)
	}
	return nil
}
