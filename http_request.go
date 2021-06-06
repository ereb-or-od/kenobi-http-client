
package http_client

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)


type HttpRequest struct {
	URL        string
	Method     string
	Token      string
	AuthScheme string
	QueryParam url.Values
	FormData   url.Values
	Header     http.Header
	Time       time.Time
	Body       interface{}
	Result     interface{}
	Error      interface{}
	RawRequest *http.Request
	SRV        *SRVRecord
	UserInfo   *User
	Cookies    []*http.Cookie

	Attempt int

	isMultiPart         bool
	isFormData          bool
	setContentLength    bool
	isSaveResponse      bool
	notParseResponse    bool
	jsonEscapeHTML      bool
	trace               bool
	outputFile          string
	fallbackContentType string
	forceContentType    string
	ctx                 context.Context
	pathParams          map[string]string
	values              map[string]interface{}
	client              *HttpClient
	bodyBuf             *bytes.Buffer
	clientTrace         *clientTrace
	multipartFiles      []*File
	multipartFields     []*MultipartField
}

func (r *HttpRequest) Context() context.Context {
	if r.ctx == nil {
		return context.Background()
	}
	return r.ctx
}

func (r *HttpRequest) UseContext(ctx context.Context) *HttpRequest {
	r.ctx = ctx
	return r
}

func (r *HttpRequest) UseHeader(header, value string) *HttpRequest {
	r.Header.Set(header, value)
	return r
}

func (r *HttpRequest) UseHeaders(headers map[string]string) *HttpRequest {
	for h, v := range headers {
		r.UseHeader(h, v)
	}
	return r
}

func (r *HttpRequest) UseHeaderVerbatim(header, value string) *HttpRequest {
	r.Header[header] = []string{value}
	return r
}

func (r *HttpRequest) UseQueryParam(param, value string) *HttpRequest {
	r.QueryParam.Set(param, value)
	return r
}

func (r *HttpRequest) UseQueryParams(params map[string]string) *HttpRequest {
	for p, v := range params {
		r.UseQueryParam(p, v)
	}
	return r
}

func (r *HttpRequest) UseQueryParamsFromValues(params url.Values) *HttpRequest {
	for p, v := range params {
		for _, pv := range v {
			r.QueryParam.Add(p, pv)
		}
	}
	return r
}

func (r *HttpRequest) UseQueryString(query string) *HttpRequest {
	params, err := url.ParseQuery(strings.TrimSpace(query))
	if err == nil {
		for p, v := range params {
			for _, pv := range v {
				r.QueryParam.Add(p, pv)
			}
		}
	} else {
		r.client.log.Error("an error occurred", err)
	}
	return r
}

func (r *HttpRequest) UseFormData(data map[string]string) *HttpRequest {
	for k, v := range data {
		r.FormData.Set(k, v)
	}
	return r
}
func (r *HttpRequest) UseFormDataFromValues(data url.Values) *HttpRequest {
	for k, v := range data {
		for _, kv := range v {
			r.FormData.Add(k, kv)
		}
	}
	return r
}

func (r *HttpRequest) UseBody(body interface{}) *HttpRequest {
	r.Body = body
	return r
}
func (r *HttpRequest) UseResponse(res interface{}) *HttpRequest {
	r.Result = getPointer(res)
	return r
}

func (r *HttpRequest) UseError(err interface{}) *HttpRequest {
	r.Error = getPointer(err)
	return r
}

func (r *HttpRequest) UseFile(param, filePath string) *HttpRequest {
	r.isMultiPart = true
	r.FormData.Set("@"+param, filePath)
	return r
}

func (r *HttpRequest) UseFiles(files map[string]string) *HttpRequest {
	r.isMultiPart = true
	for f, fp := range files {
		r.FormData.Set("@"+f, fp)
	}
	return r
}

func (r *HttpRequest) UseFileReader(param, fileName string, reader io.Reader) *HttpRequest {
	r.isMultiPart = true
	r.multipartFiles = append(r.multipartFiles, &File{
		Name:      fileName,
		ParamName: param,
		Reader:    reader,
	})
	return r
}

func (r *HttpRequest) UseMultipartFormData(data map[string]string) *HttpRequest {
	for k, v := range data {
		r = r.UseMultipartField(k, "", "", strings.NewReader(v))
	}

	return r
}

func (r *HttpRequest) UseMultipartField(param, fileName, contentType string, reader io.Reader) *HttpRequest {
	r.isMultiPart = true
	r.multipartFields = append(r.multipartFields, &MultipartField{
		Param:       param,
		FileName:    fileName,
		ContentType: contentType,
		Reader:      reader,
	})
	return r
}

func (r *HttpRequest) UseMultipartFields(fields ...*MultipartField) *HttpRequest {
	r.isMultiPart = true
	r.multipartFields = append(r.multipartFields, fields...)
	return r
}

func (r *HttpRequest) UseContentLength(l bool) *HttpRequest {
	r.setContentLength = l
	return r
}

