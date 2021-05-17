package sequence

type CodeParameters struct {
	Name    string
	Prefix  string
	Suffix  string
	Padding int32
}

type EntryParameters struct {
	Meta map[string]string
}
