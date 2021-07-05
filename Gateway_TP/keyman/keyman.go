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
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

var logPrefix = "[KeyMan] "

var l0Salt = []byte("L0 Salt value")

var egressL1Keys map[string]KeyPld

func (km *KeyMan) SetEgressL1Keys(m map[string]KeyPld){
	// Used for microbenchmarking
	egressL1Keys = m
}

var ingressL1Keys map[string]KeyPld


// KeyPld is the payload sent to other TPs carrying the key and the key expiration time
type KeyPld struct {
	Key []byte
	Ttl time.Time
}

func (k KeyPld) MarshalJSON() ([]byte, error) {
	dummy := struct {
		Key []byte
		TTL time.Time
	}{
		k.Key,
		k.Ttl,
	}
	return json.Marshal(dummy)
}

func (k *KeyPld) UnmarshalJSON(b []byte) error {
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
	reqLock		sync.Mutex
	egressLock  sync.Mutex
	ingressLock sync.Mutex
	mac 		hash.Hash
}

// NewKeyMan creates a new Keyman
func NewKeyMan(masterSecret []byte, benchmarking bool) *KeyMan {
	listenIP := tpAddrToKeyServerAddr(config.TPAddr)
	km := &KeyMan{
		keyLength:      config.KeyLength,
		keyTTL:         config.KeyTTL,
		ms: 			masterSecret,
		listenIP:   	listenIP,
		listenPort: 	config.ServerPort,
	}
	err := km.RefreshL0()
	if err!=nil{
		log.Println(logPrefix+"ERROR: Did not refresh L0")
	}	
	egressL1Keys = make(map[string]KeyPld)
	ingressL1Keys = make(map[string]KeyPld)
	if benchmarking==false{
		km.serveL1()
	}
	
	return km
}

func (km *KeyMan) serveL1(){
	log.Println(logPrefix+"*** Key Server Ready ***")
	addr := km.listenIP+":"+strconv.Itoa(km.listenPort)
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/api/Get-L1-Key", km.L1ReqHandler)
	log.Println(logPrefix+fmt.Sprintf("Listening at: https://%s/", addr))
	go func(){
		log.Fatal(http.ListenAndServeTLS(addr, config.ServerCert, config.ServerKey, nil))
	}()
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
	log.Println(logPrefix+string(buf))
}

// Handle Requests for L1 Keys
func (km *KeyMan)L1ReqHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(logPrefix+"L1ReqHandler triggered")
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	km.ingressLock.Lock()
	defer km.ingressLock.Unlock()
	var l1_key *KeyPld
	k, found := ingressL1Keys[string(buf)]
	if found && k.Ttl.After(time.Now()){
		log.Println(logPrefix+"Key in ingressL1Keys and not expired")
		l1_key = &k
	}else{
		log.Println(logPrefix+"Key not in ingressL1Keys or expired")
		l1_key, err = km.DeriveL1Key(string(buf))
		if err!=nil{
			log.Println(logPrefix+"ERROR: Failed to derive L1 Key")
			return
		}
		ingressL1Keys[string(buf)] = *l1_key
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
		err := km.RefreshL0()
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

func (km *KeyMan) GetKey(srcTP string, destTP string, zone uint32)([]byte, error){
	/*
	Use this function to create K_A->B:Z
	*/
	// Figure out if ingress or egress L1 key is needed
	//log.Println(logPrefix+"GetKey(srcTP="+srcTP+", destTP="+destTP+", zone="+strconv.Itoa(int(zone))+")")
	if srcTP == config.TPAddr{
		// Egress Key needed
		k, found := egressL1Keys[destTP]
		if found && k.Ttl.After(time.Now()){
			//log.Println(logPrefix+"Key in egressL1Keys and not expired")
			return km.DeriveL2Key(k.Key, zone)
		}else{
			log.Println(logPrefix+"Key not in egressL1Keys or expired")
			l1_key, err := km.FetchL1FromRemote(destTP)
			if err!=nil{
				log.Println(logPrefix+"ERROR: Failed to fetch L1 key from remote: "+destTP)
				return nil, errors.New("Failed to fetch L1 key from remote: "+destTP)
			}
			km.egressLock.Lock()
			egressL1Keys[destTP] = *l1_key
			km.egressLock.Unlock()
			return km.DeriveL2Key(l1_key.Key, zone)
		}
	}else if destTP ==config.TPAddr{
		// Ingress Key needed
		k, found := ingressL1Keys[srcTP]
		if found && k.Ttl.After(time.Now()){
			//log.Println(logPrefix+"Key in ingressL1Keys and not expired")
			return km.DeriveL2Key(k.Key, zone)
		}else{
			log.Println(logPrefix+"Key not in ingressL1Keys or expired")
			l1_key, err := km.DeriveL1Key(srcTP)
			if err!=nil{
				log.Println(logPrefix+"ERROR: Failed to derive L1 Key")
				return nil, errors.New("Failed to derive L1 Key")
			}
			ingressL1Keys[srcTP] = *l1_key
			return km.DeriveL2Key(l1_key.Key, zone)
		}
	}else{
		// Not a key this TP can derive
		log.Println(logPrefix+"Src nor Dest are this TP --> can'g derive L2 Key")
		return nil, errors.New("Src nor Dest are this TP --> can'g derive L2 Key")
	}
}

func (km *KeyMan) RefreshL0() error {
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

func (km *KeyMan) DeriveL1Key(remote string) (*KeyPld, error) {
	log.Println(logPrefix+"Deriving L1 Key for "+remote)
	_, t, err := km.getL0Key()
	if err != nil {
		return nil, err
	}
	io.WriteString(km.mac, remote)
	sum := km.mac.Sum(nil)
	km.mac.Reset()
	key := &KeyPld{
		Key: sum,
		Ttl: t,
	}
	log.Println(logPrefix+"L1 Key:")
	log.Println(key)
	return key, nil
}

func (km *KeyMan) FetchL1FromRemote(remote string) (*KeyPld, error) {
	//TODO: this lock is bad because it serializes all requests, not just the ones going to the same destination
	log.Println(logPrefix+"Fetching L1 Key from "+remote)
	km.reqLock.Lock()
	defer km.reqLock.Unlock()
	//Init insecure connection to remote TP
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	key_server_addr := tpAddrToKeyServerAddr(remote)
	resp, err := client.Post(fmt.Sprintf("https://%s:%s/api/Get-L1-Key", key_server_addr, strconv.Itoa(config.ServerPort)), "text/plain", strings.NewReader(config.TPAddr))
	if err != nil {
		log.Println(err)
		return nil, errors.New("ERROR")
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	key := &KeyPld{}
	err = decoder.Decode(key)
	if err != nil {
		log.Println(err)
		return nil, errors.New("ERROR")
	}
	log.Println(logPrefix+"L1 Key:")
	log.Println(key)
	return key, nil
}

func (km *KeyMan) DeriveL2Key(l1 []byte, zone uint32) ([]byte, error) {
	//log.Println(logPrefix+"Deriving L2 Key for zone "+strconv.Itoa(int(zone)))
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