package main

import (
    "github.com/gashirar/co2mini"
    "fmt"
    "log"
	"bytes"
	"encoding/json"
	"net/http"
    "time"
    "encoding/base64"

	"gopkg.in/raintank/schema.v1"
)

var url = "<YOUR URL>"
var user = "<YOUR USER>"
var apiKey = "<YOUR APIKEY>"

// createPoint creates a datapoint, i.e. a MetricData structure, and makes sure the id is set.
func createPoint(name string, interval int, val float64, time int64) *schema.MetricData {
	md := schema.MetricData{
		Name:     name,       // in graphite style format. should be same as Metric field below (used for partitioning, schema matching, indexing)
		Metric:   name,       // in graphite style format. should be same as Name field above (used to generate Id)
		Interval: interval,   // the resolution of the metric
		Value:    val,        // float64 value
		Unit:     "",         // not needed or used yet
		Time:     time,       // unix timestamp in seconds
		Mtype:    "gauge",    // not used yet. but should be one of gauge rate count counter timestamp
		Tags:     []string{}, // not needed or used yet. can be single words or key=value pairs
	}
	md.SetId()
	return &md
}

type GraphiteConfig struct {
    Url string
    User string
    ApiKey string
}

type Graphite struct {
    Config GraphiteConfig
}

func (g *Graphite) Send(metric string, interval int, val float64, t int64) (resp *http.Response, err error) {
    metrics := schema.MetricDataArray{}
    metrics = append(metrics, createPoint(metric, interval, float64(val), t))

    // encode as json
    data, err := json.Marshal(metrics)
    if err != nil {
        return nil , err
    }

    client := &http.Client{}

    req, err := http.NewRequest("POST", g.Config.Url, bytes.NewBuffer(data))
    if err != nil {
        return nil , err
    }

    basicAuthHeader := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", g.Config.User, g.Config.ApiKey)))
    req.Header.Add("Authorization", "Basic "+basicAuthHeader)
    req.Header.Add("Content-Type", "application/json")
    resp, err = client.Do(req)
    if err != nil {
        return nil , err
    }
    return resp, nil
}


func main() {
    var co2mini co2mini.Co2mini
    var co2 int
    var temp float64

    config := GraphiteConfig{Url: url, User: user, ApiKey: apiKey}
    graphite := Graphite{Config: config}

    if err := co2mini.Connect(); err != nil {
        log.Fatal("[ERROR] Cannot connect co2mini.", err)
    }

    go func() {
        if err := co2mini.Start(); err != nil {
            log.Fatal("[ERROR] Cannot start.", err)
        }
    }()

    for {
        select {
        case co2 = <-co2mini.Co2Ch:
            now := time.Now().Unix()
            resp, err := graphite.Send("co2mini.co2", 1, float64(co2), now)
            if err != nil {
                panic(err)
            }
            buf := make([]byte, 4096)
            n, err := resp.Body.Read(buf)
            fmt.Println(string(buf[:n]))
        case temp = <-co2mini.TempCh:
            now := time.Now().Unix()
            resp, err := graphite.Send("co2mini.co2", 1, float64(temp), now)
            if err != nil {
                panic(err)
            }
            buf := make([]byte, 4096)
            n, err := resp.Body.Read(buf)
            fmt.Println(string(buf[:n]))
        }
    }
}
