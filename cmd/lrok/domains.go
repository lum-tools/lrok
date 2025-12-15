package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lum-tools/lrok/internal/api"
	"github.com/spf13/cobra"
)

var domainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "Manage custom domains (Pro feature)",
	Long: `Manage your own domains for lrok tunnels.

Custom domains allow you to use your own domain names (e.g., tunnel.mycompany.com)
for lrok tunnels instead of the default *.t.lum.tools subdomains.

This feature requires a Pro or Team subscription.
Get started: https://platform.lum.tools/upgrade`,
}

var domainsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List your custom domains",
	RunE:    runDomainsList,
}

var domainsAddCmd = &cobra.Command{
	Use:   "add <domain>",
	Short: "Add a custom domain",
	Long: `Add a custom domain to your account.

You'll need to verify ownership by adding a DNS record.
After verification, create a CNAME pointing to frp.lum.tools.

Example:
  lrok domains add tunnel.mycompany.com`,
	Args: cobra.ExactArgs(1),
	RunE: runDomainsAdd,
}

var domainsVerifyCmd = &cobra.Command{
	Use:   "verify <domain>",
	Short: "Verify domain ownership via DNS",
	Long: `Check if DNS verification records are properly configured.

After adding a domain, you must verify ownership by adding the
required DNS records. This command checks for those records.`,
	Args: cobra.ExactArgs(1),
	RunE: runDomainsVerify,
}

var domainsDeleteCmd = &cobra.Command{
	Use:     "delete <domain>",
	Aliases: []string{"rm", "remove"},
	Short:   "Remove a custom domain",
	Long: `Remove a custom domain from your account.

This will stop any tunnels currently using this domain.`,
	Args: cobra.ExactArgs(1),
	RunE: runDomainsDelete,
}

var domainsCertCmd = &cobra.Command{
	Use:   "cert <domain>",
	Short: "Check TLS certificate status",
	Long: `Check the TLS certificate status for a custom domain.

After domain verification, lrok automatically provisions a 
TLS certificate via Let's Encrypt for HTTPS support.

Example:
  lrok domains cert tunnel.mycompany.com`,
	Args: cobra.ExactArgs(1),
	RunE: runDomainsCert,
}

func init() {
	domainsCmd.AddCommand(domainsListCmd)
	domainsCmd.AddCommand(domainsAddCmd)
	domainsCmd.AddCommand(domainsVerifyCmd)
	domainsCmd.AddCommand(domainsDeleteCmd)
	domainsCmd.AddCommand(domainsCertCmd)
}

func runDomainsList(cmd *cobra.Command, args []string) error {
	apiKey, err := getAPIKey()
	if err != nil {
		return err
	}

	client := api.NewDomainsClient(apiKey)
	result, err := client.ListDomains()
	if err != nil {
		return fmt.Errorf("failed to fetch domains: %w", err)
	}

	if len(result.Domains) == 0 {
		fmt.Println("📋 No custom domains configured")
		fmt.Println()
		fmt.Println("Add your first domain:")
		fmt.Println("  lrok domains add tunnel.example.com")
		return nil
	}

	fmt.Println()
	fmt.Println("Your Custom Domains")
	fmt.Println("───────────────────────────────────────────────────────────────────")
	fmt.Printf("%-30s %-15s %-12s %-14s\n", "DOMAIN", "STATUS", "TLS", "VERIFIED AT")

	for _, d := range result.Domains {
		status := formatStatus(d.Status)
		tlsStatus := formatTLSStatus(d.TLSStatus)
		verifiedAt := "-"
		if d.VerifiedAt != nil {
			verifiedAt = d.VerifiedAt.Format("2006-01-02")
		}
		fmt.Printf("%-30s %-15s %-12s %-14s\n", d.Domain, status, tlsStatus, verifiedAt)
	}

	fmt.Println()
	fmt.Printf("Total: %d domains (%s tier: %d max)\n", result.Count, result.Tier, result.MaxAllowed)

	return nil
}

func formatStatus(status string) string {
	switch status {
	case "active":
		return "✅ active"
	case "verified":
		return "✅ verified"
	case "verifying":
		return "🔄 verifying"
	case "pending":
		return "⏳ pending"
	default:
		return status
	}
}

func formatTLSStatus(status string) string {
	switch status {
	case "issued", "active":
		return "🔐 active"
	case "pending", "provisioning":
		return "🔄 pending"
	case "failed":
		return "❌ failed"
	case "none", "":
		return "⏳ none"
	default:
		return status
	}
}

