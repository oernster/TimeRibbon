# Decisions and trade-offs

The deliberate choices TimeRibbon rests on: what was chosen, what was given up
for it and why. Each entry is the decision as the product makes it today.
The detail behind each one, with the tests that hold it, lives in
[ARCHITECTURE.md](ARCHITECTURE.md) and the specification
([REQUIREMENTS.md](REQUIREMENTS.md)); [TESTING.md](TESTING.md) says where the
tests stop and why; [TECH_DEBT.md](TECH_DEBT.md) holds what is still open and
what only looks like debt.

## The product as a whole

### Places, never people; a clock, not a planner

Each clock is a place: a time zone with a label. TimeRibbon shows the time,
weekday and date there and plans nothing.

- **Rather than:** contacts or friends' names; calendars, meetings, reminders,
  alarms or time conversion.
- **Gains:** nothing personal is kept; the settings hold zones and labels
  alone.
- **Costs:** those jobs need other tools.

### Go and Wails with a web page

The application is Go on Wails, drawing a React and TypeScript page in the
platform's own web view.

- **Rather than:** a Python and Qt desktop stack.
- **Gains:** one executable with no runtime to install; the same stack draws
  the setup program.
- **Costs:** Wails offers one window and falls short at placing it, so the
  desktop is reached directly on every platform.

### Requirements before code

Every requirement was written down and agreed before the first line of code,
each naming the test that verifies it. A later change arrives as a numbered
amendment rather than an edit that hides what came before.

- **Rather than:** building first and describing afterwards.
- **Gains:** a ruled-out idea stays ruled out; a requirement counts as met only
  once its test has been seen to fail without the code.
- **Costs:** the specification is work of its own to keep true; it grows with
  every amendment.

### One choice for every clock

The 12-hour or 24-hour format and the date format are single choices that
every clock follows. No clock shows seconds.

- **Rather than:** a format per clock; a seconds display.
- **Gains:** every cell reads the same way; the ribbon changes once a minute
  rather than every second.
- **Costs:** a 12-hour clock beside a 24-hour one is not possible.

## Privacy and the network

### One network request, held by a test

The update check is the only thing that reaches the network. A structural
test fails for any of the Go code outside the update check that could open a
connection; a second fails should that exemption point at nothing. The page
makes no request of its own.

- **Rather than:** a promise that the network is used sparingly.
- **Gains:** "nothing else touches the network" is a test result rather than
  a sentence.
- **Costs:** any new outward feature has to go through the update check or
  change the test that forbids it.

### Update checks: shortly after start, then daily

A check runs shortly after launch and then once a day. It asks GitHub for the
latest published release without signing in, gives up quickly and never reads
more than a release's answer could need. It says nothing unless that release
is newer than the running copy and not one the user chose to skip. A check
asked for from Help always answers, even when GitHub cannot be reached. A
version it cannot read is never newer.

- **Rather than:** no check at all; one that reports every outcome.
- **Gains:** a new release is found without nagging; a draft, a prerelease or
  a malformed tag never prompts.
- **Costs:** one unprompted request to GitHub a day.

### It never installs an update itself

Download hands this platform's file to the browser; where the release has no
file for it, the release's page goes instead. Installing it is the user's to
do.

- **Rather than:** downloading and running the new version from inside the
  application.
- **Gains:** TimeRibbon never fetches or writes an executable; the one network
  request stays a small one.
- **Costs:** every update is a manual install.

### The browser opens through the desktop, not through Wails

The donation page and an offered download are handed to the system's own
opener on each platform, which reports a refusal. Settings or Help shows the
refusal with the address.

- **Rather than:** Wails' own call, which reports nothing, so a machine with
  no browser left the donate button doing nothing with nothing said.
- **Gains:** a refusal is seen; TimeRibbon itself fetches neither page.
- **Costs:** one opener per platform to keep.

### The map's pictures are built in

The sun map's day and night pictures are NASA's Blue Marble and Black Marble,
reduced and carried inside the executable, credited in About.

