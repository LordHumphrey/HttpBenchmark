package DnsQuery

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

// HTTP makes a DnsQuery query over HTTP(s)
type HTTP struct {
	DnsHttpConfig
	TLSConfig *tls.Config
	UserAgent string
	Method    string
	NoPMTUd   bool

	conn *http.Client
}

func (h *HTTP) Exchange(m *dns.Msg) (*dns.Msg, error) {
	if h.conn == nil || !h.ReuseConn {
		h.conn = &http.Client{
			Timeout: h.Timeout,
			Transport: &http.Transport{
				TLSClientConfig: h.TLSConfig,
				MaxConnsPerHost: 1,
				MaxIdleConns:    1,
				Proxy:           http.ProxyFromEnvironment,
			},
		}
		//log.Debug("Using HTTP/2")
		//h.conn.Transport = &http2.Transport{
		//	TLSClientConfig: h.TLSConfig,
		//	AllowHTTP:       true,
		//}
	}

	buf, err := m.Pack()
	if err != nil {
		return nil, fmt.Errorf("packing message: %w", err)
	}

	var queryURL string
	var req *http.Request
	switch h.Method {
	case http.MethodGet:
		queryURL = h.Server + "?dns=" + base64.RawURLEncoding.EncodeToString(buf)
		req, err = http.NewRequest(http.MethodGet, queryURL, nil)
		if err != nil {
			return nil, fmt.Errorf("creating http request to %s: %w", queryURL, err)
		}
	case http.MethodPost:
		queryURL = h.Server
		req, err = http.NewRequest(http.MethodPost, queryURL, bytes.NewReader(buf))
		if err != nil {
			return nil, fmt.Errorf("creating http request to %s: %w", queryURL, err)
		}
		req.Header.Set("Content-Type", "application/dns-message")
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", h.Method)
	}

	req.Header.Set("Accept", "application/dns-message")
	if h.UserAgent != "" {
		log.Debugf("Setting User-Agent to %s", h.UserAgent)
		req.Header.Set("User-Agent", h.UserAgent)
	}

	log.Debugf("[http] sending %s request to %s", h.Method, queryURL)
	resp, err := h.conn.Do(req)
	if resp != nil && resp.Body != nil {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Errorf("error closing response body: %s", err)
			}
		}(resp.Body)
	}
	if err != nil {
		return nil, fmt.Errorf("requesting %s: %w", queryURL, err)
	}

	var body []byte
	if resp != nil {
		if resp.Body != nil {
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("reading %s: %w", queryURL, err)
			}
		} else {
			return nil, fmt.Errorf("response body is nil")
		}
	} else {
		return nil, fmt.Errorf("response is nil")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status code %d from %s", resp.StatusCode, queryURL)
	}

	response := dns.Msg{}
	if err := response.Unpack(body); err != nil {
		return nil, fmt.Errorf("unpacking DnsQuery response from %s: %w", queryURL, err)
	}

	return &response, nil
}

func (h *HTTP) Close() error {
	h.conn.CloseIdleConnections()
	return nil
}
