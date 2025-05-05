package forgeron

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ScreenFingerprint represents screen-related fingerprint data
type ScreenFingerprint struct {
	AvailHeight      int     `json:"availHeight"`
	AvailWidth       int     `json:"availWidth"`
	AvailTop         int     `json:"availTop"`
	AvailLeft        int     `json:"availLeft"`
	ColorDepth       int     `json:"colorDepth"`
	Height           int     `json:"height"`
	PixelDepth       int     `json:"pixelDepth"`
	Width            int     `json:"width"`
	DevicePixelRatio float64 `json:"devicePixelRatio"`
	PageXOffset      int     `json:"pageXOffset"`
	PageYOffset      int     `json:"pageYOffset"`
	InnerHeight      int     `json:"innerHeight"`
	OuterHeight      int     `json:"outerHeight"`
	OuterWidth       int     `json:"outerWidth"`
	InnerWidth       int     `json:"innerWidth"`
	ScreenX          int     `json:"screenX"`
	ClientWidth      int     `json:"clientWidth"`
	ClientHeight     int     `json:"clientHeight"`
	HasHDR           bool    `json:"hasHDR"`
}

// UserAgentBrand represents a single brand in the user agent data
type UserAgentBrand struct {
	Brand   string `json:"brand"`
	Version string `json:"version"`
}

// UserAgentData represents the complete user agent data
type UserAgentData struct {
	Brands          []UserAgentBrand `json:"brands"`
	Mobile          bool             `json:"mobile"`
	Platform        string           `json:"platform"`
	Architecture    string           `json:"architecture"`
	Bitness         string           `json:"bitness"`
	FullVersionList []UserAgentBrand `json:"fullVersionList"`
	Model           string           `json:"model"`
	PlatformVersion string           `json:"platformVersion"`
	UAFullVersion   string           `json:"uaFullVersion"`
}

// NavigatorFingerprint represents navigator-related fingerprint data
type NavigatorFingerprint struct {
	UserAgent           string         `json:"userAgent"`
	UserAgentData       *UserAgentData `json:"userAgentData"`
	DoNotTrack          *string        `json:"doNotTrack"`
	AppCodeName         string         `json:"appCodeName"`
	AppName             string         `json:"appName"`
	AppVersion          string         `json:"appVersion"`
	OSCpu               string         `json:"oscpu"`
	Webdriver           string         `json:"webdriver"`
	Language            string         `json:"language"`
	Languages           []string       `json:"languages"`
	Platform            string         `json:"platform"`
	DeviceMemory        *int           `json:"deviceMemory"`
	HardwareConcurrency int            `json:"hardwareConcurrency"`
	Product             string         `json:"product"`
	ProductSub          string         `json:"productSub"`
	Vendor              string         `json:"vendor"`
	VendorSub           string         `json:"vendorSub"`
	MaxTouchPoints      int            `json:"maxTouchPoints"`
	ExtraProperties     map[string]any `json:"extraProperties"`
}

// VideoCard represents video card information
type VideoCard struct {
	Renderer string `json:"renderer"`
	Vendor   string `json:"vendor"`
}

// Battery represents battery information
type Battery struct {
	Charging        bool    `json:"charging"`
	ChargingTime    *int    `json:"chargingTime"`
	DischargingTime *int    `json:"dischargingTime"`
	Level           float64 `json:"level"`
}

// MediaDevice represents a single media device
type MediaDevice struct {
	DeviceID string `json:"deviceId"`
	Kind     string `json:"kind"`
	Label    string `json:"label"`
	GroupID  string `json:"groupId"`
}

// MultimediaDevices represents all media devices
type MultimediaDevices struct {
	Speakers []MediaDevice `json:"speakers"`
	Micros   []MediaDevice `json:"micros"`
	Webcams  []MediaDevice `json:"webcams"`
}

