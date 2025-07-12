package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/agent-platform/adk"
	"github.com/yeeaiclub/a2a-go/sdk/client"
	"github.com/yeeaiclub/a2a-go/sdk/types"
)

// RemoteAgentConnection represents a remote agent's connection and card
// (In a real implementation, this would manage HTTP/gRPC clients, etc.)
type RemoteAgentConnection struct {
	AgentCard *types.AgentCard
	AgentURL  string
	Client    *client.A2AClient
}

// RoutingAgent is responsible for delegating tasks to remote agents
// and managing agent cards and connections.
type RoutingAgent struct {
	taskCallback           func(taskID string, update any)
	remoteAgentConnections map[string]*RemoteAgentConnection
	cards                  map[string]*types.AgentCard
	agents                 string
	mu                     sync.Mutex
}

// NewRoutingAgent creates and initializes a RoutingAgent
func NewRoutingAgent(remoteAgentAddresses []string, taskCallback func(taskID string, update any)) *RoutingAgent {
	ra := &RoutingAgent{
		taskCallback:           taskCallback,
		remoteAgentConnections: make(map[string]*RemoteAgentConnection),
		cards:                  make(map[string]*types.AgentCard),
	}
	ra.initComponents(remoteAgentAddresses)
	return ra
}

// initComponents initializes remote agent connections and cards using a2a-go
func (ra *RoutingAgent) initComponents(remoteAgentAddresses []string) {
	for _, addr := range remoteAgentAddresses {
		if addr == "" {
			continue
		}
		httpClient := &http.Client{}
		resolver := client.NewA2ACardResolver(httpClient, addr)
		card, err := resolver.GetAgentCard()
		if err != nil {
			log.Printf("Failed to get agent card from %s: %v", addr, err)
			continue
		}
		a2aClient := client.NewClient(httpClient, addr+"/api")
		ra.remoteAgentConnections[card.Name] = &RemoteAgentConnection{
			AgentCard: &card,
			AgentURL:  addr,
			Client:    a2aClient,
		}
		ra.cards[card.Name] = &card
	}
	// Build agents string for display
	var agentInfo []string
	for _, card := range ra.cards {
		b, _ := json.Marshal(card)
		agentInfo = append(agentInfo, string(b))
	}
	ra.agents = fmt.Sprintf("%s", agentInfo)
}

// ListRemoteAgents returns all remote agent card info
func (ra *RoutingAgent) ListRemoteAgents() []map[string]string {
	var result []map[string]string
	for _, card := range ra.cards {
		result = append(result, map[string]string{
			"name":        card.Name,
			"description": card.Description,
		})
	}
	return result
}

// SendMessage sends a message to a specified agent using a2a-go
func (ra *RoutingAgent) SendMessage(agentName, task string, state map[string]interface{}) (*adk.Event, error) {
	ra.mu.Lock()
	defer ra.mu.Unlock()
	conn, ok := ra.remoteAgentConnections[agentName]
	if !ok || conn.Client == nil {
		return nil, fmt.Errorf("agent %s not found or not initialized", agentName)
	}
	msg := &types.Message{
		TaskID: "1", // You can generate unique IDs as needed
		Role:   types.User,
		Parts: []types.Part{
			&types.TextPart{Kind: "text", Text: task},
		},
	}
	resp, err := conn.Client.SendMessage(types.MessageSendParam{Message: msg})
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("agent error: %v", resp.Error)
	}
	taskObj, err := types.MapTo[types.Task](resp.Result)
	if err != nil {
		return nil, err
	}
	return &adk.Event{
		Content: &adk.Content{
			Parts: []*adk.Part{
				{Text: fmt.Sprintf("Agent %s result: %s", agentName, taskObj.Id)},
			},
		},
	}, nil
}

// RootInstruction returns the routing agent's instruction description
func (ra *RoutingAgent) RootInstruction(state map[string]interface{}) string {
	activeAgent := "None"
	if v, ok := state["active_agent"]; ok {
		activeAgent = fmt.Sprintf("%v", v)
	}
	return fmt.Sprintf(`
**Role:** You are an expert Routing Delegator. Your primary function is to accurately delegate user inquiries regarding weather or accommodations to the appropriate specialized remote agents.

**Agent Roster:**
* Available Agents: %s
* Currently Active Agent: %s
`, ra.agents, activeAgent)
}

// CheckActiveAgent checks the currently active agent
func (ra *RoutingAgent) CheckActiveAgent(state map[string]interface{}) string {
	if v, ok := state["active_agent"]; ok {
		return fmt.Sprintf("%v", v)
	}
	return "None"
}
