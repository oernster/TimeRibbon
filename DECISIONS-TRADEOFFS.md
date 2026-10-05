# Decisions and trade-offs

The deliberate choices TimeRibbon rests on, as the product makes them today: what was chosen, what
was given up and why. The detail and the tests behind each live in [ARCHITECTURE.md](ARCHITECTURE.md)
and [REQUIREMENTS.md](REQUIREMENTS.md); [TECH_DEBT.md](TECH_DEBT.md) holds what is still open.

## The product as a whole

### Places, never people; a clock, not a planner

Each clock is a time zone with a label. TimeRibbon shows the time and date there and plans nothing.

- **Rather than:** contacts; calendars, meetings, reminders, alarms or time conversion.
- **Gains:** nothing personal is kept.
- **Costs:** those jobs need other tools.

### Go and Wails with a web page

- **Rather than:** a Python and Qt desktop stack.
- **Gains:** one executable with no runtime to install; the same stack draws the setup program.
- **Costs:** Wails offers one window and falls short at placing it, so the desktop is reached
  directly on every platform.

### The ribbon's desktop behaviour is a module of its own

Everything about being a ribbon on a desktop (placing, dragging, snapping, the tab, the grip, opacity,
the tray, scaling, one copy at a time, sign-in, the update check, setup) lives in `ribbonkit`, which
names no product; TimeRibbon holds the clocks, the sun map and the composition. WeatherRibbon is the
second product built on it.

- **Rather than:** copying TimeRibbon's desktop code into each new ribbon.
- **Gains:** a desktop fix lands once for every ribbon; two ribbons running together keep off each
  other through a folder they share (FR-412).
- **Costs:** a boundary to keep clean (structural tests hold the kit free of TimeRibbon), names chosen
  for any product rather than this one and an npm link for the page's half. It is carved inside this
  repository first, under TimeRibbon's own gate, then lifted into its own.

### Requirements before code

Every requirement was agreed before the first line of code, each naming its test; a later change
arrives as a numbered amendment.

- **Rather than:** building first and describing afterwards.
- **Gains:** a ruled-out idea stays ruled out; a requirement counts as met only once its test has
  failed without the code.
- **Costs:** the specification is work of its own and grows with every amendment.

### One format for every clock, no seconds

- **Rather than:** a format per clock; a seconds display.
- **Gains:** every cell reads alike; the ribbon changes once a minute.
- **Costs:** a 12-hour clock beside a 24-hour one is not possible.

## Privacy and the network

### One network request, held by a test

The update check is the only thing that reaches the network. Structural tests fail for any other Go
code that could open a connection, start a program or load a network library; they fail too for any
request from the page.

- **Rather than:** a promise that the network is used sparingly.
- **Gains:** "nothing else touches the network" is a test result.
- **Costs:** any new outward feature has to change the test that forbids it.

### Update checks shortly after start, then daily

The check asks GitHub for the latest published release without signing in, gives up quickly and
reads little. It speaks only of a newer release not skipped; a check from Help always answers. A
version it cannot read is never newer.

- **Rather than:** no check; one that reports every outcome.
- **Gains:** a new release is found without nagging; a draft, prerelease or malformed tag never
  prompts.
- **Costs:** one unprompted request to GitHub a day.

### It never installs an update itself

Download hands this platform's file (else the release's page) to the browser.

- **Rather than:** downloading and running the new version from inside the application.
- **Gains:** TimeRibbon never fetches or writes an executable.
- **Costs:** every update is a manual install.

### The browser opens through the desktop, not through Wails

- **Rather than:** Wails' own call, which reports nothing, so a machine with no browser left the
  donate button silent.
- **Gains:** a refusal is shown with the address.
- **Costs:** one opener per platform to keep.

### Built-in pictures and time zone rules

The sun map's pictures are NASA's Blue Marble and Black Marble, reduced and carried inside. The tz
database is built in too; Windows has no rules the application can read, while macOS and Linux read
their own first.

- **Rather than:** fetched imagery or weather; the machine's rules alone; offsets by hand.
- **Gains:** the map works offline; daylight saving follows the rules everywhere.
- **Costs:** fixed pictures add to the executable; on Windows a government's change of clocks shows
  only from the next release.

## Clocks and time

### East from Greenwich, never by hand

London first, then places ahead of UTC by how far ahead, then those behind it. The order is worked out
at every refresh; ties keep the order they were added.

- **Rather than:** ordering by hand, built first then withdrawn; earliest local time first, which put
  New York ahead of London.
- **Gains:** the ribbon reads round the world one way and never needs tidying.
- **Costs:** clocks cannot be arranged by hand.

### A zone's letters where it has them, else its offset

