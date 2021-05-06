package api

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

/*-------------------------*/

func TestDockerStatus(t *testing.T) {

	tt := []struct {
		name   string
		needle string // a substring that should be found in the output content
		status int
		err    string
	}{
		{
			name:   "DockerStatus API",
			needle: "waziup.wazigate-system",
			status: 200,
		},
	}

	for _, tc := range tt {

		t.Run(tc.name, func(t *testing.T) {

			req, err := http.NewRequest("GET", "localhost", nil)
			if err != nil {
				t.Fatalf("Could not create the request: %v", err)
			}

			rec := httptest.NewRecorder()

			DockerStatus(rec, req, nil)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.status {
				t.Errorf("Expected status %v but got: %v", tc.status, res.StatusCode)
			}

			content, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("Could not read the content: %v", err)
			}

			if tc.err != "" && tc.err != string(content) {
				t.Errorf("Expected error message %q; got %q", tc.err, string(content))
			}

			if tc.needle != "" && !strings.Contains(string(content), tc.needle) {
				t.Fatalf("Expected to find %q in the content", tc.needle)
			}

		})
	}

}

/*-------------------------*/
