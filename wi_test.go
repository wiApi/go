package wi

import (
	"net/http"
	"testing"
	"time"
)

func TestNew_defaults(t *testing.T) {
	c := New("test-key")
	if c.apiKey != "test-key" {
		t.Errorf("apiKey = %q, want %q", c.apiKey, "test-key")
	}
	if c.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, defaultBaseURL)
	}
	if c.http == nil {
		t.Fatal("http client is nil")
	}
	if c.http.Timeout != defaultTimeout {
		t.Errorf("timeout = %v, want %v", c.http.Timeout, defaultTimeout)
	}
}

func TestNew_withBaseURL(t *testing.T) {
	c := New("k", WithBaseURL("https://custom.example.com/"))
	// WithBaseURL trims trailing slashes
	if c.baseURL != "https://custom.example.com" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
}

func TestNew_withBaseURL_noTrailingSlash(t *testing.T) {
	c := New("k", WithBaseURL("https://custom.example.com"))
	if c.baseURL != "https://custom.example.com" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
}

func TestNew_withTimeout(t *testing.T) {
	c := New("k", WithTimeout(5*time.Second))
	if c.http.Timeout != 5*time.Second {
		t.Errorf("timeout = %v, want 5s", c.http.Timeout)
	}
}

func TestNew_withHTTPClient(t *testing.T) {
	hc := &http.Client{Timeout: 99 * time.Second}
	c := New("k", WithHTTPClient(hc))
	if c.http != hc {
		t.Error("http client was not replaced")
	}
}

func TestSession_id(t *testing.T) {
	c := New("k")
	s := c.Session("my-instance")
	if s == nil {
		t.Fatal("Session() returned nil")
	}
	if s.id != "my-instance" {
		t.Errorf("id = %q, want %q", s.id, "my-instance")
	}
}

func TestSession_clientRef(t *testing.T) {
	c := New("k")
	s := c.Session("abc")
	if s.client != c {
		t.Error("SessionClient.client does not point to parent Client")
	}
}

func TestSession_distinct(t *testing.T) {
	c := New("k")
	s1 := c.Session("a")
	s2 := c.Session("b")
	if s1 == s2 {
		t.Error("Session() returned same pointer for different ids")
	}
	if s1.id == s2.id {
		t.Error("sessions share the same id")
	}
}