- **Rather than:** fetching imagery, live cloud or weather.
- **Gains:** the map works with no network; drawing it sends nothing.
- **Costs:** the pictures are fixed, with no clouds, weather or moon; they add
  to the size of the executable.

### Time zone rules built in, the system's own first

The tz database is built into the executable. Windows keeps no zone rules the
application can read, so there the built-in rules are the ones used. macOS and
Linux read the system's own rules first, falling back on the built-in ones
only for a zone they lack.

- **Rather than:** relying on the machine's rules alone; offsets written by
  hand.
- **Gains:** daylight saving follows the rules themselves on every platform;
  no rule is written by hand.
- **Costs:** on Windows a government's change of clocks after a release is
  shown correctly only from the next release.

## Clocks and time

### The order is east from Greenwich, never by hand

London comes first; then places ahead of UTC by how far ahead; then the places
behind it, which going east reaches last. The order is worked out at every
refresh, so daylight saving can move a clock. Clocks keeping the same time
stay in the order they were added.

- **Rather than:** ordering by hand, which was built first and then withdrawn
  with its move controls and drag; earliest local time first, which put New
  York ahead of London.
- **Gains:** the ribbon reads round the world in one direction; it never needs
  tidying.
- **Costs:** clocks cannot be arranged by hand.

### A zone's letters where it has them, else its offset

Each clock is marked with its zone's abbreviation where the tz database gives
one in letters, such as EDT; otherwise with UTC and the signed offset, such as
UTC+5:45.

- **Rather than:** an offset for every zone; the bare numbers the database
  gives where a zone has no letters.
- **Gains:** a familiar name where one exists; a readable offset where none
  does.
- **Costs:** none recorded.

### The minute changes on the minute

Every reading of the clocks carries the time left to the next minute; the
next reading is taken then. The page keeps no periodic timer for its clocks,
which a test holds.

- **Rather than:** a timer ticking at a fixed interval.
- **Gains:** the minute changes when it should, without drift; between
  minutes the ribbon does nothing.
- **Costs:** none recorded.

### A change of clock is heard, not waited out

On Windows the system's own notices of a time change and a resume from sleep
bring a fresh reading at once. macOS and Linux send no such notice, so the
wall clock is compared every few seconds with a clock that does not move with
it; a gap between the two counts as a jump.

- **Rather than:** waiting for the next minute to come round.
- **Gains:** after a sleep or a change of time every cell is correct within
  seconds.
- **Costs:** off Windows, a small check repeated for as long as the
  application runs.

### Search matches the start of a word

What is typed must begin a word of a place's name, country or zone id. Names
beginning with it come first, then later words of a name, then countries and
zones.

- **Rather than:** matching anywhere inside a name, sorted alphabetically,
  which began a search for `l` with Adelaide and Algiers.
- **Gains:** the place meant is near the top.
- **Costs:** a fragment from inside a word finds nothing.

### Places are zones; the label is the user's

The places offered are the tz database's own zones with their countries and
cities, generated from its files; a test holds every one to resolving. A label
is short free text. A town with no zone of its own takes its zone's clock and
the label typed for it.

- **Rather than:** a gazetteer of every city.
- **Gains:** nothing to look up; every place offered is one the rules know.
- **Costs:** a label naming another town is not looked up, so its mark on the
  sun map stands at the zone's city.

## The settings file

### One readable file, written whole

Every choice and clock is kept in one indented JSON file. A save writes a new
file beside it and puts it in place only once it is complete. A key the
running version does not know is written back as it was found. Derived values
such as offsets and abbreviations are never stored.

- **Rather than:** a database.
- **Gains:** a person can read and repair it; a crash part way through a
  write leaves the previous file whole.
- **Costs:** none recorded.

### A damaged file is kept aside, never overwritten

A file that is not JSON is renamed aside and a notice on the ribbon says so.
Should that rename fail, saving is refused from then on. One clock that cannot
be read is kept in the file as it was and shown as an invalid clock while the
others work; it is never given another zone.

- **Rather than:** starting afresh over the old file; dropping what cannot be
  read.
- **Gains:** the only copy of somebody's clocks is never lost to a fault.
- **Costs:** the ribbon starts from the defaults until the file is repaired.

