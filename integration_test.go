package forgeron

import (
	"strings"
	"testing"
)

// newGeneratorOrFatal creates a FingerprintGenerator or fails the test
func newGeneratorOrFatal(t *testing.T, opts ...FingerprintOption) *FingerprintGenerator {
	t.Helper()
	gen, err := NewFingerprintGenerator(opts...)
	if err != nil {
		t.Fatalf("NewFingerprintGenerator() error = %v", err)
	}
	return gen
}

// TestGenerateBasic checks that a fingerprint can be generated with no constraints
func TestGenerateBasic(t *testing.T) {
	gen := newGeneratorOrFatal(t)
	fp, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if fp.Navigator.UserAgent == "" {
		t.Error("UserAgent is empty")
	}
	if fp.Screen.Width == 0 || fp.Screen.Height == 0 {
		t.Errorf("Screen dimensions are zero: %dx%d", fp.Screen.Width, fp.Screen.Height)
	}
	if len(fp.Navigator.Languages) == 0 {
		t.Error("Languages is empty")
	}
	if fp.Navigator.Language == "" {
		t.Error("Language is empty")
	}
	if fp.Navigator.Language != fp.Navigator.Languages[0] {
		t.Errorf("Language (%q) does not match first entry of Languages (%q)", fp.Navigator.Language, fp.Navigator.Languages[0])
	}
	if len(fp.Headers) == 0 {
		t.Error("Headers map is empty")
	}
	if fp.Navigator.HardwareConcurrency == 0 {
		t.Error("HardwareConcurrency is zero")
	}
}

// TestGeneratePerBrowser verifies fingerprints can be generated for each supported browser
func TestGeneratePerBrowser(t *testing.T) {
	browsers := []string{"chrome", "firefox", "safari", "edge"}
	for _, browser := range browsers {
		t.Run(browser, func(t *testing.T) {
			gen := newGeneratorOrFatal(t)
			fp, err := gen.Generate(WithHeaderConstraints(HeaderConstraints{
				Browsers: []string{browser},
			}))
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			ua := strings.ToLower(fp.Navigator.UserAgent)
			switch browser {
			case "chrome":
				if !strings.Contains(ua, "chrome") {
					t.Errorf("expected chrome UA, got: %s", fp.Navigator.UserAgent)
				}
			case "firefox":
				if !strings.Contains(ua, "firefox") {
					t.Errorf("expected firefox UA, got: %s", fp.Navigator.UserAgent)
				}
			case "safari":
				if !strings.Contains(ua, "safari") {
					t.Errorf("expected safari UA, got: %s", fp.Navigator.UserAgent)
				}
			case "edge":
				if !strings.Contains(ua, "edg") {
					t.Errorf("expected edge UA, got: %s", fp.Navigator.UserAgent)
				}
			}
		})
	}
}

// TestGeneratePerOS verifies fingerprints can be generated for each supported OS
func TestGeneratePerOS(t *testing.T) {
	oses := []string{"windows", "macos", "linux"}
	for _, os := range oses {
		t.Run(os, func(t *testing.T) {
			gen := newGeneratorOrFatal(t)
			fp, err := gen.Generate(WithHeaderConstraints(HeaderConstraints{
				OS: []string{os},
			}))
			if err != nil {
				t.Fatalf("Generate() for OS %q error = %v", os, err)
			}
			if fp.Navigator.UserAgent == "" {
				t.Errorf("OS %q: UserAgent is empty", os)
			}
			ua := strings.ToLower(fp.Navigator.UserAgent)
			switch os {
			case "windows":
				if !strings.Contains(ua, "windows") {
					t.Errorf("expected Windows UA, got: %s", fp.Navigator.UserAgent)
				}
			case "macos":
				if !strings.Contains(ua, "mac") && !strings.Contains(ua, "macintosh") {
					t.Errorf("expected macOS UA, got: %s", fp.Navigator.UserAgent)
				}
			case "linux":
				if !strings.Contains(ua, "linux") {
					t.Errorf("expected Linux UA, got: %s", fp.Navigator.UserAgent)
				}
			}
		})
	}
}

