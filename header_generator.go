package forgeron

import (
	"encoding/json"
	"fmt"
	"strings"
)

// BrowserSpec represents a browser specification with name, min/max version, and HTTP version
type BrowserSpec struct {
	Name        string
	MinVersion  int
	MaxVersion  int
	HTTPVersion string
}

// httpBrowser represents an HTTP browser object with name, version, complete string, and HTTP version
type httpBrowser struct {
	Name           *string
	Version        []int
	CompleteString string
	HTTPVersion    string
}

// IsHTTP2 returns true if the browser uses HTTP/2
func (h *httpBrowser) IsHTTP2() bool {
	return h.HTTPVersion == "2"
}

// Supported Browsers, OS, Devices, and HTTP versions
var (
	SupportedBrowsers = []string{"chrome", "firefox", "safari", "edge"}
	SupportedOS       = []string{"windows", "macos", "linux", "android", "ios"}
	SupportedDevices  = []string{"desktop", "mobile"}
	SupportedHTTP     = []string{"1", "2"}
)

// missingValueToken is used to indicate a missing value in the headers
const missingValueToken = "*MISSING_VALUE*"

// http1SecFetchAttributes defines the default Sec-Fetch headers for HTTP/1.1
var http1SecFetchAttributes = map[string]string{
	"Sec-Fetch-Mode": "same-site",
	"Sec-Fetch-Dest": "navigate",
	"Sec-Fetch-Site": "?1",
	"Sec-Fetch-User": "document",
}

// http2SecFetchAttributes defines the default Sec-Fetch headers for HTTP/2
var http2SecFetchAttributes = map[string]string{
	"sec-fetch-mode": "same-site",
	"sec-fetch-dest": "navigate",
	"sec-fetch-site": "?1",
	"sec-fetch-user": "document",
}

// HeaderConstraints represents the configuration constraints for header generation
type HeaderConstraints struct {
	BrowserSpecs []*BrowserSpec
	Browsers     []string
	OS           []string
	Devices      []string
	Locales      []string
	HTTPVersion  string
	Strict       bool
}

// HeaderGenerator generates HTTP headers based on browser fingerprint
type HeaderGenerator struct {
	headerGeneratorNetwork *bayesianNetwork
	inputGeneratorNetwork  *bayesianNetwork
	headersOrder           map[string][]string
	uniqueBrowsers         []*httpBrowser
	options                HeaderConstraints
}

// defaultHeaderOptions returns the default header constraints
func defaultHeaderOptions() HeaderConstraints {
	return HeaderConstraints{
		Browsers:    SupportedBrowsers,
		OS:          SupportedOS,
		Devices:     SupportedDevices,
		Locales:     []string{"en-US"},
		HTTPVersion: "2",
		Strict:      false,
	}
}

// mergeOptions merges user options with defaults, preserving non-zero values from user options
func (g *HeaderGenerator) mergeOptions(userOptions HeaderConstraints) (HeaderConstraints, error) {
	merged := g.options // Start with defaults
	var validationErrors []error

	// Helper function to validate and merge a field
	validateAndMerge := func(userValues, supported []string, setter func([]string)) {
		if len(userValues) == 0 {
			return
		}
		valid, err := filterValidValues(userValues, supported)
		if err != nil {
			validationErrors = append(validationErrors, err)
		}
		if len(valid) > 0 {
			setter(valid)
		}
	}

	// Validate and merge each field
	validateAndMerge(userOptions.Browsers, SupportedBrowsers, func(v []string) { merged.Browsers = v })
	validateAndMerge(userOptions.OS, SupportedOS, func(v []string) { merged.OS = v })
	validateAndMerge(userOptions.Devices, SupportedDevices, func(v []string) { merged.Devices = v })

	// Handle locales
	if len(userOptions.Locales) > 0 {
		merged.Locales = userOptions.Locales
		if len(merged.Locales) > 10 {
			merged.Locales = merged.Locales[:10]
		}
	}

	// Handle HTTP version
	if userOptions.HTTPVersion != "" {
		if err := validateAgainstSupported(userOptions.HTTPVersion, SupportedHTTP); err != nil {
			validationErrors = append(validationErrors, err)
		} else {
			merged.HTTPVersion = userOptions.HTTPVersion
		}
	}

	merged.Strict = userOptions.Strict
	merged.BrowserSpecs = userOptions.BrowserSpecs

	if len(validationErrors) > 0 {
		return merged, fmt.Errorf("validation errors: %v", validationErrors)
	}

	return merged, nil
}