### A save that fails keeps the change

A change whose save fails stays in effect, with a notice on the ribbon, until
a later save succeeds.

- **Rather than:** undoing the change; failing silently.
- **Gains:** the ribbon keeps working on a disk that refuses writes; the
  failure is never hidden.
- **Costs:** a change made while saves fail is lost if the application ends
  before one succeeds.

### The file is a contract from the first release

Every later release reads every file the first release wrote to the same
settings. No key it wrote is renamed, dropped or given another meaning; no
stored word changes. Later keys are added after it. A test reads a frozen file
of the first release and requires every key in it to be read.

- **Rather than:** reshaping the file as the product grows.
- **Gains:** an upgrade never loses anybody's clocks.
- **Costs:** a key named badly once is named so for good. The rename to
  TimeRibbon kept the shape but moved the folder; nothing is read from the
  former one.

## The ribbon on the desktop

### One window for the ribbon and every panel

Settings, About, Licence and the update panel are shown by resizing the
ribbon's one window, centred on its display, then returning it to where the
ribbon was.

- **Rather than:** a second window per panel, which Wails does not offer.
- **Gains:** one window to place, hide and show.
- **Costs:** a panel replaces the ribbon while it is open; a move or a change
  of length during a panel is held until it closes.

### Displays and placing through the system's own calls

Displays are read and the window placed through each platform's own desktop
calls. What macOS and Linux share is written once for both, each toolkit
supplying only the rest.

- **Rather than:** Wails' own screen list and position calls, which know
  nothing of work areas or which display is which and mix two ways of
  counting position; a copy of the shared code per platform.
- **Gains:** the ribbon goes exactly where it is put, on any display; the two
  platforms that share code cannot drift apart.
- **Costs:** three platforms' worth of desktop code; macOS and Linux need cgo
  against their toolkit.

### No taskbar or Dock button

Wails always gives its window a taskbar button, so TimeRibbon takes that away
before the window is first shown: on Windows by changing the window's style,
on Linux by marking the window, on macOS by making the application an
accessory once launching has finished.

- **Rather than:** accepting a taskbar or Dock button.
- **Gains:** the ribbon stays out of the way; its icon lives in the tray or
  the menu bar.
- **Costs:** it rests on what Wails does inside, so it is checked by hand on
  any upgrade of Wails.

### Native popup menus

Both the tray menu and the ribbon's right-click menu are the system's own,
built from one list of items and words shared by every platform.

- **Rather than:** a menu drawn in the page.
- **Gains:** the ribbon's small window never clips a menu.
- **Costs:** each platform draws its menus its own way.

### Vertical, against the right edge, on a first run

A first run shows a vertical ribbon flush against the right edge of the
primary display, centred along it. Each orientation has a home edge, the right
for vertical and the top for horizontal; choosing an orientation sends the
ribbon there, as does a stored display that has gone.

- **Rather than:** horizontal by default, which the first build had.
- **Gains:** a ribbon that has lost its place always comes back somewhere
  predictable.
- **Costs:** none recorded.

### Re-centred only when its length changes

When a clock, a notice or a change of style or size changes the ribbon's
length, it is centred along that length on its display with its position
across kept. A change of scale does not re-centre it (see the grip, below). A
ribbon against the right or bottom edge stays flush there when it grows or
shrinks across. Otherwise only a drag, Position or a change of orientation
moves it.

- **Rather than:** growing from its corner, which would pull a ribbon off the
  right or bottom edge.
- **Gains:** a ribbon centred on an edge stays centred and flush as clocks come
  and go.
- **Costs:** a ribbon dragged off-centre is re-centred at the next change of
  length.

### Dragging is the platform's own

A press on empty ribbon that moves past the system's drag distance is handed
to the platform's own move loop, through the same message Wails' drag regions
send. On Windows each step of a drag is held inside the display under the
pointer. macOS and Linux give no say while a drag lasts, so a ribbon left
partly off every display is put back once it stands still.

- **Rather than:** writing a move loop of TimeRibbon's own, which would fight
  the desktop.
