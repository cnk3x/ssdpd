# ssdpd
让电脑端能发现设备并显示在网络列表里

## 查看可用参数
```shell
docker run --rm cnk3x/ssdpd /ssdpd -h

# ssdpd - 0.3 - 让电脑端能发现设备并显示在网络列表里
# 
# 命令格式:
#     ssdpd [...参数选项]
# 
# 参数选项:
#         --alive-tick         duration  通告 Alive 时间间隔 (默认: "5m0s")
#     -c, --config             string    配置文件路径, 参数优先级: 文件>启动参数>环境变量
#     -n, --friendly-name      string    友好名称 (默认: "主机名")
#         --interfaces         []string  指定网卡介质
#     -l, --location           string    指定服务描述的的访问地址
#         --manufacturer       string    制造商 (默认: "cnk3x")
#         --manufacturer-url   string    制造商链接 (默认: "https://github.com/cnk3x")
#         --maxage             int       ssdp maxage (默认: "1900")
#         --model-description  string    型号描述 (默认: "SSDPD/NAS/WEB")
#     -m, --model-name         string    型号 (默认: "SSDPD")
#         --model-number       string    型号 (默认: "SSDPD v0")
#         --model-type         string    型号类型 (默认: "NAS")
#         --model-url          string    型号链接 (默认: "https://github.com/cnk3x/ssdpd")
#     -p, --port               int       指定描述访问的端口， 0为随机端口 (默认: "0")
#     -u, --presentation-url   string    设备网页，双击设备默认跳转到该参数指定的地址 (默认: "https://github.com/cnk3x/ssdpd")
#         --serial-number      string    序列号
#         --server             string    设备服务名称 (默认: "cnk3x/ssdpd")
#         --udn                string    唯一识别符
#     -v, --verbose                      是否输出详细日志 (默认: "false")
```

## 使用
```shell

# 第一种： docker
docker run --name ssdpd --net host cnk3x/ssdpd \
    /ssdpd \
    --friendly-name '我的NAS (超级牛)' \
    --presentation-url https://www.qq.com \
    --manufacturer 我制造的 \
    --manufacturer-url http://制造商链接 \
    --model-name 我的型号 \
    --model-url http://型号的链接 \
    --model-number '我的型号 v0.1'

# 第二种： 下载使用（仅 linux amd64）
wget -O ssdpd https://github.com/cnk3x/ssdpd/releases/download/v0.3.0/ssdpd.linux.amd64
chmod +x ssdpd
./ssdpd \
    --friendly-name '我的NAS (超级牛)' \
    --presentation-url https://www.qq.com \
    --manufacturer 我制造的 \
    --manufacturer-url http://制造商链接 \
    --model-name 我的型号 \
    --model-url http://型号的链接 \
    --model-number '我的型号 v0.1'

# 第三种： 编译安装到GOPATh
go install -v github.com/cnk3x/ssdpd/cmd/ssdpd@latest
cd $(go env GOPATH)/bin
./ssdpd \
    --friendly-name '我的NAS (超级牛)' \
    --presentation-url https://www.qq.com \
    --manufacturer 我制造的 \
    --manufacturer-url http://制造商链接 \
    --model-name 我的型号 \
    --model-url http://型号的链接 \
    --model-number '我的型号 v0.1'
```

### 效果
![](./images/ssdpd-1.png)