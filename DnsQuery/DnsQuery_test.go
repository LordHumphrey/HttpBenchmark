package DnsQuery

import (
	"fmt"
	"testing"
	"time"
)

func TestDo_DNS_Query(t *testing.T) {
	flags := QueryDNSFlags{
		Name:             "vm.gtimg.cn",
		Server:           "223.5.5.5",
		Types:            []string{"AAAA"},
		ClientSubnet:     "1.1.1.1/24",
		Timeout:          10 * time.Second,
		Pad:              false,
		HTTPMethod:       "GET",
		ReuseConn:        true,
		RecursionDesired: true,
		Class:            1,
	}

	query, err := DoDnsQuery(flags)
	fmt.Print(query)
	if err != nil {
		return
	}
}
