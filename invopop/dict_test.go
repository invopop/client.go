package invopop

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDictUnmarshalJSON(t *testing.T) {
	ex := `{
		"foo": "bar",
		"data": {
			"supplier": {
				"emails": {
					"0": {
						"addr": "must be a valid email address"
					}
				}
			}
		}
	}`
	dict := new(Dict)
	err := json.Unmarshal([]byte(ex), dict)
	require.NoError(t, err)
	assert.Equal(t, "bar", dict.Get("foo").Message())
	assert.Equal(t, "must be a valid email address", dict.Get("data.supplier.emails.0.addr").Message())
	assert.Empty(t, dict.Get("data.missing").Message())
	assert.Empty(t, dict.Get("random").Message())
}

func TestDictFlatten(t *testing.T) {
	ex := `{
		"foo": "bar",
		"data": {
			"supplier": {
				"emails": {
					"0": {
						"addr": "must be a valid email address"
					}
				}
			}
		}
	}`
	dict := new(Dict)
	err := json.Unmarshal([]byte(ex), dict)
	require.NoError(t, err)
	out := dict.Flatten()
	assert.Equal(t, "bar", out["foo"])
	assert.Equal(t, "must be a valid email address", out["data.supplier.emails.0.addr"])
}

func TestDictAdd(t *testing.T) {
	d := NewDict()
	assert.Nil(t, d.Get(""))
	d.Add("foo", "bar")
	assert.Equal(t, "bar", d.Get("foo").Message())

	d.Add("plural", map[string]any{
		"zero":  "no mice",
		"one":   "%s mouse",
		"other": "%s mice",
	})
	assert.Equal(t, "no mice", d.Get("plural.zero").Message())
	assert.Equal(t, "%s mice", d.Get("plural.other").Message())

	d.Add("bad", 10) // ignore
	assert.Nil(t, d.Get("bad"))

	d.Add("self", d)
	assert.Equal(t, "bar", d.Get("self.foo").Message())
}

func TestDictMerge(t *testing.T) {
	ex := `{
		"foo": "bar",
		"baz": {
			"qux": "quux",
			"plural": {
				"zero": "no mice",
				"one": "%s mouse",
				"other": "%s mice"
			}
		}
	}`
	d1 := new(Dict)
	require.NoError(t, json.Unmarshal([]byte(ex), d1))

	ex2 := `{
		"foo": "baz",
		"extra": "value"
	}`
	d2 := new(Dict)
	require.NoError(t, json.Unmarshal([]byte(ex2), d2))

	d1.Merge(nil) // does nothing

	d3 := new(Dict)
	d3.Merge(d2)
	assert.Equal(t, "value", d3.Get("extra").Message())

	d1.Merge(d2)
	assert.Equal(t, "bar", d1.Get("foo").Message(), "should not overwrite")
	assert.Equal(t, "value", d1.Get("extra").Message())
}
