package DnsQuery

import (
	"crypto/tls"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func httpTransport() *HTTP {
	return &HTTP{
		DnsHttpConfig: DnsHttpConfig{
			Server:  "https://dns.alidns.com/dns-query",
			Timeout: 2 * time.Second,
		},
		TLSConfig: &tls.Config{},
		UserAgent: "",
		Method:    http.MethodGet,
		NoPMTUd:   false,
	}
}
func validQuery() *dns.Msg {
	msg := dns.Msg{}
	msg.RecursionDesired = true
	msg.Id = dns.Id()
	msg.Question = []dns.Question{{
		Name:   "baidu.com",
		Qtype:  dns.StringToType["A"],
		Qclass: dns.ClassINET,
	}}
	return &msg
}

func TestTransportHTTPPOST(t *testing.T) {
	tp := httpTransport()
	tp.Method = http.MethodPost
	reply, err := tp.Exchange(validQuery())
	assert.Nil(t, err)
	assert.Greater(t, len(reply.Answer), 0)
}
