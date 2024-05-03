package Utils

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
			case "img", "script", "source":
				linkAttr = "src"
			}

			if linkAttr != "" {
				for _, a := range n.Attr {
					if a.Key == linkAttr {
						// Ignore "javascript:;" links
						if a.Val != "javascript:;" && (strings.HasPrefix(a.Val, "https") || strings.HasPrefix(a.Val, "http") || strings.HasPrefix(a.Val, "//")) {
							// If the link starts with "//", prepend "http:" to make it a valid URL
							if strings.HasPrefix(a.Val, "//") {
								a.Val = "https:" + a.Val
							}
							// Ignore links that contain "gov.cn"
							if !strings.Contains(a.Val, "gov.cn") {
								links = append(links, a.Val)
							}
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

// GetAndParseLinks retrieves the HTML content from the given URL, writes it to a file, and parses all the links.
func GetAndParseLinks(url string) ([]string, error) {
	htmlContent, err := getHTML(url)
	if err != nil {
		return nil, err
	}

	// Write the HTML content to a file
	//err = writeHTMLToFile(htmlContent, url)
	//if err != nil {
	//	log.Error("error writing HTML content to file: ", err)
	//	return nil, err
	//}

	links, err := parseLinks(htmlContent)
	if err != nil {
		return nil, err
	}

	return links, nil
}

func writeHTMLToFile(htmlContent, urlStr string) error {
	// Parse the URL to get the host
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	// Use the host as the filename
	filename := u.Host + ".html"

	// Create the HtmlContent directory if it does not exist
	err = os.MkdirAll("HtmlContent", os.ModePerm)
	if err != nil {
		return err
	}

	// Write the HTML content to a file in the HtmlContent directory
	return os.WriteFile(filepath.Join("HtmlContent", filename), []byte(htmlContent), 0644)
}
