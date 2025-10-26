package main

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// webCurl represents a curl event
type webCurl struct {
	status  int
	content string
	isHttps bool
}

// WebTracker represents the Web tracking service
type WebTracker struct {
	Config *Config
}

const WebNoExec uint8 = 0
const WebCont uint8 = 1
const WebContExact uint8 = 2
const WebCode uint8 = 3
const WebNotCode uint8 = 4
const WebHTTPS uint8 = 5

// NewWebTracker returns a new WebTracker with given Config
func NewWebTracker(config *Config) *WebTracker {
	return &WebTracker{Config: config}
}

// TrackWebSources starts async tracking for Web Pages
func (w *WebTracker) TrackWebSources(noExec, debug bool) uint {
	var counter uint = 0
	for _, config := range w.Config.WebTrackingConfigs {
		go w.trackWebSource(noExec, debug, config)
		counter++
	}
	return counter
}

// trackWebSource tracks web sources. Meant to be executed periodically
func (w *WebTracker) trackWebSource(noExec, debug bool, config WebTarget) uint8 {
	if debug {
		w.Config.log("Web curl for: " + config.Target)
	}
	wC := w.curl(debug, config)

	returnValue := WebNoExec
	// Check for Status Code
	if config.StatusCode != 0 {
		if config.OnCodeIdentical && wC.status == config.StatusCode {
			returnValue = WebCode
		} else if wC.status != config.StatusCode {
			returnValue = WebNotCode
		}
	}
	// Check for HTTPS failure
	if !wC.isHttps && config.OnHTTPSFails {
		returnValue = WebHTTPS
	}
	// Check for Content match
	if len(config.Content) > 0 {
		if config.ContentIsExact {
			// Check for identity
			if strings.Contains(wC.content, config.Content) && len(wC.content) == len(config.Content) {
				returnValue = WebContExact
			}
		} else {
			// Check for substring
			if strings.Contains(wC.content, config.Content) {
				returnValue = WebCont
			}
		}
	}

	if returnValue != WebNoExec {
		w.Config.log("Executing on web tracking for: " + config.Target)
		w.Config.exec(debug, CalleeWeb, config.CommandId, noExec)
	}
	return returnValue
}

func (w *WebTracker) curl(debug bool, config WebTarget) webCurl {
	res := webCurl{
		status:  -1,
		content: "",
		isHttps: false,
	}

	for i := 0; i <= config.RetryCount; i++ {
		response, err := http.Get(config.Target)
		if err != nil {
			w.Config.log(err.Error())
		} else {
			res.status = response.StatusCode
			res.isHttps = response.TLS != nil
			res.content = w.readBody(response)
			if debug {
				w.Config.log("Status: " + strconv.Itoa(res.status))
				w.Config.log("isHttps: " + strconv.FormatBool(res.isHttps))
				w.Config.log("Content: " + res.content)
			}
			return res
		}
		time.Sleep(config.RetryDelay)
	}
	return res
}

func (w *WebTracker) readBody(response *http.Response) string {
	defer func(Body io.ReadCloser, config2 *Config) {
		err := Body.Close()
		if err != nil {
			config2.log(err.Error())
		}
	}(response.Body, w.Config)
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		w.Config.log(err.Error())
		return ""
	}
	return string(bodyBytes)
}
