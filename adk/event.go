package adk

type Event struct {
	Content      *Content
	Actions      *Actions
	ErrorMessage string
}

type Content struct {
	Parts []*Part
}

type Part struct {
	Text             string
	FunctionCall     *FunctionCall
	FunctionResponse *FunctionResponse
}

type FunctionCall struct {
	Name string
	Args map[string]interface{}
}

type FunctionResponse struct {
	Name     string
	Response interface{}
}

type Actions struct {
	Escalate bool
}

func (e *Event) IsFinalResponse() bool {
	return e.Actions != nil && e.Actions.Escalate
}
