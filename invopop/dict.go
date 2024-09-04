package invopop

import (
	"encoding/json"
	"strings"
)

// Dict helps manage a nested map of strings to either messages or
// or other dictionaries. This is useful for accessing error messages
// provided by endpoints that include a "fields" property.
//
// This is based on the Dict model included by the
// [ctxi18n](https://github.com/invopop/ctxi18n) project.
type Dict struct {
	msg     string
	entries map[string]*Dict
}

// NewDict instantiates a new dict object.
func NewDict() *Dict {
	return &Dict{
		entries: make(map[string]*Dict),
	}
}

// Add adds a new key value pair to the dictionary.
func (d *Dict) Add(key string, value any) {
	switch v := value.(type) {
	case string:
		d.entries[key] = &Dict{msg: v}
	case map[string]any:
		nd := NewDict()
		for k, row := range v {
			nd.Add(k, row)
		}
		d.entries[key] = nd
	case *Dict:
		d.entries[key] = v
	default:
		// ignore
	}
}

// Message returns the dictionary message or an empty string
// if the dictionary is nil.
func (d *Dict) Message() string {
	if d == nil {
		return ""
	}
	return d.msg
}

// Get recursively retrieves the dictionary at the provided key location.
func (d *Dict) Get(key string) *Dict {
	if d == nil {
		return nil
	}
	if key == "" {
		return nil
	}
	n := strings.SplitN(key, ".", 2)
	entry, ok := d.entries[n[0]]
	if !ok {
		return nil
	}
	if len(n) == 1 {
		return entry
	}
	return entry.Get(n[1])
}

// Merge combines the entries of the second dictionary into this one. If a
// key is duplicated in the second diction, the original value takes priority.
func (d *Dict) Merge(d2 *Dict) {
	if d2 == nil {
		return
	}
	if d.entries == nil {
		d.entries = make(map[string]*Dict)
	}
	for k, v := range d2.entries {
		if d.entries[k] == nil {
			d.entries[k] = v
			continue
		}
		d.entries[k].Merge(v)
	}
}

// Flatten returns a simple flat map of the dictionary entries. This might make
// it easier to list out all the error messages for user interfaces.
func (d *Dict) Flatten() map[string]string {
	if d == nil {
		return nil
	}
	if d.msg != "" {
		return map[string]string{"": d.msg}
	}
	m := make(map[string]string)
	for k, v := range d.entries {
		for kk, vv := range v.Flatten() {
			x := k
			if kk != "" {
				x += "." + kk
			}
			m[x] = vv
		}
	}
	return m
}

// UnmarshalJSON attempts to load the dictionary data from a JSON byte slice.
func (d *Dict) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if data[0] == '"' {
		d.msg = string(data[1 : len(data)-1])
		return nil
	}
	d.entries = make(map[string]*Dict)
	return json.Unmarshal(data, &d.entries)
}