- **Rather than:** an offset for every zone; the database's bare numbers.
- **Gains:** a familiar name where one exists, a readable offset where none does.
- **Costs:** none recorded.

### The minute changes on the minute; a clock change is heard

Each reading carries the time to the next minute and the next is taken then; the page keeps no
periodic timer. Windows' own notices of a time change or a resume bring a fresh reading at once;
macOS and Linux send none, so the wall clock is compared every few seconds with a monotonic clock.

- **Rather than:** a fixed ticking timer; waiting for the next minute.
- **Gains:** no drift; every cell correct within seconds of a sleep or a change of time.
- **Costs:** off Windows, a small repeated check.

### Search matches the start of a word; places are zones

The places offered are the tz database's zones with their countries and cities. A label is the user's
own text, so a town without a zone takes its zone's clock.

- **Rather than:** matching anywhere, which began `l` with Adelaide; a gazetteer of every city.
- **Gains:** the place meant is near the top; every place offered is one the rules know.
- **Costs:** a fragment inside a word finds nothing; a label naming another town is not looked up,
  so its map mark stands at the zone's city.

## The settings file

### One readable file, written whole, a contract from the first release

Every choice and clock lives in one indented JSON file, written beside the old one and swapped in
once complete. Unknown keys are written back as found. No key the first release wrote is renamed,
dropped or given another meaning; a test reads a frozen first-release file.

- **Rather than:** a database; reshaping the file as the product grows.
- **Gains:** a person can read and repair it; a crash mid-write leaves the old file whole; an upgrade
  never loses anybody's clocks.
- **Costs:** a key named badly once is named so for good.

### A damaged file is kept aside; a failed save keeps the change

A file that is not JSON is renamed aside with a notice; should that fail, saving stops. An unreadable
clock is kept as it was and shown as invalid. A change whose save fails stays in effect with a notice.

- **Rather than:** starting afresh over the old file; undoing the change; failing silently.
- **Gains:** the only copy of somebody's clocks is never lost; failure is never hidden.
- **Costs:** a change made while saves fail is lost if the application ends first.

## The ribbon on the desktop

### One window for the ribbon and every panel

- **Rather than:** a second window per panel, which Wails does not offer.
- **Gains:** one window to place, hide and show.
- **Costs:** a panel replaces the ribbon while open; a move or change of length waits for it to close.

### Displays and placing through the system's own calls

What macOS and Linux share is written once for both.

- **Rather than:** Wails' screen list and position calls, which know nothing of work areas; a copy of
  the shared code per platform.
- **Gains:** the ribbon goes exactly where it is put; the two platforms cannot drift apart.
- **Costs:** three platforms' worth of desktop code; cgo on macOS and Linux.

### No taskbar or Dock button; native menus

The window's taskbar or Dock button is removed before it is first shown. Both menus are the system's
own, built from one shared list.

- **Rather than:** accepting the button; a menu drawn in the page.
- **Gains:** the ribbon stays out of the way; its small window never clips a menu.
- **Costs:** it rests on Wails' internals, so it is checked by hand on any Wails upgrade.

### A home edge per orientation; vertical on the right at first

Choosing an orientation sends the ribbon to its home edge (right for vertical, top for horizontal),
as does a display that has gone.

- **Rather than:** horizontal by default, as the first build had.
- **Gains:** a ribbon that has lost its place comes back somewhere predictable.
- **Costs:** none recorded.

### Re-centred only when its length changes

A clock, a notice or a change of style or size centres the ribbon along its length, its position
across kept; against the right or bottom edge it stays flush. A change of scale does not re-centre it.

- **Rather than:** growing from its corner on every change, which would pull it off the far edges.
- **Gains:** a ribbon centred on an edge stays centred as clocks come and go.
- **Costs:** a ribbon dragged off-centre is re-centred at the next change of length.

### Dragging is the platform's own

A press that moves past the system's drag distance is handed to the platform's move loop through the
message Wails' drag regions send. Windows holds the ribbon on a display throughout; macOS and Linux
give no say, so it is put back once it stands still.

- **Rather than:** a move loop of TimeRibbon's own, which would fight the desktop.
- **Gains:** the drag behaves as every other window's; a press on a control never starts one.
- **Costs:** the message is internal to Wails; off Windows the ribbon can hang off a display until let
  go.

### It fits its clocks, then scrolls; the page measures

The ribbon is as long as its clocks until the display's edge, then scrolls, thickened by the scroll
bar. The page measures the font's widest time and date, the scroll bar and the scale it is drawn at;
Go sizes the window from that.

- **Rather than:** shrinking or wrapping clocks; fixed widths, which cut dates short on a friend's
  machine; sizing from the display's scaling, which missed Windows' text size and KDE's fractional
  scale.
