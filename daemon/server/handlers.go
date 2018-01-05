package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Arvinderpal/embd-project/pkg/machine"
	"github.com/Arvinderpal/embd-project/pkg/option"

	"github.com/gorilla/mux"
)

func (router *Router) driverStart(w http.ResponseWriter, r *http.Request) {
	logger.Debugf("starting driver(s)\n") // REMOVE
	confB, err := ioutil.ReadAll(r.Body)
	if err != nil {
		processServerError(w, r, err)
		return
	}

	if err := router.daemon.StartDrivers(confB); err != nil {
		processServerError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (router *Router) driverStop(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	machineID, exists := vars["machineID"]
	if !exists {
		processServerError(w, r, errors.New("server received empty machine id"))
		return
	}
	driverType, exists := vars["driverType"]
	if !exists {
		processServerError(w, r, errors.New("server received empty driver type"))
		return
	}
	driverID, exists := vars["driverID"]
	if !exists {
		processServerError(w, r, errors.New("server received empty driver id"))
		return
	}
	logger.Debugf("Recieved %s (id: %s) driver stop request on machine %s", driverType, driverID, machineID)

	if err := router.daemon.StopDriver(machineID, driverType, driverID); err != nil {
		processServerError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)

}

func (router *Router) machineCreate(w http.ResponseWriter, r *http.Request) {

	d := json.NewDecoder(r.Body)
	var mh machine.Machine
	if err := d.Decode(&mh); err != nil {
		processServerError(w, r, err)
		return
	}
	logger.Debugf("machineCreate: %q", mh.MachineID)
	if err := router.daemon.MachineJoin(mh); err != nil {
		processServerError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (router *Router) machineDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	machineID, exists := vars["machineID"]
	if !exists {
		processServerError(w, r, errors.New("server received empty machine id"))
		return
	}
	logger.Debugf("machineDelete: %q", machineID)
	if err := router.daemon.MachineLeave(machineID); err != nil {
		processServerError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (router *Router) machineGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	machineID, exists := vars["machineID"]
	if !exists {
		processServerError(w, r, errors.New("server received empty machine ID"))
		return
	}
	logger.Debugf("machineGet: %q", machineID)
	mh, err := router.daemon.MachineGet(machineID)
	if err != nil {
		processServerError(w, r, fmt.Errorf("error while getting machine: %s", err))
		return
	}
	if mh == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err := json.NewEncoder(w).Encode(mh); err != nil {
		processServerError(w, r, err)
		return
	}
	// w.WriteHeader(http.StatusOK)
}

func (router *Router) machineUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	machineID, exists := vars["machineID"]
	if !exists {
		processServerError(w, r, errors.New("server received empty machine id"))
		return
	}
	logger.Debugf("machineUpdate: %s", machineID)
	// var opts option.OptionMap
	// if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
	// 	processServerError(w, r, err)
	// 	return
	// }
	// if err := router.daemon.MachineUpdate(machineID, opts); err != nil {
	// 	processServerError(w, r, err)
	// 	return
	// }
	w.WriteHeader(http.StatusAccepted)
}

func (router *Router) machinesGet(w http.ResponseWriter, r *http.Request) {
	mhs, err := router.daemon.MachinesGet()
	if err != nil {
		processServerError(w, r, fmt.Errorf("error while getting machines: %s", err))
		return
	}
	if mhs == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err := json.NewEncoder(w).Encode(mhs); err != nil {
		processServerError(w, r, err)
		return
	}
	// w.WriteHeader(http.StatusOK)
}

func (router *Router) ping(w http.ResponseWriter, r *http.Request) {
	if resp, err := router.daemon.Ping(); err != nil {
		processServerError(w, r, err)
	} else {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			processServerError(w, r, err)
		}
	}
}

func (router *Router) globalStatus(w http.ResponseWriter, r *http.Request) {
	if resp, err := router.daemon.GlobalStatus(); err != nil {
		processServerError(w, r, err)
	} else {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			processServerError(w, r, err)
		}
	}
}

func (router *Router) update(w http.ResponseWriter, r *http.Request) {
	var opts option.OptionMap
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		processServerError(w, r, err)
		return
	}
	if err := router.daemon.Update(opts); err != nil {
		processServerError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
