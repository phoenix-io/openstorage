package server

import (
	"encoding/json"
	"fmt"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
	_ "io"
	"net/http"
	_ "os"
	_ "path"
)

// Implementation of the Docker volumes plugin specification.
type volumeManager struct {
	restBase
}

// To keep the context of volumeManager
//type  volumeManagerRequest volumeRequest
type volumeManagerRequest struct {
	Name string
	Opts map[string]string
}

//api.VolumeCreateRequest
type volumeManagerResponse volumeResponse

func newVolumeManagerPlugin(name string) restServer {
	return &volumeManager{restBase{name: name, version: "0.3"}}
}

func (d *volumeManager) String() string {
	return d.name
}

func (d *volumeManager) driverNotFound(request string, id string, e error, w http.ResponseWriter) error {
	err := fmt.Errorf("Failed to locate volume: " + e.Error())
	d.logReq(request, id).Warn(http.StatusNotFound, " ", err.Error())
	return err
}

func (d *volumeManager) Routes() []*Route {
	d.logReq("Route", "").Info("Route Set")
	return []*Route{
		&Route{verb: "POST", path: volDriverPath("Create"), fn: d.Create},
		&Route{verb: "POST", path: volDriverPath("Remove"), fn: d.Remove},
		&Route{verb: "POST", path: volDriverPath("Name"), fn: d.Name},
		&Route{verb: "POST", path: "/Plugin.Activate", fn: d.handshake},
	}
}

func (d *volumeManager) emptyResponse(w http.ResponseWriter) {
	json.NewEncoder(w).Encode(&volumeManagerResponse{})
}

func (d *volumeManager) decode(method string, w http.ResponseWriter, r *http.Request) (*volumeManagerRequest, error) {
	var request volumeManagerRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		e := fmt.Errorf("Unable to decode JSON payload")
		d.sendError(method, "", w, e.Error()+":"+err.Error(), http.StatusBadRequest)
		return nil, e
	}
	d.logReq(method, "").Debug("")
	return &request, nil
}

func (d *volumeManager) handshake(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(&handshakeResp{
		[]string{VolumeDriver},
	})
	if err != nil {
		d.sendError("handshake", "", w, "encode error", http.StatusInternalServerError)
		return
	}
	d.logReq("handshake", "").Debug("Handshake PLugin Mgr completed")
}

func (d *volumeManager) Create(w http.ResponseWriter, r *http.Request) {
	method := "create"

	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}

	d.logReq(method, "").Debug("Create for volumeManager", request)

	driver, err := volume.Get(d.name)
	if err != nil {
		d.driverNotFound(d.name, "", nil, w)
		return
	}

	volLocator := api.VolumeLocator{Name: request.Name}
	volOptions := api.CreateOptions{}
	volSpecs := api.VolumeSpec{}

	ID, err := driver.Create(volLocator, &volOptions, &volSpecs)
	fmt.Println(ID, err)

	json.NewEncoder(w).Encode(&volumeManagerResponse{})
}

func (d *volumeManager) Remove(w http.ResponseWriter, r *http.Request) {
	method := "remove"

	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}

	d.logReq(method, "").Debug("Remove", request)

	json.NewEncoder(w).Encode(&volumeManagerResponse{})
	return
}

func (d *volumeManager) Name(w http.ResponseWriter, r *http.Request) {
	method := "name"

	request, err := d.decode(method, w, r)
	if err != nil {
		return
	}

	d.logReq(method, "").Debug("Name", request)

	json.NewEncoder(w).Encode(&volumeManagerResponse{})
}
