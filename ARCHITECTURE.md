# TimeStrip Architecture

A small Windows desktop application showing a strip of clocks, one per chosen place. It reads the
Windows clock and nothing else from outside itself; the time zone rules are built into the
executable. It makes no network request: no Go file of this module imports a network package, which
`TestTheModuleImportsNoNetworkPackage` holds. The one address it knows, the donation page, is handed
to the desktop's browser rather than fetched.

The requirements are in [REQUIREMENTS.md](REQUIREMENTS.md); the FR, NFR and CON numbers below are
its.

## Invariant

`UI -> Application -> Domain <- Infrastructure`

Dependencies point inward. The Domain is the stable core and depends on nothing. Every rule below is
enforced by a test under `tests/structural`, not by convention. The table is the whole set: a rule
claimed in prose and enforced nowhere reads exactly like one that holds, so a guard not listed here
does not exist.

| Invariant | Enforcing test | File |
|---|---|---|
| Domain imports nothing from this module outside `internal/domain` | `TestDomainHasNoOutwardImports` | `boundary_test.go` |
| Domain is pure: no network, filesystem, process, random or tz database package; no wall clock read and no zone loaded (FR-207, CON-5) | `TestDomainIsPure` | `boundary_test.go` |
| Application never imports infrastructure or Wails | `TestApplicationDoesNotImportInfrastructure` | `boundary_test.go` |
| Infrastructure never imports Wails, the setup program's own package apart | `TestWailsStaysOutOfInfrastructure` | `boundary_test.go` |
| Only `main.go`, `app.go` and `window_life.go` import both the application and infrastructure | `TestCompositionRootIsWhitelisted` | `boundary_test.go` |
| No source file exceeds 400 lines: the Go, the front end's TypeScript and CSS, the setup page | `TestNoFileExceedsLineLimit` | `boundary_test.go` |
| No source file sits in the danger band of 381 to 400 lines | `TestNoFileInDangerBand` | `boundary_test.go` |
| Every exported type carries a doc comment | `TestEveryExportedTypeIsDocumented` | `boundary_test.go` |
| A file's lines are counted as an editor numbers them | `TestLineCountCountsTheLinesAnEditorShows` | `linecount_test.go` |
| No Go file of this module imports a network package (NFR-S-1) | `TestTheModuleImportsNoNetworkPackage` | `network_test.go` |
| The network package check recognises `net`, `crypto/tls` and `golang.org/x/net` paths and passes look-alikes | `TestNetworkPackageRecognitionIsExact` | `network_test.go` |
| The setup page loads every script beside it | `TestTheSetupPageLoadsEveryScript` | `setup_test.go` |
| No setup page file spells the product's name, which the setup program hands it | `TestTheSetupPageNeverWritesTheProductsName` | `setup_test.go` |
| About credits exactly the modules the application and the setup program link (FR-607) | `TestEveryLinkedModuleIsCredited` | `credits_test.go` |
| The wire is stated identically in `dto.go` and `frontend/src/wire.ts` | `TestTheWireIsStatedAlikeOnBothSides` | `wire_test.go` |
| The page listens for every event `app.go` emits and keys every panel it names | `TestThePageNamesEveryEventGoEmits` | `wire_test.go` |
| The setup page listens for every event `installer/app.go` emits | `TestTheSetupPageNamesEveryEventSetupEmits` | `wire_test.go` |
| Each `wails.json` names its executable as `internal/product` does | `TestEachWailsConfigNamesItsExecutableAsTheProductDoes` | `names_test.go` |
| The Licence panel is sized for the LICENSE's widest line, so it shows unwrapped | `TestTheLicencePanelIsSizedForTheLicencesWidestLine` | `licence_test.go` |

## Layers

