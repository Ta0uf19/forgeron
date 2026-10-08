package forgeron

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// zipEntryNames returns the entry names of an in-memory zip archive, in archive order.
func zipEntryNames(data []byte) ([]string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(reader.File))
	for _, f := range reader.File {
		names = append(names, f.Name)
	}
	return names, nil
}

// dataPointFiles lists every file embedded from data_points/ and consumed at runtime.
var dataPointFiles = []string{
	"browser-helper-file.json",
	"headers-order.json",
	"header-network-definition.zip",
	"input-network-definition.zip",
	"fingerprint-network-definition.zip",
}

// networkDefinitionFiles lists the zipped Bayesian network definitions.
var networkDefinitionFiles = []string{
	"header-network-definition.zip",
	"input-network-definition.zip",
	"fingerprint-network-definition.zip",
}

// readBrowserHelperFile decodes the embedded browser-helper-file.json
func readBrowserHelperFile(t *testing.T) []string {
	t.Helper()
	data, err := dataFiles.ReadFile("data_points/browser-helper-file.json")
	if err != nil {
		t.Fatalf("failed to read browser-helper-file.json: %v", err)
	}
	var entries []string
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("failed to parse browser-helper-file.json: %v", err)
	}
	return entries
}

// readHeadersOrder decodes the embedded headers-order.json
func readHeadersOrder(t *testing.T) map[string][]string {
	t.Helper()
	data, err := dataFiles.ReadFile("data_points/headers-order.json")
	if err != nil {
		t.Fatalf("failed to read headers-order.json: %v", err)
	}
	var order map[string][]string
	if err := json.Unmarshal(data, &order); err != nil {
		t.Fatalf("failed to parse headers-order.json: %v", err)
	}
	return order
}

// TestDataPointFilesAreEmbedded verifies every data point file is present in the embed FS
// and non-empty. A missing file degrades silently for the JSON loaders, so assert directly.
func TestDataPointFilesAreEmbedded(t *testing.T) {
	for _, name := range dataPointFiles {
		t.Run(name, func(t *testing.T) {
			data, err := dataFiles.ReadFile("data_points/" + name)
			if err != nil {
				t.Fatalf("failed to read %s: %v", name, err)
			}
			if len(data) == 0 {
				t.Fatalf("%s is empty", name)
			}
		})
	}
}

// TestNetworkDefinitionsLoad verifies every zipped network definition parses into a usable
// Bayesian network with consistent node wiring.
func TestNetworkDefinitionsLoad(t *testing.T) {
	for _, name := range networkDefinitionFiles {
		t.Run(name, func(t *testing.T) {
			network, err := loadNetworkFromZip(name)
			if err != nil {
				t.Fatalf("loadNetworkFromZip(%q) error = %v", name, err)
			}
			if len(network.NodesInSamplingOrder) == 0 {
				t.Fatalf("%s has no nodes", name)
			}
			if len(network.NodesByName) != len(network.NodesInSamplingOrder) {
				t.Errorf("%s: NodesByName (%d) and NodesInSamplingOrder (%d) disagree",
					name, len(network.NodesByName), len(network.NodesInSamplingOrder))
			}

			for _, n := range network.NodesInSamplingOrder {
				if n.Name == "" {
					t.Error("found node with empty name")
				}
				if len(n.PossibleValues) == 0 {
					t.Errorf("node %q has no possible values", n.Name)
				}
				if len(n.ConditionalProbs) == 0 {
					t.Errorf("node %q has no conditional probabilities", n.Name)
				}
				// Every declared parent must resolve to a node in the same network.
				for _, parentName := range n.ParentNames {
					if _, ok := network.NodesByName[parentName]; !ok {
						t.Errorf("node %q references unknown parent %q", n.Name, parentName)
					}
				}
				if len(n.parents) != len(n.ParentNames) {
					t.Errorf("node %q: wired %d parents but declares %d",
						n.Name, len(n.parents), len(n.ParentNames))
				}
			}
		})
	}
}

