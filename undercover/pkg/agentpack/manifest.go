package agentpack

// Visibility controls who can discover and install an agent.
type Visibility string

const (
	VisibilityPublic Visibility = "public"
	VisibilityOrg    Visibility = "org"
)

// PricingModel is the monetisation strategy for an agent.
type PricingModel string

const (
	PricingFree PricingModel = "free"
	PricingPaid PricingModel = "paid"
)

// SymbolType determines how the evaluation rule is triggered.
type SymbolType string

const (
	SymbolTypeRealtime SymbolType = "realtime"
	SymbolTypeTemporal SymbolType = "temporal"
)

// Pricing holds monetisation configuration from agent.yaml.
type Pricing struct {
	Model    PricingModel `yaml:"model"`
	PriceUSD float64      `yaml:"price_usd,omitempty"`
}

// SymbolDef is one symbol declared in agent.yaml.
type SymbolDef struct {
	ID          string     `yaml:"id"`
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	Type        SymbolType `yaml:"type"`
	WindowHours int        `yaml:"window_hours,omitempty"`
}

// Manifest is the parsed agent.yaml.
type Manifest struct {
	Name        string      `yaml:"name"`
	Handle      string      `yaml:"handle"`
	Version     string      `yaml:"version"`
	Description string      `yaml:"description"`
	Visibility  Visibility  `yaml:"visibility"`
	Pricing     *Pricing    `yaml:"pricing,omitempty"`
	Symbols     []SymbolDef `yaml:"symbols"`
}

// ThemeSymbol is one entry in a theme.yaml symbols map.
type ThemeSymbol struct {
	Name  string `yaml:"name"`
	Asset string `yaml:"asset"`
}

// ThemeManifest is the parsed themes/<name>/theme.yaml.
// The key in Symbols is the symbol ID from the agent manifest.
type ThemeManifest struct {
	Symbols map[string]ThemeSymbol `yaml:"symbols"`
}

// Package is the fully loaded and validated agent package.
type Package struct {
	// Dir is the root directory of the loaded package.
	Dir string

	// Manifest is the parsed agent.yaml.
	Manifest Manifest

	// Rules maps symbol ID to the SQL query loaded from rules/<id>.sql.
	Rules map[string]string

	// Themes maps theme name to its manifest.
	// "default" is always present; named themes are optional.
	Themes map[string]ThemeManifest
}
