
package http_client

import (
	"errors"
	"github.com/ereb-or-od/kenobi-json-marshaller/interfaces"
	logging "github.com/ereb-or-od/kenobi-logger/pkg/interfaces"
	"net"
	"net/http"
	"net/http/cookiejar"

	"golang.org/x/net/publicsuffix"
)

var (
	errLoggerMustBeSpecified = errors.New("logger must be specified")
	errMarshallerMustBeSpecified = errors.New("marshaller must be specified")
)

func New(logger logging.Logger, marshaller interfaces.Marshaller) (*HttpClient , error){
	cookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	return createClient(logger, marshaller, &http.Client{
		Jar: cookieJar,
	})
}

func NewWithClient(logger logging.Logger, marshaller interfaces.Marshaller, hc *http.Client) (*HttpClient, error) {
	return createClient(logger, marshaller, hc)
}

func NewWithLocalAddr(logger logging.Logger, marshaller interfaces.Marshaller, localAddr net.Addr)  (*HttpClient, error) {
	cookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	return createClient(logger, marshaller, &http.Client{
		Jar:       cookieJar,
		Transport: createTransport(localAddr),
	})
}