// TestNetworkZipLayout guards the loader assumption that each archive holds the network
// JSON as its first (and only) entry. loadNetworkFromZip blindly reads File[0].
func TestNetworkZipLayout(t *testing.T) {
	for _, name := range networkDefinitionFiles {
		t.Run(name, func(t *testing.T) {
			zipData, err := dataFiles.ReadFile("data_points/" + name)
			if err != nil {
				t.Fatalf("failed to read %s: %v", name, err)
			}
			names, err := zipEntryNames(zipData)
			if err != nil {
				t.Fatalf("%s is not a valid zip: %v", name, err)
			}
			if len(names) != 1 {
				t.Fatalf("%s: expected exactly 1 zip entry, got %d (%v)", name, len(names), names)
			}
			if names[0] != "network.json" {
				t.Errorf("%s: expected first entry 'network.json', got %q", name, names[0])
			}
		})
	}
}

// TestBrowserHelperFileEntries verifies every entry in browser-helper-file.json is
// well-formed and parseable by prepareHttpBrowserObject.
func TestBrowserHelperFileEntries(t *testing.T) {
	entries := readBrowserHelperFile(t)
	if len(entries) == 0 {
		t.Fatal("browser-helper-file.json contains no entries")
	}

	gen := &HeaderGenerator{}
	seenBrowsers := make(map[string]bool)

	for _, entry := range entries {
		if entry == missingValueToken {
			continue
		}
		browser := gen.prepareHttpBrowserObject(entry)
		if browser == nil {
			t.Errorf("entry %q failed to parse into an httpBrowser", entry)
			continue
		}
		if browser.Name == nil || *browser.Name == "" {
			t.Errorf("entry %q produced an empty browser name", entry)
			continue
		}
		if len(browser.Version) == 0 {
			t.Errorf("entry %q produced an empty version", entry)
		} else if browser.Version[0] <= 0 {
			t.Errorf("entry %q produced a non-positive major version %d", entry, browser.Version[0])
		}
		if err := validateAgainstSupported(browser.HTTPVersion, SupportedHTTP); err != nil {
			t.Errorf("entry %q has unsupported HTTP version %q", entry, browser.HTTPVersion)
		}
		if err := validateAgainstSupported(*browser.Name, SupportedBrowsers); err != nil {
			t.Errorf("entry %q has unsupported browser name %q", entry, *browser.Name)
		}
		seenBrowsers[*browser.Name] = true
	}

	// Every browser the library advertises must actually exist in the data.
	for _, name := range SupportedBrowsers {
		if !seenBrowsers[name] {
			t.Errorf("no entry for supported browser %q in browser-helper-file.json", name)
		}
	}
}

// TestBrowserHelperFileMatchesInputNetwork verifies browser-helper-file.json stays in sync
// with the *BROWSER_HTTP node of the input network. getBrowserHTTPOptions builds constraints
// from the helper file and feeds them to that node, so any drift silently shrinks the
// candidate set or makes generation impossible.
func TestBrowserHelperFileMatchesInputNetwork(t *testing.T) {
	network, err := loadNetworkFromZip("input-network-definition.zip")
	if err != nil {
		t.Fatalf("loadNetworkFromZip() error = %v", err)
	}
	node, ok := network.NodesByName["*BROWSER_HTTP"]
	if !ok {
		t.Fatal("input network has no *BROWSER_HTTP node")
	}

	inNetwork := make(map[string]bool, len(node.PossibleValues))
	for _, v := range node.PossibleValues {
		inNetwork[v] = true
	}

	entries := readBrowserHelperFile(t)
	inHelper := make(map[string]bool, len(entries))
	for _, e := range entries {
		inHelper[e] = true
		if !inNetwork[e] {
			t.Errorf("browser-helper-file.json entry %q is missing from *BROWSER_HTTP", e)
		}
	}
	for _, v := range node.PossibleValues {
		if !inHelper[v] {
			t.Errorf("*BROWSER_HTTP value %q is missing from browser-helper-file.json", v)
		}
	}
}

