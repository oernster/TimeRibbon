package application

import (
	"testing"
	"time"

	"github.com/oernster/timestrip/internal/domain/settings"
)

// FR-102, FR-202: cells in order, each with its own zone's day.
func TestSnapshotFollowsClockOrderWithEachZonesDate(t *testing.T) {
	t.Parallel()
	r := newRig(t, withEntries(
		settings.Entry{ID: "ny", Zone: "America/New_York", Label: "New York"},
		settings.Entry{ID: "syd", Zone: "Australia/Sydney", Label: "Sydney"},
	))
	cells := r.service.Snapshot().Cells
	if len(cells) != 2 || cells[0].ID != "ny" || cells[1].ID != "syd" {
		t.Fatalf("cells %+v", cells)
	}
	if cells[0].Date != "Sunday, 27 September" || cells[0].Time != "16:37" || cells[0].ZoneMark != "EDT" {
		t.Errorf("New York %+v", cells[0])
	}
	if cells[1].Date != "Monday, 28 September" || cells[1].Time != "06:37" || cells[1].ZoneMark != "AEST" {
		t.Errorf("Sydney %+v", cells[1])
	}
}

// FR-705, FR-706: the acceptance example, plus an entry that could not be read.
func TestOneBadClockLeavesTheOthersWorking(t *testing.T) {
	t.Parallel()
	r := newRig(t, withEntries(
		settings.Entry{ID: "a", Zone: "Europe/London", Label: "London"},
		settings.Entry{ID: "b", Zone: "Not/AZone"},
		settings.Entry{ID: "c", Zone: "Asia/Kolkata", Label: "Kolkata"},
		settings.Entry{ID: "d", Unreadable: "zone is not text", Original: `{"zone":7}`},
	))
	cells := r.service.Snapshot().Cells
	if cells[0].Time != "21:37" || cells[2].Time != "02:07" {
		t.Errorf("working clocks: %+v %+v", cells[0], cells[2])
	}
	if cells[1].Problem != "Unknown time zone: Not/AZone" || cells[1].Label != "Not/AZone" {
		t.Errorf("unknown zone: %+v", cells[1])
	}
	if cells[3].Problem != "This clock could not be read: zone is not text" {
		t.Errorf("unreadable: %+v", cells[3])
	}
}

// FR-706: an invalid clock is never given another zone's time.
func TestInvalidClockIsNeverGivenAnotherZone(t *testing.T) {
	t.Parallel()
	r := newRig(t, withEntries(settings.Entry{ID: "b", Zone: "Not/AZone", Label: "Grandma"}))
	cell := r.service.Snapshot().Cells[0]
	if cell.Time != "" || cell.Date != "" || cell.ZoneMark != "" || cell.Label != "Grandma" {
		t.Errorf("got %+v", cell)
	}
}

// FR-208, FR-107.
func TestSnapshotCarriesTheNextRefreshTheChoicesAndTheLayout(t *testing.T) {
	t.Parallel()
	r := newRig(t, settings.Defaults())
	snapshot := r.service.Snapshot()
	if want := snapshot.Now.Truncate(time.Minute).Add(time.Minute); !snapshot.NextRefresh.Equal(want) {
		t.Errorf("next refresh %s, want %s", snapshot.NextRefresh, want)
	}
	if snapshot.Layout != testLayout || snapshot.Style != settings.Digital || len(snapshot.Cells) != 0 {
		t.Errorf("snapshot %+v", snapshot)
	}
}
