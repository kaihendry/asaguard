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
	Weights              map[string]int    `json:"weights"`
	Banner               BannerConfig      `json:"banner"`
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
	Weights:              defaultWeights(),
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
