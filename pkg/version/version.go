package version

// Default build-time variable.
// This value isoverridden via ldflags.
var (
	Version = "unknown-version" //nolint:gochecknoglobals // ok
)