// MimeType represents a MIME type in a plugin
type MimeType struct {
	Type          string `json:"type"`
	Suffixes      string `json:"suffixes"`
	Description   string `json:"description"`
	EnabledPlugin string `json:"enabledPlugin"`
}

// Plugin represents a single plugin
type Plugin struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Filename    string     `json:"filename"`
	MimeTypes   []MimeType `json:"mimeTypes"`
}

// PluginsData represents all plugins data
type PluginsData struct {
	Plugins   []Plugin `json:"plugins"`
	MimeTypes []string `json:"mimeTypes"`
}

// Fingerprint represents the complete browser fingerprint
type Fingerprint struct {
	Screen            ScreenFingerprint    `json:"screen"`
	Navigator         NavigatorFingerprint `json:"navigator"`
	Headers           map[string]string    `json:"headers"`
	VideoCodecs       map[string]string    `json:"videoCodecs"`
	AudioCodecs       map[string]string    `json:"audioCodecs"`
	PluginsData       PluginsData          `json:"pluginsData"`
	Battery           *Battery             `json:"battery"`
	VideoCard         *VideoCard           `json:"videoCard"`
	MultimediaDevices *MultimediaDevices   `json:"multimediaDevices"`
	Fonts             []string             `json:"fonts"`
	MockWebRTC        bool                 `json:"mockWebRTC"`
	Slim              bool                 `json:"slim"`
}

// Screen represents screen dimension constraints
type Screen struct {
	MinWidth  *int
	MaxWidth  *int
	MinHeight *int
	MaxHeight *int
}

// IsSet returns true if any screen constraints are set
func (s *Screen) IsSet() bool {
	return s.MinWidth != nil || s.MaxWidth != nil || s.MinHeight != nil || s.MaxHeight != nil
}

// Validate validates the screen constraints
func (s *Screen) Validate() error {
	if s.MinWidth != nil && s.MaxWidth != nil && *s.MinWidth > *s.MaxWidth {
		return fmt.Errorf("minWidth cannot be greater than maxWidth")
	}
	if s.MinHeight != nil && s.MaxHeight != nil && *s.MinHeight > *s.MaxHeight {
		return fmt.Errorf("minHeight cannot be greater than maxHeight")
	}
	return nil
}

// FingerprintGenerator generates browser fingerprints using a Bayesian network
type FingerprintGenerator struct {
	network           *bayesianNetwork
	headerGenerator   *HeaderGenerator
	headerConstraints HeaderConstraints
	screen            *Screen
	strict            bool
	mockWebRTC        bool
	slim              bool
}

// FingerprintOption represents an option for configuring the fingerprint generator
type FingerprintOption func(*FingerprintGenerator)

// NewFingerprintGenerator creates a new fingerprint generator with the given options
func NewFingerprintGenerator(opts ...FingerprintOption) (*FingerprintGenerator, error) {
	hgen, err := NewHeaderGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to create header generator: %w", err)
	}
	generator := &FingerprintGenerator{
		network:         newBayesianNetwork(),
		headerGenerator: hgen,
	}

	// Apply options
	for _, opt := range opts {
		opt(generator)
	}

	// Load the fingerprint network definition
	if err := generator.loadNetwork(); err != nil {
		return nil, fmt.Errorf("failed to load fingerprint network: %w", err)
	}

	return generator, nil
}

// WithScreen sets the screen constraints for the fingerprint generator
func WithScreen(screen *Screen) FingerprintOption {
	return func(g *FingerprintGenerator) {
		g.screen = screen
	}
}

// WithStrict sets the strict mode for the fingerprint generator
func WithStrict(strict bool) FingerprintOption {
	return func(g *FingerprintGenerator) {
		g.strict = strict
	}
}

// WithMockWebRTC sets the mock WebRTC mode for the fingerprint generator
func WithMockWebRTC(mockWebRTC bool) FingerprintOption {
	return func(g *FingerprintGenerator) {
		g.mockWebRTC = mockWebRTC
	}
}

