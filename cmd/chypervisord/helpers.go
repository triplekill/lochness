package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mistifyio/lochness"
	"github.com/pborman/uuid"
)

// getHypervisorHelper gets the hypervisor object and handles sending a response
// in case of error
func getHypervisorHelper(hr HTTPResponse, r *http.Request) (*lochness.Hypervisor, bool) {
	ctx := GetContext(r)
	vars := mux.Vars(r)
	hypervisorID, ok := vars["hypervisorID"]
	if !ok {
		hr.JSONMsg(http.StatusBadRequest, "missing hypervisor id")
		return nil, false
	}
	if uuid.Parse(hypervisorID) == nil {
		hr.JSONMsg(http.StatusBadRequest, "invalid hypervisor id")
		return nil, false
	}
	hypervisor, err := ctx.Hypervisor(hypervisorID)
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return nil, false
	}
	return hypervisor, true
}

// saveHypervisorHelper saves the hypervisor object and handles sending a
// response in case of error
func saveHypervisorHelper(hr HTTPResponse, hypervisor *lochness.Hypervisor) bool {
	if err := hypervisor.Validate(); err != nil {
		hr.JSONMsg(http.StatusBadRequest, err.Error())
		return false
	}
	// Save
	if err := hypervisor.Save(); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return false
	}
	return true
}

// decodeHypervisor decodes request body JSON into a hypervisor object
func decodeHypervisor(r *http.Request, hypervisor *lochness.Hypervisor) (*lochness.Hypervisor, error) {
	if hypervisor == nil {
		ctx := GetContext(r)
		hypervisor = ctx.NewHypervisor()
	}

	if err := json.NewDecoder(r.Body).Decode(hypervisor); err != nil {
		return nil, err
	}
	return hypervisor, nil
}
