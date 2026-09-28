package main

// Wails' macOS half uses UTType, which lives in the UniformTypeIdentifiers framework. The wails
// command adds that framework to every build it runs (read in Wails v2.12.0,
// pkg/commands/build/base.go); TimeRibbon builds with go build, so it is named here instead, where
// no build can leave it out.

// #cgo LDFLAGS: -framework UniformTypeIdentifiers
import "C"
