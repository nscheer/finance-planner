package planner

// AppVersion is the version of the application; the status bar shows it.
//
// build/config.yml is the single source of the application's metadata (see
// the specification, 7.8), but the Go side cannot read that file at runtime,
// so the version is repeated here. TestAppVersionMatchesBuildConfig compares
// the two and fails when they drift apart.
const AppVersion = "1.0.0"
