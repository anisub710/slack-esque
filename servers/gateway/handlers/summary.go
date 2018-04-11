package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

//fmt.errorf if you need to format the error and Errors.new if you don't
// add errors for each helper method

const headerAccessControlAllowOrigin = "Access-Control-Allow-Origin"
const headerContentType = "Content-Type"
const contentTypeJSON = "application/json; utf-8"
const contentTypeHTML = "text/html"

//PreviewImage represents a preview image for a page
type PreviewImage struct {
	URL       string `json:"url,omitempty"`
	SecureURL string `json:"secureURL,omitempty"`
	Type      string `json:"type,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Alt       string `json:"alt,omitempty"`
}

//PreviewVideo represents a preview video for a page
type PreviewVideo struct {
	URL       string `json:"url,omitempty"`
	SecureURL string `json:"secureURL,omitempty"`
	Type      string `json:"type,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
}

//PageSummary represents summary properties for a web page
type PageSummary struct {
	Type        string          `json:"type,omitempty"`
	URL         string          `json:"url,omitempty"`
	Title       string          `json:"title,omitempty"`
	SiteName    string          `json:"siteName,omitempty"`
	Description string          `json:"description,omitempty"`
	Author      string          `json:"author,omitempty"`
	Keywords    []string        `json:"keywords,omitempty"`
	Icon        *PreviewImage   `json:"icon,omitempty"`
	Images      []*PreviewImage `json:"images,omitempty"`
	Videos      []*PreviewVideo `json:"videos,omitempty"`
}

