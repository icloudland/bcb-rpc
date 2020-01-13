package bcb

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// HTTPClient is a common interface for JSONRPCClient and URIClient.
type HTTPClient interface {
	Call(method string, params map[string]interface{}, result interface{}) (interface{}, error)
}

// TODO: Deprecate support for IP:PORT or /path/to/socket
func makeHTTPDialer(remoteAddr string) (string, func(string, string) (net.Conn, error)) {
	parts := strings.SplitN(remoteAddr, "://", 2)
	var protocol, address string
	if len(parts) == 1 {
		// default to tcp if nothing specified
		protocol, address = "tcp", remoteAddr
	} else if len(parts) == 2 {
		protocol, address = parts[0], parts[1]
	} else {
		// return a invalid message
		msg := fmt.Sprintf("Invalid addr: %s", remoteAddr)
		return msg, func(_ string, _ string) (net.Conn, error) {
			return nil, errors.New(msg)
		}
	}
	// accept http as an alias for tcp
	if protocol == "http" || protocol == "https" {
		protocol = "tcp"
	}

	// replace / with . for http requests (kvstore domain)
	trimmedAddress := strings.Replace(address, "/", ".", -1)
	return trimmedAddress, func(proto, addr string) (net.Conn, error) {
		return net.Dial(protocol, address)
	}
}

// We overwrite the http.Client.Dial so we can do http over tcp or unix.
// remoteAddr should be fully featured (eg. with tcp:// or unix://)
func makeHTTPClient(remoteAddr string) (string, *http.Client) {
	address, dialer := makeHTTPDialer(remoteAddr)
	proto := "http://"
	if strings.HasPrefix(remoteAddr, "https:") {
		proto = "https://"
	}
	return proto + address, &http.Client{
		Transport: &http.Transport{
			Dial: dialer,
		},
	}
}

func makeHTTPSClient(remoteAddr string, pool *x509.CertPool, disableKeepAlive bool) (string, *http.Client) {
	//_, dialer := makeHTTPDialer(remoteAddr)

	tr := new(http.Transport)
	tr.DisableKeepAlives = disableKeepAlive
	tr.IdleConnTimeout = time.Second * 120
	if pool != nil {
		tr.TLSClientConfig = &tls.Config{RootCAs: pool}
	} else {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return remoteAddr, &http.Client{Transport: tr, Timeout: time.Duration(time.Second * 120)}
}

// JSONRPCClient takes params as a slice
type JSONRPCClient struct {
	address string
	client  *http.Client
}

// NewJSONRPCClient returns a JSONRPCClient pointed at the given address.
func NewJSONRPCClient(remote string) *JSONRPCClient {
	address, client := makeHTTPClient(remote)
	return &JSONRPCClient{
		address: address,
		client:  client,
	}
}

// 改接口支持https和http
// 当使用http时，remote使用http://ip:port/path
func NewJSONRPCClientEx(remote, certFile string, disableKeepAlive bool) *JSONRPCClient {
	var pool *x509.CertPool
	if certFile != "" {
		pool = x509.NewCertPool()
		caCert, err := ioutil.ReadFile(certFile)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}

		pool.AppendCertsFromPEM(caCert)
	}

	address, client := makeHTTPSClient(remote, pool, disableKeepAlive)
	//client.Timeout = time.Second * 5

	return &JSONRPCClient{
		address: address,
		client:  client,
	}
}

func (c *JSONRPCClient) Call(method string, params map[string]interface{}, result interface{}) (interface{}, error) {
	request, err := MapToRequest("jsonrpc-client", method, params)
	if err != nil {
		return nil, err
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		fmt.Println("lib client http_client error to json.Marshal(request)")
		return nil, err
	}

	fmt.Println(string(requestBytes))
	requestBuf := bytes.NewBuffer(requestBytes)
	// log.Info(Fmt("RPC request to %v (%v): %v", c.remote, method, string(requestBytes)))
	httpResponse, err := c.client.Post(c.address, "text/json", requestBuf)
	if err != nil {
		return nil, err
	}

	defer httpResponse.Body.Close() // nolint: errcheck

	responseBytes, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(responseBytes))
	return unmarshalResponseBytes(responseBytes, result)
}

func unmarshalResponseBytes(responseBytes []byte, result interface{}) (interface{}, error) {
	// Read response.  If rpc/core/types is imported, the result will unmarshal
	// into the correct type.
	// log.Notice("response", "response", string(responseBytes))
	var err error
	response := &RPCResponse{}
	err = json.Unmarshal(responseBytes, response)
	if err != nil {
		return nil, errors.Errorf("Error unmarshalling rpc response: %v", err)
	}
	if response.Error != nil {
		return nil, errors.Errorf("Response error: %v", response.Error)
	}
	// Unmarshal the RawMessage into the result.
	//err = cdc.UnmarshalJSON(response.Result, result)
	err = json.Unmarshal(response.Result, result)
	if err != nil {
		return nil, errors.Errorf("Error unmarshalling rpc response result: %v", err)
	}
	return result, nil
}
