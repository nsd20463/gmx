// +build: windows

package main

import (
	"encoding/json"
	"net"
)

func dial(addr string) (*conn, error) {
	c, err := net.Dial("tcp", "localhost:"+addr)
	return &conn{
		c,
		json.NewDecoder(c),
		json.NewEncoder(c),
	}, err
}
