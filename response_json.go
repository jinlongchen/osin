package osin

import (
	"encoding/json"
//	"net/http"
	"github.com/valyala/fasthttp"
)

// OutputJSON encodes the Response to JSON and writes to the http.ResponseWriter
func OutputJSON(rs *Response, r *fasthttp.RequestCtx) error {
	// Add headers
	for i, k := range rs.Headers {
		for _, v := range k {
			r.Response.Header.Add(i, v)
//			w.Header().Add(i, v)
		}
	}

	if rs.Type == REDIRECT {
		// Output redirect with parameters
		u, err := rs.GetRedirectUrl()
		if err != nil {
			return err
		}
		r.Response.Header.Add("Location", u)
		r.Response.Header.SetStatusCode(302)
		//w.Header().Add("Location", u)
		//w.WriteHeader(302)
	} else {
		// set content type if the response doesn't already have one associated with it
		if r.Response.Header.Peek("Content-Type") == nil {
			r.Response.Header.Set("Content-Type", "application/json")
		}
//		if w.Header().Get("Content-Type") == "" {
//			w.Header().Set("Content-Type", "application/json")
//		}
//		w.WriteHeader(rs.StatusCode)
		r.Response.Header.SetStatusCode(rs.StatusCode)

		encoder := json.NewEncoder(r.Response.BodyWriter())
		err := encoder.Encode(rs.Output)
		if err != nil {
			return err
		}
	}
	return nil
}
