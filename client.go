package main

import (
	"net/http"
	"log"
	"bytes"
	"strings"
	"github.com/smartystreets/cproxy"
	"fmt"
)

var explicitForwardProxyHandler = cproxy.Configure()

func ForwardServer(w http.ResponseWriter, r *http.Request) {
	//if r.Method == http.MethodConnect {
	//	//explicitForwardProxyHandler.ServeHTTP(w, r)
	//	ClientProcessTcpTunnel(w, r)
	//	return
	//}

	dump := DumpIncomingRequest(r)
	LogPretty(">>> ", strings.SplitN(string(dump), "\n", 2)[0])

	req, e := http.NewRequest("POST", *server, bytes.NewReader(dump))
	if e != nil {
		w.WriteHeader(500)
		w.Write([]byte("500 Internal Error: fail on NewRequest"))
		return
	}

	AddForwardedHeaders(req, r)
	DoRequestAndWriteBack(req, w)
}

func AddForwardedHeaders(req, originReq *http.Request) {
	fBy := "package-via-http"
	fFor := originReq.Header.Get("User-Agent")
	fHost := originReq.Host
	fProto := originReq.Proto
	//fSchema := originReq.URL.Scheme
	fSchema := "https"
	ip := originReq.RemoteAddr
	LogPretty("  >>> ", originReq.Header)

	forward := fmt.Sprintf("by=%v; for=%v; host=%v; proto=%v", fBy, fFor, fHost, fProto)
	req.Header.Set("Forwarded", forward)
	req.Header.Set("X-Forwarded-By", fBy)
	req.Header.Set("X-Forwarded-For", fFor)
	req.Header.Set("X-Forwarded-Host", fHost)
	req.Header.Set("X-Forwarded-Proto", fProto)
	req.Header.Set("X-Forwarded-Schema", fSchema)
	req.Header.Set("X-Real-IP", ip)
}

func RunClient(listen, server string) {
	log.Printf("Listening %v\n", listen)
	outerHandler := http.HandlerFunc(ForwardServer)
	http.ListenAndServe(listen, outerHandler)
}
