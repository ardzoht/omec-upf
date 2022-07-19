// SPDX-License-Identifier: Apache-2.0
// Copyright 2022-present Open Networking Foundation

package pfcpiface

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"sync"
	"time"
)

var (
	simulate = simModeDisable
)

func init() {
	flag.Var(&simulate, "simulate", "create|delete|create_continue simulated sessions")
}

type PFCPIface struct {
	conf Conf

	node *PFCPNode
	Dp   Datapath
	Upf  *Upf

	httpSrv      *http.Server
	httpEndpoint string

	uc *UpfCollector
	nc *PfcpNodeCollector

	mu sync.Mutex
}

func NewPFCPIface(conf Conf, dp Datapath) *PFCPIface {
	pfcpIface := &PFCPIface{
		conf: conf,
	}

	pfcpIface.Dp = dp

	httpPort := "8080"
	if conf.CPIface.HTTPPort != "" {
		httpPort = conf.CPIface.HTTPPort
	}

	pfcpIface.httpEndpoint = ":" + httpPort

	pfcpIface.Upf = NewUPF(&conf, pfcpIface.Dp)

	Zap_init()

	return pfcpIface
}

func (p *PFCPIface) mustInit() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.node = NewPFCPNode(p.Upf)
	httpMux := http.NewServeMux()

	setupConfigHandler(httpMux, p.Upf)

	var err error

	p.uc, p.nc, err = setupProm(httpMux, p.Upf, p.node)

	if err != nil {
		log.Fatal("setupProm failed ", err)
	}

	p.httpSrv = &http.Server{Addr: p.httpEndpoint, Handler: httpMux}
}

func (p *PFCPIface) Run() {
	if simulate.enable() {
		p.Upf.sim(simulate, &p.conf.SimInfo)

		if !simulate.keepGoing() {
			return
		}
	}

	p.mustInit()

	go func() {
		if err := p.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("http server failed ", err)
		}

		log.Info("http server closed")
	}()

	// blocking
	p.node.Serve()
}

// Stop sends cancellation signal to main Go routine and waits for shutdown to complete.
func (p *PFCPIface) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	ctxHttpShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := p.httpSrv.Shutdown(ctxHttpShutdown); err != nil {
		log.Error("Failed to shutdown http: ", err)
	}

	p.node.Stop()

	// Wait for PFCP node shutdown
	p.node.Done()
}
