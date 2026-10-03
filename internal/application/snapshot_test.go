package application

import (
	"slices"
	"testing"
	"time"

	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/settings"
)

// FR-102, FR-202: cells in order, each with its own zone's day.
func TestSnapshotFollowsClockOrderWithEachZonesDate(t *testing.T) {
	t.Parallel()
	r := newRig(t, withEntries(
		settings.Entry{ID: "ny", Zone: "America/New_York", Label: "New York"},
		settings.Entry{ID: "syd", Zone: "Australia/Sydney", Label: "Sydney"},
	))
	cells := r.service.Snapshot().Cells
	if len(cells) != 2 || cells[0].ID != "syd" || cells[1].ID != "ny" {
		t.Fatalf("cells %+v", cells)
	}
	if cells[1].Date != "Sunday, 27 September" || cells[1].Time != "16:37" || cells[1].ZoneMark != "EDT" {
		t.Errorf("New York %+v", cells[1])
	}
	if cells[0].Date != "Monday, 28 September" || cells[0].Time != "06:37" || cells[0].ZoneMark != "AEST" {
		t.Errorf("Sydney %+v", cells[0])
	}
}

// FR-612: the chosen date format reaches every cell and the snapshot.
func TestSnapshotWritesDatesInTheChosenFormat(t *testing.T) {
	t.Parallel()
	initial := withEntries(
		settings.Entry{ID: "ny", Zone: "America/New_York", Label: "New York"},
		settings.Entry{ID: "syd", Zone: "Australia/Sydney", Label: "Sydney"},
	)
	initial.DateFormat = clock.MonthDayYear
	snapshot := newRig(t, initial).service.Snapshot()
	if snapshot.DateFormat != clock.MonthDayYear || snapshot.Cells[0].Date != "Mon 09/28/2026" || snapshot.Cells[1].Date != "Sun 09/27/2026" {
		t.Errorf("format %s, dates %q and %q", snapshot.DateFormat, snapshot.Cells[0].Date, snapshot.Cells[1].Date)
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
	// The working clocks come first by time; the two that cannot be shown follow as stored.
	cells := r.service.Snapshot().Cells
	if cells[0].Time != "21:37" || cells[1].Time != "02:07" {
		t.Errorf("working clocks: %+v %+v", cells[0], cells[1])
	}
	if cells[2].Problem != "Unknown time zone: Not/AZone" || cells[2].Label != "Not/AZone" {
		t.Errorf("unknown zone: %+v", cells[2])
	}
	if cells[3].Problem != "This clock could not be read: zone is not text" {
		t.Errorf("unreadable: %+v", cells[3])
	}
}

// FR-102: the ribbon runs east from Greenwich, the reference, whatever order the clocks were added
// in: London, then places further ahead, then those behind Greenwich. Two zones keeping the same
// time stay in the order they were added.
func TestTheRibbonRunsEastFromGreenwich(t *testing.T) {
	t.Parallel()
	r := newRig(t, withEntries(
		settings.Entry{ID: "kol", Zone: "Asia/Kolkata"},
		settings.Entry{ID: "syd", Zone: "Australia/Sydney"},
		settings.Entry{ID: "lon", Zone: "Europe/London"},
		settings.Entry{ID: "ny", Zone: "America/New_York"},
		settings.Entry{ID: "lis", Zone: "Europe/Lisbon"},
	))
	var ids []string
	for _, cell := range r.service.Snapshot().Cells {
		ids = append(ids, cell.ID)
	}
	if want := []string{"lon", "lis", "kol", "syd", "ny"}; !slices.Equal(ids, want) {
		t.Errorf("order %v, want %v", ids, want)
	}
}

// FR-102 Amendment 6: every place level with or ahead of UTC comes before every place behind it,
// even one a whole day behind, added first (FR-102).
func TestEveryPlaceAheadOfUTCComesBeforeEveryPlaceBehindIt(t *testing.T) {
	t.Parallel()
	r := newRig(t, withEntries(
		settings.Entry{ID: "hon", Zone: "Pacific/Honolulu"},
		settings.Entry{ID: "kir", Zone: "Pacific/Kiritimati"},
		settings.Entry{ID: "pago", Zone: "Pacific/Pago_Pago"},
		settings.Entry{ID: "tonga", Zone: "Pacific/Tongatapu"},
		settings.Entry{ID: "lon", Zone: "Europe/London"},
	))
	var ids []string
	for _, cell := range r.service.Snapshot().Cells {
		ids = append(ids, cell.ID)
	}
	if want := []string{"lon", "tonga", "kir", "pago", "hon"}; !slices.Equal(ids, want) {
		t.Errorf("order %v, want %v", ids, want)
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
