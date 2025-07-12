package adk

type Runner interface {
	RunAsync(userID, sessionID string, newMessage *Content) <-chan *Event
}
