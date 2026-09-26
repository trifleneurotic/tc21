//go:build !windows

package main

// SetPlatformIcon is a no-op on non-Windows platforms because macOS and Linux
// automatically handle runtime window decorations via standard app bundling formats.
func SetPlatformIcon(windowTitle string) {
	// Intentionally left blank
}
