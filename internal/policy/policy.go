// Package policy is the single source of truth for what is and is not allowed.
//
// Every guard rail in asaguard is driven by a policy: which MCPs are approved,
// which network domains Claude may contact, which filesystem paths form the
// sandbox boundary, which token-spend multiple constitutes a spike, and how
// much each check contributes to the compliance score. This package loads those
// rules from ~/.config/asaguard/policy.json and merges them over safe defaults,
// so the tool works out of the box for individuals and can be tightened
// centrally for teams. Keeping policy separate from code means security
// decisions are visible, auditable, and deployable without a binary update.
package policy

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Policy struct {
	RequiredSettingsKeys []string          `json:"required_settings_keys"`
	LockedSettingsKeys   []string          `json:"locked_settings_keys"`
	ApprovedMCPs         []string          `json:"approved_mcps"`
	MCPRiskCategories    map[string]string `json:"mcp_risk_categories"` // mcp name -> high|medium|low
	AllowedDomains       []string          `json:"allowed_domains"`
	SandboxReadRoots     []string          `json:"sandbox_read_roots"`
	SandboxWriteRoots    []string          `json:"sandbox_write_roots"`
	HITLWatchlist        []string          `json:"hitl_watchlist"`
	TokenSpikeMultiple   float64           `json:"token_spike_multiple"`
	TokenPrices          TokenPrices       `json:"token_prices"`
	Weights              map[string]int    `json:"weights"`
	Banner               BannerConfig      `json:"banner"`
}

// TokenPrices holds per-million-token prices for each billing category.
// Defaults reflect Claude Sonnet 4.x pricing.
type TokenPrices struct {
	InputPerMillion      float64 `json:"input_per_million"`
	OutputPerMillion     float64 `json:"output_per_million"`
	CacheWritePerMillion float64 `json:"cache_write_per_million"`
	CacheReadPerMillion  float64 `json:"cache_read_per_million"`
}

type BannerConfig struct {
	Text      string `json:"text"`
	PolicyURL string `json:"policy_url"`
}

var defaults = Policy{
	RequiredSettingsKeys: []string{},
	LockedSettingsKeys:   []string{},
	ApprovedMCPs:         []string{},
	MCPRiskCategories:    map[string]string{},
	AllowedDomains:       []string{},
	SandboxReadRoots:     defaultSandboxRoots(),
	SandboxWriteRoots:    defaultSandboxRoots(),
	HITLWatchlist:        defaultHITLWatchlist(),
	TokenSpikeMultiple:   3.0,
	TokenPrices: TokenPrices{
		InputPerMillion:      3.00,
		OutputPerMillion:     15.00,
		CacheWritePerMillion: 3.75,
		CacheReadPerMillion:  0.30,
	},
	Weights: defaultWeights(),
	Banner: BannerConfig{
		Text:      "Security reminder: follow your organisation's AI usage policy.",
		PolicyURL: "",
	},
}

func defaultSandboxRoots() []string {
	home, _ := os.UserHomeDir()
	return []string{home}
}

func defaultHITLWatchlist() []string {
	return []string{
		"rm -rf",
		"git push",
		"curl ",
		"wget ",
		"ssh ",
		"scp ",
	}
}

func defaultWeights() map[string]int {
	checks := []string{
		"settings", "mcps", "perms", "tokens",
		"network", "bypass", "sandbox", "secrets",
	}
	w := make(map[string]int, len(checks))
	each := 100 / len(checks)
	for _, c := range checks {
		w[c] = each
	}
	return w
}

// Load reads policy.json from the config dir and merges over defaults.
func Load() (*Policy, error) {
	p := defaults
	if p.Weights == nil {
		p.Weights = defaultWeights()
	}
	if p.MCPRiskCategories == nil {
		p.MCPRiskCategories = map[string]string{}
	}

	path := configPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &p, nil
	}
	if err != nil {
		return nil, err
	}

	var override Policy
	if err := json.Unmarshal(data, &override); err != nil {
		return nil, err
	}

	merge(&p, &override)
	return &p, nil
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "asaguard", "policy.json")
}

func merge(base, override *Policy) {
	if len(override.RequiredSettingsKeys) > 0 {
		base.RequiredSettingsKeys = override.RequiredSettingsKeys
	}
	if len(override.LockedSettingsKeys) > 0 {
		base.LockedSettingsKeys = override.LockedSettingsKeys
	}
	if len(override.ApprovedMCPs) > 0 {
		base.ApprovedMCPs = override.ApprovedMCPs
	}
	if len(override.MCPRiskCategories) > 0 {
		base.MCPRiskCategories = override.MCPRiskCategories
	}
	if len(override.AllowedDomains) > 0 {
		base.AllowedDomains = override.AllowedDomains
	}
	if len(override.SandboxReadRoots) > 0 {
		base.SandboxReadRoots = override.SandboxReadRoots
	}
	if len(override.SandboxWriteRoots) > 0 {
		base.SandboxWriteRoots = override.SandboxWriteRoots
	}
	if len(override.HITLWatchlist) > 0 {
		base.HITLWatchlist = override.HITLWatchlist
	}
	if override.TokenSpikeMultiple > 0 {
		base.TokenSpikeMultiple = override.TokenSpikeMultiple
	}
	if override.TokenPrices.InputPerMillion > 0 {
		base.TokenPrices.InputPerMillion = override.TokenPrices.InputPerMillion
	}
	if override.TokenPrices.OutputPerMillion > 0 {
		base.TokenPrices.OutputPerMillion = override.TokenPrices.OutputPerMillion
	}
	if override.TokenPrices.CacheWritePerMillion > 0 {
		base.TokenPrices.CacheWritePerMillion = override.TokenPrices.CacheWritePerMillion
	}
	if override.TokenPrices.CacheReadPerMillion > 0 {
		base.TokenPrices.CacheReadPerMillion = override.TokenPrices.CacheReadPerMillion
	}
	for k, v := range override.Weights {
		base.Weights[k] = v
	}
	if override.Banner.Text != "" {
		base.Banner.Text = override.Banner.Text
	}
	if override.Banner.PolicyURL != "" {
		base.Banner.PolicyURL = override.Banner.PolicyURL
	}
}
