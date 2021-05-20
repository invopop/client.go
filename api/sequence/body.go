package sequence

type CodeParameters struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Prefix  string `json:"prefix"`
	Suffix  string `json:"suffix"`
	Padding int32  `json:"padding"`
}

type EntryParameters struct {
	ID   string            `json:"id"`
	Meta map[string]string `json:"meta"`
}
