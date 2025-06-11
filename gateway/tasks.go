package gateway

import (
	"encoding/json"

	"github.com/invopop/gobl"
)

// MarshalFields is a helper method that will encode gobl field error
// objects into JSON and update the task with the details.
func (t *TaskResult) MarshalFields(in gobl.FieldErrors) error {
	var err error
	t.Fields, err = json.Marshal(in)
	return err
}

// UnmarshalFields tries to decode the JSON field from the Fault, or
// returns nil if there is none.
func (f *Fault) UnmarshalFields() (gobl.FieldErrors, error) {
	if len(f.Fields) == 0 {
		return nil, nil
	}
	gf := make(gobl.FieldErrors)
	if err := json.Unmarshal(f.Fields, &gf); err != nil {
		return nil, err
	}
	return gf, nil
}
