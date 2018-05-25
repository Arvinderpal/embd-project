package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type TestPluginConf struct {
	CEPin             uint16 `json:"ce-pin"`              // CE pin on device
	CSPin             uint16 `json:"cs-pin"`              // CS pin on device
	Channel           byte   `json:"channel"`             // RF Channel to use [0...255]
	Address           uint16 `json:"address"`             // Node address in Octal
	Master            bool   `json:"master"`              // Is this a master node?
	PollInterval      int    `json:"poll-interval"`       // We poll RF module every poll interval [units: Millisecond]
	HeartbeatInterval int    `json:"heartbeat-interval"`  // Child nodes send periodic heartbeats
	RouterWorkerCount int    `json:"router-worker-count"` // MASTER: number of workers to use per receive queue in the router
}

func main() {
	fmt.Printf("starting testplugin")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	fmt.Printf("dump: %s\n", text)

	conf := &TestPluginConf{}
	err := json.Unmarshal([]byte(text), conf)
	if err != nil {
		fmt.Errorf("%s", err)
		return
	}
	fmt.Printf("Conf: %v \n", conf)
}
