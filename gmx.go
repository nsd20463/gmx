package gmx

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"net"
	"sync"
)

const GMX_VERSION = 0

var (
	r = &registry{
		entries: make(map[string]func() interface{}),
	}

	localsocket net.Listener
)

func init() {
	s, err := localSocket()
	if err != nil {
		log.Printf("gmx: unable to open local socket: %v", err)
		return
	}

	// register the registries keys for discovery
	Publish("keys", func() interface{} {
		return r.keys()
	})
	go serve(s, r)
	localsocket = s
}

// Publish registers the function f with the supplied key.
func Publish(key string, f func() interface{}) {
	r.register(key, f)
}

// Exit cleanly shuts down gmx.
// This is useful as a defer'ed function in main(), so the local gmx socket is cleaned up
func Exit() {
	if localsocket != nil {
		localsocket.Close()
		localsocket = nil
	}
}

func serve(l net.Listener, r *registry) {
	defer l.Close()
	for {
		c, err := l.Accept()
		if err != nil {
			return
		}
		go handle(c, r)
	}
}

func handle(nc net.Conn, reg *registry) {
	// conn makes it easier to send and receive json
	type conn struct {
		net.Conn
		*json.Encoder
		*json.Decoder
	}
	c := conn{
		nc,
		json.NewEncoder(nc),
		json.NewDecoder(nc),
	}
	defer c.Close()
	for {
		var keys []string
		if err := c.Decode(&keys); err != nil {
			if err != io.EOF {
				log.Printf("gmx: client %v sent invalid json request: %v", c.RemoteAddr(), err)
			}
			return
		}
		var result = make(map[string]interface{})
		for _, key := range keys {
			if f, ok := reg.value(key); ok {
				// invoke the function for key and store the result
				val := f()
				if val == nil {
					continue
				}
				switch val.(type) {
				case float64:
					if math.IsNaN(val.(float64)) {
						log.Printf("gmx: Got NaN for %s", key)
						continue
					}
				}
				result[key] = val
			}
		}
		if err := c.Encode(result); err != nil {
			log.Printf("gmx: could not send response to client %v: %v", c.RemoteAddr(), err)
			return
		}
	}
}

type registry struct {
	sync.Mutex // protects entries from concurrent mutation
	entries    map[string]func() interface{}
}

func (r *registry) register(key string, f func() interface{}) {
	r.Lock()
	r.entries[key] = f
	r.Unlock()
}

func (r *registry) value(key string) (func() interface{}, bool) {
	r.Lock()
	f, ok := r.entries[key]
	r.Unlock()
	return f, ok
}

func (r *registry) keys() []string {
	r.Lock()
	var k = make([]string, len(r.entries))
	for e := range r.entries {
		k = append(k, e)
	}
	r.Unlock()
	return k
}