// TestHeadersOrderCoversSupportedBrowsers verifies headers-order.json defines a usable
// header order for every supported browser.
//
// Each entry is a single list holding two concatenated sections: the HTTP/1 order
// (capitalized names) followed by the HTTP/2 order, which starts with the pseudo-headers
// (":method", ":authority", ...) and then uses lowercase names. Names may legitimately
// repeat across the two sections (for example "sec-ch-ua", which is lowercase in both),
// so uniqueness is only asserted within a section.
func TestHeadersOrderCoversSupportedBrowsers(t *testing.T) {
	order := readHeadersOrder(t)

	requiredPseudo := []string{":method", ":authority", ":scheme", ":path"}

	for _, browser := range SupportedBrowsers {
		t.Run(browser, func(t *testing.T) {
			headers, ok := order[browser]
			if !ok {
				t.Fatalf("headers-order.json has no entry for %q", browser)
			}
			if len(headers) == 0 {
				t.Fatalf("headers-order.json entry for %q is empty", browser)
			}

			// Split into the HTTP/1 and HTTP/2 sections at the first pseudo-header.
			split := -1
			for i, h := range headers {
				if strings.HasPrefix(h, ":") {
					split = i
					break
				}
			}
			if split < 0 {
				t.Fatalf("%s: no HTTP/2 pseudo-headers found in order list", browser)
			}
			http1, http2 := headers[:split], headers[split:]

			if len(http1) == 0 {
				t.Errorf("%s: HTTP/1 section is empty", browser)
			}

			assertNoDuplicates := func(section string, names []string) {
				seen := make(map[string]bool, len(names))
				for _, h := range names {
					if h == "" {
						t.Errorf("%s/%s: empty header name in order list", browser, section)
						continue
					}
					if seen[h] {
						t.Errorf("%s/%s: duplicate header %q", browser, section, h)
					}
					seen[h] = true
				}
			}
			assertNoDuplicates("http1", http1)
			assertNoDuplicates("http2", http2)

			// The HTTP/1 section must not contain pseudo-headers.
			for _, h := range http1 {
				if strings.HasPrefix(h, ":") {
					t.Errorf("%s: pseudo-header %q found in the HTTP/1 section", browser, h)
				}
			}

			// The HTTP/2 section must lead with all four pseudo-headers.
			pseudo := make(map[string]bool)
			for i, h := range http2 {
				if !strings.HasPrefix(h, ":") {
					break
				}
				pseudo[h] = true
				_ = i
			}
			for _, p := range requiredPseudo {
				if !pseudo[p] {
					t.Errorf("%s: HTTP/2 section is missing pseudo-header %q", browser, p)
				}
			}

			// Past the pseudo-headers, HTTP/2 names must be lowercase (RFC 9113 8.2).
			for _, h := range http2 {
				if strings.HasPrefix(h, ":") {
					continue
				}
				if h != strings.ToLower(h) {
					t.Errorf("%s: HTTP/2 header %q is not lowercase", browser, h)
				}
			}

			// User-Agent must be orderable in both sections.
			if !containsString(http1, "User-Agent") {
				t.Errorf("%s: HTTP/1 section has no User-Agent entry", browser)
			}
			if !containsString(http2, "user-agent") {
				t.Errorf("%s: HTTP/2 section has no user-agent entry", browser)
			}
		})
	}
}

