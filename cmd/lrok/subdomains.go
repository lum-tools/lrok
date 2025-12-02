package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lum-tools/lrok/internal/config"
	"github.com/spf13/cobra"
)

var subdomainDomain string

// API response structures
type SubdomainInfo struct {
	ID         int    `json:"id"`
	Subdomain  string `json:"subdomain"`
	Domain     string `json:"domain"`
	FullDomain string `json:"full_domain"`
	CreatedAt  string `json:"created_at"`
}

type ListSubdomainsResponse struct {
	Subdomains []SubdomainInfo `json:"subdomains"`
	Count      int             `json:"count"`
	MaxAllowed int             `json:"max_allowed"`
}

type ReserveSubdomainResponse struct {
	Success   bool   `json:"success"`
	Subdomain string `json:"subdomain"`
	Domain    string `json:"domain"`
	Reason    string `json:"reason"`
}

type ReleaseSubdomainResponse struct {
	Success   bool   `json:"success"`
	Subdomain string `json:"subdomain,omitempty"`
	Domain    string `json:"domain,omitempty"`
	Reason    string `json:"reason"`
}

var subdomainsCmd = &cobra.Command{
	Use:   "subdomains",
	Short: "Manage reserved subdomains",
	Long: `Manage your reserved subdomains on lum.tools platform.

Reserved subdomains allow you to claim permanent tunnel URLs that only you can use.
Each user can reserve up to 5 subdomains.

Subcommands:
  list      List all your reserved subdomains
  reserve   Reserve a new subdomain
  delete    Release a subdomain reservation`,
}

var subdomainsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your reserved subdomains",
	Long: `List all subdomains reserved by your account.

Examples:
  lrok subdomains list
  lrok subdomains list --api-key lum_xxxxx`,
	RunE: runSubdomainsList,
}

var subdomainsReserveCmd = &cobra.Command{
	Use:   "reserve <subdomain>",
	Short: "Reserve a subdomain for your account",
	Long: `Reserve a subdomain to permanently claim a tunnel URL.

The subdomain must:
  - Use lowercase alphanumeric characters and hyphens only
  - Not start or end with a hyphen
  - Not already be reserved by another user

Examples:
  lrok subdomains reserve myapp
  lrok subdomains reserve my-awesome-api
  lrok subdomains reserve prod-service --domain t.lum.tools`,
	Args: cobra.ExactArgs(1),
	RunE: runSubdomainsReserve,
}

var subdomainsDeleteCmd = &cobra.Command{
	Use:   "delete <subdomain>",
	Short: "Release a subdomain reservation",
	Long: `Release a subdomain reservation, making it available for others.

Warning: Once released, the subdomain may be reserved by another user.

Examples:
  lrok subdomains delete myapp
  lrok subdomains delete old-service
  lrok subdomains delete myapp --domain t.lum.tools`,
	Args: cobra.ExactArgs(1),
	RunE: runSubdomainsDelete,
}

func init() {
	// Add flags to list command
	subdomainsListCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "lum.tools platform API key")

	// Add flags to reserve command
	subdomainsReserveCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "lum.tools platform API key")
	subdomainsReserveCmd.Flags().StringVar(&subdomainDomain, "domain", "t.lum.tools", "Domain suffix")

	// Add flags to delete command
	subdomainsDeleteCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "lum.tools platform API key")
	subdomainsDeleteCmd.Flags().StringVar(&subdomainDomain, "domain", "t.lum.tools", "Domain suffix")

	// Add subcommands to parent command
	subdomainsCmd.AddCommand(subdomainsListCmd)
	subdomainsCmd.AddCommand(subdomainsReserveCmd)
	subdomainsCmd.AddCommand(subdomainsDeleteCmd)
}

func getAPIKey() (string, error) {
	// Priority: flag > env var > config file
	if apiKey != "" {
		return apiKey, nil
	}

	if key := os.Getenv("LUM_API_KEY"); key != "" {
		return key, nil
	}

	if key := os.Getenv("FRP_API_KEY"); key != "" {
		return key, nil
	}

	if key, err := config.GetAPIKey(); err == nil && key != "" {
		return key, nil
	}

	return "", fmt.Errorf(`no API key configured

Set your API key using one of:
  lrok login <API_KEY>
  export LUM_API_KEY=<API_KEY>
  lrok subdomains list --api-key <API_KEY>`)
}

func runSubdomainsList(cmd *cobra.Command, args []string) error {
	key, err := getAPIKey()
	if err != nil {
		return err
	}

	fmt.Println("📋 Fetching your reserved subdomains...")

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", "https://platform.lum.tools/api/v1/me/subdomains", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch subdomains: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid or expired API key")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var result ListSubdomainsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Display results
	maxAllowedStr := fmt.Sprintf("%d", result.MaxAllowed)
	if result.MaxAllowed >= 999999 {
		maxAllowedStr = "Unlimited"
	}
	fmt.Printf("\n✅ Reserved Subdomains (%d/%s)\n", result.Count, maxAllowedStr)
	fmt.Println(strings.Repeat("─", 60))

	if result.Count == 0 {
		fmt.Println("No subdomains reserved yet.")
		fmt.Println("\nReserve one with: lrok subdomains reserve <subdomain>")
	} else {
		for _, s := range result.Subdomains {
			// Parse and format the date
			createdTime := s.CreatedAt
			if t, err := time.Parse(time.RFC3339, s.CreatedAt); err == nil {
				createdTime = t.Format("2006-01-02 15:04")
			}

			fmt.Printf("  • %s\n", s.FullDomain)
			fmt.Printf("    Created: %s\n\n", createdTime)
		}
	}

	return nil
}

