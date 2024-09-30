package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/Shopify/go-lua"

	_ "github.com/joho/godotenv/autoload"
)

// "Vendored" from stdlib
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

func main() {
	l := lua.NewState()

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("No PORT environment variable provided")
	}

	lua.OpenLibraries(l)
	if err := lua.DoFile(l, "config.lua"); err != nil {
		panic(err)
	}

	l.Global("TARGET")
	t, _ := l.ToString(-1)
	l.Global("API_KEY")
	apiKey, _ := l.ToString(-1)

	target, _ := url.Parse(t)

	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			req := r.Out

			reqUrl := &url.URL{}

			l.Global("Handle")
			l.PushString(r.In.URL.Path)
			l.PushString(r.In.URL.RawQuery)
			l.Call(2, 0)

			targetQuery := target.RawQuery
			reqUrl.Scheme = target.Scheme
			reqUrl.Host = target.Host
			reqUrl.Path, reqUrl.RawPath = joinURLPath(target, req.URL)
			if targetQuery == "" || req.URL.RawQuery == "" {
				reqUrl.RawQuery = targetQuery + req.URL.RawQuery
			} else {
				reqUrl.RawQuery = targetQuery + "&" + req.URL.RawQuery
			}

			req.Header.Set("X-API-Key", apiKey)

			r.SetURL(reqUrl)
			r.SetXForwarded()
		},
	}

	done := make(chan bool)
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprint(":", port), proxy))
	}()
	log.Printf("App is listening on http://localhost:%s/\n", port)
	<-done
}
