package main

import (
	"fmt"
	"net"
)





func main() {
	// 1. Load policies from config folder at startup
	err := LoadPolicies("config/policy.json")
	if err != nil {
		fmt.Printf("[-] Warning: Failed to load policy.json: %v. Running with empty rules.\n", err)
	} else {
		fmt.Println("[+] Loaded policy rules from config/policy.json")
	}
	// 10.13.13.1 - address of my wg0(WireGuard)
	// 2. Start HTTP/WebSocket/SSE Proxy Listener on port 8080
	httpListener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		fmt.Printf("[-] Failed to start HTTP listener on 8080: %v\n", err)
		return
	}
	defer httpListener.Close()
	fmt.Println("[+] L7 HTTP/WS Proxy running on :8080...")

	// 3. Infinite loop for HTTP connections on port 8080
	for {
		conn, err := httpListener.Accept()
		if err != nil {
			fmt.Printf("[-] Connection accept error: %v\n", err)
			continue
		}

		go HandleProxyConnection(conn)
	}
}



