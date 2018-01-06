package client

import (
	"fmt"
	"net/http"
)

func (cli Client) AttachAdaptors(conf []byte) error {

	logger.Debugf("POST /adaptor")

	serverResp, err := cli.R().SetBody(conf).Post("/adaptor")
	if err != nil {
		return fmt.Errorf("error while connecting to daemon: %s", err)
	}

	if serverResp.StatusCode() != http.StatusOK &&
		serverResp.StatusCode() != http.StatusAccepted {
		return processErrorBody(serverResp.Body(), nil)
	}
	return nil
}

func (cli Client) DetachAdaptor(machineID, adaptorType, adaptorID string) error {

	logger.Debugf("DELETE /adaptor/%q/%q/%q", machineID, adaptorType, adaptorID)

	serverResp, err := cli.R().Delete("/adaptor/" + machineID + "/" + adaptorType + "/" + adaptorID)
	if err != nil {
		return fmt.Errorf("error while connecting to daemon: %s", err)
	}

	if serverResp.StatusCode() != http.StatusOK &&
		serverResp.StatusCode() != http.StatusAccepted {
		return processErrorBody(serverResp.Body(), nil)
	}
	return nil
}