// TestGenerateDesktopVsMobile verifies desktop vs mobile fingerprints differ appropriately
func TestGenerateDesktopVsMobile(t *testing.T) {
	gen := newGeneratorOrFatal(t)

	desktop, err := gen.Generate(WithHeaderConstraints(HeaderConstraints{
		Devices: []string{"desktop"},
	}))
	if err != nil {
		t.Fatalf("Generate(desktop) error = %v", err)
	}

	mobile, err := gen.Generate(WithHeaderConstraints(HeaderConstraints{
		Devices:  []string{"mobile"},
		Browsers: []string{"chrome"},
		OS:       []string{"android"},
	}))
	if err != nil {
		t.Fatalf("Generate(mobile) error = %v", err)
	}

	if desktop.Navigator.UserAgent == "" {
		t.Error("desktop UserAgent is empty")
	}
	if mobile.Navigator.UserAgent == "" {
		t.Error("mobile UserAgent is empty")
	}

	// Mobile UA should contain "Mobile"
	if !strings.Contains(mobile.Navigator.UserAgent, "Mobile") {
		t.Errorf("mobile UA expected to contain 'Mobile', got: %s", mobile.Navigator.UserAgent)
	}
}

// TestGenerateMultipleFingerprintsAreUnique verifies that repeated calls produce varied output
func TestGenerateMultipleFingerprintsAreUnique(t *testing.T) {
	gen := newGeneratorOrFatal(t)

	const n = 20
	uas := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		fp, err := gen.Generate()
		if err != nil {
			t.Fatalf("Generate() [%d] error = %v", i, err)
		}
		uas[fp.Navigator.UserAgent] = struct{}{}
	}

	if len(uas) < 3 {
		t.Errorf("expected at least 3 distinct user agents in %d runs, got %d", n, len(uas))
	}
}

// TestGenerateScreenConstraints verifies screen dimension constraints are respected.
// NOTE: screen constraint filtering is not yet implemented (TODO in fingerprint_generator.go).
// This test is skipped until the feature is built.
func TestGenerateScreenConstraints(t *testing.T) {
	t.Skip("screen constraint filtering not yet implemented (see TODO in fingerprint_generator.go)")
}

// TestGenerateMockWebRTC verifies the MockWebRTC flag is reflected in output
func TestGenerateMockWebRTC(t *testing.T) {
	gen := newGeneratorOrFatal(t, WithMockWebRTC(true))
	fp, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if !fp.MockWebRTC {
		t.Error("expected MockWebRTC = true")
	}
}

// TestGenerateSlim verifies the Slim flag is reflected in output
func TestGenerateSlim(t *testing.T) {
	gen := newGeneratorOrFatal(t, WithSlim(true))
	fp, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if !fp.Slim {
		t.Error("expected Slim = true")
	}
}

// TestGenerateLocale verifies the Accept-Language header reflects requested locale
func TestGenerateLocale(t *testing.T) {
	tests := []struct {
		locale string
	}{
		{"fr-FR"},
		{"de-DE"},
		{"ja-JP"},
	}
	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			gen := newGeneratorOrFatal(t)
			fp, err := gen.Generate(WithHeaderConstraints(HeaderConstraints{
				Locales: []string{tt.locale},
			}))
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			acceptLang, ok := fp.Headers["Accept-Language"]
			if !ok {
				t.Fatal("missing Accept-Language header")
			}
			if !strings.Contains(acceptLang, tt.locale) {
				t.Errorf("Accept-Language %q does not contain locale %q", acceptLang, tt.locale)
			}
		})
	}
}

