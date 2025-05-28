package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	configDir  string
	savedTests = make(map[string]SavedTest)
)

type SavedTest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

func main() {
	home, _ := os.UserHomeDir()
	configDir = filepath.Join(home, ".apitester")
	os.MkdirAll(configDir, 0755)

	loadSavedTests()

	rootCmd := &cobra.Command{
		Use:   "apitester",
		Short: "A lightweight CLI alternative to Postman",
	}

	sendCmd := &cobra.Command{
		Use:   "send",
		Short: "Send an HTTP request",
		Run:   sendHandler,
	}
	sendCmd.Flags().StringP("method", "X", "GET", "HTTP method")
	sendCmd.Flags().StringP("url", "u", "", "Request URL (required)")
	sendCmd.Flags().StringP("data", "d", "", "Request body")
	sendCmd.Flags().StringToStringP("headers", "H", map[string]string{}, "Request headers")
	rootCmd.AddCommand(sendCmd)

	saveCmd := &cobra.Command{
		Use:   "save [name]",
		Short: "Save a request for later use",
		Args:  cobra.ExactArgs(1),
		Run:   saveHandler,
	}
	saveCmd.Flags().StringP("method", "X", "GET", "HTTP method")
	saveCmd.Flags().StringP("url", "u", "", "Request URL (required)")
	saveCmd.Flags().StringP("data", "d", "", "Request body")
	saveCmd.Flags().StringToStringP("headers", "H", map[string]string{}, "Request headers")
	rootCmd.AddCommand(saveCmd)

	runCmd := &cobra.Command{
		Use:   "run [name]",
		Short: "Run a saved request",
		Args:  cobra.ExactArgs(1),
		Run:   runHandler,
	}
	rootCmd.AddCommand(runCmd)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all saved requests",
		Run:   listHandler,
	}
	rootCmd.AddCommand(listCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func sendHandler(cmd *cobra.Command, args []string) {
	method, _ := cmd.Flags().GetString("method")
	url, _ := cmd.Flags().GetString("url")
	data, _ := cmd.Flags().GetString("data")
	headers, _ := cmd.Flags().GetStringToString("headers")

	if url == "" {
		fmt.Println("Error: URL is required")
		os.Exit(1)
	}

	sendRequest(method, url, data, headers)
}

func saveHandler(cmd *cobra.Command, args []string) {
	name := args[0]
	method, _ := cmd.Flags().GetString("method")
	url, _ := cmd.Flags().GetString("url")
	data, _ := cmd.Flags().GetString("data")
	headers, _ := cmd.Flags().GetStringToString("headers")

	if url == "" {
		fmt.Println("Error: URL is required")
		os.Exit(1)
	}

	savedTests[name] = SavedTest{
		URL:     url,
		Method:  strings.ToUpper(method),
		Headers: headers,
		Body:    data,
	}

	saveTestsToFile()
	fmt.Printf("Saved test '%s'\n", name)
}

func runHandler(cmd *cobra.Command, args []string) {
	name := args[0]
	test, exists := savedTests[name]
	if !exists {
		fmt.Printf("Error: No saved test named '%s'\n", name)
		os.Exit(1)
	}

	fmt.Printf("Running saved test '%s'...\n", name)
	sendRequest(test.Method, test.URL, test.Body, test.Headers)
}

func listHandler(cmd *cobra.Command, args []string) {
	if len(savedTests) == 0 {
		fmt.Println("No saved tests")
		return
	}

	fmt.Println("Saved tests:")
	for name := range savedTests {
		fmt.Printf("  - %s\n", name)
	}
}

func sendRequest(method, url, body string, headers map[string]string) {
	var reqBody io.Reader
	if body != "" {
		reqBody = bytes.NewBufferString(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	if _, ok := headers["Content-Type"]; !ok && body != "" {
		req.Header.Add("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("\nResponse:\n")
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Println("Headers:")
	for key, values := range resp.Header {
		fmt.Printf("  %s: %s\n", key, strings.Join(values, ", "))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nBody:")
	if len(respBody) > 0 {
		if json.Valid(respBody) {
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, respBody, "", "  "); err == nil {
				fmt.Println(prettyJSON.String())
			} else {
				fmt.Println(string(respBody))
			}
		} else {
			fmt.Println(string(respBody))
		}
	} else {
		fmt.Println("<empty>")
	}
}

func loadSavedTests() {
	configFile := filepath.Join(configDir, "tests.json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("Error reading config: %v\n", err)
		}
		return
	}

	if err := json.Unmarshal(data, &savedTests); err != nil {
		fmt.Printf("Error parsing saved tests: %v\n", err)
	}
}

func saveTestsToFile() {
	configFile := filepath.Join(configDir, "tests.json")
	data, err := json.MarshalIndent(savedTests, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling saved tests: %v\n", err)
		return
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		fmt.Printf("Error saving tests: %v\n", err)
	}
}