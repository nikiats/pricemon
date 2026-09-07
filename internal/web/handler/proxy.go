package handler

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func newProxy(rawURL, serviceToken string) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			r.Out.Host = target.Host
			r.SetXForwarded()
			r.Out.Header.Del("X-User-Id")
			r.Out.Header.Set("Authorization", "Bearer "+serviceToken)
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Error("upstream failed", "method", r.Method, "path", r.URL.Path, "error", err)
			writeError(w, http.StatusBadGateway, codeUpstreamUnavailable)
		},
	}, nil
}

func proxyAs(target http.Handler, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = path
		target.ServeHTTP(w, r)
	}
}
