dcos-log2es
===========

Ship DC/OS logs to elasticsearch. Currently in Proof of Concept state.

### How to build

* enable GO module 
```
export GO111MODULE=on
```
* build:
```
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w ${GO_LDFLAGS}" -v -o build/dcos-log2es-linux-amd64
```

### How to run

### Run with marathon

1. Deploy elasticsearch.

2. Deploy app with DC/OS UI. See: [marathon.json](marathon.json).


### Known Issues

* panic caused by connection(to the DC/OS Log API) is closed by random, not-known problem.
```
2019/07/23 16:53:16 No .env file found
2019/07/23 16:53:16 DC/OS Logging API: http://localhost:61001/system/v1/logs/v1/stream/?skip_prev=10
2019/07/23 16:53:16 DC/OS Logging Prefix: /
2019/07/23 16:53:16 Elasticsearch Index Pattern: filebeat-%d.%02d.%02d
2019/07/23 16:53:16 Elasticsearch URL: []
panic: unexpected end of JSON input

goroutine 1 [running]:
main.main()
	/Users/minyk/IdeaProjects/go/src/github.com/nexr/dcos-log2es/main.go:103 +0x136e
I0724 02:06:34.507697    11 executor.cpp:1017] Command exited with status 2 (pid: 13)
I0724 02:06:35.509596    12 process.cpp:927] Stopped the socket accept loop
```