//SummaryHandler handles requests for the page summary API.
//This API expects one query string parameter named `url`,
//which should contain a URL to a web page. It responds with
//a JSON-encoded PageSummary struct containing the page summary
//meta-data.
func SummaryHandler(w http.ResponseWriter, r *http.Request) {
	/*TODO: add code and additional functions to do the following:
	- Add an HTTP header to the response with the name
	 `Access-Control-Allow-Origin` and a value of `*`. This will
	  allow cross-origin AJAX requests to your server.
	- Get the `url` query string parameter value from the request.
	  If not supplied, respond with an http.StatusBadRequest error.
	- Call fetchHTML() to fetch the requested URL. See comments in that
	  function for more details.
	- Call extractSummary() to extract the page summary meta-data,
	  as directed in the assignment. See comments in that function
	  for more details
	- Close the response HTML stream so that you don't leak resources.
	- Finally, respond with a JSON-encoded version of the PageSummary
	  struct. That way the client can easily parse the JSON back into
	  an object. Remember to tell the client that the response content
	  type is JSON.

	Helpful Links:
	https://golang.org/pkg/net/http/#Request.FormValue
	https://golang.org/pkg/net/http/#Error
	https://golang.org/pkg/encoding/json/#NewEncoder
	*/

	w.Header().Add(headerAccessControlAllowOrigin, "*")
	pageURL := r.URL.Query().Get("url")

	if len(pageURL) == 0 {
		http.Error(w, "Missing url query string parameter", http.StatusBadRequest)
		return
	}

	html, err := fetchHTML(pageURL)
	if err != nil {
		http.Error(w, "Error fetching html for URL: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer html.Close()

	summary, err := extractSummary(pageURL, html)

	if err != nil {
		http.Error(w, "Error extracting summary for URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add(headerContentType, contentTypeJSON)

	if err := json.NewEncoder(w).Encode(summary); err != nil {
		http.Error(w, "Error encoding summary to JSON: "+err.Error(), http.StatusInternalServerError)
	}

}

//fetchHTML fetches `pageURL` and returns the body stream or an error.
//Errors are returned if the response status code is an error (>=400),
//or if the content type indicates the URL is not an HTML page.
func fetchHTML(pageURL string) (io.ReadCloser, error) {
	/*TODO: Do an HTTP GET for the page URL. If the response status
	code is >= 400, return a nil stream and an error. If the response
	content type does not indicate that the content is a web page, return
	a nil stream and an error. Otherwise return the response body and
	no (nil) error.

	To test your implementation of this function, run the TestFetchHTML
	test in summary_test.go. You can do that directly in Visual Studio Code,
	or at the command line by running:
		go test -run TestFetchHTML

	Helpful Links:
	https://golang.org/pkg/net/http/#Get
	*/

	response, err := http.Get(pageURL)

	if err != nil {
		return nil, fmt.Errorf("Error while getting url: %v", err)
	}
	contentType := response.Header.Get(headerContentType)

	if !strings.HasPrefix(contentType, contentTypeHTML) {
		return nil, fmt.Errorf("Content type of response is not a web page, it is: %v", contentType)
	}

	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("Error while fetching html from URL %v", response.StatusCode)
	}

	return response.Body, nil
}

//extractSummary tokenizes the `htmlStream` and populates a PageSummary
//struct with the page's summary meta-data.
func extractSummary(pageURL string, htmlStream io.ReadCloser) (*PageSummary, error) {
	/*TODO: tokenize the `htmlStream` and extract the page summary meta-data
	according to the assignment description.

	To test your implementation of this function, run the TestExtractSummary
	test in summary_test.go. You can do that directly in Visual Studio Code,
	or at the command line by running:
		go test -run TestExtractSummary

	Helpful Links:
	https://drstearns.github.io/tutorials/tokenizing/
	http://ogp.me/
	https://developers.facebook.com/docs/reference/opengraph/
	https://golang.org/pkg/net/url/#URL.ResolveReference
	*/

	tokenizer := html.NewTokenizer(htmlStream)
	extracted := map[string]string{}
	var images []*PreviewImage
	var videos []*PreviewVideo
	video := map[string]string{}
	image := map[string]string{}
	imageCount := 0
	videoCount := 0
	for {
		tokenType := tokenizer.Next()

		if tokenType == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error tokenizing HTML: %v", tokenizer.Err())
		}

		if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
			token := tokenizer.Token()
			if token.Data == "head" {
				for {

					tokenType = tokenizer.Next()
					headToken := tokenizer.Token()

					if tokenType == html.EndTagToken && headToken.Data == "head" {
						break
					}

					switch headToken.Data {
					case "meta":
						attr := headToken.Attr
						attrInfo := map[string]string{}
						isImg := false
						isVid := false

						for _, att := range attr {
							switch att.Key {
							case "property":
								if strings.HasPrefix(att.Val, "og:image") {
									isImg = true
									if att.Val == "og:image" {
										if image != nil && imageCount != 0 {
											builtImage, err := makeImages(pageURL, image)
											if err != nil {
												return nil, fmt.Errorf("Error making preview image: %v", err)
											}
											images = append(images, builtImage)
											image = map[string]string{}
										}
										imageCount++
									}
								} else if strings.HasPrefix(att.Val, "og:video") {
									isVid = true
									if att.Val == "og:video" {
										if video != nil && videoCount != 0 {
											builtVideo, err := makeVideos(pageURL, image)
											if err != nil {
												return nil, fmt.Errorf("Error making preview video: %v", err)
											}
											videos = append(videos, builtVideo)
											video = map[string]string{}
										}
										videoCount++
									}
								}
								attrInfo["property"] = att.Val

							case "content":
								attrInfo["content"] = att.Val
							case "name":
								if att.Val == "description" &&
									extracted["og:description"] == "" {
									attrInfo["property"] = "og:description"
								} else if att.Val == "keywords" {
									attrInfo["property"] = att.Val
								} else if att.Val == "author" {
									attrInfo["property"] = "og:author"
								}
							}
						}
						if isImg {
							image[attrInfo["property"]] = attrInfo["content"]
						} else if isVid {
							video[attrInfo["property"]] = attrInfo["content"]
						} else {
							extracted[attrInfo["property"]] = attrInfo["content"]
						}

					case "title":
						if extracted["og:title"] == "" {
							tokenType = tokenizer.Next()
							extracted["og:title"] = tokenizer.Token().Data
						}

					case "link":
						attr := headToken.Attr
						attrInfo := map[string]string{}
						for _, att := range attr {
							switch att.Key {
							case "rel":
								if strings.Contains(att.Val, "icon") {
									attrInfo["property"] = "icon"
								}
							case "href":
								converted, err := convertRelative(pageURL, att.Val)
								if err != nil {
									return nil, fmt.Errorf("Error converting relative URL to absolute URL: %v", err)
								}
								attrInfo["href"] = converted
							case "sizes":
								extracted["icon:sizes"] = att.Val
							case "type":
								extracted["icon:type"] = att.Val
							}

						}

						extracted[attrInfo["property"]] = attrInfo["href"]
					}

				}
			}
		} else if tokenType == html.EndTagToken && tokenizer.Token().Data == "head" {
			break
		}

	}

	if len(image) > 0 {
		builtImage, err := makeImages(pageURL, image)
		if err != nil {
			return nil, fmt.Errorf("Error making preview image: %v", err)
		}
		images = append(images, builtImage)
	}

	if len(video) > 0 {
		builtVideo, err := makeVideos(pageURL, video)
		if err != nil {
			return nil, fmt.Errorf("Error making preview video: %v", err)
		}
		videos = append(videos, builtVideo)
	}

	p, err := constructSummary(pageURL, extracted, images, videos)

	if err != nil {
		return nil, fmt.Errorf("Something wrong with construct summary: %v", err)

	}

	return p, nil
}

//constructSummary constructs pagesummary struct
func constructSummary(pageURL string, extracted map[string]string, images []*PreviewImage, videos []*PreviewVideo) (*PageSummary, error) {
	icon, err := getIcon(extracted)
	if err != nil {
		return nil, fmt.Errorf("Error getting icon: %v", err)
	}
	p := &PageSummary{
		Type:        extracted["og:type"],
		URL:         extracted["og:url"],
		Title:       extracted["og:title"],
		SiteName:    extracted["og:site_name"],
		Description: extracted["og:description"],
		Author:      extracted["og:author"],
		Keywords:    getKeywords(extracted["keywords"]),
		Icon:        icon,
		Images:      images,
		Videos:      videos,
	}

	return p, nil
}

//getKeywords converts a string of keywords into a slice
func getKeywords(keywords string) []string {
	if keywords != "" {
		splitK := strings.Split(keywords, ",")
		cleaned := []string{}
		for _, keyword := range splitK {
			cleaned = append(cleaned, strings.TrimSpace(keyword))
		}
		return cleaned
	}
	return nil
}

//getIcon gets all the relevant information extracted for an icon
// and returns a PreviewImage
func getIcon(extracted map[string]string) (*PreviewImage, error) {
	if extracted["icon"] != "" {
		dim := map[string]int{}
		if extracted["icon:sizes"] != "any" && extracted["icon:sizes"] != "" {
			splitS := strings.Split(extracted["icon:sizes"], "x")
			convertedHeight, err := convToInt(splitS[0])
			convertedWidth, errW := convToInt(splitS[1])
			if err != nil {
				return nil, fmt.Errorf("Error converting string to int: %v", err)
			}
			if errW != nil {
				return nil, fmt.Errorf("Error converting string to int: %v", errW)
			}
			dim["height"] = convertedHeight
			dim["width"] = convertedWidth
		}
		icon := &PreviewImage{
			URL:    extracted["icon"],
			Type:   extracted["icon:type"],
			Height: dim["height"],
			Width:  dim["width"],
		}
		return icon, nil
	}
	return nil, nil
}

//convertRelative converts a relative URL to an absolute URL
func convertRelative(pageURL string, otherURL string) (string, error) {
	base, err := url.Parse(pageURL)
	if err != nil {
		return "", fmt.Errorf("Error parsing base url: %v", err)
	}
	u, err := url.Parse(otherURL)
	if err != nil {
		return "", fmt.Errorf("Error parsing resource: %v", err)
	}
	return base.ResolveReference(u).String(), nil
}

//convToInt converts a string to an int
func convToInt(v string) (int, error) {
	if v != "" {
		d, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("Error converting string to int: %v", err)
		}
		return d, nil
	}
	return 0, nil
}

