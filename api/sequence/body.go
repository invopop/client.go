package sequence

type CodeParameters struct {
	Name    string `json:"name"`
	Prefix  string `json:"prefix"`
	Suffix  string `json:"suffix"`
	Padding int32  `json:"padding"`
}

type EntryParameters struct {
	Meta map[string]string `json:"meta"`
}