// WithSlim sets the slim mode for the fingerprint generator
func WithSlim(slim bool) FingerprintOption {
	return func(g *FingerprintGenerator) {
		g.slim = slim
	}
}

// WithHeaderConstraints sets the header constraints for the fingerprint generator
func WithHeaderConstraints(constraints HeaderConstraints) FingerprintOption {
	return func(g *FingerprintGenerator) {
		g.headerConstraints = constraints
	}
}

// Generate generates a new fingerprint with the given options
func (g *FingerprintGenerator) Generate(opts ...FingerprintOption) (*Fingerprint, error) {
	// Apply additional options
	for _, opt := range opts {
		opt(g)
	}

	// Generate headers first to get user agent
	headers, err := g.headerGenerator.GenerateHeaders(g.headerConstraints)
	if err != nil {
		return nil, fmt.Errorf("failed to generate headers: %w", err)
	}

	// Get user agent from headers
	userAgent := headers["User-Agent"]
	if userAgent == "" {
		return nil, fmt.Errorf("failed to find User-Agent in generated headers")
	}

	// Generate fingerprint with constraints
	constraints := map[string][]string{
		"userAgent": {userAgent},
	}

	// Add screen constraints if specified
	if g.screen != nil && g.screen.IsSet() {
		// TODO: Implement screen constraint filtering
	}

	// Generate fingerprint
	fingerprint, ok := g.network.generateConsistentSampleWhenPossible(constraints)
	if !ok {
		if g.strict {
			return nil, fmt.Errorf("could not generate fingerprint with given constraints")
		}
		// Try again without constraints
		fingerprint = g.network.generateSample(nil)
	}

	// Transform raw fingerprint into structured format
	return g.transformFingerprint(fingerprint, headers, g.mockWebRTC, g.slim)
}

