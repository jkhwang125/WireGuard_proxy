# WireGuard VPN and Policy-Based Proxy System — Project Guide

## 1. Project Overview and Architecture
- **Overview:** This system is an integrated security proxy infrastructure that securely receives client traffic via a WireGuard VPN tunnel, and filters, analyzes, verifies, and controls the traffic through a high-performance Go-based proxy and policy engine running within a Docker environment.

- **Architecture Components:**
  * Client: The source device that transmits traffic into the VPN tunnel using WireGuard configurations and key pairs.
  * WireGuard Tunnel & VPN Server: Handles encrypted tunneling and endpoint management within the Docker environment.
  * Policy Engine (policy.json, policy.go): Determines whether to allow traffic by referencing traffic control and filtering rules.
  * Validation & Proxy (handler.go, parser_http.go): Processes HTTP requests, performs CONNECT tunneling, and handles TLS interception.
  * HTTP/HTTPS Server: The final destination web server reached by the refined traffic.

## 2. Project File Structure
```text
proxy_submission/
├── run.sh             # Script to generate VPN keys and run the integrated Docker environment
├── Dockerfile         # Defines the single-container image build for the proxy and VPN server
├── docker-compose.yml # Multi-container orchestration and network settings
├── main.go            # Entrypoint for the proxy server
├── handler.go         # Core proxy handler logic
├── flowtrack.go       # IP and PORT number parsing
├── policy.go          # Policy engine logic (rule loading and evaluation)
├── config/
└─────── policy.json   # Traffic control and filtering rules definition file
├── parser_http.go     # HTTP traffic and header parsing
└── logger.go          # Management and tracking for system events and traffic logs
```

## 3. How to Build and Run the Project
- **Prerequisites:** Docker, Docker Compose, and WireGuard must be installed on the host system.
- **Key Generation and Execution:**
  The public and private keys required to activate WireGuard are automatically generated, and the automated setup and launch script will be executed.

```bash
chmod +x run.sh
./run.sh
```
The `chmod +x run.sh` command adds execution permissions to the `run.sh` file, and `./run.sh` starts the proxy and the Docker environment.

## 4. Test Methods and Results by Feature
- **TLS Traffic Decryption (HTTPS Access without Certificate Warnings):**
    * Test Method: Route HTTPS traffic to the proxy through a custom interception module (`handler.go`) utilizing a generated certificate.
    * Test Result: Confirmed that the web browser successfully accesses HTTPS websites without any security warnings.

 
- **Protocol Support (HTTP):**
    * Test Method: Transmit and proxy HTTP payloads using various methods (method, host, path).
    * Test Result: Website access was successful for HTTP traffic using non-blocked methods. However, traffic using methods, hosts, or paths designated for blocking was intercepted, a block message was displayed, and the request was redirected to a different error page.


## 5. Addressed Problems, Solutions, and Improvement/Expansion Plans
- **Addressed Problems and Solutions:**
    * Problem: Safely intercepting and validating encrypted HTTPS/TLS traffic without compromising the client's connection stability.
    * Solution: Solved by implementing a dynamic CONNECT method handler and TLS interception workflow (`handler.go`), and integrating it with the rule validation engine in `policy.json`.

- **Improvement and Expansion Plans:**
    * Optimizing the packet parsing performance of `parser_http.go` for high-load enterprise environments.
    * Real-time traffic monitoring utilizing `logger.go`.
