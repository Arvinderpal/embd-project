package client

import (
	"fmt"
	"net/http"

	"github.com/Arvinderpal/embd-project/pkg/option"
)

// Update sends a SET request to the daemon to update its configuration
func (cli Client) Update(opts option.OptionMap) error {
	serverResp, err := cli.R().SetBody(opts).Post("/update")
	if err != nil {
		return fmt.Errorf("error while connecting to daemon: %s", err)
	}

	if serverResp.StatusCode() != http.StatusOK &&
		serverResp.StatusCode() != http.StatusAccepted {
		return processErrorBody(serverResp.Body(), nil)
	}

	return nil
}
