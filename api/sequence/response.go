package sequence

type Owner struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Code struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Prefix      string `json:"prefix"`
	Padding     int32  `json:"padding"`
	Suffix      string `json:"suffix"`
	LastIndex   int64  `json:"last_index,omitempty"`
	LastEntryID string `json:"last_entry_id,omitempty"`
}

type CodeCollection struct {
	Codes []*Code `json:"codes"`
}

type Entry struct {
	ID     string            `json:"id"`
	CodeID string            `json:"code_id"`
	Idx    int64             `json:"idx"`
	Value  string            `json:"value"`
	Meta   map[string]string `json:"meta,omitempty"`
	PrevID string            `json:"prev_id,omitempty"`
}

type EntryCollection struct {
	Entries []*Entry `json:"entries"`
}
