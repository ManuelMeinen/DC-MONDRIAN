package keyman

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"gateway_tp/config"
	"gateway_tp/crypto"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"golang.org/x/crypto/pbkdf2"

	//"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var logPrefix = "[KeyMan] "

var l0Salt = []byte("L0 Salt value")

// KeyPld is the payload sent to other TPs carrying the key and the key expiration time
type keyPld struct {
	Key []byte
	Ttl time.Time
}

func (k keyPld) MarshalJSON() ([]byte, error) {
	dummy := struct {
		Key []byte
		TTL time.Time
	}{
		k.Key,
		k.Ttl,
	}
	return json.Marshal(dummy)
}

func (k *keyPld) UnmarshalJSON(b []byte) error {
	var dummy struct {
		Key []byte
		TTL time.Time
	}
	err := json.Unmarshal(b, &dummy)
	k.Key = dummy.Key
	k.Ttl = dummy.TTL
	return err
}

// KeyMan implements KeyManager interface
type KeyMan struct {
	keyLength	int
	keyTTL      time.Duration
	ms         	[]byte
	l0         	[]byte
	l0TTL     	time.Time
	l0Lock     	sync.RWMutex
	listenIP   	string
	listenPort 	int
	reqLock    	sync.Mutex
	mac 		hash.Hash
}

// NewKeyMan creates a new Keyman
func NewKeyMan(masterSecret []byte) *KeyMan {
	listenIP := tpAddrToKeyServerAddr(config.TPAddr)
	km := &KeyMan{
		keyLength:      config.KeyLength,
		keyTTL:         config.KeyTTL,
		ms: 			masterSecret,
		listenIP:   	listenIP,
		listenPort: 	config.ServerPort,
	}
	err := km.refreshL0()
	if err!=nil{
		log.Println(logPrefix+"ERROR: Did not refresh L0")
	}
	km.serveL1()
	return km
}

func (km *KeyMan) serveL1(){
	log.Println(logPrefix+"*** Key Server Ready ***")
	addr := km.listenIP+":"+strconv.Itoa(km.listenPort)
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/api/Get-L1-Key", km.L1ReqHandler)
	log.Println(logPrefix+fmt.Sprintf("Listening at: https://%s/", addr))
	go log.Fatal(http.ListenAndServeTLS(addr, config.ServerCert, config.ServerKey, nil))
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// send something to client
	io.WriteString(w, "Hello from the Key Server\n")
	log.Println(logPrefix+"IndexHandler triggered")
	defer r.Body.Close()
	// read the data
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Println(string(buf))
}

// Handle Requests for L1 Keys
func (km *KeyMan)L1ReqHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(logPrefix+"L1ReqHandler triggered")
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	l1_key, err := km.DeriveL1Key(string(buf))
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	encodeAndSend(l1_key, err, w)
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

func (km *KeyMan) getL0Key() ([]byte, time.Time, error) {
	// create new key in case we don't have a key yet or current key has expired
	if km.l0 == nil || km.l0TTL.Before(time.Now()) {
		err := km.refreshL0()
		if err != nil {
			return nil, time.Time{}, err
		}
	}
	km.l0Lock.RLock()
	defer km.l0Lock.RUnlock()
	k := make([]byte, km.keyLength)
	copy(k, km.l0)
	return k, km.l0TTL, nil
}

func (km *KeyMan) refreshL0() error {
	log.Println(logPrefix+"Refresh L0")
	km.l0Lock.Lock()
	defer km.l0Lock.Unlock()
	// check again if key indeed is missing or expired in case multiple goroutines entered the function
	if km.l0 != nil && km.l0TTL.After(time.Now()) {
		log.Println(logPrefix+"Didn't refresh L0")
		return nil
	}

	if len(km.ms) == 0 {
		return errors.New("master secret cannot be empty")
	}
	key := pbkdf2.Key(km.ms, l0Salt, 1000, km.keyLength, sha256.New)
	km.l0 = key
	km.l0TTL = time.Now().Add(km.keyTTL)
	var err error
	km.mac, err = crypto.InitMac(km.l0)
	if err != nil {
		return err
	}
	return nil
}

func (km *KeyMan) DeriveL1Key(remote string) (*keyPld, error) {
	log.Println(logPrefix+"Deriving L1 Key for "+remote)
	_, t, err := km.getL0Key()
	if err != nil {
		return nil, err
	}
	io.WriteString(km.mac, remote)
	sum := km.mac.Sum(nil)
	km.mac.Reset()
	key := &keyPld{
		Key: sum,
		Ttl: t,
	}
	log.Println(key)
	return key, nil
}

func (km *KeyMan) FetchL1FromRemote(remote string) (*keyPld, error) {
	//TODO: this lock is bad because it serializes all requests, not just the ones going to the same destination
	km.reqLock.Lock()
	defer km.reqLock.Unlock()
	//Init insecure connection to remote TP
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	key_server_addr := tpAddrToKeyServerAddr(remote)
	resp, err := client.Post(fmt.Sprintf("https://%s:%s/api/Get-L1-Key", key_server_addr, strconv.Itoa(config.ServerPort)), "text/plain", strings.NewReader(remote))
	if err != nil {
		log.Println(err)
		return nil, errors.New("ERROR")
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	key := &keyPld{}
	err = decoder.Decode(key)
	if err != nil {
		log.Println(err)
		return nil, errors.New("ERROR")
	}
	return key, nil
}

func (km *KeyMan) DeriveL2(l1 []byte, zone uint32) ([]byte, error) {
	log.Println(logPrefix+"Deriving L2 Key for zone "+strconv.Itoa(int(zone)))
	mac, err := crypto.InitMac(l1)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, zone)
	mac.Write(buf[:3])
	return mac.Sum(nil), nil
}

func tpAddrToKeyServerAddr(tpAddr string) string {
	v := strings.Split(tpAddr, ".")
	return "100."+v[1]+"."+v[2]+"."+v[3]
}