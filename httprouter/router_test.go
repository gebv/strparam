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
		{"GET", "/some/", 404, ""},

		{"GET", "/some/some", 200, "root some some"},
		// TODO: skiped because http server jumps to the dirictory higher ?? `//`
		// {"GET", "//some", 200, "root  some"},

		{"POST", "/some/some", 200, "POST root some some"},

		{"GET", "/foo", 200, "static1 page"},
		{"POST", "/foo", 200, "POST static1 page"},

		{"GET", "/foo/", 200, "foo "},
		{"POST", "/foo/", 200, "POST foo "},
		{"GET", "/foo/bar", 200, "foo bar"},
		{"GET", "/foo/bar/baz", 404, ""},
		{"POST", "/foo/bar", 200, "POST foo bar"},
		{"POST", "/foo/bar/baz", 404, ""},

		{"GET", "/foo//foo/", 200, "foo  foo "},
		{"GET", "/foo//foo", 404, ""},
		{"GET", "/foo/foo/", 404, ""},
		{"GET", "/foo/foo", 200, "foo foo"},

		{"GET", "/foo/bar/foo/", 200, "foo bar foo "},
		{"GET", "/foo/bar/foo/baz", 200, "foo bar foo baz"},
		{"GET", "/foo/bar/foo/baz/foo", 200, "foo bar foo baz foo"},
		{"GET", "/foo/bar/foo//foo", 200, "foo bar foo  foo"},
		{"GET", "/foo//foo//foo", 200, "foo  foo  foo"},
		{"GET", "/foo/bar/foo/baz/foo/", 404, ""},
		{"GET", "/foo/bar/foo/baz/foo/abc", 404, ""},
		{"GET", "/foo/bar/foo/baz/", 404, ""},

		{"GET", "/foo/bar/foo/baz/bar", 200, "foo bar foo baz bar"},
		{"GET", "/foo/bar/foo/baz/bar/", 404, ""},
		{"GET", "/foo/bar/foo/baz/bar/abc", 404, ""},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("[%s]-%q", tt.method, tt.path), func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(tt.method, tt.path, &bytes.Buffer{})
			t.Logf("Handle request path %q", request.URL.Path)
			require.NoError(t, err)
			router.ServeHTTP(recorder, request)
			if assert.EqualValues(t, tt.status, recorder.Code) {
				require.EqualValues(t, tt.body, recorder.Body.String())
			}
		})
	}
}

func Test_issues7(t *testing.T) {
	// https://github.com/gebv/strparam/issues/7
	r := NewRouter()
	r.NotFoundHandelr = fText200("not found")
	r.Add(http.MethodGet, "/{foobar}", fText200("/{foobar} %v", "foobar"))
	r.Add(http.MethodGet, "/", fText200("/"))
	r.Add(http.MethodGet, "/foo/{bar}/", fText200("/foo/{bar} %v", "bar"))
	r.Add(http.MethodGet, "/a", fText200("/a"))
	r.Add(http.MethodGet, "/a/1", fText200("/a/1"))
	r.Add(http.MethodGet, "/a/1/{param}", fText200("/a/1/{param} %v", "param"))
	r.Add(http.MethodGet, "/b", fText200("/b"))
	r.Add(http.MethodGet, "/b/1", fText200("/b/1"))
	r.Add(http.MethodGet, "/b/1/{param}", fText200("/b/1/{param} %v", "param"))

	t.Log("[INFO] schema", r.store.String())

	cases := []struct {
		in       string
		wantBody string
	}{
		{"/b/1/asd", "/b/1/{param} asd"},
		{"/b/1", "/b/1"},
		{"/b/", "not found"},
		{"/baz", "/{foobar} baz"},
	}

	for _, case_ := range cases {
		t.Run(case_.in, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request, err := http.NewRequest("GET", case_.in, nil)
			require.NoError(t, err)
			r.ServeHTTP(recorder, request)
			assert.EqualValues(t, case_.wantBody, recorder.Body.String())
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
		// {"GET", "/", 200, "home page"},
		// {"POST", "/", 404, ""},

		// {"GET", "/some", 200, "root some"},
		// {"POST", "/some", 404, ""},
		// {"GET", "/some/", 200, "root some/"},

		// {"GET", "/some/some", 200, "root some some"},

		// {"POST", "/some/some", 200, "POST root some some"},

		// {"GET", "/foo", 200, "static1 page"},
		// {"POST", "/foo", 200, "POST static1 page"},

		// {"GET", "/foo/", 200, "foo "},
		// {"POST", "/foo/", 200, "POST foo "},
		// {"GET", "/foo/bar", 200, "foo bar"},
		// {"GET", "/foo/bar/baz", 200, "foo bar/baz"},
		// {"POST", "/foo/bar", 200, "POST foo bar"},
		// {"POST", "/foo/bar/baz", 200, "POST foo bar/baz"},

		// {"GET", "/foo/bar/foo/baz", 200, "foo bar foo baz"},
		// {"GET", "/foo/bar/foo/baz/foo", 200, "foo bar foo baz foo"},
		// {"GET", "/foo/bar/foo/baz/", 200, "foo bar foo baz/"},
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
