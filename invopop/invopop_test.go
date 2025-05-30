package invopop

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/flimzy/testy"
	"resty.dev/v3"
)

func Test_put(t *testing.T) {
	tests := []struct {
		name      string
		responder testy.HTTPResponder
		path      string
		body      interface{}
		err       string
	}{
		{
			name: "unmarshalable body",
			body: json.RawMessage("this is not JSON"),
			err:  `json: error calling MarshalJSON for type json.RawMessage: invalid character 'h' in literal true (expecting 'r')`,
		},
		{
			name: "network error",
			responder: func(*http.Request) (*http.Response, error) {
				return nil, errors.New("network error")
			},
			body: map[string]string{"foo": "bar"},
			err:  `Put "": network error`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Client{
				conn: resty.NewWithClient(testy.HTTPClient(tt.responder)),
			}
			err := c.put(context.Background(), tt.path, tt.body, tt.body)
			assert.EqualError(t, err, tt.err)
		})
	}
}

func TestHTTPClient(t *testing.T) {
	hc := new(http.Client)
	c := &Client{
		conn: resty.NewWithClient(hc),
	}
	assert.Equal(t, hc, c.HTTPClient())
}