func (r *HttpRequest) UseBasicAuthentication(username, password string) *HttpRequest {
	r.UserInfo = &User{Username: username, Password: password}
	return r
}

func (r *HttpRequest) UseAuthenticationToken(token string) *HttpRequest {
	r.Token = token
	return r
}

func (r *HttpRequest) UseAuthenticationSchema(scheme string) *HttpRequest {
	r.AuthScheme = scheme
	return r
}

func (r *HttpRequest) UseOutputFile(file string) *HttpRequest {
	r.outputFile = file
	r.isSaveResponse = true
	return r
}

func (r *HttpRequest) UseSRV(srv *SRVRecord) *HttpRequest {
	r.SRV = srv
	return r
}

func (r *HttpRequest) UseDoNotParseResponse(parse bool) *HttpRequest {
	r.notParseResponse = parse
	return r
}

func (r *HttpRequest) UsePathParam(param, value string) *HttpRequest {
	r.pathParams[param] = value
	return r
}

func (r *HttpRequest) UsePathParams(params map[string]string) *HttpRequest {
	for p, v := range params {
		r.UsePathParam(p, v)
	}
	return r
}

func (r *HttpRequest) ExpectContentType(contentType string) *HttpRequest {
	r.fallbackContentType = contentType
	return r
}

func (r *HttpRequest) ForceContentType(contentType string) *HttpRequest {
	r.forceContentType = contentType
	return r
}

func (r *HttpRequest) UseJSONEscapeHTML(b bool) *HttpRequest {
	r.jsonEscapeHTML = b
	return r
}

func (r *HttpRequest) UseCookie(hc *http.Cookie) *HttpRequest {
	r.Cookies = append(r.Cookies, hc)
	return r
}

func (r *HttpRequest) UseCookies(rs []*http.Cookie) *HttpRequest {
	r.Cookies = append(r.Cookies, rs...)
	return r
}

func (r *HttpRequest) EnableTrace() *HttpRequest {
	r.trace = true
	return r
}

func (r *HttpRequest) TraceInfo() HttpTracing {
	ct := r.clientTrace

	if ct == nil {
		return HttpTracing{}
	}

	ti := HttpTracing{
		DNSLookup:      ct.dnsDone.Sub(ct.dnsStart),
		TLSHandshake:   ct.tlsHandshakeDone.Sub(ct.tlsHandshakeStart),
		ServerTime:     ct.gotFirstResponseByte.Sub(ct.gotConn),
		IsConnReused:   ct.gotConnInfo.Reused,
		IsConnWasIdle:  ct.gotConnInfo.WasIdle,
		ConnIdleTime:   ct.gotConnInfo.IdleTime,
		RequestAttempt: r.Attempt,
	}

	// Calculate the total time accordingly,
	// when connection is reused
	if ct.gotConnInfo.Reused {
		ti.TotalTime = ct.endTime.Sub(ct.getConn)
	} else {
		ti.TotalTime = ct.endTime.Sub(ct.dnsStart)
	}

	// Only calculate on successful connections
	if !ct.connectDone.IsZero() {
		ti.TCPConnTime = ct.connectDone.Sub(ct.dnsDone)
	}

	// Only calculate on successful connections
	if !ct.gotConn.IsZero() {
		ti.ConnTime = ct.gotConn.Sub(ct.getConn)
	}

	// Only calculate on successful connections
	if !ct.gotFirstResponseByte.IsZero() {
		ti.ResponseTime = ct.endTime.Sub(ct.gotFirstResponseByte)
	}

	// Capture remote address info when connection is non-nil
	if ct.gotConnInfo.Conn != nil {
		ti.RemoteAddr = ct.gotConnInfo.Conn.RemoteAddr()
	}

	return ti
}


// Get method does GET HTTP request. It's defined in section 4.3.1 of RFC7231.
func (r *HttpRequest) Get(url string) (*HttpResponse, error) {
	return r.Execute(MethodGet, url)
}

// Head method does HEAD HTTP request. It's defined in section 4.3.2 of RFC7231.
func (r *HttpRequest) Head(url string) (*HttpResponse, error) {
	return r.Execute(MethodHead, url)
}

// Post method does POST HTTP request. It's defined in section 4.3.3 of RFC7231.
func (r *HttpRequest) Post(url string) (*HttpResponse, error) {
	return r.Execute(MethodPost, url)
}

// Put method does PUT HTTP request. It's defined in section 4.3.4 of RFC7231.
func (r *HttpRequest) Put(url string) (*HttpResponse, error) {
	return r.Execute(MethodPut, url)
}

// Delete method does DELETE HTTP request. It's defined in section 4.3.5 of RFC7231.
func (r *HttpRequest) Delete(url string) (*HttpResponse, error) {
	return r.Execute(MethodDelete, url)
}

// Options method does OPTIONS HTTP request. It's defined in section 4.3.7 of RFC7231.
func (r *HttpRequest) Options(url string) (*HttpResponse, error) {
	return r.Execute(MethodOptions, url)
}

