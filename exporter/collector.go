package exporter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// A type that collects the Kibana information together to be used by
// the exporter to scrape metrics.
type KibanaCollector struct {
	// url is the base URL of the Kibana instance or the service
	url string

	// authHeader is the string that should be used as the value
	// for the "Authorization" header. If this is empty, it is
	// assumed that no authorization is needed.
	authHeader string

	// client is the http.Client that will be used to make
	// requests to collect the Kibana metrics
	client *http.Client
}

// A type that is used to unmarshal the metrics response from Kibana.
type KibanaMetrics struct {
	Status struct {
		Overall struct {
			State string `json:"state"`
		} `json:"overall"`
	} `json:"status"`
	Metrics struct {
		ConcurrentConnections int `json:"concurrent_connections"`
		Process               struct {
			UptimeInMillis int64 `json:"uptime_in_millis"`
			Memory         struct {
				Heap struct {
					TotalInBytes int64 `json:"total_in_bytes"`
					UsedInBytes  int64 `json:"used_in_bytes"`
				} `json:"heap"`
			} `json:"memory"`
		} `json:"process"`
		Os struct {
			Load struct {
				Load1m  float64 `json:"1m"`
				Load5m  float64 `json:"5m"`
				Load15m float64 `json:"15m"`
			} `json:"load"`
		} `json:"os"`
		ResponseTimes struct {
			AvgInMillis float64 `json:"avg_in_millis"`
			MaxInMillis float64 `json:"max_in_millis"`
		} `json:"response_times"`
		Requests struct {
			Disconnects int `json:"disconnects"`
			Total       int `json:"total"`
		} `json:"requests"`
	} `json:"metrics"`
}

// scrape will connect to the Kibana instance, using the details
// provided by the KibanaCollector struct, and return the metrics as a
// KibanaMetrics representation.
func (c *KibanaCollector) scrape() (error, *KibanaMetrics) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/status?extended", c.url), nil)
	if err != nil {
		return errors.New(fmt.Sprintf("could not initialize a request to scrape metrics: %s", err)), nil
	}

	if c.authHeader != "" {
		req.Header.Add("Authorization", c.authHeader)
	}

	req.Header.Add("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("error while reading Kibana status: %s", err)), nil
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("invalid response from Kibana status: %s", resp.Status)), nil

	}

	respContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("error while reading response from Kibana status: %s", err)), nil
	}

	metrics := &KibanaMetrics{}
	err = json.Unmarshal(respContent, &metrics)
	if err != nil {
		return errors.New(fmt.Sprintf("error while unmarshalling Kibana status: %s\nProblematic content:\n%s", err, respContent)), nil
	}

	return nil, metrics
}
