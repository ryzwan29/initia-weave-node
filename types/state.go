package types

type WeaveState struct {
	PreviousResponse []string
}

// NewWeaveState initializes a new WeaveState with an empty PreviousResponse slice.
func NewWeaveState() WeaveState {
	return WeaveState{
		PreviousResponse: make([]string, 0),
	}
}

// Clone creates a deep copy of WeaveState, duplicating the PreviousResponse slice.
func (w WeaveState) Clone() WeaveState {
	// Create a copy of the PreviousResponse slice
	clonedResponses := make([]string, len(w.PreviousResponse))
	copy(clonedResponses, w.PreviousResponse)

	return WeaveState{
		PreviousResponse: clonedResponses,
	}
}

// Render concatenates all responses into a single string.
func (w *WeaveState) Render() string {
	render := ""
	for _, r := range w.PreviousResponse {
		render += r
	}
	return render
}

// PopPreviousResponse removes the last response in the PreviousResponse slice.
func (w *WeaveState) PopPreviousResponse() {
	l := len(w.PreviousResponse)
	if l == 0 {
		return
	}
	w.PreviousResponse = w.PreviousResponse[:l-1]
}

// PushPreviousResponse adds a response to the end of the PreviousResponse slice.
func (w *WeaveState) PushPreviousResponse(s string) {
	w.PreviousResponse = append(w.PreviousResponse, s)
}

// PopPreviousResponseAtIndex removes a response at a specific index. Use with care.
func (w *WeaveState) PopPreviousResponseAtIndex(index int) {
	l := len(w.PreviousResponse)
	if index < 0 || index >= l {
		return
	}
	w.PreviousResponse = append(w.PreviousResponse[:index], w.PreviousResponse[index+1:]...)
}
