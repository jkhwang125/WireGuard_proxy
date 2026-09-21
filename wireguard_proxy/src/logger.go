package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// EventType defines the protocol category
type EventType string

const (
	EventHTTP  EventType = "HTTP"
	EventMySQL EventType = "MySQL"
	EventSSH   EventType = "SSH"
)

// AuditLog represents the unified JSON structure for all audit events
type AuditLog struct {
	Timestamp      string      `json:"timestamp"`
	Protocol       EventType   `json:"protocol"`
	ClientIP       string      `json:"client_ip"`
	TargetHost     string      `json:"target_host"`
	ActionTaken    string      `json:"action_taken"` // "ALLOW" or "BLOCK"
	Details        interface{} `json:"details"`      // Protocol-specific payload
}

// HTTPDetails captures HTTP-specific metadata
type HTTPDetails struct {
	Method        string            `json:"method"`
	URL           string            `json:"url"`
	Headers       map[string]string `json:"headers"`
	ContentSnippet string            `json:"content_snippet,omitempty"`
}

// MySQLDetails captures MySQL authentication and query data
type MySQLDetails struct {
	Username      string `json:"username"`
	AuthStatus    string `json:"auth_status"` // "SUCCESS" or "FAILED"
	Query         string `json:"query,omitempty"`
	ErrorResponse string `json:"error_response,omitempty"`
}

// SSHDetails captures SSH authentication, SFTP, and port-forwarding data
type SSHDetails struct {
	Username       string `json:"username"`
	AuthStatus     string `json:"auth_status"` // "SUCCESS" or "FAILED"
	ChannelType    string `json:"channel_type"` // "shell", "sftp", "direct-tcpip" (port forwarding)
	CommandOrPath  string `json:"command_or_path,omitempty"` // Executed command or SFTP file path
	ForwardInfo    string `json:"forward_info,omitempty"`    // Port forwarding host/port details
}

// LogEvent writes a structured JSON audit log to stdout (or a log file)
func LogEvent(protocol EventType, clientIP, targetHost, action string, details interface{}) {
	entry := AuditLog{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Protocol:    protocol,
		ClientIP:    clientIP,
		TargetHost:  targetHost,
		ActionTaken: action,
		Details:     details,
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode audit log: %v\n", err)
		return
	}

	
	// 1. Print live to stdout for terminal visibility
	fmt.Println(string(jsonData))

	// 2. Ensure the log directory exists
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
		return
	}

	// 3. Open the log file in append mode (creates it if it doesn't exist)
	logPath := filepath.Join(logDir, "audit.log")
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open audit log file: %v\n", err)
		return
	}
	defer file.Close()

	// 4. Write JSON payload followed by a newline character
	if _, err := file.WriteString(string(jsonData) + "\n"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to audit log file: %v\n", err)
	}
}

// --- Example Usage in your Proxy Handlers ---
/*func main() {
	// 1. Example HTTP Audit Log (Blocked Admin Path)
	LogEvent(EventHTTP, "192.168.1.100", "api.internal", "BLOCK", HTTPDetails{
		Method: "POST",
		URL:    "/api/v1/admin/users/delete",
		Headers: map[string]string{
			"User-Agent":   "Mozilla/5.0",
			"Content-Type": "application/json",
		},
		ContentSnippet: `{"user_id": 42, "force": true}`,
	})

	// 2. Example MySQL Audit Log (Blocked Query)
	LogEvent(EventMySQL, "192.168.1.100", "db.internal:3306", "BLOCK", MySQLDetails{
		Username:      "app_user",
		AuthStatus:    "SUCCESS",
		Query:         "DROP DATABASE production;",
		ErrorResponse: "Query blocked by policy: DROP operations are prohibited.",
	})

	// 3. Example SSH Audit Log (Port Forwarding / SFTP Action)
	LogEvent(EventSSH, "192.168.1.100", "gateway.internal:22", "ALLOW", SSHDetails{
		Username:      "jkhwang",
		AuthStatus:    "SUCCESS",
		ChannelType:   "direct-tcpip",
		ForwardInfo:   "Target bound to 127.0.0.1:5432 via tunnel",
	})
}
*/
