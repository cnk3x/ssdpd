package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/cnk3x/flags"
	"github.com/cnk3x/ssdpd"
)

var (
	VERSION     = "0.3"
	DESCRIPTION = "让电脑端能发现设备并显示在网络列表里"
	ENV_PREFIX  = "SSDPD_"
)

func main() {
	sOpts := ssdpd.Options{PresentationURL: "https://github.com/cnk3x/ssdpd", Port: 0}.Fill()

	var fOpts = &flags.Options{EnvPrefix: ENV_PREFIX, Version: VERSION, Description: DESCRIPTION}
	if err := flags.Bind(&sOpts, fOpts); err != nil {
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := ssdpd.AdvertiseDevice(ctx, sOpts); err != nil {
		log.Printf("停止服务: %v", err)
	}
}
