package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/agent-platform/adk"
	agent "github.com/agent-platform/routing"
)

var (
	sessionService = &adk.MockSessionService{}
	runner         = &adk.MockRunner{}
	routingAgent   *agent.RoutingAgent
)

type ChatRequest struct {
	Message   string `json:"message"`
	AgentName string `json:"agent_name,omitempty"`
}

type AgentListResponse struct {
	Agents []map[string]string `json:"agents"`
}

func main() {
	// Load configuration from TOML file
	config, err := LoadConfig("")
	if err != nil {
		log.Printf("Warning: Failed to load config.toml, falling back to environment variables: %v", err)
		// Fallback to environment variables if config file is not found
		routingAgent = agent.NewRoutingAgent([]string{
			os.Getenv("AIR_AGENT_URL"),
			os.Getenv("WEA_AGENT_URL"),
		}, nil)
	} else {
		// Initialize RoutingAgent with URLs from TOML config
		routingAgent = agent.NewRoutingAgent(config.GetAgentURLs(), nil)
		log.Printf("Loaded %d agents from config.toml", len(config.Agents))
	}

	// Try to find static directory - first try relative to current working directory
	staticDir := "./static"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		// If not found, try relative to the server directory
		execDir, err := os.Executable()
		if err == nil {
			serverDir := filepath.Dir(execDir)
			staticDir = filepath.Join(serverDir, "..", "static")
		}
	}

	// Serve static files
	http.Handle("/", http.FileServer(http.Dir(staticDir)))
	http.HandleFunc("/api/chat", chatHandler)
	http.HandleFunc("/api/agents", agentsHandler)

	log.Println("Server started at :8083")
	log.Printf("Serving static files from: %s", staticDir)
	log.Fatal(http.ListenAndServe(":8083", nil))
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	var req ChatRequest
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	sessionService.CreateSession("routing_app", "default_user", "default_session")

	// If agent_name is provided, use RoutingAgent to send message
	if req.AgentName != "" {
		state := map[string]interface{}{"active_agent": req.AgentName}
		event, err := routingAgent.SendMessage(req.AgentName, req.Message, state)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(event)
		flusher.Flush()
		return
	}

	// Default: use mock runner for streaming
	eventCh := runner.RunAsync("default_user", "default_session", &adk.Content{
		Parts: []*adk.Part{{Text: req.Message}},
	})

	encoder := json.NewEncoder(w)
	for event := range eventCh {
		encoder.Encode(event)
		flusher.Flush()
	}
}

func agentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	agents := routingAgent.ListRemoteAgents()
	json.NewEncoder(w).Encode(AgentListResponse{Agents: agents})
}
