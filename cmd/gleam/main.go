package main

import (
	"os"

	"github.com/mikespook/golib/log"
	"github.com/mikespook/golib/pid"
	"github.com/mikespook/golib/signal"

	"github.com/mikespook/gleam"
)

const (
	CONFIG_FILE = "GLEAM_CONFIG"
)

func main() {
	// prepare the configuration
	config, err := InitConfig()
	if err != nil {
		log.Error(err)
		return
	}

	// make PID file
	if config.Pid != "" {
		p, err := pid.New(config.Pid)
		if err != nil {
			log.Error(err)
			return
		}
		defer p.Close()
	}

	log.Message("Starting...")
	if config.Ca != "" || config.Cert != "" || config.Key != "" {
		log.Messagef("Setting TLS (CA=%s; Cert=%s; Key=%s)...", config.Ca, config.Cert, config.Key)
	}
	g, err := gleam.New(config.Etcd, config.Script, config.Cert, config.Key, config.Ca)
	if err != nil {
		log.Error(err)
		return
	}
	defer g.Close()
	g.ErrHandler = func(err error) {
		log.Error(err)
	}
	log.Messagef("Watching(Name = %s)...", config.Name)
	g.WatchNode(config.Name)
	for _, r := range config.Region {
		log.Messagef("Watching(Region = %s)...", r)
		g.WatchRegion(r)
	}
	go g.Serve()
	// signal handler
	sh := signal.NewHandler()
	sh.Bind(os.Interrupt, func() bool { return true })
	go func() {
		g.Wait()
		if err := signal.Send(os.Getpid(), os.Interrupt); err != nil {
			panic(err)
		}
	}()

	sh.Loop()
	log.Message("Exit!")
}