- **Domain** (`internal/domain`: `clock`, `placement`, `settings`): pure Go. Time arrives as an
  argument and a zone arrives already resolved, so the domain holds no tz database and reads no
  clock. `clock` turns an instant and a zone into what a cell shows: the local time in either
  format, the weekday and date, the zone mark (the abbreviation where the tz database gives one of
  letters, else `UTC` and the signed offset) and the hand angles; `NextRefresh` names the next minute
  boundary. It also derives a zone's default label. `placement` decides where the strip goes, in
  physical pixels: the default place, a stored placement restored on its monitor at that monitor's
  DPI, the least move that brings a strip wholly inside a work area (`Clamp`, `Recover`) and `Fit`,
  the strip's length along its orientation. `settings` is the user's choices as one value; every
  operation answers a new value and leaves the old one as it was.
- **Application** (`internal/application`): one `Service` holding every use case over six ports
  (`Store`, `Zones`, `Clock`, `IDs`, `Monitors`, `StartupEntry`, in `ports.go`). It builds the
  snapshot the strip draws (its cells ordered east from Greenwich, places behind UTC last, worked out afresh each
  time since daylight saving moves it), adds, edits and removes clocks, searches places, changes settings,
  arranges the strip (`Launch`, `Rearrange`, `Moved`, `Centred`) and answers the tray and context
  menus. A change that cannot be saved stays in effect and raises a notice until a later save
  succeeds (FR-707). It never imports infrastructure or Wails.
- **Infrastructure** (`internal/infrastructure`): the adapters behind the ports and the Windows
  integration. `store` (the settings file), `zones` (resolution through the embedded tz database and
  the place catalogue), `monitors` (displays through Win32), `startup` (the Start with Windows
  value), `system` (the wall clock and new clock ids), `appdata` (the settings folder), `runlog` (the
  run's log), `desktop` (the tray, native menus, the move fence, the desktop's broadcasts and
  handing an address to the default browser through the shell, which reports a refusal) and
  `setup` (the install policy behind the setup program).
- **UI**: the React front end plus the Wails facade in package `main`, which calls the service and
  maps what it answers into the shapes in `dto.go`.
- **Outside the layers**: `internal/product` holds the product's name, the setup program's name, the
  window class, the donation address, the version `build.ps1` stamps, the author, the copyright line
  and the credits. Every layer reads it, so it belongs to none.
- **Tools**, never shipped: `tools/genplaces` writes the place catalogue from the tz database's
  `zone.tab` and `iso3166.tab`; `tools/payload` packs the built application for the setup program;
  `tools/versioninfo` writes each executable's Windows version resource from `VERSION` and
  `internal/product` before its build, in place of Wails' template, which carried its fallback
  version and a placeholder copyright;
  `tools/genicons.py` writes every icon from the masters in `assets/`.

## Composition root

`main.go` is the composition root. It opens the run log and points standard error at it before
anything can fail, builds the adapters, injects them into the service by constructor, starts the tray
and hands the facade to Wails. The cell sizes (`layout`) and the panel size (`panelSize`) have their
one home there. No service is held in a package-level variable and there is no service locator.

