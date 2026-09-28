package desktop

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"

	"github.com/oernster/timeribbon/internal/application"
)

// testTrayItems is a menu shaped as the tray's is: plain items, a check item, a submenu, then Exit.
var testTrayItems = []application.MenuItem{
	{Action: application.ActionHide, Label: "Hide"},
	{Action: application.ActionAlwaysOnTop, Label: "Always_on top", Checkable: true, Checked: true},
	{Label: "Help", Children: []application.MenuItem{{Action: application.ActionAbout, Label: "About"}}},
	{Action: application.ActionExit, Label: "Exit"},
}

// FR-502, FR-508: ids depth first from one, a separator above Exit, a submenu holding its items,
// a check item marked, underscores kept literal; only chosen items carry an action.
func TestTheTrayMenuIsLaidOutForTheHost(t *testing.T) {
	t.Parallel()
	root, actions := layoutOf(testTrayItems)
	var labels []string
	for _, child := range root.Children {
		node := child.Value().(menuNode)
		if label, ok := node.Properties[menuLabel]; ok {
			labels = append(labels, label.Value().(string))
		} else {
			labels = append(labels, node.Properties[menuType].Value().(string))
		}
	}
	if want := []string{"Hide", "Always__on top", "Help", menuSeparator, "Exit"}; !slices.Equal(labels, want) {
		t.Errorf("labels %q, want %q", labels, want)
	}
	check := root.Children[1].Value().(menuNode)
	if check.Properties[menuToggleType].Value() != menuCheckmark || check.Properties[menuToggleState].Value() != menuToggleOn {
		t.Errorf("check item %+v", check.Properties)
	}
	want := map[int32]application.MenuAction{1: application.ActionHide, 2: application.ActionAlwaysOnTop, 4: application.ActionAbout, 6: application.ActionExit}
	if len(actions) != len(want) {
		t.Errorf("actions %v, want %v", actions, want)
	}
	for id, action := range want {
		if actions[id] != action {
			t.Errorf("id %d: %q, want %q", id, actions[id], action)
		}
	}
	if about, ok := find(root, 4); !ok || about.Properties[menuLabel].Value() != "About" {
		t.Errorf("id 4 found %v: %+v", ok, about)
	}
}

// The icon is averaged down and sent alpha first, in network byte order.
func TestTheTrayIconIsAveragedDownToARGB(t *testing.T) {
	t.Parallel()
	source := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	source.Set(0, 0, color.NRGBA{R: 200, A: 255})
	source.Set(1, 0, color.NRGBA{R: 100, A: 255})
	source.Set(0, 1, color.NRGBA{B: 40, A: 255})
	source.Set(1, 1, color.NRGBA{B: 60, A: 255})
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, source); err != nil {
		t.Fatal(err)
	}
	got, err := pixmapOf(encoded.Bytes(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte{255, 75, 0, 25}; got.Width != 1 || got.Height != 1 || !bytes.Equal(got.Data, want) {
		t.Errorf("got %dx%d %v, want 1x1 %v", got.Width, got.Height, got.Data, want)
	}
	if _, err := pixmapOf([]byte("not a png"), 1); err == nil {
		t.Error("an image that is not a PNG was read")
	}
}

// FR-501 to FR-503 against the desktop's real tray host: the icon registers; the host's calls read
// the menu, choose from it and see it change when what it says changes.
func TestTheTrayIsHostedAndAnswersTheHost(t *testing.T) {
	var hidden atomic.Bool
	d := New(func() []application.MenuItem {
		if hidden.Load() {
			return testTrayItems[1:]
		}
		return testTrayItems
	}, io.Discard)
	d.UseIcon(nil)
	if err := d.Start(); err != nil {
		t.Fatalf("the tray did not start: %v", err)
	}
	defer d.Stop()
	client, err := dbus.ConnectSessionBus()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	// Ubuntu's host lists an icon registered by bus name under that name (read in its
	// statusNotifierWatcher.js, 2026-09-28). It lists it once the icon is built, after registering
	// returns, so the list is read until it appears.
	var items []string
	for deadline := time.Now().Add(settleLimit); time.Now().Before(deadline); time.Sleep(settlePause) {
		registered, err := client.Object(watcherName, watcherPath).GetProperty(watcherName + ".RegisteredStatusNotifierItems")
		if err != nil {
			t.Fatal(err)
		}
		if items, _ = registered.Value().([]string); slices.Contains(items, d.tray.name) {
			break
		}
	}
	t.Logf("the host lists %v", items)
	if !slices.Contains(items, d.tray.name) {
		t.Errorf("the host does not list %s", d.tray.name)
	}
	menu := client.Object(d.tray.name, menuPath)
	var revision uint32
	var layout menuNode
	if err := menu.Call(menuInterface+".GetLayout", 0, int32(0), int32(-1), []string{}).Store(&revision, &layout); err != nil {
		t.Fatal(err)
	}
	if len(layout.Children) != len(testTrayItems)+1 {
		t.Errorf("the host read %d entries", len(layout.Children))
	}
	if err := menu.Call(menuInterface+".Event", 0, int32(1), menuClicked, dbus.MakeVariant(""), uint32(0)).Err; err != nil {
		t.Fatal(err)
	}
	if event := <-d.Events(); event.Kind != EventMenu || event.Action != application.ActionHide {
		t.Errorf("clicking id 1 gave %+v", event)
	}
	hidden.Store(true)
	var changed bool
	if err := menu.Call(menuInterface+".AboutToShow", 0, int32(0)).Store(&changed); err != nil || !changed {
		t.Errorf("a changed menu was not reported changed (%v)", err)
	}
	if err := client.Object(d.tray.name, itemPath).Call(itemInterface+".Activate", 0, int32(0), int32(0)).Err; err != nil {
		t.Fatal(err)
	}
	if event := <-d.Events(); event.Kind != EventIconClicked {
		t.Errorf("activating the icon gave %+v", event)
	}
}
