package exporter

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// KibanaCollector collects the Kibana information together to be used by
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

// KibanaMetrics is used to unmarshal the metrics response from Kibana.
type KibanaMetrics struct {
	Status struct {
		Overall struct {
			State string `json:"state"`
		} `json:"overall"`
	} `json:"status"`
	Metrics struct {
		ConcurrentConnections int `json:"concurrent_connections"`
		Process               struct {
			UptimeInMillis float64 `json:"uptime_in_millis"`
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

// TestConnection checks whether the connection to Kibana is healthy
func (c *KibanaCollector) TestConnection() bool {
	log.Debug().
		Msg("checking for kibana status")

	_, err := c.scrape()
	if err != nil {
		log.Info().
			Msgf("test connection to kibana failed: %s", err)
		return false
	}

	return true
}

// WaitForConnection is a method to block until Kibana becomes available
func (c *KibanaCollector) WaitForConnection() {
	for {
		if !c.TestConnection() {
			log.Info().
				Msg("waiting for kibana to be responsive")

			// hardcoded since it's unlikely this is user controlled
			time.Sleep(10 * time.Second)
			continue
		}

		log.Info().
			Msg("kibana is up")
		return
	}
}

// NewCollector builds a KibanaCollector struct
func NewCollector(kibanaURI, kibanaUsername, kibanaPassword string, kibanaSkipTLS bool) (*KibanaCollector, error) {
	collector := &KibanaCollector{}
	collector.url = kibanaURI

	if strings.HasPrefix(kibanaURI, "https://") {
		log.Debug().
			Msgf("kibana URL is a TLS one: %s", kibanaURI)

		if kibanaSkipTLS {
			log.Info().
				Msgf("skipping TLS verification for Kibana URL: %s", kibanaURI)
		}

		//#nosec G402 -- user defined
		tConf := &tls.Config{
			InsecureSkipVerify: kibanaSkipTLS,
		}

		tr := &http.Transport{
			TLSClientConfig: tConf,
		}

		collector.client = &http.Client{
			Transport: tr,
		}
	} else {
		log.Debug().
			Msgf("kibana URL is a plain text one: %s", kibanaURI)

		collector.client = &http.Client{}
		if kibanaSkipTLS {
			log.Info().
				Msgf("kibana.skip-tls is enabled for an http URL, ignoring: %s", kibanaURI)
		}
	}

	if kibanaUsername != "" && kibanaPassword != "" {
		log.Debug().
			Msg("using authenticated requests with Kibana")

		creds := fmt.Sprintf("%s:%s", kibanaUsername, kibanaPassword)
		encCreds := base64.StdEncoding.EncodeToString([]byte(creds))
		collector.authHeader = fmt.Sprintf("Basic %s", encCreds)
	} else {
		log.Info().
			Msg("Kibana username or password is not provided, assuming unauthenticated communication")
	}

	return collector, nil
}

// scrape will connect to the Kibana instance, using the details
// provided by the KibanaCollector struct, and return the metrics as a
// KibanaMetrics representation.
func (c *KibanaCollector) scrape() (*KibanaMetrics, error) {
	log.Debug().
		Msg("building request for api/status from kibana")

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/status", c.url), nil)
	if err != nil {
		return nil, fmt.Errorf("could not initialize a request to scrape metrics: %s", err)
	}

	if c.authHeader != "" {
		log.Debug().
			Msg("adding auth header")
		req.Header.Add("Authorization", c.authHeader)
	}

	req.Header.Add("Accept", "application/json")

	log.Debug().
		Msg("requesting api/status from kibana")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while reading Kibana status: %s", err)
	}

	// CWE-703
	defer func() {
		if err = resp.Body.Close(); err != nil {
			log.Warn().Msgf("error while closing response body: %s", err)
		}
	}()

	log.Debug().
		Msg("processing api/status response")

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response from Kibana status: %s", resp.Status)
	}

	respContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error while reading response from Kibana status: %s", err)
	}

	metrics := &KibanaMetrics{}
	err = json.Unmarshal(respContent, &metrics)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshalling Kibana status: %s\nProblematic content:\n%s", err, respContent)
	}

	return metrics, nil
}
