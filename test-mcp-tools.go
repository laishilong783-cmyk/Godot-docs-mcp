package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type rpcRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id,omitempty"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func main() {
	cmd := exec.Command(os.Args[1])
	cmd.Env = os.Environ()

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	cmd.Start()

	go io.Copy(io.Discard, stderr)

	reader := bufio.NewReader(stdout)

	send := func(req rpcRequest) *rpcResponse {
		data, _ := json.Marshal(req)
		fmt.Fprintln(stdin, string(data))
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		var resp rpcResponse
		json.Unmarshal([]byte(line), &resp)
		return &resp
	}

	// 1. Initialize
	send(rpcRequest{JSONRPC: "2.0", ID: 0, Method: "initialize", Params: map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo":      map[string]string{"name": "test", "version": "1.0"},
	}})

	// 2. Initialized notification
	data, _ := json.Marshal(rpcRequest{JSONRPC: "2.0", Method: "notifications/initialized"})
	fmt.Fprintln(stdin, string(data))

	// 3. List tools
	tools := send(rpcRequest{JSONRPC: "2.0", ID: 1, Method: "tools/list"})
	fmt.Println("=== Tools ===")
	printResult(tools)

	// 4. Search: CharacterBody2D
	search := send(rpcRequest{JSONRPC: "2.0", ID: 2, Method: "tools/call", Params: map[string]interface{}{
		"name": "godot_docs_search",
		"arguments": map[string]interface{}{
			"query":   "CharacterBody2D move_and_slide",
			"version": "4.4",
			"limit":   5,
		},
	}})
	fmt.Println("\n=== Search: CharacterBody2D move_and_slide ===")
	printResult(search)

	// 5. Get class
	class := send(rpcRequest{JSONRPC: "2.0", ID: 3, Method: "tools/call", Params: map[string]interface{}{
		"name": "godot_docs_get_class",
		"arguments": map[string]interface{}{
			"class_name":      "CharacterBody2D",
			"version":         "4.4",
			"include_members": true,
		},
	}})
	fmt.Println("\n=== Class: CharacterBody2D ===")
	printResult(class)

	// 6. Get method
	method := send(rpcRequest{JSONRPC: "2.0", ID: 4, Method: "tools/call", Params: map[string]interface{}{
		"name": "godot_docs_get_method",
		"arguments": map[string]interface{}{
			"class_name":  "CharacterBody2D",
			"method_name": "move_and_slide",
			"version":     "4.4",
		},
	}})
	fmt.Println("\n=== Method: CharacterBody2D.move_and_slide ===")
	printResult(method)

	// 7. Get page
	page := send(rpcRequest{JSONRPC: "2.0", ID: 5, Method: "tools/call", Params: map[string]interface{}{
		"name": "godot_docs_get_page",
		"arguments": map[string]interface{}{
			"path":      "classes/class_characterbody2d.rst",
			"version":   "4.4",
			"max_chars": 2000,
		},
	}})
	fmt.Println("\n=== Page: classes/class_characterbody2d.rst ===")
	printResult(page)

	// 8. Search TileMapLayer
	tile := send(rpcRequest{JSONRPC: "2.0", ID: 6, Method: "tools/call", Params: map[string]interface{}{
		"name": "godot_docs_search",
		"arguments": map[string]interface{}{
			"query":   "TileMapLayer set_cell",
			"version": "4.4",
			"limit":   5,
		},
	}})
	fmt.Println("\n=== Search: TileMapLayer set_cell ===")
	printResult(tile)

	cmd.Process.Kill()
}

func printResult(resp *rpcResponse) {
	if resp == nil {
		fmt.Println("null")
		return
	}
	if resp.Error != nil {
		fmt.Printf("Error: %s\n", resp.Error.Message)
		return
	}
	var pretty bytes.Buffer
	json.Indent(&pretty, resp.Result, "", "  ")
	fmt.Println(pretty.String())
}
