package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/Arvinderpal/embd-project/common/message"
	"github.com/Arvinderpal/embd-project/common/seguepb"
)

var (
	jsonFile = flag.String("json_file", "/home/awander/go/src/github.com/Arvinderpal/embd-project/scripts/configs/messages/led-toggle.json", "input file in json format")
)

const jsonStr = `{
    "id": {
	    "type": 3, 
	    "subtype": "blah",
	    "version": 1
	},
	"data": {
		"on": true,
		"value": 42
	}
}`

func main() {
	flag.Parse()

	iMsg := jsonWithKnownMessageType()
	// Let's conver to external format so it ready for trasmission
	eMsg, err := message.ConvertToExternalFormat(*iMsg)
	if err != nil {
		panic(err)
		return
	}
	fmt.Printf("%v\n", eMsg)

	iMsg2 := jsonWithUnknownMessageType()

	// Let's conver to external format so it ready for trasmission
	eMsg2, err := message.ConvertToExternalFormat(*iMsg2)
	if err != nil {
		panic(err)
		return
	}

	fmt.Printf("%v\n", eMsg2)

}

func jsonWithKnownMessageType() *message.Message {

	mhFile, err := os.Open(*jsonFile)
	defer mhFile.Close()
	if err != nil {
		panic(fmt.Sprintf("Error opening file: %s", err.Error()))
		return nil
	}
	jsonParser := json.NewDecoder(mhFile)

	iMsg := &message.Message{
		Data: &seguepb.LEDSwitchData{},
	}
	if err = jsonParser.Decode(iMsg); err != nil {
		panic(fmt.Sprintf("Error parsing machine file: %s", err))
		return nil
	}
	fmt.Printf("%v\n", iMsg)
	return iMsg
}

// this approach is useful when we don't know the message type beforehand.
func jsonWithUnknownMessageType() *message.Message {

	mhFile, err := os.Open(*jsonFile)
	defer mhFile.Close()
	if err != nil {
		panic(fmt.Sprintf("Error opening file: %s", err.Error()))
		return nil
	}
	jsonParser := json.NewDecoder(mhFile)

	// iMsg := message.Message{
	// 	Data: &seguepb.LEDSwitchData{},
	// }
	// if err = jsonParser.Decode(&iMsg); err != nil {
	// 	panic(fmt.Sprintf("Error parsing machine file: %s", err))
	// 	return
	// }
	// fmt.Printf("%v\n", iMsg)

	var dataRaw json.RawMessage
	iMsg2 := &message.Message{
		Data: &dataRaw,
	}
	if err = jsonParser.Decode(iMsg2); err != nil {
		panic(fmt.Sprintf("Error parsing machine file: %s", err))
		return nil
	}
	switch iMsg2.ID.Type {
	case seguepb.MessageType_LEDSwitch:
		ledData := &seguepb.LEDSwitchData{}
		if err := json.Unmarshal(dataRaw, ledData); err != nil {
			panic(err)
			return nil
		}
		iMsg2.Data = ledData
		fmt.Printf("%v\n", ledData)
	}
	return iMsg2
}

// Alternatives: (This did not quite work for me)
// JSON to Proto
// var sMsg seguepb.Message
// var dataRaw json.RawMessage
// if err := json.Unmarshal([]byte(jsonStr), &dataRaw); err != nil {
// 	panic(err)
// 	return
// }
// if err = jsonParser.Decode(&dataRaw); err != nil {
// 	panic(fmt.Sprintf("Error parsing json file: %s", err))
// 	return
// }
// err = jsonpb.Unmarshal([]byte(jsonStr), &sMsg)
// if err != nil {
// 	panic(fmt.Sprintf("Error converting JSON to proto: %s", err))
// }
