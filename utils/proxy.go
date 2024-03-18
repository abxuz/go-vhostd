package utils

import "net/http"

type ReverseProxyDirector = func(req *http.Request) (resp *http.Response, header http.Header, err error)

type ReverseProxyTransport struct {
	Director       ReverseProxyDirector
	HttpTransport  http.RoundTripper
	Http3Transport http.RoundTripper
}

func (t *ReverseProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, header, err := t.Director(req)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		if req.URL.Scheme == "http3" {
			req.URL.Scheme = "https"
			resp, err = t.Http3Transport.RoundTrip(req)
		} else {
			resp, err = t.HttpTransport.RoundTrip(req)
		}
		if err != nil {
			return resp, err
		}
	}

	if resp.Header == nil {
		resp.Header = header
		return resp, nil
	}

	for k, vs := range header {
		for _, v := range vs {
			resp.Header.Add(k, v)
		}
	}
	return resp, nil
}
