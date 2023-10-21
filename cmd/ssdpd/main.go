package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/cnk3x/ssdpd"
	"github.com/cnk3x/ssdpd/cmd/ssdpd/flags"
	"github.com/valyala/fasttemplate"
)

var (
	VERSION     = "0.3.1"
	DESCRIPTION = "让电脑端能发现设备并显示在网络列表里"
	ENV_PREFIX  = "SSDPD_"
)

func main() {
	sOpts := ssdpd.Options{}

	var fOpts = &flags.Options{EnvPrefix: ENV_PREFIX, Version: VERSION, Description: DESCRIPTION}
	if err := flags.Bind(&sOpts, fOpts); err != nil {
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	repl := NewRepl()
	sOpts.Location = repl(sOpts.Location)
	sOpts.FriendlyName = repl(sOpts.FriendlyName)
	sOpts.Manufacturer = repl(sOpts.Manufacturer)
	sOpts.ManufacturerURL = repl(sOpts.ManufacturerURL)
	sOpts.ModelName = repl(sOpts.ModelName)
	sOpts.ModelNumber = repl(sOpts.ModelNumber)
	sOpts.ModelURL = repl(sOpts.ModelURL)
	sOpts.ModelDescription = repl(sOpts.ModelDescription)
	sOpts.ModelType = repl(sOpts.ModelType)
	sOpts.PresentationURL = repl(sOpts.PresentationURL)
	sOpts.Server = repl(sOpts.Server)

	if sOpts.FriendlyName == "" {
		host, _ := os.Hostname()
		sOpts.FriendlyName = host
	}

	if sOpts.AliveTick == 0 {
		sOpts.AliveTick = 300 * time.Second
	}

	//最少间隔10秒
	if sOpts.AliveTick < 10*time.Second {
		sOpts.AliveTick = 10 * time.Second
	}

	if sOpts.Manufacturer == "" {
		sOpts.Manufacturer = "cnk3x"
	}

	if sOpts.ManufacturerURL == "" {
		sOpts.ManufacturerURL = "https://github.com/cnk3x"
	}

	if sOpts.ModelName == "" {
		sOpts.ModelName = "SSDPD"
	}

	if sOpts.ModelNumber == "" {
		sOpts.ModelNumber = "SSDPD v0"
	}

	if sOpts.ModelURL == "" {
		sOpts.ModelURL = "https://github.com/cnk3x/ssdpd"
	}

	if sOpts.ModelType == "" {
		sOpts.ModelType = "NAS"
	}

	if sOpts.ModelDescription == "" {
		sOpts.ModelDescription = "SSDPD/NAS/WEB"
	}

	if sOpts.MaxAge <= 0 {
		sOpts.MaxAge = 1900
	}

	if sOpts.Server == "" {
		sOpts.Server = "cnk3x/ssdpd"
	}

	if err := ssdpd.AdvertiseDevice(ctx, sOpts); err != nil {
		log.Printf("停止服务: %v", err)
	}
}

func NewRepl() func(string) string {
	mu := &sync.Mutex{}
	data := map[string]string{}

	findTag := func(tag string) []byte {
		mu.Lock()
		defer mu.Unlock()
		val, ok := data[tag]
		if !ok {
			switch tag {
			case "host":
				val, _ = os.Hostname()
			case "ip":
				addrs, _ := net.InterfaceAddrs()
				val = ssdpd.FromAddrs4(addrs).String()
			}
			data[tag] = val
		}
		return []byte(val)
	}

	replTag := func(w io.Writer, tag string) (int, error) { return w.Write(findTag(tag)) }

	return func(s string) string { return fasttemplate.ExecuteFuncString(s, "{", "}", replTag) }
}
