package sequence

type Owner struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Code struct {
	Id          string `json:"id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Owner       *Owner `json:"owner"`
	Name        string `json:"name"`
	Prefix      string `json:"prefix"`
	Padding     int32  `json:"padding"`
	Suffix      string `json:"suffix"`
	LastIndex   int64  `json:"last_index,omitempty"`
	LastEntryId string `json:"last_entry_id,omitempty"`
}

type CodeCollection struct {
	Codes []*Code `json:"codes"`
}

type Entry struct {
	Id        string            `json:"id"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
	CodeId    string            `json:"code_id"`
	Idx       int64             `json:"idx"`
	Value     string            `json:"value"`
	Meta      map[string]string `json:"meta,omitempty"`
	PrevId    string            `json:"prev_id,omitempty"`
	Sigs      []string          `json:"sigs"`
}

type EntryCollection struct {
	Entries []*Entry `json:"entries"`
}
