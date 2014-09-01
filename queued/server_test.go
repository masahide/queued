package queued

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bmizerany/assert"
)

func TestServer(t *testing.T) {
	s := NewServer(&Config{DbPath: "./server1", Store: "memory"})
	defer s.ItemStore.Drop()
	defer s.QueueStore.Drop()

	// Enqueue
	body := strings.NewReader("bar")
	req, _ := http.NewRequest("POST", "/foo", body)
	req.Header.Add("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	assert.Equal(t, w.Code, 201)

	// Invalid complete (must dequeue first)
	req, _ = http.NewRequest("DELETE", "/foo/1", nil)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)

	assert.Equal(t, w.Code, 400)

	// Info
	req, _ = http.NewRequest("GET", "/foo/1", nil)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)

	assert.Equal(t, w.Code, 200)
	assert.Equal(t, w.Header().Get("Content-Type"), "text/plain")

	// Dequeue
	req, _ = http.NewRequest("POST", "/foo/dequeue?wait=30&timeout=30", nil)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)

	assert.Equal(t, w.Code, 200)
	assert.Equal(t, w.Header().Get("Content-Type"), "text/plain")

	// Stats
	req, _ = http.NewRequest("GET", "/foo", nil)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)

	assert.Equal(t, w.Code, 200)

	// Complete
	req, _ = http.NewRequest("DELETE", "/foo/1", nil)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)

	assert.Equal(t, w.Code, 204)

	// Info not found
	req, _ = http.NewRequest("GET", "/foo/1", nil)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)

	assert.Equal(t, w.Code, 404)
}

func TestServerAuth(t *testing.T) {
	s := NewServer(&Config{DbPath: "./server2", Auth: "secret", Store: "memory"})
	defer s.ItemStore.Drop()
	defer s.QueueStore.Drop()

	body := strings.NewReader("bar")

	req, _ := http.NewRequest("POST", "/foo", body)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	assert.Equal(t, w.Code, 401)

	req, _ = http.NewRequest("POST", "/foo", body)
	req.SetBasicAuth("", "secret")
	w = httptest.NewRecorder()
	s.ServeHTTP(w, req)

	assert.Equal(t, w.Code, 201)
}
