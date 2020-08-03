package httprouter

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDemoRoutes(r *Router) {
	r.Add(http.MethodGet, "/{key}", fText200("root %s", "key"))
	r.Add(http.MethodGet, "/{key}/some", fText200("root %s some", "key"))
	r.Add(http.MethodGet, "/", fText200("home page"))
	r.Add(http.MethodPost, "/{key}/some", fText200("POST root %s some", "key"))

	r.Add(http.MethodGet, "/foo", fText200("static1 page"))
	r.Add(http.MethodPost, "/foo", fText200("POST static1 page"))

	r.Add(http.MethodGet, "/foo/{bar}", fText200("foo %s", "bar"))
	r.Add(http.MethodPost, "/foo/{bar}", fText200("POST foo %s", "bar"))

	r.Add(http.MethodGet, "/foo/{bar}/foo/{baz}", fText200("foo %s foo %s", "bar", "baz"))
	r.Add(http.MethodPost, "/foo/{bar}/foo/{baz}", fText200("POST foo %s foo %s", "bar", "baz"))
	r.Add(http.MethodGet, "/foo/{bar}/foo/{baz}/foo", fText200("foo %s foo %s foo", "bar", "baz"))
	r.Add(http.MethodPost, "/foo/{bar}/foo/{baz}/foo", fText200("POST foo %s foo %s foo", "bar", "baz"))
	r.Add(http.MethodGet, "/foo/{bar}/foo/{baz}/bar", fText200("foo %s foo %s bar", "bar", "baz"))
	r.Add(http.MethodPost, "/foo/{bar}/foo/{baz}/bar", fText200("POST foo %s foo %s bar", "bar", "baz"))
}

func Test_TableTestRoutes(t *testing.T) {
	router := NewRouter()
	router.ErrorHandler = func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}
	router.NotFoundHandelr = func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}
	ts := httptest.NewServer(router)
	defer ts.Close()

	setupDemoRoutes(router)

	t.Log("[INFO] schema", router.store.String())

	tests := []struct {
		method string
		path   string
		status int
		body   string
	}{
		{"GET", "/", 200, "home page"},
		{"POST", "/", 404, ""},

		{"GET", "/some", 200, "root some"},
		{"POST", "/some", 404, ""},
		{"GET", "/some/", 200, "root some/"},

		{"GET", "/some/some", 200, "root some some"},

		{"POST", "/some/some", 200, "POST root some some"},

		{"GET", "/foo", 200, "static1 page"},
		{"POST", "/foo", 200, "POST static1 page"},

		{"GET", "/foo/", 200, "foo "},
		{"POST", "/foo/", 200, "POST foo "},
		{"GET", "/foo/bar", 200, "foo bar"},
		{"GET", "/foo/bar/baz", 200, "foo bar/baz"},
		{"POST", "/foo/bar", 200, "POST foo bar"},
		{"POST", "/foo/bar/baz", 200, "POST foo bar/baz"},

		{"GET", "/foo/bar/foo/baz", 200, "foo bar foo baz"},
		{"GET", "/foo/bar/foo/baz/foo", 200, "foo bar foo baz foo"},
		{"GET", "/foo/bar/foo/baz/", 200, "foo bar foo baz/"},
		{"GET", "/foo/bar/foo/baz/bar", 200, "foo bar foo baz bar"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("[%s]-%q", tt.method, tt.path), func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(tt.method, tt.path, &bytes.Buffer{})
			require.NoError(t, err)
			router.ServeHTTP(recorder, request)
			if assert.EqualValues(t, tt.status, recorder.Code) {
				require.EqualValues(t, tt.body, recorder.Body.String())
			}
		})
	}
}

func BenchmarkSimpleRouting(b *testing.B) {
	router := NewRouter()
	router.ErrorHandler = func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}
	router.NotFoundHandelr = func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}
	ts := httptest.NewServer(router)
	defer ts.Close()

	setupDemoRoutes(router)

	rand.Seed(time.Now().Unix())

	tests := []struct {
		method string
		path   string
		status int
		body   string
	}{
		{"GET", "/", 200, "home page"},
		{"POST", "/", 404, ""},

		{"GET", "/some", 200, "root some"},
		{"POST", "/some", 404, ""},
		{"GET", "/some/", 200, "root some/"},

		{"GET", "/some/some", 200, "root some some"},

		{"POST", "/some/some", 200, "POST root some some"},

		{"GET", "/foo", 200, "static1 page"},
		{"POST", "/foo", 200, "POST static1 page"},

		{"GET", "/foo/", 200, "foo "},
		{"POST", "/foo/", 200, "POST foo "},
		{"GET", "/foo/bar", 200, "foo bar"},
		{"GET", "/foo/bar/baz", 200, "foo bar/baz"},
		{"POST", "/foo/bar", 200, "POST foo bar"},
		{"POST", "/foo/bar/baz", 200, "POST foo bar/baz"},

		{"GET", "/foo/bar/foo/baz", 200, "foo bar foo baz"},
		{"GET", "/foo/bar/foo/baz/foo", 200, "foo bar foo baz foo"},
		{"GET", "/foo/bar/foo/baz/", 200, "foo bar foo baz/"},
		{"GET", "/foo/bar/foo/baz/bar", 200, "foo bar foo baz bar"},
	}

	rw := &nilResponseWriter{}

	for _, tt := range tests {
		b.Run(fmt.Sprintf("[%s]-%q", tt.method, tt.path), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				request, err := http.NewRequest(tt.method, tt.path, &bytes.Buffer{})
				if err != nil {
					b.Fatal(err)
				}
				router.ServeHTTP(rw, request)
			}
		})
	}
}

type nilResponseWriter struct{}

func (r *nilResponseWriter) Header() http.Header {
	return nil
}

func (r *nilResponseWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (r *nilResponseWriter) WriteHeader(statusCode int) {
}

func fText200(text string, keys ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		routeParams := ParsedParamsFromCtx(req.Context())
		args := []interface{}{}
		for _, key := range keys {
			args = append(args, routeParams[key])
		}
		fmt.Fprintf(w, text, args...)
		w.WriteHeader(http.StatusOK)
	}
}
