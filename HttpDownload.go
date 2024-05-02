package main

import (
	"HttpBenchmark/Common"
	"HttpBenchmark/Utils"
	"context"
	"crypto/tls"
	_ "embed"
	log "github.com/sirupsen/logrus"
	"strconv"

	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

//go:embed Utils/all_cn_cidr.txt
var cidrData string

type DownloadHttpConfig struct {
	Common.HttpBaseConfig
	url                   *url.URL
	RemoteIP              *net.IP
	RemotePort            int
	PostBody              string
	Referer               string
	XForwardFor           string
	SingleIpDownloadTimes int
	DownloadSpeed         int64
	TotalDownloadedBytes  int64
}
type DownloadHttpConfigOption func(*DownloadHttpConfig)

func WithUrl(url *url.URL) DownloadHttpConfigOption {
	return func(config *DownloadHttpConfig) {
		config.url = url
	}
}

func WithRemoteIP(ip *net.IP) DownloadHttpConfigOption {
	return func(config *DownloadHttpConfig) {
		config.RemoteIP = ip
	}
}

func WithRemotePort(port int) DownloadHttpConfigOption {
	return func(config *DownloadHttpConfig) {
		config.RemotePort = port
	}
}

func WithReferer(referer string) DownloadHttpConfigOption {
	return func(config *DownloadHttpConfig) {
		config.Referer = referer
	}
}

func WithXForwardFor(xForwardFor string) DownloadHttpConfigOption {
	return func(config *DownloadHttpConfig) {
		config.XForwardFor = xForwardFor
	}
}

func WithSingleIpDownloadTimes(times int) DownloadHttpConfigOption {
	return func(config *DownloadHttpConfig) {
		config.SingleIpDownloadTimes = times
	}
}

func NewDownloadHttpConfig(opts ...DownloadHttpConfigOption) *DownloadHttpConfig {
	downloadHttpConfig := &DownloadHttpConfig{
		HttpBaseConfig:        *Common.NewHttpBaseConfig(),
		RemotePort:            443,
		XForwardFor:           Utils.GenerateRandomIPAddress(),
		SingleIpDownloadTimes: 128,
		DownloadSpeed:         0,
		TotalDownloadedBytes:  0,
	}
	for _, opt := range opts {
		opt(downloadHttpConfig)
	}
	return downloadHttpConfig
}

func (downloadHttpConfig *DownloadHttpConfig) DoHttpDownload(wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			go downloadHttpConfig.DoHttpDownload(wg)
			log.Fatalln("Recovered in DoHttpDownload", r)
		}
	}()
	log.Debugf("Download %s started", downloadHttpConfig.RemoteIP.String())
	for i := 0; i < downloadHttpConfig.SingleIpDownloadTimes; i++ {
		log.Debugf("\rDownload times: %d ", i+1)
		startTime := time.Now() // 记录开始时间
		transport := downloadHttpConfig.createTransport()
		request := downloadHttpConfig.createHttpRequest()

		client := downloadHttpConfig.createHttpClient(transport)

		response, err := client.Do(request)

		if err != nil {
			log.Println("Error in client.Do:", err)
			break
		} else {
			if response.StatusCode != 200 {
				// 打印响应主体
				responseBody, _ := io.ReadAll(response.Body)
				log.Errorln("Response body: ", string(responseBody))
				break
			}
		}
		var written int64
		written, err = io.Copy(io.Discard, response.Body)
		if err != nil {
			continue
		} else {
			elapsed := time.Since(startTime) // 计算时间差
			elapsedSeconds := elapsed.Seconds()
			if elapsedSeconds != 0 {
				downloadHttpConfig.DownloadSpeed = (written / 1024) / (int64(elapsedSeconds) + 1)
			} else {
				downloadHttpConfig.DownloadSpeed = 0 // 或者其他默认值
			}
			downloadHttpConfig.TotalDownloadedBytes += written
			log.Debugf("Download %s %d bytes,took %ss", downloadHttpConfig.RemoteIP.String(), written, elapsed.String())
		}
		err = response.Body.Close()
		if err != nil {
			log.Errorln("Error in Body.Close: %s", err)
			continue
		}
	}
	wg.Done()
	log.Infof("Download %s done", downloadHttpConfig.RemoteIP.String())
}

func (downloadHttpConfig *DownloadHttpConfig) createHttpClient(transport *http.Transport) *http.Client {
	client := &http.Client{
		Transport: transport,
		Timeout:   downloadHttpConfig.Timeout * 2,
	}
	return client
}

func (downloadHttpConfig *DownloadHttpConfig) createHttpRequest() *http.Request {
	var request *http.Request
	var requestErr error
	request, requestErr = http.NewRequest("GET", downloadHttpConfig.url.String(), nil)
	if requestErr != nil {
		log.Fatalf("Error creating new request: %s", requestErr)
		return nil
	} else {
		request.Header.Add("Cookie", Utils.GenerateRRandStringBytesMaskImper(12))
		request.Header.Add("User-Agent", downloadHttpConfig.HttpBaseConfig.HTTPUserAgent)
		request.Header.Add("Referer", downloadHttpConfig.Referer)
		if downloadHttpConfig.XForwardFor != "" {
			request.Header.Add("X-Forwarded-For", downloadHttpConfig.XForwardFor)
			request.Header.Add("X-Real-IP", downloadHttpConfig.XForwardFor)
		}
	}
	return request
}

func (downloadHttpConfig *DownloadHttpConfig) createTransport() *http.Transport {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	if downloadHttpConfig.HttpBaseConfig.LocalIP != nil {
		dialer.LocalAddr = &net.TCPAddr{
			IP: downloadHttpConfig.HttpBaseConfig.LocalIP,
		}
	}

	// Generate a random TLS configuration
	tlsConfig := &tls.Config{InsecureSkipVerify: true}

	// TODO socks5
	transport := &http.Transport{
		Proxy:                 nil,
		TLSClientConfig:       tlsConfig,
		MaxIdleConns:          downloadHttpConfig.SingleIpDownloadTimes + 16,
		MaxIdleConnsPerHost:   downloadHttpConfig.SingleIpDownloadTimes * 2,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     false,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Override the addr with your own remote IP and port
			addr = net.JoinHostPort(downloadHttpConfig.RemoteIP.String(), strconv.Itoa(downloadHttpConfig.RemotePort))
			return dialer.DialContext(ctx, network, addr)
		},
	}
	if downloadHttpConfig.url.Scheme == "https" {
		transport.DialTLSContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Override the addr with your own remote IP and port
			addr = net.JoinHostPort(downloadHttpConfig.RemoteIP.String(), strconv.Itoa(downloadHttpConfig.RemotePort))
			return tls.DialWithDialer(dialer, network, addr, tlsConfig)
		}
	} else {
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Override the addr with your own remote IP and port
			addr = net.JoinHostPort(downloadHttpConfig.RemoteIP.String(), strconv.Itoa(downloadHttpConfig.RemotePort))
			return dialer.DialContext(ctx, network, addr)
		}
	}
	return transport
}
