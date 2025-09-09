package alt_test

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/ONSdigital/dis-redirect-proxy/proxy/alt"
)

func TestChain(t *testing.T) {

	us1mux := http.NewServeMux()
	us1mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("this is upstream 1")) })
	us1mux.HandleFunc("/nf/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) })
	upstream1 := httptest.NewServer(us1mux)
	defer upstream1.Close()
	url1, err := url.Parse(upstream1.URL)
	if err != nil {
		t.Fatal(err)
	}

	us2mux := http.NewServeMux()
	us2mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("this is upstream 2")) })
	us2mux.HandleFunc("/nf/alt/nf", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) })
	upstream2 := httptest.NewServer(us2mux)
	defer upstream2.Close()
	url2, err := url.Parse(upstream2.URL)
	if err != nil {
		t.Fatal(err)
	}

	proxy1 := httputil.NewSingleHostReverseProxy(url1)
	proxy2 := httputil.NewSingleHostReverseProxy(url2)

	us3handler := func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("this is upstream 3")) }

	mux := http.NewServeMux()
	mux.Handle("/upstream1", proxy1)
	mux.Handle("/upstream2", proxy2)
	mux.Handle("/nf/direct", proxy1)

	altProxy := alt.Try(proxy1).WhenStatus(http.StatusNotFound).Then(proxy2).WhenStatus(http.StatusNotFound).Then(http.HandlerFunc(us3handler))
	mux.Handle("/nf/alt/", altProxy)
	mux.Handle("/found/alt", altProxy)

	testCall(t, mux, "/upstream1", "this is upstream 1", http.StatusOK)
	testCall(t, mux, "/upstream2", "this is upstream 2", http.StatusOK)
	testCall(t, mux, "/nf/direct", "", http.StatusNotFound)
	testCall(t, mux, "/found/alt", "this is upstream 1", http.StatusOK)
	testCall(t, mux, "/nf/alt/moo", "this is upstream 2", http.StatusOK)
	testCall(t, mux, "/nf/us2", "", http.StatusNotFound)
	testCall(t, mux, "/nf/alt/nf", "this is upstream 3", http.StatusOK)

}

func testCall(t *testing.T, mux *http.ServeMux, path, wantBody string, wantStatus int) {
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
	if w.Code != wantStatus {
		t.Errorf("status wasn't %d: %d", wantStatus, w.Code)
	}
	if wantBody != "" && w.Body.String() != wantBody {
		t.Errorf("body wasn't '%s': %s", wantBody, w.Body.String())
	}
}
