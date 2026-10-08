package application

import (
	"slices"
	"testing"
	"time"

	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/timeribbon/internal/domain/clock"
	"github.com/oernster/timeribbon/internal/domain/settings"
	"github.com/oernster/timeribbon/internal/domain/sun"
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

// FR-612: the chosen date format reaches every cell and the snapshot.
func TestSnapshotWritesDatesInTheChosenFormat(t *testing.T) {
	t.Parallel()
	initial := withEntries(
		settings.Entry{ID: "ny", Zone: "America/New_York", Label: "New York"},
		settings.Entry{ID: "syd", Zone: "Australia/Sydney", Label: "Sydney"},
	)
	initial.DateFormat = clock.MonthDayYear
	snapshot := newRig(t, initial).service.Snapshot()
	if snapshot.DateFormat != clock.MonthDayYear || snapshot.Cells[0].Date != "Sun 09/27/2026" || snapshot.Cells[1].Date != "Mon 09/28/2026" {
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

// orderPlaces are the cities the ordering tests stand on, where the real catalogue puts them.
var orderPlaces = []Place{
	{Zone: "America/Toronto", At: sun.Point{Latitude: 43.65, Longitude: -79.3833}},
	{Zone: "Europe/London", At: londonAt},
	{Zone: "Europe/Amsterdam", At: sun.Point{Latitude: 52.3667, Longitude: 4.9}},
	{Zone: "Asia/Shanghai", At: sun.Point{Latitude: 31.2333, Longitude: 121.4667}},
	{Zone: "Australia/Sydney", At: sun.Point{Latitude: -33.8667, Longitude: 151.2167}},
	{Zone: "Pacific/Tongatapu", At: sun.Point{Latitude: -21.1333, Longitude: -175.2}},
	{Zone: "Pacific/Pago_Pago", At: sun.Point{Latitude: -14.2667, Longitude: -170.7}},
	{Zone: "Pacific/Honolulu", At: sun.Point{Latitude: 21.3069, Longitude: -157.8583}},
	{Zone: "Pacific/Kiritimati", At: sun.Point{Latitude: 1.8667, Longitude: -157.3333}},
}

// orderOf answers the ids of the cells a rig over orderPlaces shows for entries, in the ribbon's order.
func orderOf(t *testing.T, entries ...settings.Entry) []string {
	t.Helper()
	var ids []string
	for _, cell := range newRigOver(t, withEntries(entries...), orderPlaces).service.Snapshot().Cells {
		ids = append(ids, cell.ID)
	}
	return ids
}

// FR-102: the ribbon runs west to east as the sun map draws the places, whatever order the clocks
// were added in, so the cells and the dots read the same way. Two clocks in one city stay in the
// order they were added; a zone with no city (UTC) stands at the meridian its offset keeps.
func TestTheRibbonRunsWestToEastLikeTheMap(t *testing.T) {
	t.Parallel()
	got := orderOf(t,
		settings.Entry{ID: "lon", Zone: "Europe/London"},
		settings.Entry{ID: "ams", Zone: "Europe/Amsterdam"},
		settings.Entry{ID: "sha", Zone: "Asia/Shanghai"},
		settings.Entry{ID: "syd", Zone: "Australia/Sydney"},
		settings.Entry{ID: "tor", Zone: "America/Toronto"},
		settings.Entry{ID: "utc", Zone: "UTC"},
		settings.Entry{ID: "bri", Zone: "Europe/London"},
	)
	if want := []string{"tor", "lon", "bri", "utc", "ams", "sha", "syd"}; !slices.Equal(got, want) {
		t.Errorf("order %v, want %v", got, want)
	}
}

// FR-102: across the date line the map decides, not the clock: Tonga, a day ahead, is furthest west;
// a zone of UTC+14 with no city stands at -150, east of Kiritimati and west of everything else.
func TestTheDateLineFollowsTheMap(t *testing.T) {
	t.Parallel()
	got := orderOf(t,
		settings.Entry{ID: "lon", Zone: "Europe/London"},
		settings.Entry{ID: "plus14", Zone: "Etc/GMT-14"},
		settings.Entry{ID: "hon", Zone: "Pacific/Honolulu"},
		settings.Entry{ID: "kir", Zone: "Pacific/Kiritimati"},
		settings.Entry{ID: "pago", Zone: "Pacific/Pago_Pago"},
		settings.Entry{ID: "tonga", Zone: "Pacific/Tongatapu"},
	)
	if want := []string{"tonga", "pago", "hon", "kir", "plus14", "lon"}; !slices.Equal(got, want) {
		t.Errorf("order %v, want %v", got, want)
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

// FR-623: the snapshot shows the scale the ribbon is drawn at, the preview while the grip is dragged
// and the kept scale once it is chosen, with the bounds the grip may reach; the kept one is saved.
func TestTheSnapshotShowsTheScaleTheRibbonIsDrawnAt(t *testing.T) {
	t.Parallel()
	r := newRig(t, clocks(2))
	if err := r.service.PreviewScale(ribbon.MaxScale); err != nil {
		t.Fatal(err)
	}
	if got := r.service.Snapshot(); got.Scale != ribbon.MaxScale || got.MinScale != ribbon.MinScale || got.MaxScale != ribbon.MaxScale {
		t.Errorf("previewing shows %v within %d to %d", got.Scale, got.MinScale, got.MaxScale)
	}
	if err := r.service.SetScale(ribbon.MinScale); err != nil || r.service.Snapshot().Scale != ribbon.MinScale || r.store.last(t).Scale != ribbon.MinScale {
		t.Errorf("keeping a scale shows %v (%v)", r.service.Snapshot().Scale, err)
	}
}