// Patch method does PATCH HTTP request. It's defined in section 2 of RFC5789.
func (r *HttpRequest) Patch(url string) (*HttpResponse, error) {
	return r.Execute(MethodPatch, url)
}

func (r *HttpRequest) Send() (*HttpResponse, error) {
	return r.Execute(r.Method, r.URL)
}

func (r *HttpRequest) Execute(method, url string) (*HttpResponse, error) {
	var addrs []*net.SRV
	var resp *HttpResponse
	var err error

	if r.isMultiPart && !(method == MethodPost || method == MethodPut || method == MethodPatch) {
		// No OnError hook here since this is a request validation error
		return nil, fmt.Errorf("multipart content is not allowed in HTTP verb [%v]", method)
	}

	if r.SRV != nil {
		_, addrs, err = net.LookupSRV(r.SRV.Service, "tcp", r.SRV.Domain)
		if err != nil {
			r.client.onErrorHooks(r, nil, err)
			return nil, err
		}
	}

	r.Method = method
	r.URL = r.selectAddr(addrs, url, 0)

	if r.client.RetryCount == 0 {
		r.Attempt = 1
		resp, err = r.client.execute(r)
		r.client.onErrorHooks(r, resp, unwrapNoRetryErr(err))
		return resp, unwrapNoRetryErr(err)
	}

	err = Backoff(
		func() (*HttpResponse, error) {
			r.Attempt++

			r.URL = r.selectAddr(addrs, url, r.Attempt)

			resp, err = r.client.execute(r)
			if err != nil {
				r.client.log.Error(fmt.Sprintf("%s, Attempt %d", err, r.Attempt), err)
			}

			return resp, err
		},
		Retries(r.client.RetryCount),
		WaitTime(r.client.RetryWaitTime),
		MaxWaitTime(r.client.RetryMaxWaitTime),
		RetryConditions(r.client.RetryConditions),
		RetryHooks(r.client.RetryHooks),
	)

	r.client.onErrorHooks(r, resp, unwrapNoRetryErr(err))

	return resp, unwrapNoRetryErr(err)
}

type SRVRecord struct {
	Service string
	Domain  string
}


func (r *HttpRequest) fmtBodyString(sl int64) (body string) {
	body = "***** NO CONTENT *****"
	if !isPayloadSupported(r.Method, r.client.AllowGetMethodPayload) {
		return
	}

	if _, ok := r.Body.(io.Reader); ok {
		body = "***** BODY IS io.Reader *****"
		return
	}

	// multipart or form-data
	if r.isMultiPart || r.isFormData {
		bodySize := int64(r.bodyBuf.Len())
		if bodySize > sl {
			body = fmt.Sprintf("***** REQUEST TOO LARGE (size - %d) *****", bodySize)
			return
		}
		body = r.bodyBuf.String()
		return
	}

	// request body data
	if r.Body == nil {
		return
	}
	var prtBodyBytes []byte
	var err error

	contentType := r.Header.Get(hdrContentTypeKey)
	kind := kindOf(r.Body)
	if canJSONMarshal(contentType, kind) {
		prtBodyBytes, err = json.MarshalIndent(&r.Body, "", "   ")
	} else if IsXMLType(contentType) && (kind == reflect.Struct) {
		prtBodyBytes, err = xml.MarshalIndent(&r.Body, "", "   ")
	} else if b, ok := r.Body.(string); ok {
		if IsJSONType(contentType) {
			bodyBytes := []byte(b)
			out := acquireBuffer()
			defer releaseBuffer(out)
			if err = json.Indent(out, bodyBytes, "", "   "); err == nil {
				prtBodyBytes = out.Bytes()
			}
		} else {
			body = b
		}
	} else if b, ok := r.Body.([]byte); ok {
		body = fmt.Sprintf("***** BODY IS byte(s) (size - %d) *****", len(b))
		return
	}

	if prtBodyBytes != nil && err == nil {
		body = string(prtBodyBytes)
	}

	if len(body) > 0 {
		bodySize := int64(len([]byte(body)))
		if bodySize > sl {
			body = fmt.Sprintf("***** REQUEST TOO LARGE (size - %d) *****", bodySize)
		}
	}

	return
}

func (r *HttpRequest) selectAddr(addrs []*net.SRV, path string, attempt int) string {
	if addrs == nil {
		return path
	}

	idx := attempt % len(addrs)
	domain := strings.TrimRight(addrs[idx].Target, ".")
	path = strings.TrimLeft(path, "/")

	return fmt.Sprintf("%s://%s:%d/%s", r.client.scheme, domain, addrs[idx].Port, path)
}

func (r *HttpRequest) initValuesMap() {
	if r.values == nil {
		r.values = make(map[string]interface{})
	}
}

var noescapeJSONMarshal = func(v interface{}) ([]byte, error) {
	buf := acquireBuffer()
	defer releaseBuffer(buf)
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	return buf.Bytes(), err
}
