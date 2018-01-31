package pipeline

import (
	"encoding/json"
	"fmt"

	"github.com/Arvinderpal/matra/common/programapi"
	"github.com/Arvinderpal/matra/pkg/programs"
)

// UnmarshalProgramConfs will unmarshall program confs from JSON format.
// func UnmarshalProgramConfs(confB []byte) ([]programapi.ProgramConf, error) {
// 	var returnPConfs []programapi.ProgramConf

// 	// We first unmarshall into a PipelineConfEnvelope to get pipeline fields
// 	// such as id, hook type and container, and also to determine the exact
// 	// number of programs specified.
// 	env := PipelineConfEnvelope{}
// 	// if err := json.NewDecoder(r).Decode(&env); err != nil {
// 	if err := json.Unmarshal(confB, &env); err != nil {
// 		return nil, err
// 	}
// 	// The pipeline contains 1 or more program configurations.
// 	for i, pcEnv := range env.Confs {
// 		// This is a little trick to make unmarshalling of the ProgramConf a
// 		// little easier. There may be an alternative (easier) way...
// 		var jc json.RawMessage
// 		rawPCEnv := programapi.ProgramConfEnvelope{
// 			Conf: &jc,
// 		}
// 		// ProgramConfEnvelope (pcEnv) is type Conf:map[string]interface{}.
// 		// In order to Unmarshall, we have to put back into a json []byte.
// 		bytes, err := json.Marshal(pcEnv)
// 		if err != nil {
// 			return nil, fmt.Errorf("Marshal of program conf %d (into []byte) failed: %s\n", i, err)
// 		}
// 		if err := json.Unmarshal(bytes, &rawPCEnv); err != nil {
// 			return nil, fmt.Errorf("Unmarshal of program conf %d (raw) failed: %s", i, err)
// 		}

// 		pConf, err := programs.NewProgramConf(pcEnv.Type, env.Type, env.ID, env.Container)
// 		if err != nil {
// 			return nil, err
// 		}

// 		if err := json.Unmarshal(jc, &pConf); err != nil {
// 			return nil, fmt.Errorf("Unmarshal of %s failed: %s", pcEnv.Type, err)
// 		}
// 		err = pConf.ValidateConf()
// 		if err != nil {
// 			return nil, err
// 		}
// 		returnPConfs = append(returnPConfs, pConf)
// 	}
// 	return returnPConfs, nil
// }

func NewProgramConfs(env PipelineConfEnvelope, epNInfo *programapi.EndpointNetworkInfo) ([]programapi.ProgramConf, error) {
	var returnPConfs []programapi.ProgramConf

	// The pipeline contains 1 or more program configurations.
	for i, pcEnv := range env.Confs {
		// This is a little trick to make unmarshalling of the ProgramConf a
		// little easier. There may be an alternative (easier) way...
		var jc json.RawMessage
		rawPCEnv := programapi.ProgramConfEnvelope{
			Conf: &jc,
		}
		// ProgramConfEnvelope (pcEnv) is type Conf:map[string]interface{}.
		// In order to Unmarshall, we have to put back into a json []byte.
		bytes, err := json.Marshal(pcEnv)
		if err != nil {
			return nil, fmt.Errorf("Marshal of program conf %d (into []byte) failed: %s\n", i, err)
		}
		if err := json.Unmarshal(bytes, &rawPCEnv); err != nil {
			return nil, fmt.Errorf("Unmarshal of program conf %d (raw) failed: %s", i, err)
		}

		pConf, err := programs.NewProgramConf(pcEnv.Type, env.Type, env.ID, env.Container, epNInfo)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(jc, &pConf); err != nil {
			return nil, fmt.Errorf("Unmarshal of %s failed: %s", pcEnv.Type, err)
		}
		err = pConf.ValidateConf()
		if err != nil {
			return nil, err
		}
		returnPConfs = append(returnPConfs, pConf)
	}
	return returnPConfs, nil
}
