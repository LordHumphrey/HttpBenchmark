package main

import (
	"HttpBenchmark/Common"
	"HttpBenchmark/DnsQuery"
	"HttpBenchmark/Utils"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

func main() {
	log.Debugf("start...")
	parallelDownloads, httpBaseConfig, downloadHttpConfig := parseArgs()

	for {
		parsedURL, err := url.Parse(downloadHttpConfig.url.String())
		if err != nil {
			log.Println("Error creating new request:", err)
			return
		}
		subNetIpList, getSubNetIpErr := Utils.GetIpSubnetFromEmbedFile(cidrData, *parallelDownloads)
		if getSubNetIpErr != nil {
			log.Errorln("Get Ip from fail")
		}
		for _, subNetIp := range subNetIpList {
			queryRes := doDnsQuery(httpBaseConfig, parsedURL.Host, subNetIp)
			var waitGroup sync.WaitGroup

			tasks := createDownloadTasks(downloadHttpConfig, queryRes, *parallelDownloads)
			go calculateTotalDownloadedAndSpeed(tasks)
			executeDownloadTasks(tasks, &waitGroup)
			waitGroup.Wait()
		}
	}
}

func doDnsQuery(httpBaseConfig *Common.HttpBaseConfig, host, subNetIp string) []*net.IP {
	queryDNSFlags := DnsQuery.NewQueryDNSFlags()
	queryDNSFlags.Name = host
	queryDNSFlags.ClientSubnet = subNetIp
	queryDNSFlags.HttpBaseConfig = *httpBaseConfig
	queryRes, err := DnsQuery.DoDnsQuery(*queryDNSFlags)
	if err != nil {
		log.Error("Error in DoDnsQuery:", err)
	}
	return queryRes
}

func parseArgs() (*int, *Common.HttpBaseConfig, *DownloadHttpConfig) {

	httpBaseConfig := Common.NewHttpBaseConfig()
	downloadHttpConfig := NewDownloadHttpConfig()
	singleIpDownloadTimes := flag.Int("SingleIpDownloadTimes", downloadHttpConfig.SingleIpDownloadTimes, "The number of single ip download times")
	httpMethod := flag.String("httpMethod", httpBaseConfig.HTTPMethod, "The HTTP method to use")
	postBody := flag.String("postBody", downloadHttpConfig.PostBody, "The HTTP post body")
	reuseConn := flag.Bool("reuseConn", httpBaseConfig.ReuseConn, "Whether to reuse the connection")
	timeout := flag.Duration("timeout", httpBaseConfig.Timeout, "The timeout duration")
	referer := flag.String("referer", downloadHttpConfig.Referer, "The HTTP referer")
	xForwardFor := flag.String("xForwardFor", downloadHttpConfig.XForwardFor, "The X-Forwarded-For HTTP header")

	httpBaseConfig.HTTPMethod = *httpMethod
	httpBaseConfig.ReuseConn = *reuseConn
	httpBaseConfig.Timeout = *timeout

	localIP := flag.String("localIP", "", "The local IP to use")
	targetUrl := flag.String("url", "", "The URL to download")
	parallelDownloads := flag.Int("parallel", 16, "The number of parallel downloads")

	flag.Parse()

	if localIP == nil {
		log.Fatalln("Please provide a local IP")
	} else if !isValidLocalIP(*localIP) {
		log.Fatalln("Please provide a valid local IP")
	} else {
		httpBaseConfig.LocalIP = net.ParseIP(*localIP)
		log.Debugf("Local IP: %s", *localIP)
	}
	if *targetUrl == "" || *parallelDownloads <= 0 {
		log.Fatalln("Please provide a local IP, a URL, and a positive number for parallel downloads")
	}
	downloadHttpConfig.url, _ = url.Parse(*targetUrl)
	downloadHttpConfig.PostBody = *postBody
	downloadHttpConfig.Referer = *referer
	downloadHttpConfig.XForwardFor = *xForwardFor
	downloadHttpConfig.SingleIpDownloadTimes = *singleIpDownloadTimes

	downloadHttpConfig.HttpBaseConfig = *httpBaseConfig

	return parallelDownloads, httpBaseConfig, downloadHttpConfig
}

func createDownloadTasks(config *DownloadHttpConfig, queryRes []*net.IP, parallelDownloads int) []*DownloadHttpConfig {
	tasks := make([]*DownloadHttpConfig, parallelDownloads)
	queryResLen := len(queryRes)

	for i := 0; i < parallelDownloads; i++ {
		queryResponseIp := queryRes[i%queryResLen]
		downloadHttp := NewDownloadHttpConfig()
		downloadHttp.Referer = config.url.String()
		downloadHttp.LocalIP = config.LocalIP
		downloadHttp.url = config.url
		downloadHttp.RemoteIP = queryResponseIp
		if config.url.Scheme == "https" {
			downloadHttp.RemotePort = 443
		} else {
			downloadHttp.RemotePort = 80
		}
		tasks[i] = downloadHttp
	}
	return tasks
}

func executeDownloadTasks(tasks []*DownloadHttpConfig, waitGroup *sync.WaitGroup) {
	for _, task := range tasks {
		waitGroup.Add(1)
		go task.DoHttpDownload(waitGroup)
	}
}

func calculateTotalDownloadedAndSpeed(tasks []*DownloadHttpConfig) {
	for {
		totalDownloadSpeed := int64(0)
		totalDownloaded := int64(0)
		for _, task := range tasks {
			totalDownloadSpeed += task.DownloadSpeed
			totalDownloaded += task.TotalDownloadedBytes
		}

		// Convert totalDownloadSpeed from KB/s to Mbps
		totalDownloadSpeedInMbps := (totalDownloadSpeed * 8) / 1024

		// Clear the console
		clearConsole()

		fmt.Printf("Total download speed: %d Mbps, Total downloaded: %s", totalDownloadSpeedInMbps, Utils.FormatBytes(totalDownloaded))
		time.Sleep(3 * time.Second)
	}
}

func clearConsole() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			log.Error("Error in clearConsole:", err)
		}
	case "linux", "darwin":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			log.Error("Error in clearConsole:", err)
		}
	}
}

func isValidLocalIP(ipStr string) bool {
	// Parse the input string to an IP
	ip := net.ParseIP(ipStr)
	if ip == nil {
		log.Errorf("invalid IP format")
		return false
	}

	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Error("error getting network interfaces")
		return false
	}

	// Check each interface
	for _, i := range interfaces {
		addresses, err := i.Addrs()
		if err != nil {
			log.Error("error getting addresses")
			return false
		}
		// Check each address
		for _, addr := range addresses {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Check if the parsed IP matches the current address
			if ip.Equal(net.ParseIP(ipStr)) {
				return true
			}
		}
	}
	return false
}