// NewHeaderGenerator creates a new header generator
func NewHeaderGenerator() (*HeaderGenerator, error) {
	generator := &HeaderGenerator{
		options: defaultHeaderOptions(),
	}

	// Load headers order and unique browsers
	generator.loadHeadersOrder()
	generator.loadUniqueBrowsers()
	// Load networks
	err := generator.loadInputGeneratorNetwork()
	if err != nil {
		return nil, err
	}
	err = generator.loadHeaderNetwork()
	if err != nil {
		return nil, err
	}

	return generator, nil
}

// GenerateHeaders generates HTTP headers based on the given options
func (g *HeaderGenerator) GenerateHeaders(options HeaderConstraints) (map[string]string, error) {
	// Merge user constraints with defaults
	constraints, err := g.mergeOptions(options)
	if err != nil {
		return nil, err
	}

	// Prepare input constraints
	inputConstraints, err := g.prepareConstraints(constraints)
	if err != nil {
		return nil, err
	}

	// Generate input values using the input generator network (randomized)
	inputSample, ok := g.inputGeneratorNetwork.generateConsistentSampleWhenPossible(inputConstraints)
	if !ok {
		// fallback to default values
		if constraints.HTTPVersion == "1" {
			// Try with HTTP/2
			constraints.HTTPVersion = "2"
			headers, err := g.GenerateHeaders(constraints)
			if err != nil {
				return nil, err
			}
			return pascalizeHeaders(headers), nil
		}
		// If the input generation failed and strict mode is enabled, return an error
		if constraints.Strict {
			return nil, fmt.Errorf("no headers based on this input can be generated. Please relax or change some of the requirements you specified")
		}

		// TODO: we can remove one by one
		// Relax constraints
		relaxedConstraints := constraints
		relaxedConstraints.Locales = nil
		relaxedConstraints.Devices = nil
		headers, err := g.GenerateHeaders(relaxedConstraints)
		if err != nil {
			return nil, err
		}
		return pascalizeHeaders(headers), nil
	}

	// Generate headers using the header network
	sample := g.headerGeneratorNetwork.generateSample(inputSample)

	// Generate headers from sample
	headers := g.generateHeadersFromSample(sample)

	// Add Accept-Language header
	if len(constraints.Locales) > 0 {
		acceptLanguage := g.generateAcceptLanguageHeader(constraints.Locales)
		if sample["*HTTP_VERSION"] == "2" {
			headers["accept-language"] = acceptLanguage
		} else {
			headers["Accept-Language"] = acceptLanguage
		}
	}

	// Add Sec-Fetch headers if needed
	browser := g.prepareHttpBrowserObject(sample["*BROWSER_HTTP"])
	if browser != nil && g.shouldAddSecFetch(browser) {
		if browser.IsHTTP2() {
			for k, v := range http2SecFetchAttributes {
				headers[k] = v
			}
		} else {
			for k, v := range http1SecFetchAttributes {
				headers[k] = v
			}
		}
	}

	// TODO: implement header reordering
	// Pascalize headers for HTTP/2
	if constraints.HTTPVersion == "2" {
		return pascalizeHeaders(headers), nil
	}

	return nil, nil
}

// getPossibleAttributeValues returns the possible values for each attribute
func (g *HeaderGenerator) getPossibleAttributeValues(options HeaderConstraints) map[string][]string {
	values := make(map[string][]string)

	// Get browser HTTP options
	values["*BROWSER_HTTP"] = g.getBrowserHTTPOptions(options)

	// Get OS options
	if len(options.OS) > 0 {
		values["*OPERATING_SYSTEM"] = options.OS
	}

	// Get device options
	if len(options.Devices) > 0 {
		values["*DEVICE"] = options.Devices
	}

	return values
}

// getBrowserHTTPOptions returns the browser HTTP options based on the given constraints
func (g *HeaderGenerator) getBrowserHTTPOptions(options HeaderConstraints) []string {
	var result []string

	// BrowserSpecs are specified
	if len(options.BrowserSpecs) > 0 {
		for _, browser := range options.BrowserSpecs {
			for _, uniqueBrowser := range g.uniqueBrowsers {
				if uniqueBrowser.Name != nil && *uniqueBrowser.Name == browser.Name {
					if browser.HTTPVersion == uniqueBrowser.HTTPVersion || browser.HTTPVersion == "" {
						// Check version constraints if specified
						if browser.MinVersion > 0 && uniqueBrowser.Version[0] < browser.MinVersion {
							continue
						}
						if browser.MaxVersion > 0 && uniqueBrowser.Version[0] > browser.MaxVersion {
							continue
						}
						result = append(result, uniqueBrowser.CompleteString)
					}
				}
			}
		}
		return result
	}

	// Otherwise, use browser strings
	for _, browserName := range options.Browsers {
		browser := &BrowserSpec{
			Name:        browserName,
			HTTPVersion: options.HTTPVersion,
		}
		for _, uniqueBrowser := range g.uniqueBrowsers {
			if uniqueBrowser.Name != nil && *uniqueBrowser.Name == browser.Name {
				if browser.HTTPVersion == uniqueBrowser.HTTPVersion || browser.HTTPVersion == "" {
					result = append(result, uniqueBrowser.CompleteString)
				}
			}
		}
	}

	return result
}

