package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockFileServer struct {
	http.Handler
	RootFn func() string
}

func (fs mockFileServer) Root() string {
	return fs.RootFn()
}

func TestDirAliasHandler(t *testing.T) {
	da := newDirAliases()
	da.FileServerFactory = fileServerFactory(func(root string) FileServer {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, r.URL.Path)
		})
		return &mockFileServer{
			Handler: h,
			RootFn: func() string {
				return root
			},
		}
	})

	tests := []struct {
		Alias  string
		Path   string
		Status int
	}{
		{"foo", "/bar", http.StatusOK},
		{"baz", "/", http.StatusOK},
		{"", "/", http.StatusNotFound},
	}
	for _, test := range tests {
		da.Put(test.Alias, "whatever")

		w := httptest.NewRecorder()
		r, err := http.NewRequest("GET", fmt.Sprintf("/%s%s", test.Alias, test.Path), nil)
		if err != nil {
			t.Fatal(err)
		}
		da.ServeHTTP(w, r)

		if w.Code != test.Status {
			t.Errorf("expected http status %d, got %d", test.Status, w.Code)
			continue
		}
		if w.Code != http.StatusOK {
			continue
		}
		path := w.Body.String()
		if path != test.Path {
			t.Errorf("expected %v, got %v", test.Path, path)
		}
	}
}