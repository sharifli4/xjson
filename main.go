package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kenan/xjson/config"
	"github.com/kenan/xjson/internal/api"
	"github.com/kenan/xjson/internal/ui"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			initConfig()
			return
		case "auth":
			authenticate()
			return
		case "help", "-h", "--help":
			printHelp()
			return
		}
	}

	run()
}

func printHelp() {
	fmt.Println(`xjson - API Response Inspector

A terminal-based tool for inspecting API responses.

Usage:
  xjson          Start the inspector
  xjson init     Create a default config file
  xjson auth     Authenticate with the API
  xjson help     Show this help message

Keybindings:
  j/k, ↑/↓       Scroll up/down
  n/p            Next/previous item
  /              Search
  r              Refresh
  t              Back to timeline
  ?              Toggle help
  q              Quit

Config file: ~/.xjson.yaml`)
}

func initConfig() {
	path := config.DefaultConfigPath()

	// Check if config already exists
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("Config file already exists at %s\n", path)
		fmt.Println("Edit it to add your API credentials.")
		return
	}

	if err := config.CreateDefault(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created config file at %s\n", path)
	fmt.Println("\nEdit the file and add your X API credentials:")
	fmt.Println("  client_id: Your OAuth 2.0 Client ID")
	fmt.Println("  client_secret: Your OAuth 2.0 Client Secret (if using confidential client)")
	fmt.Println("  bearer_token: Your Bearer Token (for app-only auth)")
	fmt.Println("\nGet credentials at: https://developer.twitter.com/en/portal/dashboard")
}

func authenticate() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		fmt.Println("Run 'xjson init' to create a config file first.")
		os.Exit(1)
	}

	if cfg.ClientID == "" || cfg.ClientID == "YOUR_CLIENT_ID" {
		fmt.Println("Please configure your client_id in ~/.xjson.yaml")
		os.Exit(1)
	}

	auth := api.NewAuthenticator(cfg.ClientID, cfg.ClientSecret, cfg.RedirectURL)
	if doAuth(auth) {
		fmt.Println("You can now run 'xjson' to start the app.")
	}
}

func run() {
	cfg, err := config.Load()
	if err != nil {
		// No config - create one and start auth
		fmt.Println("No config found. Creating one...")
		if err := config.CreateDefault(); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created config at %s\n\n", config.DefaultConfigPath())
		fmt.Println("You need X API credentials to continue.")
		fmt.Println("Get them at: https://developer.twitter.com/en/portal/dashboard")
		fmt.Println("\nEdit ~/.xjson.yaml and add your client_id, then run again.")
		os.Exit(0)
	}

	var client *api.Client

	// Try OAuth first
	if cfg.ClientID != "" && cfg.ClientID != "YOUR_CLIENT_ID" {
		auth := api.NewAuthenticator(cfg.ClientID, cfg.ClientSecret, cfg.RedirectURL)

		if auth.HasStoredToken() {
			// Try existing token
			httpClient, err := auth.HTTPClient(context.Background())
			if err == nil {
				client = api.NewClient(httpClient)
			}
		}

		// No valid token - start auth flow automatically
		if client == nil {
			fmt.Println("Authentication required. Starting OAuth flow...")
			if doAuth(auth) {
				// Auth succeeded, get client
				httpClient, err := auth.HTTPClient(context.Background())
				if err == nil {
					client = api.NewClient(httpClient)
				}
			}
		}
	}

	// Fall back to bearer token
	if client == nil && cfg.BearerToken != "" && cfg.BearerToken != "YOUR_BEARER_TOKEN" {
		client = api.NewClientWithBearerToken(cfg.BearerToken)
	}

	if client == nil {
		fmt.Println("\nNo valid authentication.")
		fmt.Println("Please add credentials to ~/.xjson.yaml:")
		fmt.Println("  - client_id: Your OAuth 2.0 Client ID")
		fmt.Println("  - bearer_token: Or use a Bearer Token instead")
		fmt.Println("\nGet credentials at: https://developer.twitter.com/en/portal/dashboard")
		os.Exit(1)
	}

	app := ui.NewApp(client)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func doAuth(auth *api.Authenticator) bool {
	// Start local server for callback
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	server := &http.Server{Addr: ":8080"}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no code in callback")
			fmt.Fprintln(w, "Error: No authorization code received")
			return
		}

		codeChan <- code
		fmt.Fprintln(w, `
			<html><body style="font-family: monospace; padding: 40px; background: #1a1a2e; color: #0f0;">
			<h2>Authorization successful!</h2>
			<p>You can close this window and return to the terminal.</p>
			</body></html>
		`)
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Get auth URL
	authURL, verifier, err := auth.StartAuthFlow()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting auth: %v\n", err)
		return false
	}

	fmt.Println("┌─────────────────────────────────────────────────────────┐")
	fmt.Println("│  Open this URL in your browser to authenticate:        │")
	fmt.Println("└─────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println(authURL)
	fmt.Println()
	fmt.Println("Waiting for authorization...")

	// Wait for callback
	select {
	case code := <-codeChan:
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err := auth.CompleteAuthFlow(ctx, code, verifier)
		server.Shutdown(context.Background())

		if err != nil {
			fmt.Fprintf(os.Stderr, "\nAuth error: %v\n", err)
			return false
		}

		fmt.Println("\n✓ Authentication successful! Starting app...")
		time.Sleep(1 * time.Second)
		return true

	case err := <-errChan:
		server.Shutdown(context.Background())
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return false

	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		fmt.Println("\nTimeout waiting for authorization")
		return false
	}
}
