package agentpack

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	reHandle  = regexp.MustCompile(`^[a-z0-9-]+$`)
	reSemver  = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`)
)

// Load reads, parses, and validates an agent package rooted at dir.
// It returns an error if the manifest is invalid, any declared symbol is
// missing its rule or default theme asset, or any named theme is incomplete.
func Load(dir string) (*Package, error) {
	manifest, err := loadManifest(filepath.Join(dir, "agent.yaml"))
	if err != nil {
		return nil, err
	}

	if err := validateManifest(manifest); err != nil {
		return nil, err
	}

	rules, err := loadRules(dir, manifest.Symbols)
	if err != nil {
		return nil, err
	}

	themes, err := loadThemes(dir, manifest.Symbols)
	if err != nil {
		return nil, err
	}

	return &Package{
		Dir:      dir,
		Manifest: manifest,
		Rules:    rules,
		Themes:   themes,
	}, nil
}

func loadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read agent.yaml: %w", err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("parse agent.yaml: %w", err)
	}

	return m, nil
}

func validateManifest(m Manifest) error {
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("manifest: name is required")
	}
	if !reHandle.MatchString(m.Handle) {
		return fmt.Errorf("manifest: handle %q must match %s", m.Handle, reHandle)
	}
	if !reSemver.MatchString(m.Version) {
		return fmt.Errorf("manifest: version %q must be a valid semver (e.g. 1.0.0)", m.Version)
	}
	if strings.TrimSpace(m.Description) == "" {
		return fmt.Errorf("manifest: description is required")
	}
	if m.Visibility != VisibilityPublic && m.Visibility != VisibilityOrg {
		return fmt.Errorf("manifest: visibility must be %q or %q", VisibilityPublic, VisibilityOrg)
	}
	if err := validatePricing(m.Pricing); err != nil {
		return err
	}
	if len(m.Symbols) == 0 {
		return fmt.Errorf("manifest: at least one symbol is required")
	}
	seen := make(map[string]bool, len(m.Symbols))
	for _, s := range m.Symbols {
		if err := validateSymbolDef(s); err != nil {
			return err
		}
		if seen[s.ID] {
			return fmt.Errorf("manifest: duplicate symbol id %q", s.ID)
		}
		seen[s.ID] = true
	}
	return nil
}

func validatePricing(p *Pricing) error {
	if p == nil {
		return nil
	}
	if p.Model != PricingFree && p.Model != PricingPaid {
		return fmt.Errorf("pricing: model must be %q or %q", PricingFree, PricingPaid)
	}
	if p.Model == PricingPaid && p.PriceUSD < 0.99 {
		return fmt.Errorf("pricing: price_usd must be at least 0.99 for paid agents")
	}
	return nil
}

func validateSymbolDef(s SymbolDef) error {
	if !reHandle.MatchString(s.ID) {
		return fmt.Errorf("symbol: id %q must match %s", s.ID, reHandle)
	}
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("symbol %q: name is required", s.ID)
	}
	if strings.TrimSpace(s.Description) == "" {
		return fmt.Errorf("symbol %q: description is required", s.ID)
	}
	if s.Type != SymbolTypeRealtime && s.Type != SymbolTypeTemporal {
		return fmt.Errorf("symbol %q: type must be %q or %q", s.ID, SymbolTypeRealtime, SymbolTypeTemporal)
	}
	if s.Type == SymbolTypeTemporal && s.WindowHours < 1 {
		return fmt.Errorf("symbol %q: window_hours is required for temporal symbols", s.ID)
	}
	return nil
}

func loadRules(dir string, symbols []SymbolDef) (map[string]string, error) {
	rules := make(map[string]string, len(symbols))
	for _, s := range symbols {
		path := filepath.Join(dir, "rules", s.ID+".sql")
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("symbol %q: missing rule file rules/%s.sql", s.ID, s.ID)
		}
		rules[s.ID] = strings.TrimSpace(string(data))
	}
	return rules, nil
}

func loadThemes(dir string, symbols []SymbolDef) (map[string]ThemeManifest, error) {
	themesDir := filepath.Join(dir, "themes")
	entries, err := os.ReadDir(themesDir)
	if err != nil {
		return nil, fmt.Errorf("themes/: %w", err)
	}

	themes := make(map[string]ThemeManifest)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == "default" {
			tm, err := loadDefaultTheme(themesDir, symbols)
			if err != nil {
				return nil, err
			}
			themes["default"] = tm
		} else {
			tm, err := loadNamedTheme(filepath.Join(themesDir, name), name, symbols)
			if err != nil {
				return nil, err
			}
			themes[name] = tm
		}
	}

	if _, ok := themes["default"]; !ok {
		return nil, fmt.Errorf("themes/default/ is required but not found")
	}

	return themes, nil
}

// loadDefaultTheme validates that themes/default/ contains an SVG for every
// symbol ID and returns a synthetic ThemeManifest using the manifest defaults.
func loadDefaultTheme(themesDir string, symbols []SymbolDef) (ThemeManifest, error) {
	defaultDir := filepath.Join(themesDir, "default")
	tm := ThemeManifest{Symbols: make(map[string]ThemeSymbol, len(symbols))}

	for _, s := range symbols {
		asset := s.ID + ".svg"
		if _, err := os.Stat(filepath.Join(defaultDir, asset)); err != nil {
			return ThemeManifest{}, fmt.Errorf("themes/default/: missing asset %s for symbol %q", asset, s.ID)
		}
		tm.Symbols[s.ID] = ThemeSymbol{Name: s.Name, Asset: asset}
	}

	return tm, nil
}

// loadNamedTheme parses a theme.yaml and validates it covers all symbols.
func loadNamedTheme(themeDir, name string, symbols []SymbolDef) (ThemeManifest, error) {
	data, err := os.ReadFile(filepath.Join(themeDir, "theme.yaml"))
	if err != nil {
		return ThemeManifest{}, fmt.Errorf("themes/%s/: missing theme.yaml", name)
	}

	var tm ThemeManifest
	if err := yaml.Unmarshal(data, &tm); err != nil {
		return ThemeManifest{}, fmt.Errorf("themes/%s/theme.yaml: %w", name, err)
	}

	for _, s := range symbols {
		entry, ok := tm.Symbols[s.ID]
		if !ok {
			return ThemeManifest{}, fmt.Errorf("themes/%s/theme.yaml: missing entry for symbol %q", name, s.ID)
		}
		if _, err := os.Stat(filepath.Join(themeDir, entry.Asset)); err != nil {
			return ThemeManifest{}, fmt.Errorf("themes/%s/: missing asset %s for symbol %q", name, entry.Asset, s.ID)
		}
	}

	return tm, nil
}