- **Gains:** every clock stays reachable and whole at any font, scaling or text size.
- **Costs:** some clocks are out of sight on a long ribbon; the page measures again on every change
  of size, style or format.

### Shown only once it is sized

- **Rather than:** showing it at once and growing it in view, which on a fractionally scaled Linux
  desktop often left it cut off.
- **Gains:** the ribbon first appears whole.
- **Costs:** it appears a moment later, up to a second where the page never reports.

### Closing hides; one copy runs; a launch toggles

- **Rather than:** a second launch only showing it; several copies.
- **Gains:** a single launcher button both shows and hides the ribbon.
- **Costs:** a copy left running makes a newer build's first launch only toggle the old ribbon.

### Linux on X11, with WebKit's faster paths off

GTK runs through X11 (XWayland on Wayland). WebKit's DMABUF renderer is off unless the user chose a
value; hardware acceleration is off as Wails would choose.

- **Rather than:** Wayland, where a window may not choose where it stands; leaving the renderer on
  and telling NVIDIA users to turn it off.
- **Gains:** the ribbon stands where placed; NVIDIA's own driver draws the page.
- **Costs:** under XWayland the tab relies on GTK's own pointer events; machines that could use the
  faster path lose it, which clocks redrawn once a minute do not need.

### A Linux tray icon of its own

- **Rather than:** a tray library that could not rebuild its menu as it opens and kept global state.
- **Gains:** the menu always shows the current ticks, with no sandbox permission of its own.
- **Costs:** the tray code is TimeRibbon's own to maintain.

## The unpinned ribbon

### A thin tab, only against an edge, always on top

Unpinned, the ribbon shrinks to a thin accent tab shortly after the pointer leaves, only while flush
against an edge along its orientation; elsewhere it shows in full with the choice kept. A drop near
an edge snaps flush. It stays above other windows and opens on a resting pointer without taking the
keyboard, keeping the tab's frame and the page's colour as it grows.

- **Rather than:** a tab wherever it stands (reversed); a tab big enough to press; one other windows
  can cover; giving the ordinary frame back as it opens, which flashed a caption, a stretched band and
  white.
- **Gains:** the tab costs almost no room, is never lost and opens cleanly.
- **Costs:** a drop near an edge snaps whether meant or not; touch reports no resting pointer, so
  touch users keep it pinned.

### The pointer is read as each platform was measured to need

