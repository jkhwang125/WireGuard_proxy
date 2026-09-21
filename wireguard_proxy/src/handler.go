package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	
)
func HandleProxyConnection(clientConn net.Conn) {
    defer clientConn.Close()

    clientReader := bufio.NewReader(clientConn)

    // 1. Parse and inspect the incoming request stream
    parsedReq, req, err := ParseAndInspectHTTP(clientReader)
    if err != nil {
        return
    }

    

    // --- 2. POLICY EVALUATION (RUNS FIRST FOR EVERYTHING) ---
    var allowed bool
    var action ActionConfig
    var reason string

    if parsedReq.Method == "CONNECT"{
        fmt.Printf("\n[>] INCOMING: CONNECT %s\n", parsedReq.URL)
        allowed, action, reason = EvaluateHTTPSPolicy(parsedReq, req)
    } else {
        fmt.Printf("\n[>] INCOMING: %s %s\n", parsedReq.Method, parsedReq.URL)
        allowed, action, reason = EvaluateHTTPPolicy(parsedReq, req)
    }

    // --- 3. HANDLE POLICY BLOCKS ---
    if !allowed {
        fmt.Printf("[!] POLICY BLOCKED: %s\n", reason)
        
        // HTTP/2 or HTTPS CONNECT block -> Send 403
        if parsedReq.Method == "CONNECT" {
	//if action.Type == "BLOCK"{
            status := action.StatusCode
            //if status == 0 { status = 403 }
	    htmlbody := fmt.Sprintf(action.ResponseBody)
            blockResp := fmt.Sprintf("HTTP/1.1 %d Forbidden\r\nContent-Type: text/html; charset=utf-8\r\nContent-Length: %d\r\n\r\n%s\n", status, len(htmlbody), htmlbody)
            clientConn.Write([]byte(blockResp))
	    clientConn.Close()
            return
        }

        // Plain HTTP Redirect block
        if action.Type == "BLOCK_REDIRECT" && action.RedirectURL != "" {
		status := action.StatusCode
		htmlbody:=fmt.Sprintf(action.ResponseBody)
		redirectResp := fmt.Sprintf("HTTP/1.1 %d Found\r\nContent-Type: text/html\r\nContent-Length: %d\r\n\r\n%s\n", status, len(htmlbody), htmlbody)
		clientConn.Write([]byte(redirectResp))
		return
	}

        // Plain HTTP Forbidden block
        blockResp := fmt.Sprintf("HTTP/1.1 403 Forbidden\r\nContent-Type: text/plain\r\n\r\n[L7 Proxy Block] %s\n", reason)
        clientConn.Write([]byte(blockResp))
        return
    }

    // --- 4. EXECUTE ALLOWED CONNECTIONS ---
    if parsedReq.Method == "CONNECT" {
        targetHost := parsedReq.URL
        if targetHost == "" && req != nil {
            targetHost = req.Host
        }

        targetHost = strings.TrimLeft(targetHost, "/")
        targetHost = strings.TrimPrefix(targetHost, "http://")
        targetHost = strings.TrimPrefix(targetHost, "https://")

        fmt.Printf("[<] FORWARDING TO: %s\n", targetHost)

        serverConn, err := net.Dial("tcp", targetHost)
        if err != nil {
            fmt.Printf("[-] Failed to connect to target: %v\n", err)
            clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
            return
        }
        defer serverConn.Close()

        _, err = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
        if err != nil {
            return
        }

        go func() {
            io.Copy(serverConn, clientReader)
        }()
        io.Copy(clientConn, serverConn)
        return
    }

    // Standard HTTP Forwarding
    targetHost := req.URL.Host
    if targetHost == "" {
        targetHost = req.Host
    }
    if targetHost == "" || strings.Contains(targetHost, "localhost:8080") || strings.Contains(targetHost, "127.0.0.1:8080") || strings.Contains(targetHost, ":8080") {
    	fmt.Printf("[!] Loop prevention triggered for target: %s\n", targetHost)
    	//http.Error(clientConn, "Bad Gateway: Proxy Loop Prevented", http.StatusBadGateway)
	_, _ = clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\nProxy Loop Prevented\n"))
	return
    }
    if targetHost == "" {
        fmt.Println("[-] Error: Could not determine target host")
        return
    }
    if _, _, err := net.SplitHostPort(targetHost); err != nil {
        targetHost = net.JoinHostPort(targetHost, "80")
    }

    fmt.Printf("[<] FORWARDING TO: %s\n", targetHost)

    serverConn, err := net.Dial("tcp", targetHost)
    if err != nil {
        fmt.Printf("[-] Failed to connect to target: %v\n", err)
        return
    }
    defer serverConn.Close()

    req.RequestURI = req.URL.RequestURI()
    err = req.Write(serverConn)
    if err != nil {
        return
    }

    _, _ = io.Copy(clientConn, serverConn)

}