// generateHeadersFromSample generates headers from a sample and removes unwanted headers
func (g *HeaderGenerator) generateHeadersFromSample(sample map[string]string) map[string]string {
	headers := make(map[string]string)
	for k, v := range sample {
		if !strings.HasPrefix(k, "*") && v != missingValueToken {
			headers[k] = v
		}
	}
	return headers
}

// generateAcceptLanguageHeader generates the Accept-Language header
func (g *HeaderGenerator) generateAcceptLanguageHeader(locales []string) string {
	var parts []string
	for i, locale := range locales {
		if i >= 10 { // Limit to 10 languages
			break
		}
		q := 1.0 - float64(i)*0.1
		parts = append(parts, fmt.Sprintf("%s;q=%.1f", locale, q))
	}
	return strings.Join(parts, ", ")
}

// shouldAddSecFetch determines if Sec-Fetch headers should be added based on the user agent
func (g *HeaderGenerator) shouldAddSecFetch(browser *httpBrowser) bool {
	if browser == nil || browser.Name == nil || len(browser.Version) == 0 {
		return false
	}

	switch *browser.Name {
	case "chrome":
		return browser.Version[0] >= 76
	case "firefox":
		return browser.Version[0] >= 90
	case "edge":
		return browser.Version[0] >= 79
	default:
		return false
	}
}

// loadHeadersOrder loads the headers order from the headers-order.json file
func (g *HeaderGenerator) loadHeadersOrder() {
	data, err := dataFiles.ReadFile("data_points/headers-order.json")
	if err != nil {
		fmt.Printf("Warning: failed to read headers-order.json: %v\n", err)
		return
	}
	var headersOrder map[string][]string
	if err := json.Unmarshal(data, &headersOrder); err != nil {
		fmt.Printf("Warning: failed to parse headers-order.json: %v\n", err)
		return
	}
	g.headersOrder = headersOrder
}

// loadHeaderNetwork loads the header generator network
func (g *HeaderGenerator) loadHeaderNetwork() error {
	network, err := loadNetworkFromZip("header-network-definition.zip")
	if err != nil {
		return err
	}
	g.headerGeneratorNetwork = network
	return nil
}

// loadInputGeneratorNetwork loads the input generator input-network-definition
func (g *HeaderGenerator) loadInputGeneratorNetwork() error {
	network, err := loadNetworkFromZip("input-network-definition.zip")
	if err != nil {
		return err
	}
	g.inputGeneratorNetwork = network
	return nil
}

// loadUniqueBrowsers loads the unique browsers from the browser-helper-file.json
func (g *HeaderGenerator) loadUniqueBrowsers() {
	g.uniqueBrowsers = make([]*httpBrowser, 0)

	data, err := dataFiles.ReadFile("data_points/browser-helper-file.json")
	if err != nil {
		fmt.Printf("Warning: failed to read browser-helper-file.json: %v\n", err)
		return
	}

	var browserStrings []string
	if err := json.Unmarshal(data, &browserStrings); err != nil {
		fmt.Printf("Warning: failed to parse browser-helper-file.json: %v\n", err)
		return
	}

	// Convert browser strings to httpBrowser
	g.uniqueBrowsers = make([]*httpBrowser, 0, len(browserStrings))
	for _, browserStr := range browserStrings {
		if browserStr == missingValueToken {
			continue
		}
		browser := g.prepareHttpBrowserObject(browserStr)
		if browser != nil {
			g.uniqueBrowsers = append(g.uniqueBrowsers, browser)
		}
	}
}

