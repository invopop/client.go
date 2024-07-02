package snippets

// NoteMessage contains the minimum snippet details of a message.
type NoteMessage struct {
	UUID  string `json:"uuid,omitempty"`
	Title string `json:"title"`
}
