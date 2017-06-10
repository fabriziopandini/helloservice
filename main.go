package main // import "github.com/fabriziopandini/helloservice"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var startTime time.Time

func main() {

	startTime = time.Now()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", doHostname)
	router.HandleFunc("/echo", doEcho)
	router.HandleFunc("/echoheaders", doEchoheaders)
	router.HandleFunc("/hostname", doHostname)
	router.HandleFunc("/fqdn", doFQDN)
	router.HandleFunc("/ip", doIP)
	router.HandleFunc("/env", doEnv)
	router.HandleFunc("/healthz", doHealthz)
	router.HandleFunc("/healthz-fail/{failAfter:[0-9]+}", doFailHealthz)
	router.HandleFunc("/exit/{exitCode:[0-9]+}", doExit)

	serve := ":8080"
	if port, err := strconv.Atoi(os.Getenv("TEST_SERVICE_PORT")); err == nil && port != 0 {
		serve = fmt.Sprintf(":%d", port)
	}
	fmt.Printf("Serving %s\n", serve)

	log.Fatal(http.ListenAndServe(serve, router))
}

func doEcho(w http.ResponseWriter, r *http.Request) {
	body := make(map[string]string)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body["echo"] = buf.String()

	writeJSONResponse(w, body)
}

func doEchoheaders(w http.ResponseWriter, r *http.Request) {
	body := make(map[string]interface{})
	headers := make(map[string]string)

	for k, v := range r.Header {
		headers[k] = strings.Join(v, ", ")
	}
	body["headers"] = headers

	writeJSONResponse(w, body)
}

func doHostname(w http.ResponseWriter, r *http.Request) {
	hostname, err := os.Hostname()
	if err != nil {
		writeJSONError(w, err)
		return
	}

	body := make(map[string]interface{})
	body["hostname"] = hostname

	writeJSONResponse(w, body)
}

func doEnv(w http.ResponseWriter, r *http.Request) {
	body := make(map[string]interface{})
	env := make(map[string]string)

	for _, e := range os.Environ() {
		token := strings.Split(e, "=")
		k := token[0]
		v := token[1]
		env[k] = v
	}
	body["env"] = env

	writeJSONResponse(w, body)
}

func doIP(w http.ResponseWriter, r *http.Request) {
	body := make(map[string]interface{})
	ip := []string{}

	ifaces, err := net.Interfaces()
	if err != nil {
		writeJSONError(w, err)
		return
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			writeJSONError(w, err)
			return
		}
		for _, addr := range addrs {
			var ipAddr net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ipAddr = v.IP
			case *net.IPAddr:
				ipAddr = v.IP
			}
			ip = append(ip, ipAddr.String())
		}
	}
	body["ip"] = ip

	writeJSONResponse(w, body)
}

func doFQDN(w http.ResponseWriter, r *http.Request) {
	body := make(map[string]interface{})
	body["fqdn"] = getFQDN()

	writeJSONResponse(w, body)
}

func getFQDN() string {
	// from https://github.com/ShowMax/go-fqdn/blob/master/fqdn.go
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}

	addrs, err := net.LookupIP(hostname)
	if err != nil {
		return hostname
	}

	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			ip, err := ipv4.MarshalText()
			if err != nil {
				return hostname
			}
			hosts, err := net.LookupAddr(string(ip))
			if err != nil || len(hosts) == 0 {
				return hostname
			}
			fqdn := hosts[0]
			return strings.TrimSuffix(fqdn, ".") // return fqdn without trailing dot
		}
	}
	return hostname
}

func doExit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exitCode, _ := strconv.Atoi(vars["exitCode"])

	os.Exit(exitCode)
}

func doHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	body := make(map[string]string)
	body["uptime"] = fmt.Sprintf("%.1f s", time.Since(startTime).Seconds())
	body["status"] = "200 OK"
	writeJSONResponse(w, body)
}

func doFailHealthz(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	failAfter, _ := strconv.ParseFloat(vars["failAfter"], 64)
	if failAfter == 0 {
		failAfter = 10.0
	}

	uptime := time.Since(startTime).Seconds()

	body := make(map[string]interface{})
	body["uptime"] = fmt.Sprintf("%.1f s", uptime)
	if uptime < failAfter {
		w.WriteHeader(http.StatusOK)
		body["status"] = fmt.Sprintf("200 OK, %.1f seconds before failing", failAfter-uptime)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		body["status"] = fmt.Sprintf("500 INTERNAL SERVER ERROR, failed since %.1f seconds", uptime-failAfter)
	}

	writeJSONResponse(w, body)
}

func writeJSONResponse(w http.ResponseWriter, data interface{}) {

	jData, err := json.Marshal(data)
	if err != nil {
		writeJSONError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

func writeJSONError(w http.ResponseWriter, err error) {
	jData, err := json.Marshal(err)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}
