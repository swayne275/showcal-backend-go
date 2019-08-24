package clientapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// cause ioutil.ReadAll() in function under test to error
// TODO get error case working
type errReader int

func (errReader) Read(buf []byte) (n int, err error) {
	return 0, errors.New("test error")
}

func TestGetRequestBody(t *testing.T) {
	cases := []struct {
		name    string
		input   *http.Request
		want    string
		wantErr bool
	}{
		{
			name:    "'test' in body",
			input:   httptest.NewRequest(http.MethodGet, "/junk", strings.NewReader("test")),
			want:    "test",
			wantErr: false,
		},
		{
			name:    "test error",
			input:   httptest.NewRequest(http.MethodGet, "/junk", errReader(0)),
			want:    "",
			wantErr: true,
		},
	}

	for _, c := range cases {
		got, err := getRequestBody(*c.input)
		gotErr := (err != nil)

		if gotErr != c.wantErr {
			t.Errorf("incorrect output error for '%s': expected '%t', got '%t'",
				c.name, c.wantErr, gotErr)
		}

		if string(got) != c.want {
			t.Errorf("incorrect output for '%s': expected '%x', got '%x'",
				c.name, c.want, got)
		}
	}
}

func TestSetupCors(t *testing.T) {
	headers := []struct {
		key  string
		want string
	}{
		{
			key:  "Access-Control-Allow-Origin",
			want: "*",
		},
		{
			key:  "Access-Control-Allow-Methods",
			want: "POST, GET, OPTIONS",
		},
		{
			key:  "Access-Control-Allow-Headers",
			want: "Accept, Content-Type, Content-Length, Accept-Encoding",
		},
	}

	w := httptest.NewRecorder()
	setupCors(w)
	resp := w.Result()

	for _, c := range headers {
		got := resp.Header.Get(c.key)
		if got != c.want {
			t.Errorf("incorrect output for key '%s': expected '%s', got '%s'",
				c.key, c.want, got)
		}
	}
}
