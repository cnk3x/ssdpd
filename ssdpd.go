package ssdpd

import (
	"context"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/koron/go-ssdp"
)

type Options struct {
	ConfigFile       string        `json:"-" flag:"config,file" short:"c" usage:"配置文件路径, 参数优先级: 文件>启动参数>环境变量"`
	Interfaces       []string      `json:"interfaces" usage:"指定网卡介质"`
	Location         string        `json:"localtion" short:"l" usage:"指定服务描述的的访问地址"`
	Port             int           `json:"port" short:"p" usage:"指定描述访问的端口， 0为随机端口"`
	FriendlyName     string        `json:"friendly_name" short:"n" usage:"友好名称"`
	Manufacturer     string        `json:"manufacturer" usage:"制造商"`
	ManufacturerURL  string        `json:"manufacturer_url" usage:"制造商链接"`
	ModelName        string        `json:"model_name" short:"m" usage:"型号"`
	ModelNumber      string        `json:"model_number" usage:"型号"`
	ModelURL         string        `json:"model_url" usage:"型号链接"`
	ModelDescription string        `json:"model_description" usage:"型号描述"`
	ModelType        string        `json:"model_type" usage:"型号类型"`
	SerialNumber     string        `json:"serial_number" usage:"序列号"`
	UDN              string        `json:"udn" usage:"唯一识别符"`
	PresentationURL  string        `json:"presentation_url" short:"u" usage:"设备网页，双击设备默认跳转到该参数指定的地址"`
	Server           string        `json:"server" usage:"设备服务名称"`
	Verbose          bool          `json:"verbose" short:"v" usage:"是否输出详细日志"`
	AliveTick        time.Duration `json:"alive_tick" usage:"通告 Alive 时间间隔"`
	MaxAge           int           `json:"maxage" flag:"maxage" usage:"ssdp maxage"`
}

func (sOpts Options) Fill() Options {
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

	return sOpts
}

// 发送局域网SSDP广播, 让电脑端能发现设备并显示在网络列表里。
func AdvertiseDevice(ctx context.Context, sOpts Options) (err error) {
	sOpts = sOpts.Fill()

	if sOpts.Port <= 1024 {
		l, le := net.Listen("tcp", ":0")
		if le != nil {
			sOpts.Port = 20261
		} else {
			sOpts.Port = l.Addr().(*net.TCPAddr).Port
			l.Close()
		}
	}

	ni, p4 := findNI(sOpts.Interfaces...)
	if ni != nil {
		if sOpts.Location == "" {
			sOpts.Location = fmt.Sprintf("%s:%d", p4.String(), sOpts.Port)
		}

		ssdp.Interfaces = []net.Interface{*ni}
		sOpts.Interfaces = []string{ni.Name}

		if sOpts.SerialNumber == "" {
			sOpts.SerialNumber = strings.ReplaceAll(ni.HardwareAddr.String(), ":", "")
		}
	}

	if sOpts.SerialNumber == "" {
		id := uuid.New()
		sOpts.SerialNumber = hex.EncodeToString(id[10:])
	}

	if sOpts.UDN == "" {
		sOpts.UDN = uuid.New().URN()
	}

	location := sOpts.Location
	if location != "" {
		if !strings.HasPrefix(location, "http") && !strings.Contains(location, "://") {
			location = "http://" + location
		}
	}

	if sOpts.Verbose {
		ssdp.Logger = log.Default()
	}

	go http.ListenAndServe(fmt.Sprintf(":%d", sOpts.Port), handleDesc(sOpts))
	if ni != nil {
		log.Printf("device desc xml location: http://%s:%d", p4.String(), sOpts.Port)
	} else {
		log.Printf("device desc xml location: %s", location)
	}

	var ad *ssdp.Advertiser
	if ad, err = ssdp.Advertise("upnp:rootdevice", fmt.Sprintf("%s::upnp:rootdevice", sOpts.UDN), location, sOpts.Server, sOpts.MaxAge); err != nil {
		return
	}

	defer ad.Close()
	defer ad.Bye()

	aliveTick := time.NewTicker(sOpts.AliveTick)
	defer aliveTick.Stop()
	for {
		select {
		case <-aliveTick.C:
			ad.Alive()
		case <-ctx.Done():
			return
		}
	}
}