The facade is `app.go` (the calls the page makes) and `window_life.go` (startup, showing, hiding,
closing and the desktop's events), split only to keep each file small; the structural whitelist names
all three files. The facade holds the service through `stripService`, an interface in `app.go`. It
holds each call into Wails and the desktop as a field, pointed by `newApp` at the real calls in
`wails_calls.go` and `window_life.go`. That is what lets the facade's tests stand in for all three
and read what it decided. `identity.go` answers About and Licence; `dto.go` holds the wire; `launch.go` holds
the window's options; `bindings_on.go` and `bindings_off.go` tell the run `wails build` makes to
generate bindings, which carries the `bindings` build tag, not to write the log, read the settings or
show a tray icon.

## Dependency direction

```
             +------------------------------+
   Wails/UI  | app.go, window_life.go       |
             +--------------+---------------+
                            | calls
             +--------------v---------------+
             |  application: Service, ports |
             +------+----------------^------+
          depends on|                | implements
             +------v-----+          |
             |   domain   |          |
             +------------+          |
                     +---------------+----------------+
                     |        infrastructure          |
                     | store, zones, monitors,        |
                     | startup, system, appdata,      |
                     | runlog, desktop, setup         |
                     +--------------------------------+
```

## One window

Wails v2 offers one window, so the strip, Settings, About and Licence share it (CON-6). The strip is
the window at the size its clocks need. Opening a panel resizes the window to `panelSize`, centred on
the strip's display and never larger than its work area (`Service.Centred`); closing one returns the
window to where the strip was last left. While a panel is open, a move of the window is not recorded
as the strip's and a change of content is fitted when the panel closes.

The window opens hidden. `startup` finds its handle by the class `TimeStripStrip`, takes it off the
taskbar, fences its moves and places it, all before the page is shown, so it never appears blank or
in the wrong place. Wails always marks its window as an application window, which forces a taskbar
button; `HideFromTaskbar` takes that style off and marks it a tool window once, before it is shown.

## The strip's size and place

**Size (FR-105, FR-106).** `stripSize` counts the cells the page draws: each notice, then each clock
(the Add clock prompt standing in for them when there are none). Along the orientation the strip is that many cells plus
padding, while that fits the work area of its display; beyond that it is the work area's length and
its cells scroll. Across, it is one cell plus padding, plus the thickness of the scroll bar when the
cells scroll, so the bar never covers them. The bar is the web engine's, not one Windows reports, so
the page measures it once it has loaded and hands it to Go through `SetScrollbar`. The strip hides
overflow on both axes and scrolls only along its own; hiding one axis alone let the browser turn the
other into a second scroll bar, measured in Edge on 2026-09-27. A plain wheel moves a scrolling
horizontal strip along.

Every change that can alter the cells (a clock added or removed, the style or orientation changed, a
notice raised by a failed save or dismissed, the scroll bar reported) refits the strip where it
stands. Where the refit changes the strip's length, it is centred along that length on its display
with its position across kept; the place is stored (FR-104) and the service remembers the length it last
arranged to tell. Nothing else re-centres it, so a drag holds until the length next changes. Sizes
are computed in DIP and scaled to the display's DPI, so a strip moved between displays
at different scaling keeps its size in DIP (FR-407).

**Place (FR-403 to FR-406).** Coordinates are physical pixels on the virtual desktop. Wails'
`WindowSetPosition` places a window relative to the work area of the monitor it is on while
`WindowGetPosition` answers absolute coordinates. Its screen list carries no origin, device name or
work area either. So displays are read through `EnumDisplayMonitors` and `GetMonitorInfoW` (`monitors`)
and the window is placed with `SetWindowPos` (`desktop.Place`). With nothing stored the strip goes
16 DIP inside the primary work area's right edge, centred vertically. The end of a drag is heard
through a WinEvent hook on `EVENT_SYSTEM_MOVESIZEEND`; the placement is stored as the monitor's device
name, its work area, its DPI and the strip's offset from the work area's corner. At launch it is
restored on that monitor, the offset scaled by any change of DPI; where that monitor is gone it goes
to the default place on the primary. A display change refits the strip where it is.

**The drag (FR-401, FR-402).** A press on empty strip area that moves past Windows' own drag
distance (`SM_CXDRAG`, `SM_CYDRAG`) hands the press to Windows' move loop through
`window.WailsInvoke('drag')`, the message Wails' own drag regions send. That message is internal to
Wails v2 rather than a documented call; a press on a control never starts one. While the window
moves, a window procedure placed in front of Wails' own (`desktop.KeepOnDisplays`) answers each
`WM_MOVING` by moving the proposed rectangle the least distance that keeps it inside the work area of
the display under the pointer, so the strip can be carried onto another display but never left half
off one.

## Time

Zones resolve through `time.LoadLocation` with `time/tzdata` built in, so no rule depends on what the
machine holds (CON-5); an empty zone id is refused rather than read as UTC. The place catalogue,
`internal/infrastructure/zones/places.tsv`, is written by `tools/genplaces` from tz 2025b's
`zone.tab` and `iso3166.tab` and holds 418 zones; a test holds every one of them to resolving in the
tz database Go embeds.

Each snapshot carries the milliseconds to the next minute boundary; the page takes the next snapshot
then, so each refresh is scheduled from the current time rather than from the last one
(FR-208). The hidden tray window hears `WM_TIMECHANGE` and the resume broadcasts of
`WM_POWERBROADCAST`, on which the page takes a fresh snapshot at once (FR-209).

## The settings file

`%APPDATA%\TimeStrip\settings.json`, indented JSON a person can read, carrying a format version
(FR-701). Derived values (offsets, abbreviations, times) are never stored. It is written to a
temporary file in the same folder, flushed and renamed over the old one, so a failure part way leaves
the previous file whole (FR-702). Reading is tolerant: no file means the defaults and no notice
(FR-703); a file that is not JSON is renamed to `settings.unreadable.json` and a notice says so
(FR-704). Should that rename fail, saving is refused from then on so the only copy is never
overwritten. One clock that cannot be read or names an unknown zone is kept in the file as it was and
shown in words as an invalid clock while the others work (FR-705, FR-706). A top-level key this
version does not know is written back as it was found.

**The file is a contract from 1.0.0 (NFR-C-1).** Every later 1.x reads every file 1.0.0 writes to the
same settings. No key 1.0.0 writes may be renamed, dropped or given another meaning. No stored word
(such as `12h` or `analogue`) may change. A later release may add keys. The guard is
`TestA1Point0SettingsFileIsReadWhole`, which reads the frozen fixture
`internal/infrastructure/store/testdata/settings-1.0.0.json` (every key set away from its default)
and requires every key to be read rather than merely carried. It was proved by renaming a key and by
changing a stored word: each failed it. The fixture is never regenerated from a later writer, since
what it proves is that the old shape still reads.

## The desktop

`desktop` owns a hidden top-level window on its own locked thread: the notification-area icon, the
native menus and the desktop's broadcasts. A message-only window would not hear the broadcasts. The
icon is read out of the executable itself. When Explorer restarts it re-adds the icon on the
`TaskbarCreated` message. Nothing crosses the thread boundary by callback: the desktop reports on a
buffered channel, dropping an event with a line in the log rather than blocking the thread Windows
called in on; the facade's `listen` loop acts on it. Both the window procedure and the listen loop
recover a panic and log it, so one fault cannot leave a strip that reacts to nothing.

Both menus are native popup menus, so the strip's small window never clips them. A menu item may hold
children, which become a submenu (the Help submenu of FR-508); identifiers are numbered depth first,
so a choice inside a submenu still names its action. A tray icon that cannot be created is not fatal:
the strip still runs. Closing it then quits, since nothing would bring it back.

## Help, About and Licence

About and Licence are panels of the one window (FR-607, FR-608). About shows the application icon,
the name and version, the author, the copyright line and a credit for every component the executables
ship, each naming its licence and what it does here; `TestEveryLinkedModuleIsCredited` asks the Go
tool which modules the two executables link as `wails build` builds them and holds the credits to
that list in both directions. Licence shows the `LICENSE` file embedded in the binary.

Both bodies read themselves when they hold more than fits (FR-609): still for 5 seconds on opening,
down one pixel every other 40 ms tick, still for 5 seconds at the end, back to the top at 15 pixels a
tick, still for 2 seconds, repeat. A wheel, a press, a touch, a key or focus arriving in the body
suspends the cycle for 2.5 seconds, after which it carries on from where the reader left it; focus
arriving during the opening 5 seconds does not shorten them; a body beneath a dialog marked modal
stands frozen and takes no input. The cycle is ported from PigeonPost and has one home,
`installer/frontend/dist/auto-scroll.js`: the setup page has no build step and can load a script but
import nothing, while the window's build can import that file, so both surfaces run the same script
(FR-811). It publishes one name, `window.AutoScroll`; `frontend/src/autoScroll.ts` states its shape
for the type checker and wraps it in a React hook.

## The setup program

Delivery is a second Wails application. `installer/` is its own `main` package in the same module,
embedding the built application as a zip and the setup page as assets, so one file is the whole
distribution. `build.ps1` packs the built application and `LICENSE` into `installer/payload.zip`
through `tools/payload`, builds the setup program with the version from `VERSION`, then writes the
empty placeholder zip back whether or not that build succeeded, so the real payload never reaches a
commit. The payload is embedded as a string rather than a byte slice, which Go keeps in the read-only
image rather than charging to the process.

`internal/infrastructure/setup` holds the install policy: the places, the payload extraction with its
fence against an entry that leaves the install folder (checked for every entry before any is written,
FR-803), the version comparison, the Apps list record, the shortcuts (through the Windows shell's COM
object, via go-ole), the Start with Windows value (through `startup`, the one Settings writes, so the
two cannot disagree), the process handling and the step log. `installer/app.go` is a facade over it.
Each step is weighted by the time it was measured to take, so the progress bar moves as the work does.

Which screen opens is decided by one reading of the machine:

| Reading | Screen |
|---|---|
| started with `-uninstall` | Uninstall |
| nothing installed | Install |
| the version carried is newer | Update |
| the version carried is older | Go back |
| the versions match | Installed: Repair, Reinstall, Uninstall |

Versions compare by major, minor then patch as numbers. Install, Update, Go back and Reinstall are one
act: the files, setup copied beside them as `uninstall.exe`, the Apps list record, then the boxes.
Repair writes the files again keeping the shortcuts and Start with Windows as they are on the machine.
Every place written is per user: the files under `%LOCALAPPDATA%\Programs\TimeStrip`, the Start Menu
shortcut under `%APPDATA%`, the Desktop shortcut on the user's own Desktop, the record and Start with
Windows under `HKCU`. Windows never asks for administrator rights (FR-810). Uninstall removes the
shortcuts, Start with Windows and the record. It deletes `%APPDATA%\TimeStrip` (the settings, the log
and the web view's data) only when **Also forget my settings** is ticked, then hands the install folder to a hidden PowerShell that deletes it once setup
has exited.

Setup refuses to write while the application runs and offers to close it, waiting up to 5 seconds
for it to go (FR-807). **The setup page names nothing**: it has no build step, so nothing compiles or
type checks it. The product's name arrives on the state it is handed;
`TestTheSetupPageNeverWritesTheProductsName` holds that. Its keyboard ring (`setup-ring.js`) is the
window's model written again for a page that cannot import it; `frontend/src/setupRing.test.ts`
loads the shipped script and presses keys against it. The window is painted the page's own ground
before the page loads, read from the page's stylesheet, so it never flashes the wrong colour. Setup's
own web view data and step log sit under the temporary folder.

## Data locations

| What | Where |
|---|---|
| Settings | `%APPDATA%\TimeStrip\settings.json`; `settings.unreadable.json` beside it when a damaged file was kept aside |
| Run log | `%APPDATA%\TimeStrip\TimeStrip.log`, started afresh once it passes 1 MB |
| The window's web view data | `%APPDATA%\TimeStrip\WebView2`, named in `launch.go` inside the settings folder so uninstalling with **Also forget my settings** removes it; nothing of TimeStrip's own is kept there |
| Time zone rules and the place catalogue | built into the executable |
| Installed files | `%LOCALAPPDATA%\Programs\TimeStrip`, with `uninstall.exe` |
| Shortcuts | the user's Start Menu Programs folder and Desktop |
| Start with Windows | the value `TimeStrip` under `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, holding the quoted path and no arguments |
| Apps list record | `HKCU\...\Uninstall\TimeStrip` |
| Setup's step log and web view data | `TimeStripSetup.log` and `TimeStripSetup` in the temporary folder |

## Errors

Errors are wrapped with context at each boundary using `%w`; `errors.Is` sentinels mark the ones a
caller acts on.

- **Before the window, a run is ended only by a failure to run the window at all.** Standard error is
  pointed at the log as the first act of the run (`runlog.Keep`), so even the Go runtime's own panic
  report is kept. A settings folder that cannot be found falls back to a folder in the temporary
  folder; settings that cannot be read, a tray icon that cannot be made and a missing executable path
  are logged and the strip still opens. The embedded place catalogue failing to parse is a build
  defect, which a test holds.
- **Shown on the strip, which keeps working:** a settings file kept aside and a save that failed, as
  notices with OK; an invalid clock, in words in its own cell.
- **Refused beneath the control that was pressed:** every page call that Go can refuse. Each `api`
  wrapper takes a refusal handler as its last argument and answers null rather than rejecting, so a
  call written without a handler does not compile and no refusal is dropped on the page.
- **Logged and carried on:** a desktop event nobody was reading, a panic in the desktop's thread or in
  handling one of its events, a failure to place, hide from the taskbar or fence the strip.

## Quality enforcement

- The structural tests in the table above run as part of the suite.
- The wire is written twice by necessity: Go structs with json tags in `dto.go` and TypeScript
  interfaces in `frontend/src/wire.ts`. Wails' generated bindings are gitignored and imported by
  nothing, so they are not the contract; a structural test compares the two statements.
- `test.ps1` checks formatting, vets, runs staticcheck at the version it pins, runs the whole Go suite
  and the front end's lint, type check and tests, holds `internal/domain` and `internal/application`
  to 100% coverage, then holds each other gated package at a floor measured from what it reaches.
  [TESTING.md](TESTING.md) tabulates every figure.
- `build.ps1` runs `test.ps1` before it builds and offers no switch to skip it. It pins cgo off for
  the gate and the build alike.

## Design decisions

| Decision | Why | Rejected alternative |
|---|---|---|
| Go with Wails and a web front end | One executable with no runtime to install; the same stack draws the setup program | A Python and Qt desktop stack |
| The tz database built into the executable | Every machine shows the same rules, whatever it has installed; no DST rule is written by hand | Reading the machine's zone files; offsets written by hand |
| Displays and placement through Win32 | Wails' screen list has no origin, device name or work area; its position calls mix relative and absolute coordinates | Wails' own position calls |
| One window for the strip and every panel | Wails v2 offers one window | A second window per panel |
| Native popup menus | The strip's window is small; a menu drawn in the page would be clipped by it | A menu drawn in the page |
| The scroll bar measured by the page | It is the web engine's bar, which Windows' scroll bar metric does not describe | A thickness written into the Go code |
| The self-reading cycle in one script beside the setup page | The setup page cannot import; the window's build can import the file, so both run the same code | The cycle written twice, once in TypeScript and once for the setup page |
| The strip hides from the taskbar by swapping its window style | Wails always marks its window as an application window | Accepting a taskbar button |
| The web view's data folder named inside the settings folder | Left to Wails it was `%APPDATA%\TimeStrip.exe`, beside the settings folder, which forgetting the settings on uninstall did not reach | Deleting Wails' default folder by name at uninstall, which hangs on a rule Wails does not promise |
| Settings in one JSON file, written whole | A person can read and repair it; a crash mid-write cannot damage it | A database |
| Everything per user | Nothing needs administrator rights, so nothing asks for them | A machine-wide install |

See also [TESTING.md](TESTING.md) for the test suite and [DEVELOPMENT.md](DEVELOPMENT.md) for
building from source.