- **Gains:** the drag behaves as every other window's does; a press on a
  control never starts one; the ribbon can be carried onto any display but
  never left half off one.
- **Costs:** the message is internal to Wails rather than documented, so
  dragging is checked by hand on any upgrade of Wails; on macOS and Linux the
  ribbon can hang off a display until let go.

### It fits its clocks, then scrolls

The ribbon is as long as its clocks until it reaches the edge of the display;
then its clocks scroll. It is made thicker by the scroll bar, so the bar never
covers them. A plain wheel moves a horizontal ribbon along.

- **Rather than:** shrinking the clocks; wrapping them onto several rows.
- **Gains:** every clock stays its own size and reachable.
- **Costs:** on a long ribbon some clocks are out of sight.

### The page measures; the window follows

Only the page knows the font it really draws with and the scale it is really
drawn at, so it measures both and Go sizes the window from what it reports. A
cell is as wide as the widest time and date its formats can show; the scroll
bar is the thickness the page measures; the display's own scaling stands in
only until the page has reported.

- **Rather than:** fixed cell widths, which fitted only the fonts they were
  tried with, so on a friend's machine the dates were cut short; sizing from
  the display's scaling alone, which cut the page off wherever it missed part
  of the scale: the user's text size on Windows, a fractional scale handed over
  as font size alone on Linux.
- **Gains:** no time or date is ever cut short, whatever the font; the window
  fits the page at any scaling and text size.
- **Costs:** the page measures again whenever the size, style or a format
  changes.

### Shown only once it is sized

A launched ribbon stays hidden until the page is ready and has reported its
scale and its widest text; a page that never reports is shown shortly after
anyway.

- **Rather than:** showing it as soon as the page is ready, then growing it
  in view, which on a fractionally scaled Linux desktop often left it cut off
  until the next change.
- **Gains:** the ribbon first appears whole, at its final size.
- **Costs:** the ribbon appears a moment later; a page that never reports
  delays it by up to a second.

### Closing hides; one copy runs; a launch toggles

Alt+F4 or a close hides the ribbon; only Exit ends the application, except
where there is no tray icon to bring the ribbon back. One copy runs per user.
Launching TimeRibbon again shows or hides the ribbon as the tray icon's click
does, so one launcher button does both.

- **Rather than:** a second launch only showing the ribbon; several copies.
- **Gains:** a single button, such as a Stream Deck's, both shows and hides
  it.
- **Costs:** a copy left running holds the lock, so launching a newer build
  over it only shows or hides the old ribbon until the old copy exits.

### Linux draws through X11

GTK is sent through X11; on a Wayland desktop TimeRibbon runs through
XWayland.

- **Rather than:** Wayland, where a window may not choose where it stands,
  which the ribbon must.
- **Gains:** the ribbon stands exactly where it is placed.
- **Costs:** under XWayland neither the page nor the X server sees the pointer
  leave, so the tab relies on GTK's own events.

### WebKit's faster drawing paths off on every Linux machine

The web view's DMABUF renderer is turned off before it starts, unless the
user has already chosen a setting for it, which is then left alone. Its
hardware acceleration is off too, as Wails itself would choose.

- **Rather than:** leaving them on and telling NVIDIA users to turn the
  renderer off; turning it off in the Flatpak alone, which a build run outside
  it would miss.
- **Gains:** on NVIDIA's own driver, where the renderer drew only the window's
  background, the page draws.
- **Costs:** machines where they work lose a faster path, which clocks redrawn
  once a minute do not need.

### A Linux tray icon of its own

The Linux tray icon is TimeRibbon's own, speaking the desktop's tray protocol
directly, in a way that needs no sandbox permission of its own. Its menu is
rebuilt as the tray is about to show it.

- **Rather than:** an existing tray library, which offered no way to rebuild
  the menu as it opens and kept its state global.
- **Gains:** the menu always shows the current ticks.
- **Costs:** the tray code is TimeRibbon's own to maintain.

## The unpinned ribbon

### A tab only against an edge

