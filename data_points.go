package forgeron

// DataPointsVersion is the upstream apify/fingerprint-suite version that the files
// embedded from data_points/ were taken from.
//
// The files are published identically by the npm packages header-generator and
// fingerprint-generator at this version, and by the PyPI package
// apify-fingerprint-datapoints (whose own version line differs; 2.1.88 corresponds to
// apify-fingerprint-datapoints 0.15.0).
//
// When syncing new data, bump this constant and update the digests in
// dataPointChecksums (see data_points_test.go) in the same commit. TestDataPointChecksums
// fails until both are consistent with the files on disk, so the two cannot drift apart
// silently.
const DataPointsVersion = "2.1.88"
