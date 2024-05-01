package DnsQuery

import (
	"crypto/tls"
	"fmt"
	"github.com/miekg/dns"
	tlsutil "github.com/natesales/q/util/tls"
	log "github.com/sirupsen/logrus"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// createQuery creates a slice of DnsQuery queries
func createQuery(flags QueryDNSFlags, rrTypes []uint16) []dns.Msg {
	var queries []dns.Msg
	for _, qType := range rrTypes {
		req := dns.Msg{}
		// Query for each requested RR type
		req.Id = dns.Id()
		req.Authoritative = flags.AuthoritativeAnswer
		req.AuthenticatedData = flags.AuthenticData
		req.CheckingDisabled = flags.CheckingDisabled
		req.RecursionDesired = flags.RecursionDesired
		req.RecursionAvailable = flags.RecursionAvailable
		req.Zero = flags.Zero
		req.Truncated = flags.Truncated

		if flags.DNSSEC || flags.NSID || flags.Pad || flags.ClientSubnet != "" {
			opt := &dns.OPT{
				Hdr: dns.RR_Header{
					Name:   ".",
					Class:  flags.UDPBuffer,
					Rrtype: dns.TypeOPT,
				},
			}

			if flags.DNSSEC {
				opt.SetDo()
			}

			if flags.NSID {
				opt.Option = append(opt.Option, &dns.EDNS0_NSID{
					Code: dns.EDNS0NSID,
				})
			}

			if flags.Pad {
				paddingOpt := new(dns.EDNS0_PADDING)

				msgLen := req.Len()
				padLen := 128 - msgLen%128

				// Truncate padding to fit in UDP buffer
				if msgLen+padLen > int(opt.UDPSize()) {
					padLen = int(opt.UDPSize()) - msgLen
					if padLen < 0 { // Stop padding
						padLen = 0
					}
				}

				log.Debugf("Padding with %d bytes", padLen)
				paddingOpt.Padding = make([]byte, padLen)
				opt.Option = append(opt.Option, paddingOpt)
			}

			if flags.ClientSubnet != "" {
				ip, ipNet, err := net.ParseCIDR(flags.ClientSubnet)
				if err != nil {
					log.Fatalf("parsing subnet %s", flags.ClientSubnet)
				}
				var mask int
				if ipNet != nil {
					mask, _ = ipNet.Mask.Size()
				}
				log.Debugf("EDNS0 client subnet %s/%d", ip, mask)

				ednsSubnet := &dns.EDNS0_SUBNET{
					Code:          dns.EDNS0SUBNET,
					Address:       ip,
					Family:        1, // IPv4
					SourceNetmask: uint8(mask),
				}

				if ednsSubnet.Address.To4() == nil {
					ednsSubnet.Family = 2 // IPv6
				}
				opt.Option = append(opt.Option, ednsSubnet)
			}

			req.Extra = append(req.Extra, opt)
		}

		req.Question = []dns.Question{{
			Name:   dns.Fqdn(flags.Name),
			Qtype:  qType,
			Qclass: flags.Class,
		}}

		queries = append(queries, req)
	}
	return queries
}

// newTransport creates a new transport based on local options
func newTransport(flags QueryDNSFlags, tlsConfig *tls.Config) (*Transport, error) {
	var ts Transport

	common := DnsHttpConfig{
		Server:    flags.Server,
		ReuseConn: flags.ReuseConn,
		Timeout:   flags.Timeout,
	}

	log.Debugf("Using HTTP(s) transport: %s", flags.Server)

	ts = &HTTP{
		DnsHttpConfig: common,
		TLSConfig:     tlsConfig,
		UserAgent:     flags.HTTPUserAgent,
		Method:        flags.HTTPMethod,
		NoPMTUd:       !flags.PMTUD,
	}

	return &ts, nil
}

// parseRRTypes parses a list of RR types in string format ("A", "AAAA", etc.) or integer format (1, 28, etc.)
func parseRRTypes(t []string) ([]uint16, error) {
	rrTypes := make(map[uint16]bool, len(t))
	var rrTypesSlice []uint16
	for _, rrType := range t {
		typeCode, ok := dns.StringToType[strings.ToUpper(rrType)]
		if ok {
			rrTypes[typeCode] = true
			rrTypesSlice = append(rrTypesSlice, typeCode)
		} else {
			typeCode, err := strconv.Atoi(rrType)
			if err != nil {
				return nil, fmt.Errorf("%s is not a valid RR type", rrType)
			}
			log.Debugf("using RR type %d as integer", typeCode)
			rrTypes[uint16(typeCode)] = true
			rrTypesSlice = append(rrTypesSlice, uint16(typeCode))
		}
	}
	return rrTypesSlice, nil
}
func DoDnsQuery(queryDNSFlags QueryDNSFlags) ([]*net.IP, error) {
	queryDNSFlags.Server, _ = parseServer(queryDNSFlags.Server)

	rrTypesSlice, err := parseRRTypes(queryDNSFlags.Types)
	if err != nil {
		log.Fatalf("parsing RR types: %v", err)
	}
	msgLists := createQuery(queryDNSFlags, rrTypesSlice)
	tlsConfig, _ := setTLSConfig(&queryDNSFlags)
	transport, err := newTransport(queryDNSFlags, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating new transport: %v", err)
	}
	var replies []*dns.Msg
	for _, msg := range msgLists {
		response, err := (*transport).Exchange(&msg)
		if err != nil {
			return nil, fmt.Errorf("error exchanging message: %v", err)
		}
		replies = append(replies, response)
	}
	var ipResultList []*net.IP
	for _, reply := range replies {
		for _, answer := range reply.Answer {
			switch rr := answer.(type) {
			case *dns.A:
				ipResultList = append(ipResultList, &rr.A)
				log.Debugf("A Record") // Access the A field of the dns.A struct
			case *dns.AAAA:
				ipResultList = append(ipResultList, &rr.AAAA)
				log.Debugf("AAAA Record") // Access the AAAA field of the dns.AAAA struct
			default:
				log.Debugf("Unknown type")
			}
		}
	}
	return ipResultList, nil
}

// parseServer is a revised version of parseServer that uses the URL package for parsing
func parseServer(s string) (string, error) {
	// Remove IPv6 scope ID if present
	var scopeId string
	v6scopeRe := regexp.MustCompile(`\[[a-fA-F0-9:]+%[a-zA-Z0-9]+]`)
	if v6scopeRe.MatchString(s) {
		v6scopeRemoveRe := regexp.MustCompile(`(%[a-zA-Z0-9]+)`)
		matches := v6scopeRemoveRe.FindStringSubmatch(s)
		if len(matches) > 1 {
			scopeId = matches[1]
			s = v6scopeRemoveRe.ReplaceAllString(s, "")
		}
		log.Tracef("Removed IPv6 scope ID %s from server %s", scopeId, s)
	}

	// Check if server starts with a scheme, if not, default to plain
	schemeRe := regexp.MustCompile(`^[a-zA-Z0-9]+://`)
	if !schemeRe.MatchString(s) {
		// Enclose in brackets if IPv6
		v6re := regexp.MustCompile(`^[a-fA-F0-9:]+$`)
		if v6re.MatchString(s) {
			s = "[" + s + "]"
		}
		s = "plain://" + s
	}

	// Parse server as URL
	tu, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("parsing %s as URL: %s", s, err)
	}

	// Set default port
	if tu.Port() == "" {
		if tu.Scheme == "https" {
			setPort(tu, 443)
		} else {
			setPort(tu, 80)
		}

	}

	tu.Path = "/dns-query"

	server := tu.String()

	// Add IPv6 scope ID back to server
	if scopeId != "" {
		server = strings.Replace(server, "]", scopeId+"]", 1)
	}

	return server, nil
}