Unpinned, the ribbon shrinks to a thin tab in the scheme's accent shortly
after the pointer leaves. It does so only while flush against an edge along
its orientation, inner edges between displays included. Anywhere else it
shows in full as though pinned, the choice kept. A drop near such an edge
snaps flush; the last edge is remembered. It is pinned unless chosen
otherwise.

- **Rather than:** a tab wherever the ribbon stands, which was the first
  ruling and was reversed.
- **Gains:** dragging it back onto an edge brings the tab back by itself.
- **Costs:** a drop near an edge snaps to it whether or not that was meant.

### The pointer is read as each platform was measured to need

Resting on the tab opens the ribbon without taking the keyboard from what has
it; moving away collapses it again. Windows and macOS read where the pointer
is at short intervals, Windows against the window's cut shape; Linux listens
for GTK's own events of the pointer arriving and leaving.

- **Rather than:** the page's own pointer events everywhere. Measured, the
  page on macOS is blind while the application is inactive; under XWayland
  neither the page nor a reading of the pointer sees it leave.
- **Gains:** the tab opens and closes reliably on all three, never stealing
  the keyboard.
- **Costs:** three ways of reading the pointer; a repeated reading on Windows
  and macOS.

### A thin tab that stays on top, opening cleanly

The tab is far thinner than the least size a pressed target should have,
since it is rested on rather than pressed. Unpinned, the ribbon stays above
other windows. As it opens it keeps the tab's frame, the page is told first
and the window grows once the page has drawn, in the page's own background
colour.

- **Rather than:** a tab big enough to press; one other windows can cover;
  handing the ribbon back its ordinary frame as it opens, which flashed a
  close button, a stretched band and white, each seen by screen capture.
- **Gains:** the tab costs almost no room and is never lost; the ribbon opens
  with no caption, band or white.
- **Costs:** a touch screen reports no resting pointer, so touch users keep
  the ribbon pinned; an opening waits briefly on the page.

## The sun map

### A pull out with a lane of its own

The map slides out beside the ribbon from a handle, below or above a
horizontal ribbon and beside a vertical one. While the map is on, the ribbon
is deeper by a lane where the handle stands. One remembered choice serves both
orientations.

- **Rather than:** a handle standing on the cells, which covered the middle
  clock's name on a friend's machine; a map always out for a horizontal
  ribbon.
- **Gains:** the handle covers no clock; every map opens and closes the same
  way.
- **Costs:** the ribbon is a little deeper while the map is on.

### The window is cut to the ribbon and its map on Windows

On Windows the window is cut to the ribbon and the map before every placing,
so the desktop shows and takes clicks round a map shorter or longer than the
ribbon. macOS and Linux stay rectangles.

- **Rather than:** a rectangle covering the desktop beside the map.
- **Gains:** nothing is hidden that the ribbon does not draw.
- **Costs:** on macOS and Linux the window covers that space.

### Labels move aside

Each label is measured once drawn, then placed in clock order at the first
spot round its dot that stands clear of every dot and earlier label; to the
right where none does.

- **Rather than:** every label to the right of its dot, printing over its
  neighbours.
- **Gains:** nearby cities stay readable.
- **Costs:** where no spot is clear a label can still overlap another.

## The interface

### Every choice applies at once, offered in both places

Each choice takes effect the moment it is made, with no Save step. Settings
offers every choice the menus do, built from the same items; a test fails for
any menu choice Settings lacks. Size, the formats, the theme and opacity are
in Settings alone; commands stay on the menus.

- **Rather than:** two lists written separately; style and orientation on the
  menus alone, as they once were.
- **Gains:** the menus and Settings cannot disagree.
- **Costs:** a new choice has to fit both.

### Settings is wide and as tall as its content

Settings opens wide with its choices in columns and grows to the height of
its content, capped by the display's work area, where it scrolls. The other
panels stay narrow, for their text.

- **Rather than:** a fixed height, which had to be scrolled even on a large
  display; a tall narrow panel.
- **Gains:** on a display with room nothing scrolls.
- **Costs:** the window changes height as clocks are added or removed.

### See-through, never invisible