// TestGenerateChromeHasUserAgentData verifies Chrome fingerprints include UserAgentData
func TestGenerateChromeHasUserAgentData(t *testing.T) {
	gen := newGeneratorOrFatal(t)
	fp, err := gen.Generate(WithHeaderConstraints(HeaderConstraints{
		Browsers: []string{"chrome"},
	}))
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if fp.Navigator.UserAgentData == nil {
		t.Fatal("expected UserAgentData for Chrome, got nil")
	}
	if len(fp.Navigator.UserAgentData.Brands) == 0 {
		t.Error("UserAgentData.Brands is empty")
	}
	if fp.Navigator.UserAgentData.Platform == "" {
		t.Error("UserAgentData.Platform is empty")
	}
}

// TestGenerateFirefoxHasNoUserAgentData verifies Firefox fingerprints have no UserAgentData
func TestGenerateFirefoxHasNoUserAgentData(t *testing.T) {
	gen := newGeneratorOrFatal(t)
	for i := 0; i < 5; i++ {
		fp, err := gen.Generate(WithHeaderConstraints(HeaderConstraints{
			Browsers: []string{"firefox"},
		}))
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if fp.Navigator.UserAgentData != nil {
			t.Errorf("expected no UserAgentData for Firefox, got: %+v", fp.Navigator.UserAgentData)
		}
	}
}

// TestGenerateScreenAvailDimensions verifies avail dimensions are <= full dimensions
func TestGenerateScreenAvailDimensions(t *testing.T) {
	gen := newGeneratorOrFatal(t)
	for i := 0; i < 10; i++ {
		fp, err := gen.Generate()
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		s := fp.Screen
		if s.AvailWidth > s.Width {
			t.Errorf("AvailWidth (%d) > Width (%d)", s.AvailWidth, s.Width)
		}
		if s.AvailHeight > s.Height {
			t.Errorf("AvailHeight (%d) > Height (%d)", s.AvailHeight, s.Height)
		}
	}
}

// TestHeaderGeneratorStandalone verifies the header generator works independently
func TestHeaderGeneratorStandalone(t *testing.T) {
	hgen, err := NewHeaderGenerator()
	if err != nil {
		t.Fatalf("NewHeaderGenerator() error = %v", err)
	}

	tests := []struct {
		name        string
		constraints HeaderConstraints
	}{
		{"defaults", HeaderConstraints{}},
		{"chrome only", HeaderConstraints{Browsers: []string{"chrome"}}},
		{"firefox only", HeaderConstraints{Browsers: []string{"firefox"}}},
		{"safari only", HeaderConstraints{Browsers: []string{"safari"}}},
		{"edge only", HeaderConstraints{Browsers: []string{"edge"}}},
		{"windows", HeaderConstraints{OS: []string{"windows"}}},
		{"macos", HeaderConstraints{OS: []string{"macos"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers, err := hgen.GenerateHeaders(tt.constraints)
			if err != nil {
				t.Fatalf("GenerateHeaders() error = %v", err)
			}
			if len(headers) == 0 {
				t.Error("generated headers map is empty")
			}
			if _, ok := headers["User-Agent"]; !ok {
				t.Error("missing User-Agent header")
			}
		})
	}
}

// TestInvalidBrowserReturnsError verifies that an entirely unsupported browser list returns an error.
// mergeOptions collects validation errors and GenerateHeaders propagates them; there is no silent fallback.
func TestInvalidBrowserReturnsError(t *testing.T) {
	gen := newGeneratorOrFatal(t)
	_, err := gen.Generate(WithHeaderConstraints(HeaderConstraints{
		Browsers: []string{"netscape"},
	}))
	if err == nil {
		t.Fatal("expected an error for unsupported browser, got nil")
	}
	if !strings.Contains(err.Error(), "not supported") {
		t.Errorf("expected 'not supported' in error message, got: %v", err)
	}
}

// TestScreenValidation verifies that invalid screen constraints are rejected
func TestScreenValidation(t *testing.T) {
	minW, maxW := 1920, 1024 // intentionally invalid: min > max
	screen := &Screen{MinWidth: &minW, MaxWidth: &maxW}
	if err := screen.Validate(); err == nil {
		t.Error("expected Validate() to return error for minWidth > maxWidth")
	}
}
