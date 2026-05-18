package updater

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckAndUpdateDevSkipsNetwork(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	orig := http.DefaultClient.Transport
	http.DefaultClient.Transport = assertNoRequest{t: t, called: &called}
	defer func() { http.DefaultClient.Transport = orig }()

	CheckAndUpdate("dev")

	if called {
		t.Fatal("CheckAndUpdate(\"dev\") made a network request")
	}
}

type assertNoRequest struct {
	t      *testing.T
	called *bool
}

func (a assertNoRequest) RoundTrip(r *http.Request) (*http.Response, error) {
	*a.called = true
	a.t.Errorf("unexpected HTTP request to %s", r.URL)
	return nil, nil
}
