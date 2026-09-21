package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"net"
	"regexp"
)

// --- STRUCTS MATCHING YOUR JSON SCHEMA ---

type PolicyConfig struct {
	Schema              string          `json:"$schema,omitempty"`
	PolicyVersion       string          `json:"policy_version,omitempty"`
	GlobalDefaultAction string          `json:"global_default_action,omitempty"`
	AuditLogging        bool            `json:"audit_logging"`
	Protocols           ProtocolsConfig `json:"protocols"`
}

type ProtocolsConfig struct {
	Http HttpProtocolConfig `json:"http"`
	Http2 HttpProtocolConfig `json:"http2"`
}

type HttpProtocolConfig struct {
	Enabled bool       `json:"enabled"`
	Rules   []HttpRule `json:"rules"`
}

type HttpRule struct {
	ID           string          `json:"id"`
	Match        HttpMatchConfig `json:"match"`
	Action       ActionConfig    `json:"action"`
	AuditMessage string          `json:"audit_message"`
}

type HttpMatchConfig struct {
	Methods     []string `json:"methods"`
	Hosts       []string `json:"hosts"`
	PathPattern string   `json:"path_pattern"`
}

type ActionConfig struct {
	Type             string            `json:"type"`
	RedirectURL      string            `json:"redirect_url,omitempty"`
	StatusCode       int               `json:"status_code,omitempty"`
	ResponseBody     string            `json:"response_body,omitempty"`
	InjectHeaders    map[string]string `json:"inject_headers,omitempty"`
	StripHeaders     []string          `json:"strip_headers,omitempty"`
	QueryReplacement string            `json:"query_replacement,omitempty"`
	CommandPrefix    string            `json:"command_prefix,omitempty"`
	ErrorCode        uint16            `json:"error_code,omitempty"`
	SQLState         string            `json:"sql_state,omitempty"`
	ErrorMessage     string            `json:"error_message,omitempty"`
}

// Global state holding the active loaded policy configuration
var loadedPolicy PolicyConfig

// LoadPolicies reads and parses rules from the config file
func LoadPolicies(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &loadedPolicy); err != nil {
		fmt.Printf("[-] Failed to read policy file: %v\n", err)
		return err
	}

	fmt.Printf("[*] Loaded HTTP rules: %d\n",
		len(loadedPolicy.Protocols.Http.Rules))

	return nil
}

// --- HTTP POLICY EVALUATION ---

func EvaluateHTTPPolicy(parsed *ParsedHTTP, req *http.Request) (bool, ActionConfig, string) {
	for _, rule := range loadedPolicy.Protocols.Http.Rules {
		if matchesRule(parsed, req, rule) {
			switch rule.Action.Type {
			case "BLOCK_PAGE", "BLOCK_REDIRECT":
				return false, rule.Action, rule.AuditMessage

			case "MODIFY":
				for key, val := range rule.Action.InjectHeaders {
					req.Header.Set(key, val)
				}
				for _, key := range rule.Action.StripHeaders {
					req.Header.Del(key)
				}
				fmt.Printf("[*] POLICY MODIFIED: %s\n", rule.AuditMessage)
			}
		}
	}
	return true, ActionConfig{}, ""
}

// --- HTTPS / HTTP2 POLICY EVALUATION ---
func EvaluateHTTPSPolicy(parsed *ParsedHTTP, req *http.Request) (bool, ActionConfig, string) {
    for _, rule := range loadedPolicy.Protocols.Http2.Rules {
        if matchesRule(parsed, req, rule) {
            switch rule.Action.Type {
            case "BLOCK", "BLOCK_PAGE", "BLOCK_REDIRECT":
                return false, rule.Action, rule.AuditMessage
            case "ALLOW":
                return true, rule.Action, rule.AuditMessage
            }
        }
    }

    // Fall back to global default action if no rule matches the CONNECT request
    if loadedPolicy.GlobalDefaultAction == "BLOCK" {
        return false, ActionConfig{Type: "BLOCK"}, "Global default policy action: BLOCK"
    }
    return true, ActionConfig{Type: "ALLOW"}, ""
}
func matchesRule(parsed *ParsedHTTP, req *http.Request, rule HttpRule) bool {
	// 1. Method Matching
    methodMatched := false
    for _, m := range rule.Match.Methods {
        if m == "*" || m == parsed.Method {
            methodMatched = true
            break
        }
    }
    if !methodMatched {
        return false
    }
	if len(rule.Match.Hosts) > 0 {
		hostMatched := false
		for _, h := range rule.Match.Hosts {
			if h == "*" || h == req.Host {
				hostMatched = true
				break
			}

			// Clean up req.Host by stripping the port if present (e.g., "restricted-archive.org:443" -> "restricted-archive.org")
			evalHost := req.Host
			if hostOnly, _, err := net.SplitHostPort(evalHost); err == nil {
				evalHost = hostOnly
			}

			if h == evalHost {
				hostMatched = true
				break
			}
		}
		if !hostMatched {
			return false
		}
	}

	if rule.Match.PathPattern != "" {
		targetPath := parsed.URL
		if parsedURL, err := url.Parse(parsed.URL); err == nil && parsedURL.Path != "" {
			targetPath = parsedURL.Path
		}
		matched, err := regexp.MatchString(rule.Match.PathPattern, targetPath)
		if err != nil || !matched {
			return false
		}
	}

	return true
}
