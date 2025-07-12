package adk

import "time"

type MockSessionService struct{}

func (m *MockSessionService) CreateSession(appName, userID, sessionID string) error {
	return nil
}

type MockRunner struct{}

func (m *MockRunner) RunAsync(userID, sessionID string, newMessage *Content) <-chan *Event {
	ch := make(chan *Event)
	go func() {
		defer close(ch)
		// Simulate a tool call
		ch <- &Event{
			Content: &Content{
				Parts: []*Part{
					{FunctionCall: &FunctionCall{Name: "weather", Args: map[string]interface{}{"city": "Beijing"}}},
				},
			},
		}
		time.Sleep(1 * time.Second)
		// Simulate a tool response
		ch <- &Event{
			Content: &Content{
				Parts: []*Part{
					{FunctionResponse: &FunctionResponse{Name: "weather", Response: map[string]interface{}{"temp": 25, "desc": "Sunny"}}},
				},
			},
		}
		time.Sleep(1 * time.Second)
		// Final response
		ch <- &Event{
			Content: &Content{
				Parts: []*Part{
					{Text: "The weather in Beijing is sunny, 25°C."},
				},
			},
		}
	}()
	return ch
}
