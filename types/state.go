package types

type WeaveState struct {
	PreviousResponse []string
}

func NewWeaveState() WeaveState {
	return WeaveState{
		PreviousResponse: make([]string, 0),
	}
}

func (w *WeaveState) Render() string {
	render := ""
	for _, r := range w.PreviousResponse {
		render += r
	}
	return render
}

func (w *WeaveState) PopPreviousResponse() {
	l := len(w.PreviousResponse)
	if l == 0 {
		return
	}

	w.PreviousResponse = w.PreviousResponse[:l-1]
}

func (w *WeaveState) PushPreviousResponse(s string) {
	w.PreviousResponse = append(w.PreviousResponse, s)
}
