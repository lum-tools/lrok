package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	PlatformBaseURL = "https://platform.lum.tools"
)

// Domain represents a custom domain in the platform
type Domain struct {
	ID                string     `json:"id"`
	Domain            string     `json:"domain"`
	Type              string     `json:"type"` // "premium" or "byod"
	Status            string     `json:"status"` // pending, verifying, verified, active
	VerificationToken string     `json:"verification_token,omitempty"`
	TLSStatus         string     `json:"tls_status,omitempty"` // pending, provisioning, active, failed
	VerifiedAt        *time.Time `json:"verified_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// DomainsListResponse represents the API response for listing domains
type DomainsListResponse struct {
	Domains    []Domain `json:"domains"`
	Count      int      `json:"count"`
	MaxAllowed int      `json:"max_allowed"`
	Tier       string   `json:"tier"`
}

// DomainCreateResponse represents the API response for creating a domain
type DomainCreateResponse struct {
	Success      bool   `json:"success"`
	Domain       Domain `json:"domain"`
	Verification struct {
		CNAMEHost   string `json:"cname_host"`
		CNAMETarget string `json:"cname_target"`
		TXTHost     string `json:"txt_host"`
		TXTValue    string `json:"txt_value"`
	} `json:"verification"`
}

// DomainVerifyResponse represents the API response for verifying a domain
type DomainVerifyResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Domain    Domain `json:"domain,omitempty"`
	NextSteps struct {
		CNAMEHost   string `json:"cname_host"`
		CNAMETarget string `json:"cname_target"`
		Usage       string `json:"usage"`
	} `json:"next_steps,omitempty"`
	Expected struct {
		CNAMEHost   string `json:"cname_host"`
		CNAMETarget string `json:"cname_target"`
		TXTHost     string `json:"txt_host"`
		TXTValue    string `json:"txt_value"`
	} `json:"expected,omitempty"`
}

// APIError represents an error response from the API
type APIError struct {
	Error      string `json:"error"`
	UpgradeURL string `json:"upgrade_url,omitempty"`
}

// DomainsClient handles API requests for domain management
type DomainsClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewDomainsClient creates a new domains API client
func NewDomainsClient(apiKey string) *DomainsClient {
	return &DomainsClient{
		baseURL:    PlatformBaseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *DomainsClient) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "lrok-cli/1.0")

	return c.httpClient.Do(req)
}

// ListDomains fetches all domains for the authenticated user
func (c *DomainsClient) ListDomains() (*DomainsListResponse, error) {
	resp, err := c.doRequest("GET", "/api/domains", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("API error (status %d)", resp.StatusCode)
		}
		return nil, fmt.Errorf("%s", apiErr.Error)
	}

	var result DomainsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// AddDomain creates a new custom domain
func (c *DomainsClient) AddDomain(domain string) (*DomainCreateResponse, error) {
	body := map[string]string{"domain": domain}
	resp, err := c.doRequest("POST", "/api/domains", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		var apiErr APIError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil && apiErr.UpgradeURL != "" {
			return nil, fmt.Errorf("%s\n\n   Upgrade at: %s", apiErr.Error, apiErr.UpgradeURL)
		}
	}

	if resp.StatusCode != http.StatusCreated {
		var apiErr APIError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("API error (status %d)", resp.StatusCode)
		}
		return nil, fmt.Errorf("%s", apiErr.Error)
	}

	var result DomainCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// VerifyDomain triggers DNS verification for a domain
func (c *DomainsClient) VerifyDomain(domainID string) (*DomainVerifyResponse, error) {
	resp, err := c.doRequest("POST", "/api/domains/"+url.PathEscape(domainID)+"/verify", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result DomainVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// DeleteDomain removes a custom domain
func (c *DomainsClient) DeleteDomain(domainID string) error {
	resp, err := c.doRequest("DELETE", "/api/domains/"+url.PathEscape(domainID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("domain not found")
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return fmt.Errorf("API error (status %d)", resp.StatusCode)
		}
		return fmt.Errorf("%s", apiErr.Error)
	}

	return nil
}
