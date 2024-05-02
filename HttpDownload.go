package main

import (
	"HttpBenchmark/DnsQuery"
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
	LocalIP               *net.IP
	url                   *url.URL
	RemoteIP              *net.IP
	RemotePort            int
	HTTPMethod            string
	PostBody              string
	ReuseConn             bool
	Timeout               time.Duration
	Referer               string
	XForwardFor           string
	SingleIpDownloadTimes int
	DownloadSpeed         int64
	TotalDownloaded       int64
}

func NewDownloadHttpConfig() *DownloadHttpConfig {
	return &DownloadHttpConfig{
		LocalIP:               nil,
		RemotePort:            443,
		ReuseConn:             true,
		HTTPMethod:            "GET",
		Timeout:               10 * time.Second,
		XForwardFor:           Utils.GenerateRandomIPAddress(),
		SingleIpDownloadTimes: 128,
		DownloadSpeed:         0,
		TotalDownloaded:       0,
	}
}

func (dc *DownloadHttpConfig) DoHttpDownload(queryDNSFlags DnsQuery.QueryDNSFlags, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			go dc.DoHttpDownload(queryDNSFlags, wg)
			log.Fatalln("Recovered in DoHttpDownload", r)
		}
	}()
	log.Debugf("Download %s started", dc.RemoteIP.String())
	for i := 0; i < dc.SingleIpDownloadTimes; i++ {
		log.Debugf("\rDownload times: %d ", i+1)
		startTime := time.Now() // 记录开始时间
		transport := dc.createTransport()
		request := dc.createHttpRequest(queryDNSFlags)

		client := dc.createHttpClient(transport)

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
				dc.DownloadSpeed = (written / 1024) / (int64(elapsedSeconds) + 1)
			} else {
				dc.DownloadSpeed = 0 // 或者其他默认值
			}
			dc.TotalDownloaded += written
			log.Debugf("Download %s %d bytes,took %ss", dc.RemoteIP.String(), written, elapsed.String())
		}
		err = response.Body.Close()
		if err != nil {
			log.Errorln("Error in Body.Close: %s", err)
			continue
		}
	}
	wg.Done()
	log.Infof("Download %s done", dc.RemoteIP.String())
}

func (dc *DownloadHttpConfig) createHttpClient(transport *http.Transport) *http.Client {
	client := &http.Client{
		Transport: transport,
		Timeout:   dc.Timeout * 2,
	}
	return client
}

func (dc *DownloadHttpConfig) createHttpRequest(queryDNSFlags DnsQuery.QueryDNSFlags) *http.Request {
	var request *http.Request
	var requestErr error
	request, requestErr = http.NewRequest("GET", dc.url.String(), nil)
	if requestErr != nil {
		log.Fatalf("Error creating new request: %s", requestErr)
		return nil
	} else {
		request.Header.Add("Cookie", Utils.GenerateRRandStringBytesMaskImper(12))
		request.Header.Add("User-Agent", queryDNSFlags.HTTPUserAgent)
		request.Header.Add("Referer", dc.Referer)
		if dc.XForwardFor != "" {
			request.Header.Add("X-Forwarded-For", dc.XForwardFor)
			request.Header.Add("X-Real-IP", dc.XForwardFor)
		}
	}
	return request
}

func (dc *DownloadHttpConfig) createTransport() *http.Transport {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	if dc.LocalIP != nil {
		dialer.LocalAddr = &net.TCPAddr{
			IP: *dc.LocalIP,
		}
	}

	// Generate a random TLS configuration
	tlsConfig := &tls.Config{InsecureSkipVerify: true}

	// TODO socks5
	transport := &http.Transport{
		Proxy:                 nil,
		TLSClientConfig:       tlsConfig,
		MaxIdleConns:          dc.SingleIpDownloadTimes + 16,
		MaxIdleConnsPerHost:   dc.SingleIpDownloadTimes * 2,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     false,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Override the addr with your own remote IP and port
			addr = net.JoinHostPort(dc.RemoteIP.String(), strconv.Itoa(dc.RemotePort))
			return dialer.DialContext(ctx, network, addr)
		},
	}
	if dc.url.Scheme == "https" {
		transport.DialTLSContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Override the addr with your own remote IP and port
			addr = net.JoinHostPort(dc.RemoteIP.String(), strconv.Itoa(dc.RemotePort))
			return tls.DialWithDialer(dialer, network, addr, tlsConfig)
		}
	} else {
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Override the addr with your own remote IP and port
			addr = net.JoinHostPort(dc.RemoteIP.String(), strconv.Itoa(dc.RemotePort))
			return dialer.DialContext(ctx, network, addr)
		}
	}
	return transport
}