func runDomainsAdd(cmd *cobra.Command, args []string) error {
	apiKey, err := getAPIKey()
	if err != nil {
		return err
	}

	domain := strings.ToLower(strings.TrimSpace(args[0]))

	client := api.NewDomainsClient(apiKey)
	result, err := client.AddDomain(domain)
	if err != nil {
		return fmt.Errorf("failed to add domain: %w", err)
	}

	fmt.Println()
	fmt.Println("📋 Domain Verification Required")
	fmt.Println()
	fmt.Println("To verify ownership, add one of these DNS records:")
	fmt.Println()
	fmt.Println("Option 1: CNAME Record")
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Printf("  Host:   %s\n", result.Verification.CNAMEHost)
	fmt.Printf("  Target: %s\n", result.Verification.CNAMETarget)
	fmt.Println()
	fmt.Println("Option 2: TXT Record")
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Printf("  Host:  %s\n", result.Verification.TXTHost)
	fmt.Printf("  Value: %s\n", result.Verification.TXTValue)
	fmt.Println()
	fmt.Println("After adding the record, verify with:")
	fmt.Printf("  lrok domains verify %s\n", result.Domain.ID)
	fmt.Println()
	fmt.Println("💡 DNS propagation can take up to 48 hours, but usually completes in minutes.")

	return nil
}

func runDomainsVerify(cmd *cobra.Command, args []string) error {
	apiKey, err := getAPIKey()
	if err != nil {
		return err
	}

	domainID := args[0]

	fmt.Println("🔍 Checking DNS records...")
	fmt.Println()

	client := api.NewDomainsClient(apiKey)
	result, err := client.VerifyDomain(domainID)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	if result.Success {
		fmt.Println("✅ Domain verified!")
		fmt.Println()
		fmt.Println("Next: Create a CNAME record pointing to your tunnel:")
		fmt.Println()
		fmt.Printf("  Host:   %s\n", result.NextSteps.CNAMEHost)
		fmt.Printf("  Target: %s\n", result.NextSteps.CNAMETarget)
		fmt.Println()
		fmt.Println("Then use your domain:")
		fmt.Printf("  %s\n", result.NextSteps.Usage)
	} else {
		fmt.Println("❌ DNS verification failed")
		fmt.Println()
		fmt.Println("Please check your DNS records:")
		fmt.Println()
		fmt.Println("Expected CNAME:")
		fmt.Printf("  Host:   %s\n", result.Expected.CNAMEHost)
		fmt.Printf("  Target: %s\n", result.Expected.CNAMETarget)
		fmt.Println()
		fmt.Println("Or TXT record:")
		fmt.Printf("  Host:  %s\n", result.Expected.TXTHost)
		fmt.Printf("  Value: %s\n", result.Expected.TXTValue)
		fmt.Println()
		fmt.Println("💡 DNS changes can take time to propagate. Try again in a few minutes.")
	}

	return nil
}

func runDomainsDelete(cmd *cobra.Command, args []string) error {
	apiKey, err := getAPIKey()
	if err != nil {
		return err
	}

	domainID := args[0]

	fmt.Printf("⚠️  This will remove the domain from your account.\n")
	fmt.Printf("   Any active tunnels using this domain will stop working.\n")
	fmt.Println()
	fmt.Printf("Type the domain ID '%s' to confirm: ", domainID)

	reader := bufio.NewReader(os.Stdin)
	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.TrimSpace(confirmation)

	if confirmation != domainID {
		fmt.Println("❌ Cancelled - confirmation did not match")
		return nil
	}

	client := api.NewDomainsClient(apiKey)
	if err := client.DeleteDomain(domainID); err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Domain removed")

	// Small delay for better UX
	time.Sleep(100 * time.Millisecond)

	return nil
}

func runDomainsCert(cmd *cobra.Command, args []string) error {
	apiKey, err := getAPIKey()
	if err != nil {
		return err
	}

	domainID := args[0]

	fmt.Println("🔐 Checking certificate status...")
	fmt.Println()

	client := api.NewDomainsClient(apiKey)
	result, err := client.GetCertificateStatus(domainID)
	if err != nil {
		return fmt.Errorf("failed to get certificate status: %w", err)
	}

	fmt.Printf("Domain: %s\n", result.Domain)
	fmt.Printf("Verified: %v\n", result.Verified)
	fmt.Println()

	switch result.CertificateStatus {
	case "issued":
		fmt.Println("✅ TLS Certificate: Active")
		if result.Certificate != nil {
			fmt.Printf("   Issuer: %s\n", result.Certificate.Issuer)
			fmt.Printf("   Type: %s\n", result.Certificate.Type)
			fmt.Printf("   Auto-renew: %v\n", result.Certificate.AutoRenew)
		}
	case "pending":
		fmt.Println("🔄 TLS Certificate: Provisioning")
		fmt.Println("   Certificate is being issued by Let's Encrypt.")
		fmt.Println("   This typically takes 1-5 minutes.")
	case "failed":
		fmt.Println("❌ TLS Certificate: Failed")
		fmt.Println("   Certificate provisioning failed.")
		fmt.Println("   Ensure your domain CNAME points to frp.lum.tools")
	case "none":
		fmt.Println("⏳ TLS Certificate: Not Started")
		if !result.Verified {
			fmt.Println("   Domain must be verified first.")
			fmt.Println("   Run: lrok domains verify <domain_id>")
		} else {
			fmt.Println("   Certificate will be provisioned automatically.")
		}
	default:
		fmt.Printf("Certificate Status: %s\n", result.CertificateStatus)
	}

	return nil
}