// transformFingerprint converts a raw fingerprint map into a structured Fingerprint
func (g *FingerprintGenerator) transformFingerprint(raw map[string]string, headers map[string]string, mockWebRTC bool, slim bool) (*Fingerprint, error) {
	// Preprocess the fingerprint data
	for key, value := range raw {
		// Handle missing values, set to empty string
		if value == "*MISSING_VALUE*" {
			raw[key] = ""
			continue
		}

		// Handle stringified objects/arrays
		if strings.HasPrefix(value, "*STRINGIFIED*") {
			// Remove the prefix and parse the JSON
			jsonStr := value[len("*STRINGIFIED*"):]
			raw[key] = jsonStr
		}
	}

	// Process Accept-Language header
	if acceptLanguage := headers["Accept-Language"]; acceptLanguage != "" {
		// Split by comma and take the first part before any semicolon
		locales := strings.Split(acceptLanguage, ",")
		languages := make([]string, 0, len(locales))
		for _, locale := range locales {
			// Split by semicolon and take the first part
			parts := strings.SplitN(locale, ";", 2)
			languages = append(languages, strings.TrimSpace(parts[0]))
		}
		raw["languages"] = fmt.Sprintf(`["%s"]`, strings.Join(languages, `","`))
	}

	// Parse screen data
	screenData, ok := raw["screen"]
	if !ok {
		return nil, fmt.Errorf("missing screen data in fingerprint")
	}
	var screen ScreenFingerprint
	if err := json.Unmarshal([]byte(screenData), &screen); err != nil {
		return nil, fmt.Errorf("failed to parse screen data: %w, data: %s", err, screenData)
	}

	// Parse navigator data
	var userAgentData *UserAgentData
	if uaData, ok := raw["userAgentData"]; ok && uaData != "" {
		if err := json.Unmarshal([]byte(uaData), &userAgentData); err != nil {
			return nil, fmt.Errorf("failed to parse user agent data: %w, data: %s", err, uaData)
		}
	}

	// Parse extra properties
	extraProperties := make(map[string]any)
	if extraProps, ok := raw["extraProperties"]; ok && extraProps != "" {
		if err := json.Unmarshal([]byte(extraProps), &extraProperties); err != nil {
			return nil, fmt.Errorf("failed to parse extra properties: %w, data: %s", err, extraProps)
		}
	}

	navigator := NavigatorFingerprint{
		UserAgent:           raw["userAgent"],
		UserAgentData:       userAgentData,
		DoNotTrack:          parseStringPtr(raw["doNotTrack"]),
		AppCodeName:         raw["appCodeName"],
		AppName:             raw["appName"],
		AppVersion:          raw["appVersion"],
		OSCpu:               raw["oscpu"],
		Webdriver:           raw["webdriver"],
		Platform:            raw["platform"],
		DeviceMemory:        parseIntPtr(raw["deviceMemory"]),
		Product:             raw["product"],
		ProductSub:          raw["productSub"],
		Vendor:              raw["vendor"],
		VendorSub:           raw["vendorSub"],
		ExtraProperties:     extraProperties,
		HardwareConcurrency: parseInt(raw["hardwareConcurrency"]),
		MaxTouchPoints:      parseInt(raw["maxTouchPoints"]),
	}

	// Parse languages
	var languages []string
	if err := json.Unmarshal([]byte(raw["languages"]), &languages); err != nil {
		return nil, fmt.Errorf("failed to parse languages: %w, data: %s", err, raw["languages"])
	}
	navigator.Languages = languages
	if len(languages) > 0 {
		navigator.Language = languages[0]
	}

	// Parse video card if present
	var videoCard *VideoCard
	if vc, ok := raw["videoCard"]; ok && vc != "" {
		videoCard = &VideoCard{}
		if err := json.Unmarshal([]byte(vc), videoCard); err != nil {
			return nil, fmt.Errorf("failed to parse video card data: %w, data: %s", err, vc)
		}
	}

	// Parse battery if present
	var battery *Battery
	if b, ok := raw["battery"]; ok && b != "" {
		battery = &Battery{}
		if err := json.Unmarshal([]byte(b), battery); err != nil {
			return nil, fmt.Errorf("failed to parse battery data: %w, data: %s", err, b)
		}
	}

	// Parse multimedia devices
	var multimediaDevices *MultimediaDevices
	if md, ok := raw["multimediaDevices"]; ok && md != "" {
		multimediaDevices = &MultimediaDevices{}
		if err := json.Unmarshal([]byte(md), multimediaDevices); err != nil {
			return nil, fmt.Errorf("failed to parse multimedia devices: %w, data: %s", err, md)
		}
	}

	// Parse plugins data
	var pluginsData PluginsData
	if pd, ok := raw["pluginsData"]; ok && pd != "" {
		if err := json.Unmarshal([]byte(pd), &pluginsData); err != nil {
			return nil, fmt.Errorf("failed to parse plugins data: %w, data: %s", err, pd)
		}
	}

	// Parse fonts
	var fonts []string
	if f, ok := raw["fonts"]; ok && f != "" {
		if err := json.Unmarshal([]byte(f), &fonts); err != nil {
			return nil, fmt.Errorf("failed to parse fonts: %w, data: %s", err, f)
		}
	}

	return &Fingerprint{
		Screen:            screen,
		Navigator:         navigator,
		Headers:           headers,
		VideoCodecs:       parseMap(raw["videoCodecs"]),
		AudioCodecs:       parseMap(raw["audioCodecs"]),
		PluginsData:       pluginsData,
		Battery:           battery,
		VideoCard:         videoCard,
		MultimediaDevices: multimediaDevices,
		Fonts:             fonts,
		MockWebRTC:        mockWebRTC,
		Slim:              slim,
	}, nil
}

// loadNetwork loads the fingerprint network definition from the embedded zip file
func (g *FingerprintGenerator) loadNetwork() error {
	network, err := loadNetworkFromZip("fingerprint-network-definition.zip")
	if err != nil {
		return err
	}
	g.network = network
	return nil
}