An opacity slider draws the ribbon's background anywhere from faint to fully
opaque, never fully transparent; the clocks on it stay solid and Settings is
always opaque. It is saved once let go.

- **Rather than:** allowing it to fade out entirely; fading the whole window,
  clocks and Settings with it.
- **Gains:** the ribbon can sit over other work, while it can always be seen
  and found again; the times stay readable at any opacity.
- **Costs:** the web view is drawn transparent, so the window's own paint has
  to be cleared below full opacity.

### Resized by a grip, within bounds

A grip in the ribbon's corner draws everything in the clocks smaller or larger
together, on top of Large or Small; a double-click returns them to their own
size. The window cannot be resized freely. As a resized window does, it grows
and shrinks from its top-left corner, so the grip stays under the pointer.

- **Rather than:** free window resizing; bounds set by the user; centring the
  ribbon again on every step of the drag, which slid the clocks along it.
- **Gains:** no shape the layout was not made for; the least size stays
  readable and the largest still fits a small display; the clocks hold still
  under the pointer while they are resized.
- **Costs:** a ribbon centred on an edge is no longer centred once resized,
  until Position or the next change of clocks centres it again.

### One home for every colour, contrast held by test

Every colour of every scheme has one home, each stated once for light and dark
together. A scheme's hue lives in the colours the ribbon itself paints. Tests
require every offered scheme to state every colour the ribbon needs and its
text to meet the accessibility standard's contrast floor in both themes.

- **Rather than:** colours written where they are used; a scheme whose hue
  lives only in the accent, which the ribbon never paints, so Ocean measured
  barely apart from Classic.
- **Gains:** ten distinct schemes, each following the theme; a colour that
  cannot be read fails the suite rather than shipping.
- **Costs:** a new colour has to be stated for every scheme.

### Long pages read themselves, from one script

About, Licence and setup's Licence scroll gently on their own when they hold
more than fits; they stop the moment the reader takes over. The cycle is one
script, shared by the setup page and the window.

- **Rather than:** static pages; the cycle written twice, once for each.
- **Gains:** long text can be read hands free; both surfaces run the same
  code.
- **Costs:** the script is plain JavaScript, typed for the page through a
  wrapper.

### The licence shown as written

Licence keeps the licence file's own line breaks and never wraps it again; its
type is sized so the widest line fits, which a test holds.

- **Rather than:** letting the page wrap the text.
- **Gains:** the licence reads exactly as it was written.
- **Costs:** the type is small.

## Building and installing

### Installed for one user, without administrator rights

On Windows everything is written under the user's own folders and registry;
on Linux the Flatpak is installed for the user's account.

- **Rather than:** a machine-wide install.
- **Gains:** nothing asks for administrator rights.
- **Costs:** each account on a machine installs separately.

### A setup program of its own on Windows

Install, update, going back, repair and removal are one bespoke program,
ported in shape from Bridge Talk, with its install logic kept apart from its
window. Which screen opens is decided by one reading of the machine. Every
file in the payload is checked to land inside the install folder before any is
written. A running copy is asked to close first. The user's settings are
removed only when asked. Everything the application keeps lives inside its
settings folder, the web view's own data included, so forgetting the settings
leaves nothing behind. The built application never reaches the repository: the
build puts its empty placeholder back whether or not it succeeded.

- **Rather than:** a generic installer; the web view's data left where Wails
  puts it, beside the settings folder, which uninstalling did not reach.
- **Gains:** one identity throughout; a hostile payload writes nothing; a
  clone builds and tests without a built executable.
- **Costs:** the setup program is TimeRibbon's own to maintain; its window's
  side acts on the machine and has no tests; an interrupted build can leave
  the payload behind until the next build.

### macOS: Apple Silicon only, signed and notarised

The macOS build runs only on an Apple Silicon Mac, with the Go compiler
rather than the Wails command. The oldest macOS it claims is read from the Go
toolchain; a build that links code made for a newer macOS is refused. The app
and the DMG are signed and notarised.

- **Rather than:** a universal build; shipping unsigned.
- **Gains:** Gatekeeper lets it open; the stated minimum is one the
  executable really meets.
