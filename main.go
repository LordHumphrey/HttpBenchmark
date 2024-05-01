package main

import (
	"HttpBenchmark/DnsQuery"
	"HttpBenchmark/Utils"
	"context"
	"crypto/tls"
	_ "embed"
	"flag"
	"fmt"
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
	singleIpDownloadTimes int
}

func NewDownloadHttpConfig() *DownloadHttpConfig {
	return &DownloadHttpConfig{
		LocalIP:               nil,
		RemotePort:            443,
		ReuseConn:             true,
		HTTPMethod:            "GET",
		Timeout:               10 * time.Second,
		XForwardFor:           Utils.GenerateRandomIPAddress(),
		singleIpDownloadTimes: 1024,
	}
}

func DoHttpDownload(queryDNSFlags DnsQuery.QueryDNSFlags, downloadHttpConfig DownloadHttpConfig, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			go DoHttpDownload(queryDNSFlags, downloadHttpConfig, wg)
			log.Fatalln("Recovered in DoHttpDownload", r)
		}
	}()

	for i := 0; i < downloadHttpConfig.singleIpDownloadTimes; i++ {
		dialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}
		if downloadHttpConfig.LocalIP != nil {
			dialer.LocalAddr = &net.TCPAddr{
				IP: *downloadHttpConfig.LocalIP,
			}
		}

		// Generate a random TLS configuration
		tlsConfig := &tls.Config{InsecureSkipVerify: true}

		// TODO socks5
		//socks5Dialer, _ := proxy.SOCKS5("tcp", "172.18.4.85:9000", nil, dialer)
		transport := &http.Transport{
			Proxy:                 nil,
			TLSClientConfig:       tlsConfig,
			MaxIdleConns:          downloadHttpConfig.singleIpDownloadTimes + 16,
			MaxIdleConnsPerHost:   downloadHttpConfig.singleIpDownloadTimes * 2,
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
		var request *http.Request
		var requestErr error
		request, requestErr = http.NewRequest("GET", downloadHttpConfig.url.String(), nil)
		if requestErr != nil {
			log.Fatalf("Error creating new request: %s", requestErr)
			break
		} else {
			request.Header.Add("Cookie", Utils.GenerateRRandStringBytesMaskImper(12))
			request.Header.Add("User-Agent", queryDNSFlags.HTTPUserAgent)
			request.Header.Add("Referer", downloadHttpConfig.Referer)
			if downloadHttpConfig.XForwardFor != "" {
				request.Header.Add("X-Forwarded-For", downloadHttpConfig.XForwardFor)
				request.Header.Add("X-Real-IP", downloadHttpConfig.XForwardFor)
			}
		}

		client := &http.Client{
			Transport: transport,
			Timeout:   downloadHttpConfig.Timeout,
		}

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
		_, err = io.Copy(io.Discard, response.Body)
		if err != nil {
			log.Errorln("Error in io.Copy:", err)
			//continue
		}
		err = response.Body.Close()
		if err != nil {
			log.Errorln("Error in Body.Close: %s", err)
			//continue
		}
	}

	wg.Done()
	log.Debugf("Download %s done", downloadHttpConfig.RemoteIP.String())
}

func main() {
	fmt.Println("start...")
	// Define the command line arguments
	localIP := flag.String("localIP", "", "The local IP to use")
	myURL := flag.String("url", "", "The URL to download")
	parallelDownloads := flag.Int("parallel", 16, "The number of parallel downloads")
	singleIpDownloadTimes := flag.Int("singleIpDownloadTimes", 1024, "The number of single ip download times")
	flag.Parse()
	// Check if the arguments were provided
	if *localIP == "" || *myURL == "" || *parallelDownloads <= 0 {
		log.Fatalln("Please provide a local IP, a URL, and a positive number for parallel downloads")
	}

	// Parse the command line arguments

	log.Debugf("Local IP: %s", *localIP)
	for {
		parsedURL, err := url.Parse(*myURL)
		if err != nil {
			log.Println("Error creating new request:", err)
			return
		}
		subNetIpList, getSubNetIpErr := Utils.GetIpSubnetFromEmbedFile(cidrData, *parallelDownloads)
		if getSubNetIpErr != nil {
			log.Errorln("Get Ip from fail")
		}
		for _, subNetIp := range subNetIpList {
			queryDNSFlags := DnsQuery.NewQueryDNSFlags()
			queryDNSFlags.Name = parsedURL.Host
			queryDNSFlags.ClientSubnet = subNetIp

			queryRes, _ := DnsQuery.DoDnsQuery(*queryDNSFlags)
			for _, queryResponseIp := range queryRes {
				downloadHttp := NewDownloadHttpConfig()
				downloadHttp.Referer = *myURL
				ip := net.ParseIP(*localIP)
				downloadHttp.LocalIP = &ip
				downloadHttp.url = parsedURL
				downloadHttp.RemoteIP = queryResponseIp
				if singleIpDownloadTimes != nil {
					log.Infof("Single IP download times: %d", *singleIpDownloadTimes)
					downloadHttp.singleIpDownloadTimes = *singleIpDownloadTimes
				}
				if parsedURL.Scheme == "https" {
					downloadHttp.RemotePort = 443
				} else {
					downloadHttp.RemotePort = 80
				}
				log.Debugf("Remote IP: %s", queryResponseIp.String())
				var waitGroup sync.WaitGroup
				for i := 0; i < *parallelDownloads; i++ {
					waitGroup.Add(1)
					log.Println("Start Download IP: ", queryResponseIp)
					go DoHttpDownload(*queryDNSFlags, *downloadHttp, &waitGroup)
				}
				waitGroup.Wait()
			}
		}
	}
}
