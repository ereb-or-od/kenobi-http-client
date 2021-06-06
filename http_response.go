
package http_client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HttpResponse struct {
	Request     *HttpRequest
	RawResponse *http.Response

	body       []byte
	size       int64
	receivedAt time.Time
}

func (r *HttpResponse) ExtractBody() []byte {
	if r.RawResponse == nil {
		return []byte{}
	}
	return r.body
}

func (r *HttpResponse) ExtractStatus() string {
	if r.RawResponse == nil {
		return ""
	}
	return r.RawResponse.Status
}

func (r *HttpResponse) ExtractStatusCode() int {
	if r.RawResponse == nil {
		return 0
	}
	return r.RawResponse.StatusCode
}

func (r *HttpResponse) ExtractProto() string {
	if r.RawResponse == nil {
		return ""
	}
	return r.RawResponse.Proto
}

func (r *HttpResponse) ExtractResult() interface{} {
	return r.Request.Result
}

func (r *HttpResponse) ExtractError() interface{} {
	return r.Request.Error
}

func (r *HttpResponse) ExtractHeader() http.Header {
	if r.RawResponse == nil {
		return http.Header{}
	}
	return r.RawResponse.Header
}

func (r *HttpResponse) ExtractCookies() []*http.Cookie {
	if r.RawResponse == nil {
		return make([]*http.Cookie, 0)
	}
	return r.RawResponse.Cookies()
}

func (r *HttpResponse) ExtractBodyAsString() string {
	if r.body == nil {
		return ""
	}
	return strings.TrimSpace(string(r.body))
}

func (r *HttpResponse) CalculateDuration() time.Duration {
	if r.Request.clientTrace != nil {
		return r.Request.TraceInfo().TotalTime
	}
	return r.receivedAt.Sub(r.Request.Time)
}

func (r *HttpResponse) ReceivedAt() time.Time {
	return r.receivedAt
}

func (r *HttpResponse) Size() int64 {
	return r.size
}

func (r *HttpResponse) ExtractRawBody() io.ReadCloser {
	if r.RawResponse == nil {
		return nil
	}
	return r.RawResponse.Body
}

func (r *HttpResponse) IsSuccess() bool {
	return r.ExtractStatusCode() > 199 && r.ExtractStatusCode() < 300
}

func (r *HttpResponse) IsError() bool {
	return r.ExtractStatusCode() > 399
}

func (r *HttpResponse) setReceivedAt() {
	r.receivedAt = time.Now()
	if r.Request.clientTrace != nil {
		r.Request.clientTrace.endTime = r.receivedAt
	}
}
func (r *HttpResponse) fmtBodyString(sl int64) string {
	if r.body != nil {
		if int64(len(r.body)) > sl {
			return fmt.Sprintf("***** RESPONSE TOO LARGE (size - %d) *****", len(r.body))
		}
		ct := r.ExtractHeader().Get(hdrContentTypeKey)
		if IsJSONType(ct) {
			out := acquireBuffer()
			defer releaseBuffer(out)
			err := json.Indent(out, r.body, "", "   ")
			if err != nil {
				return fmt.Sprintf("*** ExtractError: Unable to format response body - \"%s\" ***\n\nLog ExtractBody as-is:\n%s", err, r.ExtractBodyAsString())
			}
			return out.String()
		}
		return r.ExtractBodyAsString()
	}

	return "***** NO CONTENT *****"
}
