package Common

import (
	"net"
	"time"
)

type HttpBaseConfig struct {
	LocalIP   net.IP        `long:"local-ip" description:"Local IP address to bind to"`
	ReuseConn bool          `long:"reuse-conn" description:"Reuse connections across queries to the same server (default: true)" default:"true"`
	Timeout   time.Duration `long:"timeout" description:"Query timeout" default:"10s"`
	// HTTP
	HTTPUserAgent string `long:"http-user-agent" description:"HTTP user agent" default:""`
	HTTPMethod    string `long:"http-method" description:"HTTP method" default:"GET"`
	PMTUD         bool   `long:"pmtud" description:"PMTU discovery (default: true)"`

	// TLS parameters
	TLSInsecureSkipVerify bool     `short:"i" long:"tls-insecure-skip-verify" description:"Disable TLS certificate verification"`
	TLSServerName         string   `long:"tls-server-name" description:"TLS server name for host verification"`
	TLSMinVersion         string   `long:"tls-min-version" description:"Minimum TLS version to use" default:"1.0"`
	TLSMaxVersion         string   `long:"tls-max-version" description:"Maximum TLS version to use" default:"1.3"`
	TLSNextProtos         []string `long:"tls-next-protos" description:"TLS next protocols for ALPN"`
	TLSCipherSuites       []string `long:"tls-cipher-suites" description:"TLS cipher suites"`
	TLSCurvePreferences   []string `long:"tls-curve-preferences" description:"TLS curve preferences"`
	TLSClientCertificate  string   `long:"tls-client-cert" description:"TLS client certificate file"`
	TLSClientKey          string   `long:"tls-client-key" description:"TLS client key file"`
	TLSKeyLogFile         string   `long:"tls-key-log-file" env:"SSLKEYLOGFILE" description:"TLS key log file"`
}
type HttpBaseConfigOption func(*HttpBaseConfig)

func WithLocalIP(ip net.IP) HttpBaseConfigOption {
	return func(config *HttpBaseConfig) {
		config.LocalIP = ip
	}
}

func WithReuseConn(reuseConn bool) HttpBaseConfigOption {
	return func(config *HttpBaseConfig) {
		config.ReuseConn = reuseConn
	}
}

func WithTimeout(timeout time.Duration) HttpBaseConfigOption {
	return func(config *HttpBaseConfig) {
		config.Timeout = timeout
	}
}

func WithHTTPUserAgent(userAgent string) HttpBaseConfigOption {
	return func(config *HttpBaseConfig) {
		config.HTTPUserAgent = userAgent
	}
}

func WithHTTPMethod(method string) HttpBaseConfigOption {
	return func(config *HttpBaseConfig) {
		config.HTTPMethod = method
	}
}
func NewHttpBaseConfig(opts ...HttpBaseConfigOption) *HttpBaseConfig {
	httpBaseConfig := &HttpBaseConfig{
		LocalIP:               nil,
		ReuseConn:             true,
		Timeout:               10 * time.Second,
		HTTPUserAgent:         "",
		HTTPMethod:            "GET",
		PMTUD:                 false,
		TLSInsecureSkipVerify: false,
		TLSServerName:         "",
		TLSMinVersion:         "1.0",
		TLSMaxVersion:         "1.3",
		TLSNextProtos:         []string{},
		TLSCipherSuites:       []string{},
		TLSCurvePreferences:   []string{},
		TLSClientCertificate:  "",
		TLSClientKey:          "",
		TLSKeyLogFile:         "",
	}
	for _, opt := range opts {
		opt(httpBaseConfig)
	}
	return httpBaseConfig
}
