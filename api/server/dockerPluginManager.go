package server

import (
	"encoding/json"
	"fmt"
	_ "io"
	"net/http"
	_ "os"
	_ "path"

	//	"github.com/libopenstorage/openstorage/volume"
)

const (
	// VolumeDriver is the string returned in the handshake protocol.
	VolumeDriverMgr = "VolumeDriverMgr"
)

// Implementation of the Docker volumes plugin specification.
type driverMgr struct {
	restBase
}

func newVolumePluginMgr(name string) restServer {
	return &driverMgr{restBase{name: name, version: "0.3"}}
}

func (d *driverMgr) String() string {
	return d.name
}
func (d *driverMgr) volNotFound(request string, id string, e error, w http.ResponseWriter) error {
	err := fmt.Errorf("Failed to locate volume: " + e.Error())
	d.logReq(request, id).Warn(http.StatusNotFound, " ", err.Error())
	return err
}

func (d *driverMgr) Routes() []*Route {
	d.logReq("Route", "").Info("Route Set")
	return []*Route{
		&Route{verb: "POST", path: volDriverPath("Create"), fn: d.create},
		&Route{verb: "POST", path: volDriverPath("Remove"), fn: d.remove},
		&Route{verb: "POST", path: volDriverPath("Name"), fn: d.name1},
		&Route{verb: "POST", path: "/Plugin.Activate", fn: d.handshake},
	}
}

func (d *driverMgr) emptyResponse(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(&volumeResponse{})
}

func (d *driverMgr) decode(method string, w http.ResponseWriter, r *http.Request) (*volumeRequest, error) {
	var request volumeRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		e := fmt.Errorf("Unable to decode JSON payload")
		d.sendError(method, "", w, e.Error()+":"+err.Error(), http.StatusBadRequest)
		return nil, e
	}
	d.logReq(method, request.Name).Debug("")
	return &request, nil
}

func (d *driverMgr) handshake(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(&handshakeResp{
		[]string{VolumeDriver},
	})
	if err != nil {
		d.sendError("handshake", "", w, "encode error", http.StatusInternalServerError)
		return
	}
	d.logReq("handshake", "").Debug("Handshake PLugin Mgr completed")
}

func (d *driverMgr) create(w http.ResponseWriter, r *http.Request) {
	method := "create"

	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}

	d.logReq(method, request.Name).Debug("Create for PLugin Mgr")

	json.NewEncoder(w).Encode(&volumeResponse{})
}

func (d *driverMgr) remove(w http.ResponseWriter, r *http.Request) {
	method := "remove"

	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}

	d.logReq(method, request.Name).Debug("Remove")

	json.NewEncoder(w).Encode(&volumeResponse{})
	return
}

func (d *driverMgr) name1(w http.ResponseWriter, r *http.Request) {
	method := "name"

	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}

	d.logReq(method, request.Name).Debug("Name")

	json.NewEncoder(w).Encode(&volumeResponse{})
}