Windows and macOS read the pointer at short intervals (Windows against the window's cut shape); Linux
listens for GTK's crossing events.

- **Rather than:** the page's own events, which were measured blind on an inactive macOS application
  and under XWayland.
- **Gains:** the tab opens and closes reliably everywhere.
- **Costs:** three ways of reading the pointer.

## The sun map

### A pull out with a lane of its own

The map slides out from a handle that stands in its own lane along the ribbon, one remembered choice
for both orientations.

- **Rather than:** a handle on the cells, which covered a clock's name; a map always out.
- **Gains:** the handle covers no clock; every map opens the same way.
- **Costs:** the ribbon is a little deeper while the map is on.

### The window is cut to the ribbon and its map on Windows

- **Rather than:** a rectangle covering the desktop beside the map.
- **Gains:** nothing is hidden that the ribbon does not draw.
- **Costs:** on macOS and Linux the window stays a rectangle, its spare area painted as the ribbon
  and answering the ribbon's right-click and drag.

### Labels move aside

- **Rather than:** every label right of its dot, printing over its neighbours.
- **Gains:** nearby cities stay readable.
- **Costs:** where no spot is clear a label can still overlap.

## The interface

### Every choice applies at once, offered in both places

Settings offers every menu choice from the same items; a test fails for any it lacks. Size, formats,
theme and opacity are in Settings alone; commands stay on the menus.

- **Rather than:** two lists written separately.
- **Gains:** the menus and Settings cannot disagree.
- **Costs:** a new choice has to fit both.

### Settings is wide and as tall as its content

- **Rather than:** a fixed height that scrolled on a large display; a tall narrow panel.
- **Gains:** on a display with room nothing scrolls.
- **Costs:** the window changes height as clocks come and go.

### See-through background, solid clocks

The opacity slider fades the ribbon's background from faint to opaque, never invisible; the clocks,
the map and every panel stay solid. It is saved once let go.

- **Rather than:** fading out entirely; fading the whole window, clocks and Settings with it.
- **Gains:** the ribbon sits over other work yet can always be found; the times stay readable.
- **Costs:** the window is drawn translucent, so its own paint must be cleared below full opacity
  and the page must report its true surface colour.

### Resized by a grip, within bounds, from its corner

A corner grip draws everything in the clocks smaller or larger together, on top of Large or Small,
from 75 to 200 percent; a double-click restores them. As a window does, it grows from its top-left
corner, with the sun map held still until let go. The pointer is read from the desktop.

- **Rather than:** free resizing; bounds set by the user; re-centring on every step, which slid the
  clocks along; the page's pointer events, which jumped backwards while the window resized.
- **Gains:** no shape the layout was not made for; the clocks hold still and follow the pointer
  smoothly.
- **Costs:** a ribbon centred on an edge is off-centre once resized, until Position or the next
  change of clocks; one pointer reading per platform to keep.

### One home for every colour, contrast held by test

Every colour of every scheme is stated once for light and dark together; tests require each offered
scheme to be complete and its text to meet the contrast floor in both themes.

- **Rather than:** colours written where used; a hue carried only by the accent, which the ribbon
  never paints.
- **Gains:** ten distinct schemes; an unreadable colour fails the suite.
- **Costs:** a new colour is stated for every scheme.

### Long pages read themselves, from one script; the licence as written

About, Licence and setup's Licence scroll gently on their own and stop when the reader takes over,
through one script shared by both pages. Licence keeps the file's own line breaks, sized to fit.

- **Rather than:** static pages; the cycle written twice; wrapped licence text.
- **Gains:** hands-free reading; the licence reads exactly as written.
- **Costs:** the script is plain JavaScript typed through a wrapper; the licence type is small.

## Building and installing

### Per user, with a setup program of its own on Windows

Everything is written under the user's own folders and registry; the Flatpak installs for the user.
One bespoke program installs, updates, goes back, repairs and removes, its logic kept apart from its
window. Every payload entry is checked to land inside the install folder before any is written. The
settings, web view data included, are removed only when asked.

- **Rather than:** a machine-wide install; a generic installer.
- **Gains:** nothing asks for administrator rights; a hostile payload writes nothing; forgetting the
  settings leaves nothing behind.
- **Costs:** each account installs separately; the setup window's side acts on the machine and has
  no tests.

### macOS: Apple Silicon only, signed and notarised

- **Rather than:** a universal build; shipping unsigned.
- **Gains:** Gatekeeper lets it open; the stated minimum macOS is one the executable meets.
- **Costs:** Intel Macs are not served; signing needs a Terminal at the Mac itself.

### A Flatpak with a narrow sandbox

- **Rather than:** a native package; wider access.
- **Gains:** the application holds only what it uses.
- **Costs:** Linux users need Flatpak; the network is granted to the whole application for the update
  check.

## Engineering

### Layers with one place where they meet

Domain, application, infrastructure and interface, each depending only inward; only the composition
root and the facade see both sides. Structural tests hold every boundary.

- **Rather than:** convention; a dependency injection framework.
- **Gains:** clock, placement and settings rules are tested with no disk, network, clock or screen.
- **Costs:** more packages and explicit wiring.

### Complete coverage where it means something

The domain and application are held to 100 percent; every other gated package to the coverage it
measured.

- **Rather than:** one figure over everything; aspirational floors.
- **Gains:** a shortfall in the pure layers is a decision nobody made; every other one is named.
- **Costs:** the desktop code sits far lower and relies on checks by hand; macOS and Linux carry no
  figure.

### Small files

- **Rather than:** letting files grow.
- **Gains:** files split at real seams.
- **Costs:** many small files.

### The wire written twice and compared; refusals that cannot be dropped

The shapes crossing between Go and the page are stated in both languages and compared by test. Every
page call Go can refuse takes a refusal handler and never throws.

- **Rather than:** Wails' generated bindings; promises that reject.
- **Gains:** a contract both sides are checked against; no refusal is lost.
- **Costs:** every wire change is made twice; every call site names its handler.

### Tests with real parts; a gate that cannot be skipped

No mocking library; hand-written doubles against real interfaces. No test reaches the network or the
user's own settings. Every guard is proved by planting a violation. The Windows build runs the whole
gate first with no switch to skip it.

- **Rather than:** mocks; an optional test step.
- **Gains:** a passing test means the real behaviour holds; nothing ships that failed a check.
- **Costs:** fakes are kept by hand; the macOS and Linux checks are a practice nothing enforces.

### Every name and the version have one home

- **Rather than:** copies where needed.
- **Gains:** a rename is made once; tests forbid the former name anywhere.
- **Costs:** static files such as the site are stamped from the source.

### Desktop events never block the desktop

The desktop hands events on without waiting, dropping one with a log line when nobody reads; it
recovers any panic.

- **Rather than:** calling back into the application on the desktop's thread.
- **Gains:** one fault cannot leave a ribbon that reacts to nothing.
- **Costs:** under a flood an event can be lost.
