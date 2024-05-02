package main

import (
	"HttpBenchmark/DnsQuery"
	"HttpBenchmark/Utils"
	"flag"
	"fmt"
	"github.com/AdguardTeam/golibs/log"
	"github.com/sirupsen/logrus"
	"net"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

func main() {
	log.Debug("start...")
	localIPStr, myURL, parallelDownloads, _ := parseArgs() // SingleIpDownloadTimes removed as it's not used
	localIP := net.ParseIP(*localIPStr)

	for {
		parsedURL, err := url.Parse(*myURL)
		if err != nil {
			logrus.Println("Error creating new request:", err)
			return
		}
		subNetIpList, getSubNetIpErr := Utils.GetIpSubnetFromEmbedFile(cidrData, *parallelDownloads)
		if getSubNetIpErr != nil {
			logrus.Errorln("Get Ip from fail")
		}
		for _, subNetIp := range subNetIpList {
			queryDNSFlags := DnsQuery.NewQueryDNSFlags()
			queryDNSFlags.Name = parsedURL.Host
			queryDNSFlags.ClientSubnet = subNetIp

			queryRes := doDnsQuery(queryDNSFlags)
			var waitGroup sync.WaitGroup

			tasks := createDownloadTasks(&localIP, myURL, parallelDownloads, parsedURL, queryRes)
			go calculateTotalDownloadedAndSpeed(tasks)
			executeDownloadTasks(tasks, &waitGroup, *queryDNSFlags)
			waitGroup.Wait()
		}
	}
}

func doDnsQuery(queryDNSFlags *DnsQuery.QueryDNSFlags) []*net.IP {
	queryRes, err := DnsQuery.DoDnsQuery(*queryDNSFlags)
	if err != nil {
		log.Error("Error in DoDnsQuery:", err)
	}
	return queryRes
}

func parseArgs() (*string, *string, *int, *int) {
	localIP := flag.String("localIP", "", "The local IP to use")
	myURL := flag.String("url", "", "The URL to download")
	parallelDownloads := flag.Int("parallel", 16, "The number of parallel downloads")
	singleIpDownloadTimes := flag.Int("SingleIpDownloadTimes", 128, "The number of single ip download times")
	flag.Parse()
	if *localIP == "" || *myURL == "" || *parallelDownloads <= 0 {
		logrus.Fatalln("Please provide a local IP, a URL, and a positive number for parallel downloads")
	} else {
		logrus.Debugf("Local IP: %s", *localIP)
	}
	return localIP, myURL, parallelDownloads, singleIpDownloadTimes
}

func createDownloadTasks(localIP *net.IP, myURL *string, parallelDownloads *int, parsedURL *url.URL, queryRes []*net.IP) []*DownloadHttpConfig {
	tasks := make([]*DownloadHttpConfig, *parallelDownloads)
	queryResLen := len(queryRes)

	for i := 0; i < *parallelDownloads; i++ {
		queryResponseIp := queryRes[i%queryResLen]
		downloadHttp := NewDownloadHttpConfig()
		downloadHttp.Referer = *myURL
		downloadHttp.LocalIP = localIP
		downloadHttp.url = parsedURL
		downloadHttp.RemoteIP = queryResponseIp
		if parsedURL.Scheme == "https" {
			downloadHttp.RemotePort = 443
		} else {
			downloadHttp.RemotePort = 80
		}
		tasks[i] = downloadHttp
	}
	return tasks
}

func executeDownloadTasks(tasks []*DownloadHttpConfig, waitGroup *sync.WaitGroup, queryDNSFlags DnsQuery.QueryDNSFlags) {
	for _, task := range tasks {
		waitGroup.Add(1)
		go task.DoHttpDownload(queryDNSFlags, waitGroup)
	}
}

func calculateTotalDownloadedAndSpeed(tasks []*DownloadHttpConfig) {
	for {
		totalDownloadSpeed := int64(0)
		totalDownloaded := int64(0)
		for _, task := range tasks {
			totalDownloadSpeed += task.DownloadSpeed
			totalDownloaded += task.TotalDownloaded
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