// setPort sets the port of a url.URL
func setPort(u *url.URL, port int) {
	if strings.Contains(u.Host, ":") {
		if strings.Contains(u.Host, "[") && strings.Contains(u.Host, "]") {
			u.Host = fmt.Sprintf("%s]:%d", strings.Split(u.Host, "]")[0], port)
			return
		}
		u.Host = "[" + u.Host + "]"
	}
	u.Host = fmt.Sprintf("%s:%d", u.Host, port)
}

func setTLSConfig(flags *QueryDNSFlags) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: flags.TLSInsecureSkipVerify,
		ServerName:         flags.TLSServerName,
		MinVersion:         tlsutil.Version(flags.TLSMinVersion, tls.VersionTLS10),
		MaxVersion:         tlsutil.Version(flags.TLSMaxVersion, tls.VersionTLS13),
		NextProtos:         flags.TLSNextProtos,
		CipherSuites:       tlsutil.ParseCipherSuites(flags.TLSCipherSuites),
		CurvePreferences:   tlsutil.ParseCurves(flags.TLSCurvePreferences),
	}

	// TLS client certificate authentication
	if flags.TLSClientCertificate != "" {
		cert, err := tls.LoadX509KeyPair(flags.TLSClientCertificate, flags.TLSClientKey)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// TLS secret logging
	if flags.TLSKeyLogFile != "" {
		log.Warnf("TLS secret logging enabled")
		keyLogFile, err := os.OpenFile(flags.TLSKeyLogFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			return nil, fmt.Errorf("error opening key log file: %v", err)
		}
		tlsConfig.KeyLogWriter = keyLogFile
	}

	return tlsConfig, nil
}