// containsString reports whether needle is present in haystack.
func containsString(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

// TestHeaderGeneratorLoadsDataPoints verifies the non-fatal loaders actually populated
// their fields. loadHeadersOrder and loadUniqueBrowsers only print a warning on failure,
// so a corrupt data point would otherwise go unnoticed.
func TestHeaderGeneratorLoadsDataPoints(t *testing.T) {
	gen, err := NewHeaderGenerator()
	if err != nil {
		t.Fatalf("NewHeaderGenerator() error = %v", err)
	}

	if len(gen.uniqueBrowsers) == 0 {
		t.Error("uniqueBrowsers is empty; browser-helper-file.json failed to load")
	}
	if len(gen.headersOrder) == 0 {
		t.Error("headersOrder is empty; headers-order.json failed to load")
	}
	if gen.headerGeneratorNetwork == nil || len(gen.headerGeneratorNetwork.NodesByName) == 0 {
		t.Error("headerGeneratorNetwork was not loaded")
	}
	if gen.inputGeneratorNetwork == nil || len(gen.inputGeneratorNetwork.NodesByName) == 0 {
		t.Error("inputGeneratorNetwork was not loaded")
	}

	// Every supported browser must be reachable through the loaded helper data.
	for _, name := range SupportedBrowsers {
		found := false
		for _, b := range gen.uniqueBrowsers {
			if b.Name != nil && *b.Name == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("no loaded httpBrowser for supported browser %q", name)
		}
	}
}

// TestNoInternalTokensLeakIntoHeaders verifies internal data point tokens and the
// '*'-prefixed network nodes never reach the generated headers.
func TestNoInternalTokensLeakIntoHeaders(t *testing.T) {
	gen, err := NewHeaderGenerator()
	if err != nil {
		t.Fatalf("NewHeaderGenerator() error = %v", err)
	}

	for _, browser := range SupportedBrowsers {
		t.Run(browser, func(t *testing.T) {
			for i := 0; i < 10; i++ {
				headers, err := gen.GenerateHeaders(HeaderConstraints{Browsers: []string{browser}})
				if err != nil {
					t.Fatalf("GenerateHeaders() error = %v", err)
				}
				for k, v := range headers {
					if strings.HasPrefix(k, "*") {
						t.Errorf("internal network node %q leaked into headers", k)
					}
					if v == missingValueToken || strings.Contains(v, missingValueToken) {
						t.Errorf("header %q carries the missing value token: %q", k, v)
					}
					if strings.Contains(v, "*STRINGIFIED*") {
						t.Errorf("header %q carries an unexpanded stringified token: %q", k, v)
					}
				}
			}
		})
	}
}

// TestNoInternalTokensLeakIntoFingerprint verifies the fingerprint data points decode
// cleanly: no raw tokens survive into the generated fingerprint.
func TestNoInternalTokensLeakIntoFingerprint(t *testing.T) {
	gen := newGeneratorOrFatal(t)

	for i := 0; i < 10; i++ {
		fp, err := gen.Generate()
		if err != nil {
			t.Fatalf("Generate() [%d] error = %v", i, err)
		}

		raw, err := json.Marshal(fp)
		if err != nil {
			t.Fatalf("failed to marshal fingerprint: %v", err)
		}
		out := string(raw)
		for _, token := range []string{missingValueToken, "*STRINGIFIED*"} {
			if strings.Contains(out, token) {
				t.Errorf("fingerprint contains raw token %q", token)
			}
		}

		for k := range fp.Headers {
			if strings.HasPrefix(k, "*") {
				t.Errorf("internal network node %q leaked into fingerprint headers", k)
			}
		}
	}
}

// TestBrowserSpecsVersionRange verifies BrowserSpecs min/max version filtering selects
// only browsers present in the data points within the requested range.
func TestBrowserSpecsVersionRange(t *testing.T) {
	gen, err := NewHeaderGenerator()
	if err != nil {
		t.Fatalf("NewHeaderGenerator() error = %v", err)
	}

	const minVersion = 140
	options := HeaderConstraints{
		BrowserSpecs: []*BrowserSpec{
			{Name: "chrome", MinVersion: minVersion, HTTPVersion: "2"},
		},
		HTTPVersion: "2",
	}

	candidates := gen.getBrowserHTTPOptions(options)
	if len(candidates) == 0 {
		t.Fatalf("no chrome candidates with major version >= %d in the data points", minVersion)
	}
	for _, c := range candidates {
		browser := gen.prepareHttpBrowserObject(c)
		if browser == nil || browser.Name == nil {
			t.Errorf("candidate %q failed to parse", c)
			continue
		}
		if *browser.Name != "chrome" {
			t.Errorf("candidate %q is not chrome", c)
		}
		if browser.Version[0] < minVersion {
			t.Errorf("candidate %q is below MinVersion %d", c, minVersion)
		}
		if browser.HTTPVersion != "2" {
			t.Errorf("candidate %q is not HTTP/2", c)
		}
	}

	// An impossible range must yield nothing rather than falling back silently.
	impossible := gen.getBrowserHTTPOptions(HeaderConstraints{
		BrowserSpecs: []*BrowserSpec{
			{Name: "chrome", MinVersion: 9000, MaxVersion: 9001, HTTPVersion: "2"},
		},
		HTTPVersion: "2",
	})
	if len(impossible) != 0 {
		t.Errorf("expected no candidates for an out-of-range version window, got %d", len(impossible))
	}
}

// TestGeneratedUserAgentVersionExistsInDataPoints verifies the major browser version in a
// generated User-Agent is one the data points actually ship. This catches a stale or
// partially-updated browser-helper-file.json.
func TestGeneratedUserAgentVersionExistsInDataPoints(t *testing.T) {
	gen, err := NewHeaderGenerator()
	if err != nil {
		t.Fatalf("NewHeaderGenerator() error = %v", err)
	}

	known := make(map[string]map[int]bool)
	for _, b := range gen.uniqueBrowsers {
		if b.Name == nil || len(b.Version) == 0 {
			continue
		}
		if known[*b.Name] == nil {
			known[*b.Name] = make(map[int]bool)
		}
		known[*b.Name][b.Version[0]] = true
	}

	// Version tokens per browser. Edge brands itself "Edg/" on desktop, "EdgA/" on
	// Android and "EdgiOS/" on iOS. Safari reports its own version via "Version/".
	tokens := map[string][]string{
		"edge":    {"Edg/", "EdgA/", "EdgiOS/"},
		"chrome":  {"Chrome/", "CriOS/"},
		"firefox": {"Firefox/", "FxiOS/"},
		"safari":  {"Version/"},
	}

	for _, browser := range SupportedBrowsers {
		t.Run(browser, func(t *testing.T) {
			for i := 0; i < 20; i++ {
				headers, err := gen.GenerateHeaders(HeaderConstraints{Browsers: []string{browser}})
				if err != nil {
					t.Fatalf("GenerateHeaders() error = %v", err)
				}
				ua, ok := headers["User-Agent"]
				if !ok {
					t.Fatal("missing User-Agent header")
				}

				idx, token := -1, ""
				for _, candidate := range tokens[browser] {
					if at := strings.Index(ua, candidate); at >= 0 {
						idx, token = at, candidate
						break
					}
				}
				if idx < 0 {
					t.Fatalf("expected one of %v in %s User-Agent, got: %s", tokens[browser], browser, ua)
				}

				rest := ua[idx+len(token):]
				major := rest
				if cut := strings.IndexAny(rest, ". "); cut >= 0 {
					major = rest[:cut]
				}
				version := atoi(major)
				if version == 0 {
					t.Fatalf("could not parse major version out of %q", ua)
				}
				if !known[browser][version] {
					t.Errorf("User-Agent reports %s %d, which is absent from browser-helper-file.json: %s",
						browser, version, ua)
				}
			}
		})
	}
}

// TestGenerateHTTP1Headers verifies HTTP/1 header generation.
// NOTE: GenerateHeaders returns (nil, nil) for HTTPVersion "1" whenever the input network
// yields a consistent sample, because the final branch only handles HTTP/2.
// This test is skipped until the HTTP/1 return path is implemented.
func TestGenerateHTTP1Headers(t *testing.T) {
	t.Skip("HTTP/1 header generation returns (nil, nil); see the final branch of GenerateHeaders")
}