//makeImages returns a PreviewImage based on extracted information
func makeImages(pageURL string, extracted map[string]string) (*PreviewImage, error) {
	imageVal, dim, err := cleanMedia("og:image", pageURL, extracted)

	if err != nil {
		return nil, fmt.Errorf("Error cleaning media: %v", err)
	}

	image := &PreviewImage{
		URL:       imageVal["og:image"],
		SecureURL: imageVal["og:image:secure_url"],
		Type:      imageVal["og:image:type"],
		Width:     dim["og:image:width"],
		Height:    dim["og:image:height"],
		Alt:       imageVal["og:image:alt"],
	}
	return image, nil
}

//makeVideos returns a PreviewImage based on extracted information
func makeVideos(pageURL string, extracted map[string]string) (*PreviewVideo, error) {
	videoVal, dim, err := cleanMedia("og:video", pageURL, extracted)

	if err != nil {
		return nil, fmt.Errorf("Error cleaning media: %v", err)
	}

	video := &PreviewVideo{
		URL:       videoVal["og:video"],
		SecureURL: videoVal["og:vide:secure_url"],
		Type:      videoVal["og:video:type"],
		Width:     dim["og:video:width"],
		Height:    dim["og:video:height"],
	}

	return video, nil
}

//cleanMedia cleans up extracted information for images and videos by making relative URLs to
//absolute URLs and converts dimensions from strings to ints.
func cleanMedia(mediaType string, pageURL string, extracted map[string]string) (map[string]string, map[string]int, error) {
	cleaned := map[string]string{}
	dim := map[string]int{}
	for k, v := range extracted {
		if strings.Contains(k, mediaType) {
			if strings.Contains(k, "url") || k == mediaType {
				converted, err := convertRelative(pageURL, v)
				if err != nil {
					return nil, nil, fmt.Errorf("Error converting relative URL to absolute URL: %v", err)
				}
				extracted[k] = converted

			} else if k == mediaType+":width" || k == mediaType+":height" {
				convertedInt, err := convToInt(v)
				if err != nil {
					return nil, nil, fmt.Errorf("Error converting string to int: %v", err)
				}
				dim[k] = convertedInt
			}
			cleaned[k] = extracted[k]
		}
	}
	return cleaned, dim, nil
}
