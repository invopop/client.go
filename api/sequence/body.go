package sequence

// CodeParameters defines the require fields to create a code.
type CodeParameters struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Prefix  string `json:"prefix"`
	Suffix  string `json:"suffix"`
	Padding int32  `json:"padding"`
}

// EntryParameters defines the require fields to create a entry.
type EntryParameters struct {
	ID   string            `json:"id"`
	Meta map[string]string `json:"meta"`
}
