package handler

import (
	"controller/db"
	"controller/types"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

var ApiMap = make(map[string]http.HandlerFunc)

func Init(){
	
	/*** API used by Zone Translation Points ***/
	ApiMap["/api/get-subnets"] = GetSubnetsHandler
	ApiMap["/api/get-transitions"] = GetTransitionsHandler

	/*** API used by admin frontend ***/
	/*** READ ***/
	ApiMap["/"] = IndexHandler
	ApiMap["/api/get-all-sites"] = GetAllSitesHandler
	ApiMap["/api/get-all-zones"] = GetAllZonesHandler
	ApiMap["/api/get-all-subnets"] = GetAllSubnetsHandler
	ApiMap["/api/get-all-transitions"] = GetAllTransitionsHandler

	/*** Insert ***/
	ApiMap["/api/insert-sites"] = InsertSitesHandler
	ApiMap["/api/insert-zones"] = InsertZonesHandler
	ApiMap["/api/insert-subnets"] = InsertSubnetsHandler
	ApiMap["/api/insert-transitions"] = InsertTransitionsHandler
	
	/*** Delete ***/
	ApiMap["/api/delete-sites"] = DeleteSitesHandler
	ApiMap["/api/delete-zones"] = DeleteZonesHandler
	ApiMap["/api/delete-subnets"] = DeleteSubnetsHandler
	ApiMap["/api/delete-all-transitions"] = DeleteTransitionsHandler
	ApiMap["/api/delete-transitions"] = DeleteTransitionsHandler
		
	fmt.Println("Handlers initialized")
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// send something to client
	io.WriteString(w, "Hello from the Controller\n")
	defer r.Body.Close()
	// read the data
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Println(string(buf))
}

/*** GET Handlers (Read) ***/

// GetAllSitesHandler returns all sites information to the client
func GetAllSitesHandler(w http.ResponseWriter, r *http.Request) {
	sites, err := db.DB.GetAllSites()
	encodeAndSend(sites, err, w)
}

// GetAllZonesHandler returns all sites information to the client
func GetAllZonesHandler(w http.ResponseWriter, r *http.Request) {
	zones, err := db.DB.GetAllZones()
	encodeAndSend(zones, err, w)
}

// GetAllSubnetsHandler returns all subnet information to the client
func GetAllSubnetsHandler(w http.ResponseWriter, r *http.Request) {
	nets, err := db.DB.GetAllSubnets()
	encodeAndSend(&nets, err, w)
}

// GetAllTransitionsHandler returns all transition information to the client
func GetAllTransitionsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: the shttp package is very basic and does not allow to set the local address
	// therefore the handler sees the default ISD-AS,127.0.0.1 address as remote. The public IP of the TP is therfore sent
	// in the body. This should be checked to match the certificate
	transitions, err := db.DB.GetAllTransitions()
	encodeAndSend(transitions, err, w)
}

// GetSubnetsHandler returns the subnet information for a given TP to the client
func GetSubnetsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: the shttp package is very basic and does not allow to set the local address
	// therefore the handler sees the default ISD-AS,127.0.0.1 address as remote. The public IP of the TP is therfore sent
	// in the body. This should be checked to match the certificate
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	nets, err := db.DB.GetSubnets(string(buf))
	encodeAndSend(nets, err, w)
}

// GetTransitionsHandler returns all transition information to the client
func GetTransitionsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: the shttp package is very basic and does not allow to set the local address
	// therefore the handler sees the default ISD-AS,127.0.0.1 address as remote. The public IP of the TP is therfore sent
	// in the body. This should be checked to match the certificate
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	transitions, err := db.DB.GetTransitions(string(buf))
	encodeAndSend(transitions, err, w)
}

func encodeAndSend(data interface{}, err error, w http.ResponseWriter) {
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ") //TODO: remove after testing
	enc.Encode(data)
	
}

/*** POST Handlers (Insert) ***/

// InsertSitesHandler inserts the given sites into the backend
func InsertSitesHandler(w http.ResponseWriter, r *http.Request) {
	// decode body into site
	var sites types.Sites
	err := json.NewDecoder(r.Body).Decode(&sites)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, err.Error())
		return
	}
	err = db.DB.InsertSites(sites)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// InsertZonesHandler inserts the given zones into the backend
func InsertZonesHandler(w http.ResponseWriter, r *http.Request) {
	// decode body into site
	var zones types.Zones
	err := json.NewDecoder(r.Body).Decode(&zones)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, err.Error())
		return
	}
	err = db.DB.InsertZones(zones)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// InsertSubnetsHandler inserts the given subnets into the backend
func InsertSubnetsHandler(w http.ResponseWriter, r *http.Request) {
	// decode body into site
	var subnets types.Subnets
	err := json.NewDecoder(r.Body).Decode(&subnets)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, err.Error())
		return
	}
	err = db.DB.InsertSubnets(subnets)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// InsertTransitionsHandler inserts the given transitions into the backend
func InsertTransitionsHandler(w http.ResponseWriter, r *http.Request) {
	// decode body into transitions
	var transitions types.Transitions
	err := json.NewDecoder(r.Body).Decode(&transitions)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, err.Error())
		return
	}
	err = db.DB.InsertTransitions(transitions)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

/*** Delete Handlers (Delete) ***/

// DeleteSitesHandler deletes the given sites from the backend
func DeleteSitesHandler(w http.ResponseWriter, r *http.Request) {
	// decode body into sites
	var sites types.Sites
	err := json.NewDecoder(r.Body).Decode(&sites)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, err.Error())
		return
	}
	err = db.DB.DeleteSites(sites)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteZonesHandler deletes the given zones from the backend
func DeleteZonesHandler(w http.ResponseWriter, r *http.Request) {
	// decode body into zones
	var zones types.Zones
	err := json.NewDecoder(r.Body).Decode(&zones)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, err.Error())
		return
	}
	err = db.DB.DeleteZones(zones)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteSubnetsHandler deletes the given subnets from the backend
func DeleteSubnetsHandler(w http.ResponseWriter, r *http.Request) {
	// decode body into subnets
	var subnets types.Subnets
	err := json.NewDecoder(r.Body).Decode(&subnets)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, err.Error())
		return
	}
	err = db.DB.DeleteSubnets(subnets)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteTransitionsHandler deletes the given transitions from the backend
func DeleteTransitionsHandler(w http.ResponseWriter, r *http.Request) {
	// decode body into transitions
	var transitions types.Transitions
	err := json.NewDecoder(r.Body).Decode(&transitions)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, err.Error())
		return
	}
	err = db.DB.DeleteTransitions(transitions)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
