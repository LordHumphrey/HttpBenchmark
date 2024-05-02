package Utils

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"strings"
)

func getHTML(url string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// Set User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Mobile Safari/537.36 Edg/124.0.0.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error("error closing response body: ", err)
		}
	}(resp.Body)

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// parseLinks parses the HTML content and returns all the links.
func parseLinks(htmlContent string) ([]string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	var links []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			var linkAttr string
			switch n.Data {
			case "a", "link":
				linkAttr = "href"
			case "img", "script":
				linkAttr = "src"
			}

			if linkAttr != "" {
				for _, a := range n.Attr {
					if a.Key == linkAttr {
						// Ignore "javascript:;" links and links without "https"
						if a.Val != "javascript:;" && strings.HasPrefix(a.Val, "https") {
							links = append(links, a.Val)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return links, nil
}

// GetAndParseLinks retrieves the HTML content from the given URL and parses all the links.
func GetAndParseLinks(url string) ([]string, error) {
	htmlContent, err := getHTML(url)
	if err != nil {
		return nil, err
	}

	links, err := parseLinks(htmlContent)
	if err != nil {
		return nil, err
	}

	return links, nil
}
