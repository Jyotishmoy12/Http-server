package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║     HTTP Request Debugger - Testing Various Formats   ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝\n")

	testCases := []struct {
		name    string
		request string
	}{
		{
			name: "Browser-like Request (Chrome/Edge)",
			request: "GET / HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"User-Agent: Mozilla/5.0\r\n" +
				"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\r\n" +
				"Accept-Language: en-US,en;q=0.5\r\n" +
				"Accept-Encoding: gzip, deflate\r\n" +
				"Connection: keep-alive\r\n" +
				"\r\n",
		},
		{
			name: "Minimal Valid HTTP/1.1",
			request: "GET / HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n",
		},
		{
			name: "HTTP/1.0 Style (no Host required)",
			request: "GET / HTTP/1.0\r\n" +
				"\r\n",
		},
		{
			name: "With User-Agent Only",
			request: "GET / HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"User-Agent: TestClient/1.0\r\n" +
				"\r\n",
		},
		{
			name: "Test /yourproblem endpoint",
			request: "GET /yourproblem HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"User-Agent: TestClient/1.0\r\n" +
				"\r\n",
		},
		{
			name: "Absolute URI format",
			request: "GET http://localhost:42069/ HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n",
		},
	}

	for i, tc := range testCases {
		fmt.Printf("[Test %d/%d] %s\n", i+1, len(testCases), tc.name)
		fmt.Println("├─ Connecting to localhost:42069...")

		conn, err := net.DialTimeout("tcp", "localhost:42069", 2*time.Second)
		if err != nil {
			fmt.Printf("├─ ❌ Connection failed: %v\n", err)
			fmt.Println("└─ FAILED\n")
			continue
		}

		fmt.Println("├─ Connected!")
		fmt.Printf("├─ Sending %d bytes\n", len(tc.request))
		
		// Send the request
		n, err := conn.Write([]byte(tc.request))
		if err != nil {
			fmt.Printf("├─ ❌ Write failed: %v\n", err)
			conn.Close()
			fmt.Println("└─ FAILED\n")
			continue
		}
		fmt.Printf("├─ Wrote %d bytes successfully\n", n)

		// Give server time to process
		time.Sleep(200 * time.Millisecond)

		// Try to read response with timeout
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		reader := bufio.NewReader(conn)
		
		fmt.Println("├─ Reading response...")
		response := ""
		buf := make([]byte, 4096)
		
		readAttempts := 0
		for readAttempts < 5 {
			n, err := reader.Read(buf)
			if n > 0 {
				response += string(buf[:n])
			}
			if err != nil {
				break
			}
			readAttempts++
			if n == 0 {
				time.Sleep(100 * time.Millisecond)
			}
		}

		conn.Close()

		if len(response) > 0 {
			fmt.Println("├─ ✅ Got response!")
			fmt.Printf("├─ Response length: %d bytes\n", len(response))
			fmt.Println("│  ┌─── Response ───────────────────────")
			lines := splitLines(response)
			for _, line := range lines {
				if len(line) > 80 {
					fmt.Printf("│  │ %s...\n", line[:80])
				} else {
					fmt.Printf("│  │ %s\n", line)
				}
			}
			fmt.Println("│  └────────────────────────────────────")
			fmt.Println("└─ ✅ SUCCESS\n")
		} else {
			fmt.Println("├─ ⚠️  No response received (server may have closed connection)")
			fmt.Println("└─ FAILED\n")
		}

		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("Diagnosis:")
	fmt.Println("If no tests passed, your server's request parser likely:")
	fmt.Println("1. Is waiting for more data than what's sent")
	fmt.Println("2. Not detecting the end of headers (\\r\\n\\r\\n)")
	fmt.Println("3. Has a bug in reading from the TCP connection")
	fmt.Println("4. Not properly handling the request line or headers")
	fmt.Println()
	fmt.Println("Check your internal/request/request.go parser!")
	fmt.Println("═══════════════════════════════════════════════════════")
}

func splitLines(s string) []string {
	lines := []string{}
	current := ""
	
	for i := 0; i < len(s); i++ {
		if s[i] == '\r' && i+1 < len(s) && s[i+1] == '\n' {
			lines = append(lines, current)
			current = ""
			i++ // skip \n
		} else if s[i] == '\n' {
			lines = append(lines, current)
			current = ""
		} else if s[i] >= 32 && s[i] <= 126 {
			current += string(s[i])
		} else {
			current += fmt.Sprintf("\\x%02x", s[i])
		}
	}
	
	if current != "" {
		lines = append(lines, current)
	}
	
	return lines
}