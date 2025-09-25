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

// NewWebTracker returns a new WebTracker with given Config
func NewWebTracker(config *Config) *WebTracker {
	return &WebTracker{Config: config}
}

// TrackWebSources starts async tracking for Web Pages
func (w *WebTracker) TrackWebSources(noExec, debug bool) {
	for _, config := range w.Config.WebTrackingConfigs {
		go w.trackWebSource(noExec, debug, config)
	}
}

// trackWebSource tracks web sources. Meant to be executed periodically
func (w *WebTracker) trackWebSource(noExec, debug bool, config WebTarget) {
	if debug {
		w.Config.log("Web curl for: " + config.Target)
	}
	wC := w.curl(debug, config)

	var execFlag = false
	// Check for Status Code
	if config.StatusCode != 0 {
		if config.OnCodeIdentical {
			execFlag = wC.status == config.StatusCode
		} else {
			execFlag = wC.status != config.StatusCode
		}
	}
	// Check for HTTPS failure
	execFlag = execFlag || (!wC.isHttps && config.OnHTTPSFails)
	// Check for Content match
	if len(config.Content) > 0 {
		if config.ContentIsExact {
			// Check for identity
			execFlag = execFlag || (strings.Contains(wC.content, config.Content) && len(wC.content) == len(config.Content))
		} else {
			// Check for substring
			execFlag = execFlag || strings.Contains(wC.content, config.Content)
		}
	}

	if execFlag {
		w.Config.log("Executing on web tracking for: " + config.Target)
		w.Config.exec(CalleeWeb, noExec)
	}
}

func (w *WebTracker) curl(debug bool, config WebTarget) webCurl {
	res := webCurl{
		status:  -1,
		content: "",
		isHttps: false,
	}

	for i := 0; i < config.RetryCount; i++ {
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
		return ""
	}
	return string(bodyBytes)
}
