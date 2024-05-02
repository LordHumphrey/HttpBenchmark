package DnsQuery

import (
	"HttpBenchmark/Common"
	"github.com/miekg/dns"
	"math/rand"
	"time"
)

type QueryDNSFlags struct {
	Common.HttpBaseConfig
	Name                string   `short:"q" long:"qname" description:"Query name" default:"baidu.com"`
	Server              string   `short:"s" long:"server" description:"DnsQuery server(s)" default:"https://dns.alidns.com/dns-query"`
	Types               []string `short:"t" long:"type" description:"RR type (e.g. A, AAAA, MX, etc.) or type integer" default:"A"`
	DNSSEC              bool     `short:"d" long:"dnssec" description:"Set the DO (DNSSEC OK) bit in the OPT record" default:"false"`
	NSID                bool     `short:"n" long:"nsid" description:"Set EDNS0 NSID opt" default:"false"`
	ClientSubnet        string   `long:"subnet" description:"Set EDNS0 client subnet" default:""`
	Pad                 bool     `long:"pad" description:"Set EDNS0 padding" default:"false"`
	Class               uint16   `short:"C" description:"Set query class (default: IN 0x01)" default:"1"`
	AuthoritativeAnswer bool     `long:"aa" description:"Set AA (Authoritative Answer) flag in query" default:"false"`
	AuthenticData       bool     `long:"ad" description:"Set AD (Authentic Data) flag in query" default:"false"`
	CheckingDisabled    bool     `long:"cd" description:"Set CD (Checking Disabled) flag in query" default:"false"`
	RecursionDesired    bool     `long:"rd" description:"Set RD (Recursion Desired) flag in query (default: true)" default:"true"`
	RecursionAvailable  bool     `long:"ra" description:"Set RA (Recursion Available) flag in query" default:"false"`
	Zero                bool     `long:"z" description:"Set Z (Zero) flag in query" default:"false"`
	Truncated           bool     `long:"t" description:"Set TC (Truncated) flag in query" default:"false"`
	UDPBuffer           uint16   `long:"udp-buffer" description:"Set EDNS0 UDP size in query" default:"1232"`
}

type Transport interface {
	Exchange(*dns.Msg) (*dns.Msg, error)
	Close() error
}

// NewQueryDNSFlags creates a new QueryDNSFlags with default values
func NewQueryDNSFlags() *QueryDNSFlags {
	// IP list
	ipList := []string{
		"123.123.123.123/32",
		"123.123.123.124/32",
		"202.106.0.20/32",
		"202.106.195.68/32",
		"221.5.203.98/32",
		"221.7.92.98/32",
		"210.21.196.62/32",
		"221.5.88.88/32",
		"202.99.160.68/32",
		"202.99.166.42/32",
		"202.102.224.68/32",
		"202.102.227.68/32",
		"202.97.224.69/32",
		"202.97.224.68/32",
		"202.98.0.68/32",
		"202.98.5.68/32",
		"221.6.4.66/32",
		"221.6.4.67/32",
		"58.240.57.33/32",
		"202.99.224.68/32",
		"202.99.224.82/32",
		"202.102.128.68/32",
		"202.102.152.32/32",
		"202.102.134.68/32",
		"202.102.154.32/32",
		"202.99.192.66/32",
		"202.99.192.68/32",
		"221.11.1.67/32",
		"221.11.1.68/32",
		"210.22.70.32/32",
		"210.22.84.31/32",
		"19.6.6.61/32",
		"24.161.87.155/32",
		"202.99.104.68/32",
		"202.99.96.68/32",
		"221.12.1.227/32",
		"221.12.33.227/32",
		"202.96.69.38/32",
		"202.96.64.68/32",
	}
	// Server list
	serverList := []string{
		"https://dns.alidns.com/dns-query",
		"https://223.5.5.5/dns-query",
		"https://223.6.6.6/dns-query",
		"https://doh.pub/dns-query",
		"https://1.12.12.12/dns-query",
		"https://120.53.53.53/dns-query",
		"https://sm2.doh.pub/dns-query",
		"https://doh.360.cn/dns-query",
	}

	randSource := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &QueryDNSFlags{
		HttpBaseConfig:      *Common.NewHttpBaseConfig(),
		Name:                "baidu.com",
		Server:              serverList[randSource.Intn(len(serverList))],
		Types:               []string{"A"},
		DNSSEC:              false,
		NSID:                false,
		ClientSubnet:        ipList[randSource.Intn(len(ipList))], // Randomly select an IP from the list
		Pad:                 false,
		Class:               1,
		AuthoritativeAnswer: false,
		AuthenticData:       false,
		CheckingDisabled:    false,
		RecursionDesired:    true,
		RecursionAvailable:  false,
		Zero:                false,
		Truncated:           false,
		UDPBuffer:           1232,
	}
}