func runSubdomainsReserve(cmd *cobra.Command, args []string) error {
	subdomain := strings.ToLower(args[0])

	key, err := getAPIKey()
	if err != nil {
		return err
	}

	fmt.Printf("🔒 Reserving subdomain '%s.%s'...\n", subdomain, subdomainDomain)

	// Build request body
	requestBody := fmt.Sprintf(`{"subdomain": "%s", "domain": "%s"}`, subdomain, subdomainDomain)

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", "https://platform.lum.tools/api/v1/me/subdomains", strings.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reserve subdomain: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid or expired API key")
	}

	if resp.StatusCode != 200 {
		// Try to parse error response
		var result ReserveSubdomainResponse
		if err := json.Unmarshal(body, &result); err == nil && !result.Success {
			// Check if it's a limit error
			if strings.Contains(strings.ToLower(result.Reason), "maximum") || 
			   strings.Contains(strings.ToLower(result.Reason), "limit") ||
			   strings.Contains(strings.ToLower(result.Reason), "reached") {
				fmt.Println()
				fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
				fmt.Println("❌ Free tier limit reached: 5 reserved subdomains")
				fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
				fmt.Println()
				fmt.Println("📦 Upgrade to lrok Pro (€9/month) for:")
				fmt.Println("   • 50 reserved subdomains")
				fmt.Println("   • Unlimited concurrent tunnels")
				fmt.Println("   • TCP/UDP protocols")
				fmt.Println("   • Priority support")
				fmt.Println()
				fmt.Println("👉 Upgrade now: https://platform.lum.tools/billing")
				fmt.Println()
				fmt.Println("Or manage your subdomains:")
				fmt.Println("   lrok subdomains list")
				fmt.Println("   lrok subdomains delete <name>")
				fmt.Println()
				return fmt.Errorf("maximum subdomains reached")
			}
			fmt.Printf("\n❌ Failed to reserve subdomain\n")
			fmt.Printf("   %s\n", result.Reason)
			return fmt.Errorf("reservation failed: %s", result.Reason)
		}
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var result ReserveSubdomainResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Success {
		fmt.Printf("\n✅ Successfully reserved '%s.%s'\n", result.Subdomain, result.Domain)
		fmt.Printf("   %s\n", result.Reason)
		fmt.Printf("\nYou can now use this subdomain in your tunnels:\n")
		fmt.Printf("  lrok http 8080 --subdomain %s\n", result.Subdomain)
	} else {
		// Check if it's a limit error
		if strings.Contains(strings.ToLower(result.Reason), "maximum") || 
		   strings.Contains(strings.ToLower(result.Reason), "limit") ||
		   strings.Contains(strings.ToLower(result.Reason), "reached") {
			fmt.Println()
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println("❌ Free tier limit reached: 5 reserved subdomains")
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Println()
			fmt.Println("📦 Upgrade to lrok Pro (€9/month) for:")
			fmt.Println("   • 50 reserved subdomains")
			fmt.Println("   • Unlimited concurrent tunnels")
			fmt.Println("   • TCP/UDP protocols")
			fmt.Println("   • Priority support")
			fmt.Println()
			fmt.Println("👉 Upgrade now: https://platform.lum.tools/billing")
			fmt.Println()
			fmt.Println("Or manage your subdomains:")
			fmt.Println("   lrok subdomains list")
			fmt.Println("   lrok subdomains delete <name>")
			fmt.Println()
			return fmt.Errorf("maximum subdomains reached")
		}
		fmt.Printf("\n❌ Failed to reserve subdomain\n")
		fmt.Printf("   %s\n", result.Reason)
		return fmt.Errorf("reservation failed: %s", result.Reason)
	}

	return nil
}

func runSubdomainsDelete(cmd *cobra.Command, args []string) error {
	subdomain := strings.ToLower(args[0])

	key, err := getAPIKey()
	if err != nil {
		return err
	}

	fmt.Printf("🗑️  Releasing subdomain '%s.%s'...\n", subdomain, subdomainDomain)

	client := &http.Client{Timeout: 30 * time.Second}
	url := fmt.Sprintf("https://platform.lum.tools/api/v1/me/subdomains/%s?domain=%s", subdomain, subdomainDomain)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to release subdomain: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid or expired API key")
	}

	var result ReleaseSubdomainResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Success {
		fmt.Printf("\n✅ Successfully released '%s.%s'\n", result.Subdomain, result.Domain)
		fmt.Printf("   %s\n", result.Reason)
		fmt.Println("\n⚠️  Note: This subdomain is now available for others to reserve.")
	} else {
		fmt.Printf("\n❌ Failed to release subdomain\n")
		fmt.Printf("   %s\n", result.Reason)
		return fmt.Errorf("release failed")
	}

	return nil
}
