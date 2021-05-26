package sequence

// Code defines the structure of a code.
type Code struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Prefix      string `json:"prefix"`
	Padding     int32  `json:"padding"`
	Suffix      string `json:"suffix"`
	LastIndex   int64  `json:"last_index,omitempty"`
	LastEntryID string `json:"last_entry_id,omitempty"`
}

// CodeCollection defines the structure which holds a list of codes.
type CodeCollection struct {
	Codes []*Code `json:"codes"`
}

// Entry defines the structure of a entry.
type Entry struct {
	ID     string            `json:"id"`
	CodeID string            `json:"code_id"`
	Idx    int64             `json:"idx"`
	Value  string            `json:"value"`
	Meta   map[string]string `json:"meta,omitempty"`
	PrevID string            `json:"prev_id,omitempty"`
}

// EntryCollection defines the structure which holds a list of entries.
type EntryCollection struct {
	Entries []*Entry `json:"entries"`
}