- **Costs:** Intel Macs are not served; an Apple developer account; signing
  needs a Terminal at the Mac itself, since the keychain refuses a remote
  shell.

### A Flatpak with a narrow sandbox

The Linux build is a Flatpak granted X11, the GPU, the tray, the
single-instance lock, the autostart folder and the network for the update
check. No other part of the file system.

- **Rather than:** a native package; a sandbox with wider access.
- **Gains:** the application holds only what it uses.
- **Costs:** Linux users need Flatpak; the network is granted to the whole
  application although only the update check uses it.

## Engineering

### Layers with one place where they meet

The code is split into domain, application, infrastructure and interface,
each depending only inward. Only the composition root and the facade over it
see both the application and infrastructure; the service is built there by
constructor, held in no global and found through no locator. Structural tests
hold every boundary.

- **Rather than:** convention alone; a dependency injection framework.
- **Gains:** clock, placement and settings rules are tested with no disk,
  network, clock or screen.
- **Costs:** more packages and more explicit wiring.

### Complete coverage where it means something

The domain and application layers are held to complete coverage. Every other
gated package is held to a floor at the coverage it measured, so a lost test
fails the gate.

- **Rather than:** one figure over the whole program; floors picked as
  aspirations.
- **Gains:** anything short of complete in the pure layers is a decision
  nobody made; every shortfall elsewhere is named in TESTING.md.
- **Costs:** the desktop code sits far lower and relies on checks by hand;
  the macOS and Linux halves carry no figure at all.

### Small files

Every source file is held under a fixed size, with a band below the limit
that sends a file well clear of it rather than a line under it. Build scripts
are exempt.

- **Rather than:** letting files grow.
- **Gains:** files split at real seams.
- **Costs:** many small files; the facade is split across several for size
  alone.

### The wire is written twice and compared

The shapes crossing between Go and the page are stated in Go and again in
TypeScript; a structural test fails when the two disagree, as another does
when the page misses an event Go sends.

- **Rather than:** Wails' generated bindings, which are left out of the
  repository and used by nothing.
- **Gains:** a contract both sides are checked against.
- **Costs:** every change to the wire is made twice.

### A refusal cannot be dropped on the page

Every call the page makes that Go can refuse takes a refusal handler and
answers nothing rather than throwing; a call written without one does not
compile.

- **Rather than:** promises that reject.
- **Gains:** no refusal is lost; each is shown beneath the control pressed.
- **Costs:** every call site names its handler.

### Tests with real parts

No Go test uses a mocking library; every double is written by hand against
the real interface. The facade holds each call into Wails and the desktop as
a replaceable part, so its decisions can be tested. No test reaches the
network or writes to the user's own settings, sign-in entry or Apps list.
Every guard is proved by planting a violation and watching it fail.

- **Rather than:** a mocking library; an untested facade.
- **Gains:** a passing test means the real behaviour holds; a guard is known
  to bite.
- **Costs:** fakes are written and kept by hand.

### The gate cannot be skipped

The Windows build runs the whole gate first with no switch to skip it, built
the same way as what ships. The macOS and Linux code is checked on a machine
of its own platform.

- **Rather than:** an optional test step.
- **Gains:** nothing ships that failed a check.
- **Costs:** the macOS and Linux checks are a practice before each release
  that nothing enforces.

### Every name and the version have one home

The product's name, its app id, the donation address, the author and the
credits live in one place; the version lives in one file and is passed into
every build and the site. Tests hold the build configuration to those names
and forbid the product's former name anywhere in the repository.

- **Rather than:** copies written where they are needed.
- **Gains:** a rename is made once and cannot drift.
- **Costs:** static files such as the site have to be stamped from the
  source.

### Desktop events never block the desktop

The desktop hands its events on without waiting; when nobody is reading, an
event is dropped with a line in the log rather than holding up the thread the
system called in on. A panic there or in handling an event is recovered and
logged.

- **Rather than:** calling back into the application from the desktop's
  thread.
- **Gains:** one fault cannot leave a ribbon that reacts to nothing.
- **Costs:** under a flood an event can be lost.
