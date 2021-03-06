package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress = kingpin.Flag(
		"web.listen-address",
		"Address on which to listen for the proxy.",
	).Default(":9443").String()
	caCert = kingpin.Flag(
		"ca.cert",
		"Certificate authority cert in PEM format.",
	).Default("ca.pem").String()
	certFile = kingpin.Flag(
		"cert.file",
		"Certificate for the TLS server in PEM format",
	).Default("cert.pem").String()
	keyFile = kingpin.Flag(
		"key.file",
		"Key for the TLS server in PEM format",
	).Default("cert.pem").String()
)

// Proxy strcut
type Proxy struct {
}

// NewProxy creates a new Proxy instance
func NewProxy() *Proxy { return &Proxy{} }

func (p *Proxy) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	log.Printf("%v %v", req.Method, req.RequestURI)
	// parse the url
	exporterPort := getExporterPort(req)
	target := "http://127.0.0.1:" + exporterPort
	log.Printf("Serving reverse proxy: %s\n", target)
	log.Printf("Serving reverse proxy path: %s\n", req.URL.Path)
	log.Printf("Serving reverse proxy query: %s\n", req.URL.RawQuery)
	url, _ := url.Parse(target)

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	// Remove query string used by proxy
	q := req.URL.Query()
	q.Del("exporter-port")
	req.URL.RawQuery = q.Encode()
	log.Printf("Updated reverse proxy query: %s\n", req.URL.RawQuery)

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(wr, req)
}

func getExporterPort(req *http.Request) (exporterPort string) {
	q, _ := url.ParseQuery(req.URL.RawQuery)
	// If "proxy-port" query parameter doesn't exist
	// return empty port, which will default to 80
	// TODO: return better errors
	if _, ok := q["exporter-port"]; ok {
		ep := q["exporter-port"][0]
		log.Printf("Found request for exporter port: %s\n", ep)
		epi, err := strconv.Atoi(ep)
		if err != nil {
			log.Println("Exporter-port is not an integer, bailing")
			return ""
		}
		// Restrict proxy ports to a range
		if (epi < 9000) || (epi > 65000) {
			log.Printf("Requested port outside allowed range: '9000 - 65000")
			return ""
		}
		exporterPort = strconv.Itoa(epi)
		return exporterPort
	}
	log.Printf("Request didn't contain required query parameter, exporter-port")
	return ""
}

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	caCert, err := ioutil.ReadFile(*caCert)
	if err != nil {
		log.Fatalf("Problem reading the CA cert: %s\n", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// This TLS configuration will require connecting clients
	// to use a cert signed by our CA
	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	tlsConfig.BuildNameToCertificate()

	proxy := NewProxy()

	log.Printf("Starting metric proxy server on: %s\n", *listenAddress)
	server := &http.Server{
		Addr:      *listenAddress,
		TLSConfig: tlsConfig,
		Handler:   proxy,
	}
	if err = server.ListenAndServeTLS(*certFile, *keyFile); err != nil {
		log.Fatalf("Problem starting the proxy: %s\n", err)
	}
}
