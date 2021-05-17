package api

import (
	"bytes"
	"controller/types"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"net/http"
)

const controllerAddr = "localhost"
const controllerPort = "4433"

// Connection wraps the http client
type Connection struct {
	Client *http.Client
}

var Conn Connection

// Start a client which doesn't verify the signature
func StartInsecureClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	return client
}

/*** Getters ***/

// Get all sites from the controller
func (c Connection) GetAllSites() types.Sites {
	resp, err := c.Client.Get(fmt.Sprintf("https://%s:%s/api/get-all-sites", controllerAddr, controllerPort))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	sites := types.Sites{}
	err = decoder.Decode(&sites)
	if err != nil {
		panic(err)
	}
	return sites
}

// Get all zones from the controller
func (c Connection) GetAllZones() types.Zones {
	resp, err := c.Client.Get(fmt.Sprintf("https://%s:%s/api/get-all-zones", controllerAddr, controllerPort))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	zones := types.Zones{}
	err = decoder.Decode(&zones)
	if err != nil {
		panic(err)
	}
	return zones
}

// Get all subnets from the controller
func (c Connection) GetAllSubnets() types.Subnets {
	resp, err := c.Client.Get(fmt.Sprintf("https://%s:%s/api/get-all-subnets", controllerAddr, controllerPort))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	subnets := types.Subnets{}
	err = decoder.Decode(&subnets)
	if err != nil {
		panic(err)
	}
	return subnets
}

// Get all transitions from the controller
func (c Connection) GetAllTransitions() types.Transitions {
	resp, err := c.Client.Get(fmt.Sprintf("https://%s:%s/api/get-all-transitions", controllerAddr, controllerPort))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	transitions := types.Transitions{}
	err = decoder.Decode(&transitions)
	if err != nil {
		panic(err)
	}
	return transitions
}

// GetSubnets returns all subnets stored in the backend relevant for ZTP with address tpAddr
func (c Connection) GetSubnets(tpAddr string) types.Subnets {
	resp, err := c.Client.Post(fmt.Sprintf("https://%s:%s/api/get-subnets", controllerAddr, controllerPort), "text/plain", strings.NewReader(tpAddr))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	subnets := types.Subnets{}
	err = decoder.Decode(&subnets)
	if err != nil {
		panic(err)
	}
	return subnets
}

// GetTransitions returns all transitions of a given TP stored in the backend
func (c Connection) GetTransitions(tpAddr string) types.Transitions {
	resp, err := c.Client.Post(fmt.Sprintf("https://%s:%s/api/get-transitions", controllerAddr, controllerPort), "text/plain", strings.NewReader(tpAddr))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	transitions := types.Transitions{}
	err = decoder.Decode(&transitions)
	if err != nil {
		panic(err)
	}
	return transitions
}

/*** Insertions ***/

// InsertSites inserts sites into the Backend
func (c Connection) InsertSites(sites types.Sites) {

	b, err := json.Marshal(sites)
	if err != nil {
		panic(err)
	}

	body := bytes.NewBuffer(b)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("https://%s:%s/api/insert-sites", controllerAddr, controllerPort), body)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := c.Client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response status code: %d, text:\n%s\n", resp.StatusCode, string(b))
}

// InsertZones inserts zones into the Backend
func (c Connection) InsertZones(zones types.Zones) {
	b, err := json.Marshal(zones)
	if err != nil {
		panic(err)
	}

	body := bytes.NewBuffer(b)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("https://%s:%s/api/insert-zones", controllerAddr, controllerPort), body)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := c.Client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response status code: %d, text:\n%s\n", resp.StatusCode, string(b))
}

// InsertSubnets inserts subnets into the Backend
func (c Connection) InsertSubnets(subnets types.Subnets) {
	b, err := json.Marshal(subnets)
	if err != nil {
		panic(err)
	}

	body := bytes.NewBuffer(b)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("https://%s:%s/api/insert-subnets", controllerAddr, controllerPort), body)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := c.Client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response status code: %d, text:\n%s\n", resp.StatusCode, string(b))
}

// InsertTransitions inserts zone transitions into the Backend
func (c Connection) InsertTransitions(transitions types.Transitions) {
	b, err := json.Marshal(transitions)
	if err != nil {
		panic(err)
	}

	body := bytes.NewBuffer(b)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("https://%s:%s/api/insert-transitions", controllerAddr, controllerPort), body)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := c.Client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response status code: %d, text:\n%s\n", resp.StatusCode, string(b))
}

/*** Deletions ***/

// DeleteSites deletes branch sites from the Backend
func (c Connection) DeleteSites(sites types.Sites) {
	b, err := json.Marshal(sites)
	if err != nil {
		panic(err)
	}

	body := bytes.NewBuffer(b)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("https://%s:%s/api/delete-sites", controllerAddr, controllerPort), body)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := c.Client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response status code: %d, text:\n%s\n", resp.StatusCode, string(b))
}

// DeleteZones deletes zones from the Backend
func (c Connection) DeleteZones(zones types.Zones) {
	b, err := json.Marshal(zones)
	if err != nil {
		panic(err)
	}

	body := bytes.NewBuffer(b)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("https://%s:%s/api/delete-zones", controllerAddr, controllerPort), body)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := c.Client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response status code: %d, text:\n%s\n", resp.StatusCode, string(b))
}

// DeleteSubnets delete subnets from the Backend
func (c Connection) DeleteSubnets(subnets types.Subnets) {
	b, err := json.Marshal(subnets)
	if err != nil {
		panic(err)
	}

	body := bytes.NewBuffer(b)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("https://%s:%s/api/delete-subnets", controllerAddr, controllerPort), body)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := c.Client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response status code: %d, text:\n%s\n", resp.StatusCode, string(b))
}

// DeleteTransitions delete zone transitions into the Backend
func (c Connection) DeleteTransitions(transitions types.Transitions) {
	b, err := json.Marshal(transitions)
	if err != nil {
		panic(err)
	}

	body := bytes.NewBuffer(b)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("https://%s:%s/api/delete-transitions", controllerAddr, controllerPort), body)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := c.Client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response status code: %d, text:\n%s\n", resp.StatusCode, string(b))
}