// 设备描述接口服务
func handleDesc(opts Options) http.HandlerFunc {
	type DeviceDesc struct {
		DeviceType       string `xml:"deviceType,omitempty"`
		FriendlyName     string `xml:"friendlyName,omitempty"`
		Manufacturer     string `xml:"manufacturer,omitempty"`
		ManufacturerURL  string `xml:"manufacturerURL,omitempty"`
		ModelName        string `xml:"modelName,omitempty"`
		ModelNumber      string `xml:"modelNumber,omitempty"`
		ModelURL         string `xml:"modelURL,omitempty"`
		ModelType        string `xml:"modelType,omitempty"`
		ModelDescription string `xml:"modelDescription,omitempty"`
		SerialNumber     string `xml:"serialNumber,omitempty"`
		UDN              string `xml:"UDN,omitempty"`
		PresentationURL  string `xml:"presentationURL,omitempty"`
	}

	type SpecVersionDesc struct {
		Major string `xml:"major,omitempty"`
		Minor string `xml:"minor,omitempty"`
	}

	type Desc struct {
		XMLName     xml.Name        `xml:"root"`
		XmlNS       string          `xml:"xmlns,attr"`
		SpecVersion SpecVersionDesc `xml:"specVersion"`
		Device      DeviceDesc      `xml:"device"`
	}

	desXml, _ := xml.Marshal(Desc{
		XmlNS:       "urn:schemas-upnp-org:device-1-0",
		SpecVersion: SpecVersionDesc{Major: "1", Minor: "0"},
		Device: DeviceDesc{
			DeviceType:       "urn:schemas-upnp-org:device:Basic:1",
			FriendlyName:     opts.FriendlyName,
			Manufacturer:     opts.Manufacturer,
			ManufacturerURL:  opts.ManufacturerURL,
			ModelName:        opts.ModelName,
			ModelNumber:      opts.ModelNumber,
			ModelURL:         opts.ModelURL,
			ModelDescription: opts.ModelDescription,
			ModelType:        opts.ModelType,
			SerialNumber:     opts.SerialNumber,
			UDN:              opts.UDN,
			PresentationURL:  opts.PresentationURL,
		},
	})

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(200)
		w.Write(desXml)
	}
}

// 查找合适的IPv4地址
func p4FromAddrs(addrs []net.Addr) net.IP {
	for _, addr := range addrs {
		if in, ok := addr.(*net.IPNet); ok {
			if !in.IP.IsLoopback() && !in.IP.IsUnspecified() {
				if p4 := in.IP.To4(); p4 != nil {
					return p4
				}
			}
		}
	}
	return nil
}

// 通过名称查找可用的网络，找到合适的第一个就返回，如果没有指定名称，则从全部网络中查找
func findNI(names ...string) (ni *net.Interface, p4 net.IP) {
	contains := func(n string) bool {
		if len(names) == 0 || (len(names) == 1 && names[0] == "") {
			return true
		}
		for _, nane := range names {
			if n == nane {
				return true
			}
		}
		return false
	}

	nis, _ := net.Interfaces()
	for _, i := range nis {
		available := contains(i.Name) && i.Flags&(net.FlagRunning|net.FlagUp|net.FlagMulticast|net.FlagBroadcast) != 0 && len(i.HardwareAddr) > 0
		if available {
			addrs, _ := i.Addrs()
			if p := p4FromAddrs(addrs); p != nil && !p.IsUnspecified() && p[0] != 172 {
				ni, p4 = &i, p
				return
			}
		}
	}
	return
}