// prepareHttpBrowserObject extracts structured information about a browser and HTTP version from a string
func (g *HeaderGenerator) prepareHttpBrowserObject(httpBrowserString string) *httpBrowser {
	if httpBrowserString == missingValueToken {
		return &httpBrowser{
			Name:           nil,
			Version:        []int{},
			CompleteString: missingValueToken,
			HTTPVersion:    "",
		}
	}

	parts := strings.Split(httpBrowserString, "|")
	if len(parts) != 2 {
		return nil
	}

	browserString, httpVersion := parts[0], parts[1]
	browserParts := strings.Split(browserString, "/")
	if len(browserParts) != 2 {
		return nil
	}

	browserName, versionString := browserParts[0], browserParts[1]
	versionParts := strings.Split(versionString, ".")
	version := make([]int, len(versionParts))
	for i, part := range versionParts {
		version[i] = atoi(part)
	}

	return &httpBrowser{
		Name:           &browserName,
		Version:        version,
		CompleteString: httpBrowserString,
		HTTPVersion:    httpVersion,
	}
}

// validateAgainstSupported checks if a value exists in the supported values slice
func validateAgainstSupported(value string, supported []string) error {
	for _, s := range supported {
		if value == s {
			return nil
		}
	}
	return fmt.Errorf("value '%s' is not supported", value)
}

// filterValidValues filters a slice of values to only include those that are in the supported values
func filterValidValues(values []string, supported []string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}

	var invalidValues []string
	valid := make([]string, 0, len(values))
	for _, v := range values {
		if err := validateAgainstSupported(v, supported); err != nil {
			invalidValues = append(invalidValues, v)
			continue
		}
		valid = append(valid, v)
	}

	if len(invalidValues) > 0 {
		return valid, fmt.Errorf("the following values are not supported: %v", invalidValues)
	}
	return valid, nil
}

// prepareConstraints prepares and validates the input constraints for the input generator network
func (g *HeaderGenerator) prepareConstraints(options HeaderConstraints) (map[string][]string, error) {
	// Get possible attribute values
	possibleAttributeValues := g.getPossibleAttributeValues(options)

	// Prepare HTTP version specific values
	http1Values := make(map[string][]string)
	http2Values := make(map[string][]string)

	// Always include values for both HTTP versions
	http1Values["*BROWSER"] = options.Browsers
	http1Values["*OPERATING_SYSTEM"] = options.OS
	http1Values["*DEVICE"] = options.Devices
	http2Values["*BROWSER"] = options.Browsers
	http2Values["*OPERATING_SYSTEM"] = options.OS
	http2Values["*DEVICE"] = options.Devices

	// Filter values based on HTTP version
	constraints := make(map[string][]string)
	for key, values := range possibleAttributeValues {
		var filteredValues []string
		for _, value := range values {
			var shouldInclude bool
			if key == "*BROWSER_HTTP" {
				shouldInclude = g.filterBrowserHTTP(value, http1Values, http2Values)
				if shouldInclude {
					filteredValues = append(filteredValues, value)
				}
			} else {
				shouldInclude = g.filterOtherValues(value, http1Values, http2Values, key)
				if shouldInclude {
					filteredValues = append(filteredValues, value)
				}
			}
		}
		if len(filteredValues) > 0 {
			constraints[key] = filteredValues
		}
	}

	return constraints, nil
}

// filterBrowserHTTP filters the browser HTTP value based on the HTTP/1 and HTTP/2 values
func (g *HeaderGenerator) filterBrowserHTTP(value string, http1Values, http2Values map[string][]string) bool {
	parts := strings.Split(value, "|")
	if len(parts) != 2 {
		return false
	}
	browserString, httpVersion := parts[0], parts[1]
	browserParts := strings.Split(browserString, "/")
	if len(browserParts) != 2 {
		return false
	}
	browserName := browserParts[0]

	if httpVersion == "1" {
		if len(http1Values) == 0 {
			return true
		}
		browsers := http1Values["*BROWSER"]
		for _, browser := range browsers {
			if browser == browserName {
				return true
			}
		}
		return false
	} else {
		if len(http2Values) == 0 {
			return true
		}
		browsers := http2Values["*BROWSER"]
		for _, browser := range browsers {
			if browser == browserName {
				return true
			}
		}
		return false
	}
}

// filterOtherValues filters the other attribute values based on the HTTP/1 and HTTP/2 values
func (g *HeaderGenerator) filterOtherValues(value string, http1Values, http2Values map[string][]string, key string) bool {
	// If no values are specified for either HTTP version, include all values
	if len(http1Values) == 0 && len(http2Values) == 0 {
		return true
	}

	// Check if the value exists in either HTTP version's allowed values
	if values, ok := http1Values[key]; ok {
		for _, v := range values {
			if v == value {
				return true
			}
		}
	}

	if values, ok := http2Values[key]; ok {
		for _, v := range values {
			if v == value {
				return true
			}
		}
	}
	return false
}
