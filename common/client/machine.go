package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Arvinderpal/embd-project/pkg/machine"
	"github.com/Arvinderpal/embd-project/pkg/option"
)

// MachineJoin sends a machine POST request with mh to the daemon.
func (cli Client) MachineJoin(mh machine.Machine) error {

	logger.Debugf("POST /machine/%q", mh.MachineID)

	serverResp, err := cli.R().SetBody(mh).Post("/machine/" + mh.MachineID)
	if err != nil {
		return fmt.Errorf("error while connecting to daemon: %s", err)
	}

	if serverResp.StatusCode() != http.StatusCreated {
		return processErrorBody(serverResp.Body(), mh)
	}

	return nil
}

// MachineLeave sends a DELETE request with machineID to the daemon.
func (cli Client) MachineLeave(machineID string) error {

	logger.Debugf("DELETE /machine/%q", machineID)

	serverResp, err := cli.R().Delete("/machine/" + machineID)
	if err != nil {
		return fmt.Errorf("error while connecting to daemon: %s", err)
	}

	if serverResp.StatusCode() != http.StatusNoContent &&
		serverResp.StatusCode() != http.StatusNotFound {
		return processErrorBody(serverResp.Body(), machineID)
	}

	return nil
}

// MachineGet sends a GET request with machineID to the daemon.
func (cli Client) MachineGet(machineID string) (*machine.Machine, error) {

	logger.Debugf("GET /machine/%q", machineID)

	serverResp, err := cli.R().Get("/machine/" + machineID)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to daemon: %s", err)
	}

	if serverResp.StatusCode() != http.StatusOK &&
		serverResp.StatusCode() != http.StatusNoContent {
		return nil, processErrorBody(serverResp.Body(), machineID)
	}

	if serverResp.StatusCode() == http.StatusNoContent {
		return nil, nil
	}

	// logger.Infof("mh []byte: %s", string(serverResp.Body()))
	var mh machine.Machine
	if err := json.Unmarshal(serverResp.Body(), &mh); err != nil {
		return nil, err
	}

	return &mh, nil
}

// MachinesGet sends a GET request to the daemon.
func (cli Client) MachinesGet() ([]machine.Machine, error) {

	serverResp, err := cli.R().Get("/machines")
	if err != nil {
		return nil, fmt.Errorf("error while connecting to daemon: %s", err)
	}

	if serverResp.StatusCode() != http.StatusOK &&
		serverResp.StatusCode() != http.StatusNoContent {
		return nil, processErrorBody(serverResp.Body(), nil)
	}

	if serverResp.StatusCode() == http.StatusNoContent {
		return nil, nil
	}

	var mhs []machine.Machine
	if err := json.Unmarshal(serverResp.Body(), &mhs); err != nil {
		return nil, err
	}

	return mhs, nil
}

// MachineUpdate sends a POST request with machineID and opts to the daemon.
func (cli Client) MachineUpdate(machineID string, opts option.OptionMap) error {

	logger.Debugf("Update: POST /machine/%d", machineID)

	serverResp, err := cli.R().SetBody(opts).Post("/machine/update/" + machineID)
	if err != nil {
		return fmt.Errorf("error while connecting to daemon: %s", err)
	}

	if serverResp.StatusCode() != http.StatusOK &&
		serverResp.StatusCode() != http.StatusAccepted {
		return processErrorBody(serverResp.Body(), machineID)
	}

	return nil
}
