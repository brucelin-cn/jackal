/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package s2s

import (
	"fmt"
	"net"
	"strconv"
	"sync/atomic"

	"time"

	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/transport"
)

var listenerProvider = net.Listen

var (
	srv         *server
	initialized uint32
)

type server struct {
	cfg       *Config
	ln        net.Listener
	stmCnt    int32
	listening uint32
}

func Initialize(cfg *Config) {
	if cfg.Disabled {
		return
	}
	if !atomic.CompareAndSwapUint32(&initialized, 0, 1) {
		return
	}
	srv = &server{cfg: cfg}
	go srv.start()
}

func (s *server) start() {
	bindAddr := s.cfg.Transport.BindAddress
	port := s.cfg.Transport.Port
	address := bindAddr + ":" + strconv.Itoa(port)

	log.Infof("s2s_in: listening at %s", address)

	if err := s.listenConn(address); err != nil {
		log.Fatalf("%v", err)
	}
}

func (s *server) listenConn(address string) error {
	ln, err := listenerProvider("tcp", address)
	if err != nil {
		return err
	}
	s.ln = ln

	atomic.StoreUint32(&s.listening, 1)
	for atomic.LoadUint32(&s.listening) == 1 {
		conn, err := ln.Accept()
		if err == nil {
			keepAlive := time.Second * time.Duration(s.cfg.Transport.KeepAlive)
			go s.startStream(transport.NewSocketTransport(conn, keepAlive))
			continue
		}
	}
	return nil
}

func (s *server) startStream(tr transport.Transport) {
}

func (s *server) nextID() string {
	return fmt.Sprintf("s2s_in:%d", atomic.AddInt32(&s.stmCnt, 1))
}