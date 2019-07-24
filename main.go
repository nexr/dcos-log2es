package main

import (
	"context"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v6/estransport"
	"github.com/joho/godotenv"
	sse "github.com/r3labs/sse"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	//"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v6"
	"github.com/elastic/go-elasticsearch/v6/esapi"
)

var es *elasticsearch.Client
var cfg elasticsearch.Config

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

var dcosLogAPI = ""
var esIndexPattern = ""
var exist = false
var loggingEnabled = ""
var loggingPrefix = ""

func main() {
	dcosLogAPI, exist = os.LookupEnv("DCOS_LOG_API")
	if !exist {
		dcosLogAPI = "http://localhost:61001/system/v1/logs/v1/stream/?skip_prev=10"
	}
	esIndexPattern, exist = os.LookupEnv("DCOS_LOG_INDEX_PATTERN")
	if !exist {
		esIndexPattern = "filebeat-%d.%02d.%02d"
	}
	loggingEnabled, _ = os.LookupEnv("LOGGING_ENABLED")
	if enable, _ := strconv.ParseBool(loggingEnabled); enable {
		cfg = elasticsearch.Config{
			Logger: &estransport.TextLogger{
				Output:             os.Stdout,
				EnableResponseBody: true,
				EnableRequestBody:  true,
			},
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		}
	} else {
		cfg = elasticsearch.Config{
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		}
	}

	loggingPrefix, exist = os.LookupEnv("LOGGING_PREFIX")
	if !exist {
		loggingPrefix = "/"
	}

	log.Printf("DC/OS Logging API: %s", dcosLogAPI)
	log.Printf("DC/OS Logging Prefix: %s", loggingPrefix)
	log.Printf("Elasticsearch Index Pattern: %s", esIndexPattern)
	log.Printf("Elasticsearch URL: %s", cfg.Addresses)

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Errorf("Error creating the client: %s", err)
	}

	res, err := es.Info()
	if err != nil {
		fmt.Errorf("Error getting response: %s", err)
	}

	fmt.Println(res)

	defer res.Body.Close()

	events := make(chan *sse.Event)
	client := sse.NewClient(dcosLogAPI)
	client.SubscribeChan("messages", events)

	for {
		select {
		case e := <-events:
			dat := &DcosLog{}
			if err := json.Unmarshal(e.Data, &dat); err != nil {
				panic(err)
			}
			if dat.Fields.DCOSSPACE != "" && strings.HasPrefix(dat.Fields.DCOSSPACE, loggingPrefix) {
				dat.Fields.SYSLOGTIMESTAMP = time.Unix(0, dat.RealtimeTimestamp*1000).Format(time.RFC3339Nano)
				utcNow := time.Now().UTC()
				indexPattern := fmt.Sprintf(esIndexPattern, utcNow.Year(), utcNow.Month(), utcNow.Day())
				logString, err := json.Marshal(dat.Fields)
				req := esapi.IndexRequest{
					Index: indexPattern,
					Body:  strings.NewReader(strings.ToLower(string(logString))),
				}

				esres, err := req.Do(context.Background(), es)
				if err != nil {
					fmt.Errorf("Error getting response: %s", err)
				}
				if esres.StatusCode >= 400 {
					fmt.Println(esres.Status())
				}
			}
		}
	}
}

type DcosLog struct {
	Fields struct {
		AGENTID          string `json:"AGENT_ID"`
		CONTAINERID      string `json:"CONTAINER_ID"`
		DCOSSPACE        string `json:"DCOS_SPACE"`
		EXECUTORID       string `json:"EXECUTOR_ID"`
		FRAMEWORKID      string `json:"FRAMEWORK_ID"`
		MESSAGE          string `json:"MESSAGE"`
		STREAM           string `json:"STREAM"`
		SYSLOGIDENTIFIER string `json:"SYSLOG_IDENTIFIER"`
		SYSLOGTIMESTAMP  string `json:"@timestamp"`
	} `json:"fields"`
	Cursor             string `json:"cursor"`
	MonotonicTimestamp int64  `json:"monotonic_timestamp"`
	RealtimeTimestamp  int64  `json:"realtime_timestamp"`
}
