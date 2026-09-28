// Package application holds TimeRibbon's use cases: one named entry point per action the user can
// take, each executable from a test with no window open.
//
// It depends on the domain and on the ports declared here. Infrastructure implements the ports;
// the composition root wires them in.
package application

import (
	"time"

	"github.com/oernster/timeribbon/internal/domain/placement"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// Loaded is what the store answers at launch.
type Loaded struct {
	// Settings is what was read; the defaults when nothing was there or the file was unreadable.
	Settings settings.Settings
	// Notice is a problem the user should read, such as a file kept aside (FR-704); empty when none.
	Notice string
}

// Store keeps the settings between runs (FR-701).
type Store interface {
	// Load reads the settings. Absence is not an error (FR-703); an error is a fault reading them.
	Load() (Loaded, error)
	// Save writes the settings, replacing the previous copy atomically (FR-702).
	Save(settings.Settings) error
}

// Place is one entry of the place search (FR-302).
type Place struct {
	// Zone is the IANA zone id.
	Zone string
	// Label is the zone's default label.
	Label string
	// Country is the name of the country the zone lies in; empty where the tz database names none.
	Country string
}

// Zones resolves zone ids and lists the places a clock can be set to.
type Zones interface {
	// Resolve answers the location for zone; an error when the tz database does not know it.
	Resolve(zone string) (*time.Location, error)
	// Catalogue answers every canonical zone as a place.
	Catalogue() []Place
}

// Clock answers the current instant. Only infrastructure reads the wall clock (FR-207).
type Clock interface {
	Now() time.Time
}

// IDs answers a new stable clock id, unique for the life of the settings.
type IDs interface {
	NewID() string
}

// Monitors answers the displays as Windows reports them now.
type Monitors interface {
	Monitors() ([]placement.Monitor, error)
}

// StartupEntry is the Start with Windows value (FR-605).
type StartupEntry interface {
	Enabled() (bool, error)
	Enable() error
	Disable() error
}

// Ports gathers the collaborators a Service is built from.
type Ports struct {
	Store    Store
	Zones    Zones
	Clock    Clock
	IDs      IDs
	Monitors Monitors
	Startup  StartupEntry
	Releases ReleaseSource
	// Build is not a collaborator but the facts about the running build the update check compares
	// against, given here so the composition root states them once.
	Build Build
}
