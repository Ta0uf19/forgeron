<h1 align="center">
    Forgeron ⚒️
</h1>

<h4 align="center">
    🌐 Intelligent browser header & fingerprint generator
</h4>

## What is it?
Forgeron is a browser header and fingerprint generator that mimics realistic, statistically-informed browser fingerprints and HTTP headers that reflect real-world distributions across browsers, operating systems, and devices.

It is a reimplementation of [Apify's fingerprint-suite](https://github.com/apify/fingerprint-suite) in Go.

Forgeron is ideal for:
- Web scraping and automation where realistic client behavior is essential
- Privacy-focused browsers or anti-fingerprinting research
- Simulating traffic from a variety of platforms and regions

## Features

- Uses a Bayesian generative network to mimic actual web traffic
- Easy and simple for humans to use
- Extensive customization options for browsers, operating systems, devices, locales, and HTTP version
- Browser fingerprint generation: mimics real browser fingerprints (screen, navigator, plugins, fonts, WebRTC, etc.)
- Browser header generation: generates realistic HTTP headers (user agent, accept, sec-fetch, etc.)

## Installation

```bash
go get github.com/ta0uf19/forgeron
```

## Usage

The header generator creates realistic HTTP headers by generating:
- **User-Agent**: Browser, OS, and version information
- **Accept Headers**: Content types, languages, and encodings
- **Sec-Fetch Headers**: Security and fetch mode information
- **Connection Headers**: HTTP version and connection details
- **Additional Headers**: Cache control, upgrade requests, and more

### Basic Example

```go
// Create a header generator
generator, err := forgeron.NewHeaderGenerator()
if err != nil {
    panic(err)
}

// Generate headers with specific constraints
headers, err := generator.GenerateHeaders(forgeron.HeaderConstraints{
    Browsers: []string{"chrome"},
    OS:       []string{"windows", "macos"},
    Devices:  []string{"desktop"},
    Locales:  []string{"en-US"},
})
if err != nil {
    panic(err)
}
```

Example output:
```
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36
sec-ch-ua-mobile: ?0
Sec-Fetch-Mode: same-site
Sec-Fetch-Dest: navigate
Accept-Language: en-US;q=1.0
Sec-Fetch-Site: ?1
sec-ch-ua-platform: "macOS"
Accept-Encoding: gzip, deflate, br, zstd
sec-ch-ua: "Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"
Upgrade-Insecure-Requests: 1
Sec-Fetch-User: document
```

### Header Constraints
The header generator allows you to specify constraints for the generated headers, you can specify one or multiple constraints.
The following constraints are available:
- `Browsers`: A list of browsers to include in the generated headers. (e.g., `["chrome", "firefox", "edge", "safari"]`) or `BrowserSpecs`: A list of browser specification with min, max version 
- `OS`: A list of operating systems to include in the generated headers. (e.g., `["windows", "macos", "linux"]`)
- `Devices`: A list of devices to include in the generated headers. (e.g., `["desktop", "mobile"]`)
- `Locales`: A list of locales to include in the generated headers. (e.g., `["en-US", "fr-FR"]`)
- `HTTPVersion`: The HTTP version to use for the generated headers. (e.g., `"1.1"` or `"2"`)
- `Strict`: A boolean value indicating whether to use strict mode. If set to `true`, the generator will only use the specified constraints and will not fall back to other values.

### Browser specification

Set specificiations for browsers, including version ranges and HTTP version:
```go
// Generate headers with specific constraints
headers, err := generator.GenerateHeaders(forgeron.HeaderConstraints{
    BrowserSpecs: []*forgeron.BrowserSpec{
        {Name: "chrome", MinVersion: 100, MaxVersion: 131},
        {Name: "firefox", MinVersion: 80, HTTPVersion: "1"},
    },
})
```

### Fingerprint Generation

The fingerprint generator creates realistic browser fingerprints by generating:
- **Navigator Properties**: User agent, platform, language, hardware specs
- **Screen Information**: Resolution, color depth, available screen space
- **Plugin Data**: Installed browser plugins and their details
- **Font List**: System and browser fonts
- **WebRTC Data**: Local and public IP addresses
- **Additional Browser Properties**: Canvas, WebGL, AudioContext fingerprints

```go
// Create a fingerprint generator
generator, err := forgeron.NewFingerprintGenerator()
if err != nil {
    panic(err)
}

// Generate a fingerprint
fingerprint, err := generator.Generate(forgeron.WithHeaderConstraints(
    forgeron.HeaderConstraints{
        Browsers: []string{"chrome"},
        OS:       []string{"macos"},
    },
))
```

<details>
<summary>Example response</summary>

```
&forgeron.Fingerprint{
  Screen: forgeron.ScreenFingerprint{
    AvailHeight:      1440,
    AvailWidth:       2560,
    AvailTop:         0,
    AvailLeft:        1440,
    ColorDepth:       24,
    Height:           1440,
    PixelDepth:       24,
    Width:            2560,
    DevicePixelRatio: 2.000000,
    PageXOffset:      0,
    PageYOffset:      0,
    InnerHeight:      0,
    OuterHeight:      1440,
    OuterWidth:       2560,
    InnerWidth:       0,
    ScreenX:          1440,
    ClientWidth:      0,
    ClientHeight:     19,
    HasHDR:           false,
  },
  Navigator: forgeron.NavigatorFingerprint{
    UserAgent:     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
    UserAgentData: &forgeron.UserAgentData{
      Brands: []forgeron.UserAgentBrand{
        forgeron.UserAgentBrand{
          Brand:   "Chromium",
          Version: "128",
        },
        forgeron.UserAgentBrand{
          Brand:   "Not;A=Brand",
          Version: "24",
        },
        forgeron.UserAgentBrand{
          Brand:   "Google Chrome",
          Version: "128",
        },
      },
      Mobile:          false,
      Platform:        "macOS",
      Architecture:    "x86",
      Bitness:         "64",
      FullVersionList: []forgeron.UserAgentBrand{
        forgeron.UserAgentBrand{
          Brand:   "Chromium",
          Version: "128.0.6613.120",
        },
        forgeron.UserAgentBrand{
          Brand:   "Not;A=Brand",
          Version: "24.0.0.0",
        },
        forgeron.UserAgentBrand{
          Brand:   "Google Chrome",
          Version: "128.0.6613.120",
        },
      },
      Model:           "",
      PlatformVersion: "13.7.0",
      UAFullVersion:   "128.0.6613.120",
    },
    DoNotTrack:  (*string)(nil),
    AppCodeName: "Mozilla",
    AppName:     "Netscape",
    AppVersion:  "5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
    OSCpu:       "null",
    Webdriver:   "false",
    Language:    "en-US",
    Languages:   []string{
      "en-US",
    },
    Platform:            "MacIntel",
    DeviceMemory:        &8,
    HardwareConcurrency: 4,
    Product:             "Gecko",
    ProductSub:          "20030107",
    Vendor:              "Google Inc.",
    VendorSub:           "null",
    MaxTouchPoints:      0,
    ExtraProperties:     map[string]interface {}{
      "globalPrivacyControl": nil,
      "installedApps":        []interface {}{},
      "pdfViewerEnabled":     true,
      "vendorFlavors":        []interface {}{
        "chrome",
      },
    },
  },
  Headers: map[string]string{
    "Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
    "Accept-Encoding":           "gzip, deflate, br, zstd",
    "Accept-Language":           "en-US;q=1.0",
    "Sec-Fetch-Dest":            "navigate",
    "Sec-Fetch-Mode":            "same-site",
    "Sec-Fetch-Site":            "?1",
    "Sec-Fetch-User":            "document",
    "Upgrade-Insecure-Requests": "1",
    "User-Agent":                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
    "sec-ch-ua":                 "\"Chromium\";v=\"128\", \"Not;A=Brand\";v=\"24\", \"Google Chrome\";v=\"128\"",
    "sec-ch-ua-mobile":          "?0",
    "sec-ch-ua-platform":        "\"macOS\"",
  },
  VideoCodecs: map[string]string{
    "h264": "probably",
    "ogg":  "",
    "webm": "probably",
  },
  AudioCodecs: map[string]string{
    "aac": "probably",
    "m4a": "maybe",
    "mp3": "probably",
    "ogg": "probably",
    "wav": "probably",
  },
  PluginsData: forgeron.PluginsData{
    Plugins: []forgeron.Plugin{
      forgeron.Plugin{
        Name:        "PDF Viewer",
        Description: "Portable Document Format",
        Filename:    "internal-pdf-viewer",
        MimeTypes:   []forgeron.MimeType{
          forgeron.MimeType{
            Type:          "application/pdf",
            Suffixes:      "pdf",
            Description:   "Portable Document Format",
            EnabledPlugin: "PDF Viewer",
          },
          forgeron.MimeType{
            Type:          "text/pdf",
            Suffixes:      "pdf",
            Description:   "Portable Document Format",
            EnabledPlugin: "PDF Viewer",
          },
        },
      },
      forgeron.Plugin{
        Name:        "Chrome PDF Viewer",
        Description: "Portable Document Format",
        Filename:    "internal-pdf-viewer",
        MimeTypes:   []forgeron.MimeType{
          forgeron.MimeType{
            Type:          "application/pdf",
            Suffixes:      "pdf",
            Description:   "Portable Document Format",
            EnabledPlugin: "Chrome PDF Viewer",
          },
          forgeron.MimeType{
            Type:          "text/pdf",
            Suffixes:      "pdf",
            Description:   "Portable Document Format",
            EnabledPlugin: "Chrome PDF Viewer",
          },
        },
      },
      forgeron.Plugin{
        Name:        "Chromium PDF Viewer",
        Description: "Portable Document Format",
        Filename:    "internal-pdf-viewer",
        MimeTypes:   []forgeron.MimeType{
          forgeron.MimeType{
            Type:          "application/pdf",
            Suffixes:      "pdf",
            Description:   "Portable Document Format",
            EnabledPlugin: "Chromium PDF Viewer",
          },
          forgeron.MimeType{
            Type:          "text/pdf",
            Suffixes:      "pdf",
            Description:   "Portable Document Format",
            EnabledPlugin: "Chromium PDF Viewer",
          },
        },
      },
      forgeron.Plugin{
        Name:        "Microsoft Edge PDF Viewer",
        Description: "Portable Document Format",
        Filename:    "internal-pdf-viewer",
        MimeTypes:   []forgeron.MimeType{
          forgeron.MimeType{
            Type:          "application/pdf",
            Suffixes:      "pdf",
            Description:   "Portable Document Format",
            EnabledPlugin: "Microsoft Edge PDF Viewer",
          },
          forgeron.MimeType{
            Type:          "text/pdf",
            Suffixes:      "pdf",
            Description:   "Portable Document Format",
            EnabledPlugin: "Microsoft Edge PDF Viewer",
          },
        },
      },
      forgeron.Plugin{
        Name:        "WebKit built-in PDF",
        Description: "Portable Document Format",
        Filename:    "internal-pdf-viewer",
        MimeTypes:   []forgeron.MimeType{
          forgeron.MimeType{
            Type:          "application/pdf",
            Suffixes:      "pdf",
            Description:   "Portable Document Format",
            EnabledPlugin: "WebKit built-in PDF",
          },
          forgeron.MimeType{
            Type:          "text/pdf",
            Suffixes:      "pdf",
            Description:   "Portable Document Format",
            EnabledPlugin: "WebKit built-in PDF",
          },
        },
      },
    },
    MimeTypes: []string{
      "Portable Document Format~~application/pdf~~pdf",
      "Portable Document Format~~text/pdf~~pdf",
    },
  },
  Battery: &forgeron.Battery{
    Charging:        false,
    ChargingTime:    (*int)(nil),
    DischargingTime: (*int)(nil),
    Level:           0.090000,
  },
  VideoCard: &forgeron.VideoCard{
    Renderer: "ANGLE (NVIDIA Corporation, NVIDIA GeForce GT 650M OpenGL Engine, OpenGL 4.1)",
    Vendor:   "Google Inc. (NVIDIA Corporation)",
  },
  MultimediaDevices: &forgeron.MultimediaDevices{
    Speakers: []forgeron.MediaDevice{
      forgeron.MediaDevice{
        DeviceID: "",
        Kind:     "audiooutput",
        Label:    "",
        GroupID:  "",
      },
    },
    Micros: []forgeron.MediaDevice{
      forgeron.MediaDevice{
        DeviceID: "",
        Kind:     "audioinput",
        Label:    "",
        GroupID:  "",
      },
    },
    Webcams: []forgeron.MediaDevice{
      forgeron.MediaDevice{
        DeviceID: "",
        Kind:     "videoinput",
        Label:    "",
        GroupID:  "",
      },
    },
  },
  Fonts: []string{
    "Arial Unicode MS",
    "Batang",
    "Calibri",
    "Century",
    "Century Gothic",
    "EUROSTILE",
    "Gill Sans",
    "Haettenschweiler",
    "Helvetica Neue",
    "Lucida Bright",
    "Lucida Sans",
    "Menlo",
    "MS Mincho",
    "MS Reference Specialty",
    "MT Extra",
    "MYRIAD PRO",
    "Marlett",
    "Minion Pro",
    "Monotype Corsiva",
    "PMingLiU",
  },
  MockWebRTC: false,
  Slim:       false,
}
```
</details>

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 