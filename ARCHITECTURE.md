# TimeRibbon Architecture

A desktop application for Windows, macOS and Linux showing a ribbon of clocks, one per chosen place.
It reads the system clock, its settings file, the desktop (displays, tray, sign-in entry) and GitHub's
latest release. The time zone rules are built in; macOS and Linux read their own zone files first
([Time](#time)). The domain and application are the same code everywhere; each platform's own half
sits in files its build tags or names select.

Its one network request is the update check (FR-509): only `internal/infrastructure/update` imports a
network package; `requests_test.go` holds the other ways out (a program started, a Windows library
loaded by name, a request from either page) to the ones it names. The donation page and a release's
download go to the desktop's browser; a release's addresses are taken only as `https` on `github.com`.

FR, NFR and CON numbers are those of [REQUIREMENTS.md](REQUIREMENTS.md).

## Invariant

`UI -> Application -> Domain <- Infrastructure`

Dependencies point inward. Every rule below is a test under `tests/structural`; a guard not listed
here does not exist.

| Invariant | Enforcing test | File |
|---|---|---|
| Domain imports nothing from this module outside `internal/domain` | `TestDomainHasNoOutwardImports` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Domain is pure: no network, filesystem, process, random or tz package; no wall clock read, no zone loaded (FR-207, CON-5) | `TestDomainIsPure` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Application never imports infrastructure or Wails | `TestApplicationDoesNotImportInfrastructure` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Infrastructure never imports Wails | `TestWailsStaysOutOfInfrastructure` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Only `main.go`, `app.go` and `window_life.go` import both application and infrastructure | `TestCompositionRootIsWhitelisted` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| No source file exceeds 400 lines: Go, the front end's TypeScript and CSS, the setup page | `TestNoFileExceedsLineLimit` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| No source file sits in the danger band of 381 to 400 lines | `TestNoFileInDangerBand` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Every exported type carries a doc comment | `TestEveryExportedTypeIsDocumented` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Lines are counted as an editor numbers them | `TestLineCountCountsTheLinesAnEditorShows` | [`linecount_test.go`](tests/structural/linecount_test.go) |
| No Go file outside `internal/infrastructure/update` imports `net`, `crypto/tls` or `golang.org/x/net` (NFR-S-1) | `TestOnlyTheUpdateCheckImportsANetworkPackage` | [`network_test.go`](tests/structural/network_test.go) |
| Only the files in `processStarters` start a program or hand an address to the desktop; no Go file names a Windows library outside `systemLibraries` (NFR-S-1) | `TestOnlyNamedFilesStartAProcess` | [`requests_test.go`](tests/structural/requests_test.go) |
| Every file in `processStarters` exists | `TestEveryNamedProcessStarterExists` | [`requests_test.go`](tests/structural/requests_test.go) |
| Neither page uses a request API or names a web address, the SVG namespace aside (NFR-S-1) | `TestThePageMakesNoRequest` | [`requests_test.go`](tests/structural/requests_test.go) |
| Those checks recognise `os/exec`, `ShellExecute`, a library such as `WinHTTP.DLL`, `fetch`, `WebSocket` and an address, passing look-alikes such as `prefetch` | `TestRequestRecognitionIsExact` | [`requests_test.go`](tests/structural/requests_test.go) |
| The network exemption names a directory that exists | `TestTheNetworkExemptionNamesTheUpdatePackage` | [`network_test.go`](tests/structural/network_test.go) |
| The network check recognises `net`, `crypto/tls` and `golang.org/x/net` and passes look-alikes | `TestNetworkPackageRecognitionIsExact` | [`network_test.go`](tests/structural/network_test.go) |
| The setup page loads every script beside it | `TestTheSetupPageLoadsEveryScript` | [`setup_test.go`](tests/structural/setup_test.go) |
| No setup page file spells the product's name | `TestTheSetupPageNeverWritesTheProductsName` | [`setup_test.go`](tests/structural/setup_test.go) |
| Each platform's About credits exactly the modules its build links (FR-607) | `TestEveryLinkedModuleIsCredited` | [`credits_test.go`](tests/structural/credits_test.go) |
| No platform credits a module twice | `TestAModuleIsCreditedOncePerPlatform` | [`credits_test.go`](tests/structural/credits_test.go) |
| The wire is stated alike in `dto.go` and `frontend/src/wire.ts` | `TestTheWireIsStatedAlikeOnBothSides` | [`wire_test.go`](tests/structural/wire_test.go) |
| The page listens for every event `app.go` emits and keys every panel it names | `TestThePageNamesEveryEventGoEmits` | [`wire_test.go`](tests/structural/wire_test.go) |
| The setup page listens for every event `installer/app.go` emits | `TestTheSetupPageNamesEveryEventSetupEmits` | [`wire_test.go`](tests/structural/wire_test.go) |
| Each `wails.json` names its executable as `internal/product` does | `TestEachWailsConfigNamesItsExecutableAsTheProductDoes` | [`names_test.go`](tests/structural/names_test.go) |
| Every offered scheme has a block in `colours.css` stating each of Classic's tokens (the problem colour aside); every block is offered (FR-611) | `TestEveryOfferedSchemeHasItsOwnCompleteBlock` | [`colours_test.go`](tests/structural/colours_test.go) |
| Text, muted text and problem text meet 4.5:1 on the cell and the surface, every scheme, both themes (NFR-U-1) | `TestTextMeetsTheContrastFloorOnEverySchemeAndTheme` | [`contrast_test.go`](tests/structural/contrast_test.go) |
| Classic's dark colours are the same under the system's dark mode as under a chosen dark theme | `TestClassicDarkIsTheSameUnderTheSystemAsWhenChosen` | [`contrast_test.go`](tests/structural/contrast_test.go) |
| Contrast is computed as WCAG 2.x states it; an unreadable colour form is refused | `TestContrastIsComputedAsTheStandardStatesIt` | [`contrast_test.go`](tests/structural/contrast_test.go) |
| The Licence panel is sized for the LICENSE's widest line | `TestTheLicencePanelIsSizedForTheLicencesWidestLine` | [`licence_test.go`](tests/structural/licence_test.go) |
| No tracked or new file holds the product's former name or the word its window went by before the ribbon, the npm lock file and Python's string method of that name aside | `TestNoTrackedFileHoldsTheRetiredWord` | [`retired_test.go`](tests/structural/retired_test.go) |
| That word is found in any case and inside names while current names pass | `TestTheRetiredWordIsFoundInAnyCaseAndInsideNames` | [`retired_test.go`](tests/structural/retired_test.go) |

## Layers

- **Domain** (`internal/domain`), pure Go: time arrives as an argument and a zone already resolved.
  - `clock`: an instant and a zone become what a cell shows (time in either format, the date in the
    chosen `DateFormat`, the zone mark, the hand angles); `NextRefresh` names the next minute;
    `Samples` writes every time and date a cell can show, for the page to measure (FR-620).
  - `placement`, in physical pixels: the default place, a stored placement restored at its monitor's
    DPI, the least move into a work area (`Clamp`, `Recover`), the ribbon's length (`Fit`), centring
    along a work area (`CentredAlong`) or against an edge (`AgainstEdge`), the tab (`Tab`, FR-614),
    the edge a ribbon stands flush against (`FlushAgainst`) and the snap of a drop within `SnapReach`
    (`Snapped`, FR-410). `sunmap.go` puts the map on the side away from the ribbon's edge
    (`InnerSide`, `MapBeside`); `MapHeld` keeps it still while the grip is dragged (FR-623).
  - `settings`: the user's choices as one value, every operation answering a new one. It holds the
    pin in effect (`PinnedInEffect`, FR-619), the stay-on-top rule (`OnTop`, FR-617), the last edge
    (`LastEdge`, FR-411), the opacity bounds (20 to 100 percent, FR-622), the scale bounds (75 to
    200, FR-623) and `ScaleAfter`, the scale a drag of the grip has reached.
  - `sun`: the subsolar point from NOAA's equations (FR-906), checked against NOAA's values in
    `testdata`.
  - `hover`: told the pointer arrived or left and the time, it answers whether an unpinned ribbon is
    open and when to ask again (FR-615, FR-616).
- **Application** (`internal/application`): one `Service` over seven ports (`Store`, `Zones`, `Clock`,
  `IDs`, `Monitors`, `StartupEntry` in `ports.go`; `ReleaseSource` in `updates.go`). It builds the
  snapshot, edits clocks, searches places (`SearchPlaces`, FR-302), changes settings, arranges the
  ribbon (`Launch`, `Rearrange`, `Moved`, `ToEdge`, `ToLastEdge`, `Centred`), takes the page's
  measurements, previews and keeps the grip's scale (`PreviewScale`, `SetScale`), checks for updates
  and answers the menus. The snapshot orders cells east from Greenwich (`eastFromGreenwich`): places
  level with or ahead of UTC by offset, then those behind it, read at the snapshot's instant; ties keep
  their stored order and an unshowable clock goes last. A change that cannot be saved stays in effect
  with a notice (FR-707).
- **Infrastructure** (`internal/infrastructure`): on every platform `store`, `zones`, `system` (wall
  clock, ids), `update` and `iconscale`; per platform `monitors`, `startup`, `appdata`, `runlog` and
  `desktop` (tray, native menus, the ribbon's window, the end of a move, the desktop's broadcasts, the
  pointer, the browser opener). Windows only: `setup`. Linux only: `gtkmain`. macOS only: `cocoamain`.
- **UI**: the React front end and the Wails facade in package `main`, which maps the service's
  answers into `dto.go`.
- **Outside the layers**: `internal/product` holds the name, app id, setup program's name, window
  class, donation address, version, author, copyright line, sign-in label and credits. The domain and
  application never read it.
- **Tools**, never shipped: `genplaces` (the place catalogue), `payload` (the setup program's
  payload), `versioninfo` (each executable's version resource), `identity` (names for the Linux and
  macOS scripts), `linuxicons` (the Flatpak's icons) and `genicons.py` (every committed icon).

## Composition root

`main.go` points standard error at the run log before anything can fail, builds the adapters,
injects them into the service, prepares the platform, starts the tray and hands the facade to Wails.
`preparePlatform` does nothing on Windows; on Linux and macOS it hands the desktop the icon and ends
the run on SIGTERM or SIGINT (`quit_signal.go`). `platform_linux.go` sends GTK through X11 and turns
off the DMABUF renderer; `platform_darwin.go` links UniformTypeIdentifiers. The cell sizes (`layouts`)
and panel sizes (`panels`) live there. No service is held in a global.

The facade is `app.go` and `window_life.go`, split for size. It holds the service through the
`ribbonService` interface and each call into Wails and the desktop as a field, so its tests can stand
in for all three. Beside it: `identity.go` (About, Licence), `updates.go`, `measure.go`,
`clockscale.go`, `opacity.go`, `panel.go`, `choices.go`, `unpinned.go`, `sunmap.go`, `dto.go`,
`launch.go` (window options), `launch_show.go` (the first showing) and `bindings_on.go` /
`bindings_off.go`, which keep the binding-generation run from writing the log or showing a tray icon.

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
                     +--------------------------------+
```

## One window

Wails v2 offers one window, so the ribbon, Settings, About, Licence and the update panel share it
(CON-6). Opening a panel resizes the window to its size in `panels` (Settings 900 DIP wide, FR-625;
the others 560), centred on the ribbon's display within its work area; closing returns the ribbon to
where it was. While a panel is open a move is not recorded and a change of length waits for the close.

**Settings fits its content (FR-621).** Once `openPanel` has resolved, `panelFit.ts` measures the
panel at its own width and hands the height to `FitPanel`, which centres it again at that height,
capped by the work area. It measures again whenever the content or the panel's size changes. Measuring
before the window became the panel took the ribbon's narrower window and could overtake the opening.

**Opacity (FR-622).** On Windows the web view is transparent and the window translucent, since Wails
otherwise paints the window solid behind the page; on macOS the web view is transparent; on Linux the
window is translucent (`launch.go`). Everything is drawn inside `#root`. `app.css` mixes
`--window-opacity` into the surface, the dial's face and the tab's accent, so only backgrounds fade
and the clocks stay solid; `App` sets it to 1 while the window is a panel. The window's own paint
shows behind anything less than opaque, so `opacity.go` paints it the page's colour at full opacity
(a window catching up with a new size then shows that colour, not white) and clear below it. The page
reports that colour from a hidden `#surface-swatch`, since `#root`'s faded background is no colour to
paint a window in.

**Scale (FR-623).** Layouts stay in unscaled DIP; `ribbonSize` alone applies the scale, adding the
scroll bar after, since zoom leaves the engine's bar its own thickness. The page draws the ribbon
under CSS `zoom`. The grip sends `BeginScale`, `DragScale` and `EndScale`; Go turns the pointer's
travel into a scale through `settings.ScaleAfter`, fractional while dragging (`PreviewScale`, never
saved) and rounded when kept (`SetScale`). On Windows Go reads the cursor itself (`desktop.Cursor`),
because the page's pointer events jumped backwards while the window resized under them; elsewhere it
takes the page's reading (TECH_DEBT.md item 2). Each step refits the window and tells the page to
redraw. A change of scale keeps the top-left corner: `lengthChanged` counts only a change of length at
the same scale, so clocks re-centre the ribbon (FR-104) and scale does not. While dragging, the sun
map is held at its size and its place from the ribbon's corner (`MapHeld`), so the window's corner
stays still; it is resized once on release. A ribbon flush against the right or bottom edge still
grows left or up (`KeptFlush`).

**Hidden until placed.** The window opens hidden; `startup` finds it, takes it off the taskbar, fences
its moves and places it first. On Windows it is found by its class and `HideFromTaskbar` swaps
Wails' application-window style for a tool window's. A launched ribbon is shown once the page has
reported its scale and widest text, else after `sizeWait` (a second): shown earlier it grew in view
and at a fractional KDE scale often stayed cut off (8 of 16 launches, 2026-10-02).

**The unpinned ribbon.** The facade owns `hover`'s timer (`unpinned.go`). Opening tells the page first
and grows the window once the page reports `RibbonDrawn` (else after `drawWait`); the open ribbon keeps
the tab's frame; the window wears the page's colour (`SetBackground`). Each removed a flicker measured
on Windows (REQUIREMENTS section 2.3). The pointer is read every 50 ms on Windows (against the
window's cut shape) and macOS; Linux hears GTK's crossing events instead, since under XWayland
neither the page nor the X server sees it leave.

**The sun map** shares the window: placements are decided for the ribbon alone, `windowOf` adds the
map and `ribbonFromWindow` reads a dragged window back. On Windows the window is cut to the two
(`placement.Shape`, `desktop.Shape` over `SetWindowRgn`, FR-913) before every placing. The page lays
them out (`Surface.tsx`) at boxes Go sends in the page's units, blends day and night by solar altitude
(`sunLight.ts`) and stands labels clear (`labels.ts`, FR-914).

## The ribbon's size and place

**Size (FR-105, FR-106, FR-620).** `ribbonSize` sizes the ribbon from its notices and clocks (the Add
clock prompt when there are none): cells plus padding along the orientation up to the work area,
beyond which they scroll; one cell across, plus the scroll bar when scrolling and the handle's lane
while the sun map is on. The page measures the scroll bar (`SetScrollbar`), its `devicePixelRatio`
(`SetPixelRatio`) and the widest time and date its font draws (`measure.ts`, `SetMeasured`);
`layoutFor` widens the size's cell to that width, so the cells drawn and the window sized cannot
disagree. The ratio matters because Windows' text size enlarges the page without changing the DPI;
AppKit sizes in points; on Linux a window takes the ratio over GTK's own scale
(`desktop.PixelsPerDIP`), since KDE hands an X11 program a fractional scale as font DPI alone.

**Refits.** Every change that can alter the cells refits the ribbon where it stands; a change of
orientation sends it to that orientation's home edge instead (FR-409, `settings.HomeEdge`: top for
horizontal, right for vertical). Where the length changed, the ribbon is centred along it with its
position across kept (`placement.CentredAlong`) and that place stored (FR-104); the first arrangement
of a run never counts. A failed save's notice is one more cell, fitted once and not saved again.
Against the right or bottom edge the ribbon keeps the far edge flush (`placement.KeptFlush`), since
it is placed by its top-left corner.

**Position (FR-408)** puts the ribbon flush against an edge of the work area it overlaps most,
centred along it (`placement.AgainstEdge`).

**Place (FR-403 to FR-406).** On Windows coordinates are physical pixels on the virtual desktop.
Wails' `WindowSetPosition` is relative to the current monitor's work area while `WindowGetPosition`
is absolute; its screen list has no origin, device name or work area, so displays are read through
`EnumDisplayMonitors` and `GetMonitorInfoW` and the window placed with `SetWindowPos`. A drag's end is
heard through a WinEvent hook on `EVENT_SYSTEM_MOVESIZEEND`; the placement stores the monitor's device
name, work area, DPI and the offset from its corner, restored at launch scaled by any change of DPI.
With nothing stored (or the monitor gone) the ribbon goes to its home edge on the primary.

**The drag (FR-401, FR-402).** A press on empty ribbon that moves past the desktop's drag distance
(Windows' `SM_CXDRAG`/`SM_CYDRAG`, GTK's `gtk-dnd-drag-threshold`, Windows' 4 DIP on macOS) is handed
to the platform's move loop through `window.WailsInvoke('drag')`, internal to Wails v2. On Windows a
window procedure in front of Wails' (`desktop.KeepOnDisplays`) answers each `WM_MOVING` by keeping
the rectangle inside the display under the pointer.

## Time

Zones resolve through `time.LoadLocation` with `time/tzdata` built in (CON-5); an empty id is refused.
Windows has no zone files Go reads, so there the built-in rules are used; macOS and Linux read their
own first. The place catalogue, `internal/infrastructure/zones/places.tsv`, comes from tz 2025b's
`zone.tab` and `iso3166.tab` through `tools/genplaces` and holds 418 zones, each held to resolving by
`TestEveryPlaceResolvesInTheEmbeddedDatabase`.

Each snapshot carries the time to the next minute and the page takes the next one then (FR-208). On
Windows the tray window hears `WM_TIMECHANGE` and `WM_POWERBROADCAST` resumes (FR-209). Linux and macOS
broadcast neither, so `desktop/clockwatch.go` compares the wall clock with Go's monotonic clock every
2 seconds; a drift over 2 seconds counts as a jump.

## The settings file

`settings.json` in the settings folder ([Data locations](#data-locations)) is indented JSON with a
format version (FR-701); derived values are never stored. A save writes a temporary file in the same
folder, flushes it and renames it over the old one (FR-702). Reading is tolerant:

- No file means the defaults and no notice (FR-703).
- A file that is there but cannot be read raises a notice and refuses every save that run.
- A file that is not JSON is renamed to `settings.unreadable.json` with a notice (FR-704); where that
  name is taken, to the first free of `settings.unreadable-2.json` onwards, up to `keptAsideLimit`.
  Where the rename fails, saving is refused from then on.
- A UTF-8 byte order mark is passed over.
- A clock that cannot be read or names an unknown zone is kept as it was and shown in words as invalid
  (FR-705, FR-706); two clocks sharing an id are both kept, the later renumbered (`same-2`).
- An unknown top-level key is written back as found.

**The file is a contract from the first release (NFR-C-1).** No key the first release writes is
renamed, dropped or given another meaning; no stored word changes. Later keys (`size`, `colour`,
`skippedUpdate`, `dateFormat`, `pinned`, `lastEdge`, `sunMap`, `pullOut`, `opacity`, `scale`) are
written after the first release's (`store/decode.go`); a file without them reads as the defaults.
`TestA1Point0SettingsFileIsReadWhole` reads the frozen fixture `store/testdata/settings-1.0.0.json`,
every key set away from its default; it was proved by renaming a key and by changing a stored word.
The fixture is never regenerated.

## Colour

Every colour has one home per scheme: Classic in `frontend/src/theme.css`, the others in
`colours.css` keyed off `data-colour` (FR-611), each token stated once as `light-dark(light, dark)`.
A scheme's hue lives in the tokens the ribbon paints, because the accent reaches only Settings and the
tab.

## The desktop

On Windows `desktop` owns a hidden top-level window on its own locked thread for the tray icon, the
native menus and the broadcasts (a message-only window would not hear them). It re-adds the icon on
`TaskbarCreated`. The desktop reports on a buffered channel, dropping an event with a log line rather
than blocking Windows' thread; the facade's `listen` loop acts on it. Both recover a panic and log it.

Both menus are native popups, so the small window never clips them. Their items have one home,
`internal/application/menus.go` and `menu_choices.go`; identifiers are numbered depth first
(`desktop/menu.go`). Settings offers every menu choice from the same items: `Service.SettingsChoices`
answers them, the page hands the chosen action to `Choose` (`choices.go`), which refuses anything not
offered (FR-624). `TestEveryMenuChoiceIsOfferedBySettings` fails for a menu choice Settings lacks. On
Windows a left click on the tray icon toggles the ribbon; on Linux the tray host's activation does;
on macOS a click opens the menu. A tray icon that cannot be made is not fatal; closing then quits.

## The desktop on Linux and macOS

Both reach the desktop through cgo: GTK 3 and AppKit. What does not depend on the toolkit is written
once in `_unix.go` files: the `Desktop`, its events, the move-end settling, the clock watch, the
window registry, the browser opener, the sign-in file, the icon and signal handling.

- **One thread.** `gtkmain.Do` and `cocoamain.Do` run a function on the toolkit's loop and wait,
  raising a panic again on the caller.
- **Finding and hiding (FR-101).** The ribbon is the top-level window titled with the product's name.
  Linux marks it to skip the taskbar and switcher; macOS makes the application an accessory after
  Wails has made it regular (measured by `lsappinfo` reporting `UIElement`).
- **Coordinates.** Both count in DIP, every display reported at `placement.BaseDPI`; AppKit's
  bottom-left origin is turned over into the domain's top-left reckoning (CON-7).
- **Placing.** macOS uses one `setFrame`. Linux sets the size and awaits it (up to 500 ms) before
  moving, since the window manager clamps a move by the size the window has when it arrives
  (`TestTheRibbonReturnsFromAPanelToWhereItIsPlaced`). WebKit's DMABUF renderer is off unless the user
  set it (`TestAChosenDMABUFSettingIsKept`); hardware acceleration is off (`WebviewGpuPolicyNever`),
  as Wails would choose without Linux options.
- **The end of a move (FR-404)** is the ribbon standing still for 300 ms (`moveSettle`), heard through
  `configure-event` or `NSWindowDidMoveNotification`. A position `Place` chose is never a move
  (`TestAPassingPositionOnTheWayToAPlacementIsNotAMove`). Nothing fences the drag; `Service.Moved`
  brings back a ribbon left partly off every display.
- **Menus.** The right-click menu is a GTK popup handed a button press stamped with the X server's
  time; on macOS an `NSMenu`. The Linux tray is TimeRibbon's own StatusNotifierItem with a
  `com.canonical.dbusmenu` menu over godbus, registered by object path so it needs no bus name or
  sandbox permission; macOS uses an `NSStatusItem` rebuilt as it opens.
- **Sign-in (FR-605).** Linux writes an XDG autostart entry (the real `~/.config/autostart` under the
  Flatpak, running `flatpak run`); macOS a launchd agent. Each is removed when turned off.

## Help, About and Licence

About shows the icon, name and version, author, copyright line and a credit for every component this
platform's build ships (FR-607), from one table in `internal/product/credits.go`;
`TestEveryLinkedModuleIsCredited` holds each platform's credits to the modules its build links.
Licence shows the embedded `LICENSE` exactly as written, its type sized so the widest line fits
(FR-608). Both read themselves when they overflow (FR-609) through one script,
`installer/frontend/dist/auto-scroll.js`, shared with the setup page, which can import nothing;
`frontend/src/autoScroll.ts` types it and wraps it in a React hook.

## The update check

`internal/infrastructure/update` asks GitHub's `releases/latest`, which answers only a published
release, never a draft or prerelease. It is unauthenticated, times out after 5 seconds, never retries
and reads at most a megabyte. The service compares the tag with the stamped version as dotted
integers (anything else is never newer), picks this platform's asset by its ending and honours the
skipped release except on a manual check. Addresses are taken only as `https` on `github.com` with no
user or port (`onGitHub`). `updates.go` checks 3 seconds after start, then every 24 hours, recovering
any panic; a check with something to say shows the update panel. The addresses stay in Go: Download
and Skip ask Go to act on what it offered.

## Delivery

**macOS and Linux** build with `go build` and Wails' `desktop,production` tags, the version from
`VERSION` through `-ldflags`, names from `tools/identity`.

- `builddmg.sh` builds for Apple Silicon, assembles and signs `TimeRibbon.app` with the hardened
  runtime, notarises and staples it, then the DMG. The minimum macOS is read from the Go toolchain and
  passed through the cgo flags; a link of code built for a newer macOS is refused.
- `build_flatpak.sh` builds in the GNOME 50 SDK against WebKitGTK 4.1 (`-tags webkit2_41`). The
  sandbox gets X11 with IPC, the GPU, the tray host's bus name, the single-instance lock's bus name,
  the autostart folder and the network for the update check alone. Both scripts first stop a copy
  left running, which holds the single-instance lock (FR-506).

**The setup program** (Windows) is a second Wails application in `installer/`, embedding the built
application as a zip. `build.ps1` packs it through `tools/payload`, builds setup, then writes the empty
placeholder back whatever happened. The install policy lives in `internal/infrastructure/setup`: the
paths, the extraction with its fence against an entry leaving the install folder (every entry checked
before any is written, FR-803), the version comparison, the Apps list record, the shortcuts (COM via
go-ole), Start with Windows (through `startup`, the value Settings writes) and the step log.
`installer/app.go` is a facade over it. Steps are weighted by measured time.

| Reading of the machine | Screen |
|---|---|
| started with `-uninstall` | Uninstall |
| nothing installed | Install |
| the version carried is newer | Update |
| the version carried is older | Go back |
| the versions match | Installed: Repair, Reinstall, Uninstall |

Everything written is per user (FR-810). Repair keeps the shortcuts and Start with Windows as they
are. Uninstall deletes `%APPDATA%\TimeRibbon` only when **Also forget my settings** is ticked, then a
hidden PowerShell deletes the install folder once setup has exited. Setup offers to close a running
copy, waiting up to 5 seconds (FR-807). The setup page names nothing: the product's name arrives on
its state. Its keyboard ring (`setup-ring.js`) is the window's model written again, held by
`setupRing.test.ts`; each screen opens with nothing focused (FR-809).

## Data locations

| What | Where |
|---|---|
| Settings | `settings.json`: `%APPDATA%\TimeRibbon` on Windows, `~/Library/Application Support/TimeRibbon` on macOS, `~/.var/app/uk.codecrafter.TimeRibbon/config/TimeRibbon` for the Flatpak, `$XDG_CONFIG_HOME/TimeRibbon` else `~/.config/TimeRibbon` outside it; kept-aside copies beside it |
| Run log | `TimeRibbon.log` beside the settings, started afresh over 1 MiB |
| Web view data on Windows | `%APPDATA%\TimeRibbon\WebView2`, named in `launch.go` so forgetting the settings removes it |
| Time zone rules, place catalogue | built in; macOS and Linux read their own zone files first |
| Installed files | Windows: `%LOCALAPPDATA%\Programs\TimeRibbon` with `uninstall.exe`; macOS: where the user drags the app; Linux: the user's Flatpak installation |
| Start at sign-in | Windows: `TimeRibbon` under `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, the quoted path and no arguments. macOS: `~/Library/LaunchAgents/uk.codecrafter.TimeRibbon.plist`. Linux: `~/.config/autostart/uk.codecrafter.TimeRibbon.desktop` |
| Apps list record | `HKCU\...\Uninstall\TimeRibbon` |
| Setup's log and web view data | `TimeRibbonSetup.log` and `TimeRibbonSetup` in the temporary folder |

## Errors

Errors are wrapped with `%w` at each boundary, so `errors.Is` finds sentinels such as
`ErrNoSuchClock` beneath.

- **Before the window** only a failure to run the window or to read the embedded catalogue ends the
  run. Standard error goes to the log first (`runlog.Keep`), so even a runtime panic is kept. A
  missing settings folder falls back to the temporary folder; unreadable settings, a failed tray icon
  and a missing executable path are logged and the ribbon still opens.
- **On the ribbon:** a kept-aside file and a failed save as notices; an invalid clock in words.
- **Beneath the control pressed:** every page call Go can refuse. Each `api` wrapper takes a refusal
  handler and answers null rather than rejecting, so a call without one does not compile.
- **Logged and carried on:** a dropped desktop event, a panic in the desktop's thread or its handling,
  a failure to place, hide from the taskbar or fence the ribbon.

## Quality enforcement

- The structural tests above run with the suite.
- `test.ps1` checks formatting, vet and staticcheck, runs the Go suite and the front end's lint, type
  check and tests, holds the domain and application to 100% and every other gated package to its
  measured floor ([TESTING.md](TESTING.md)).
- `build.ps1` runs `test.ps1` first with no switch to skip it, cgo off for both.
- The Linux and macOS code is checked on its own platform ([TESTING.md](TESTING.md#on-macos-and-linux)).

## Design decisions

The architectural ones; the full set with their costs is in
[DECISIONS-TRADEOFFS.md](DECISIONS-TRADEOFFS.md).

| Decision | Why | Rejected alternative |
|---|---|---|
| Displays and placement through each desktop's own calls | Wails' screen list lacks origin, device name and work area; its position calls mix relative and absolute coordinates (CON-7) | Wails' position calls |
| One window for the ribbon and every panel | Wails v2 offers one | A second window per panel |
| Native popup menus | A page-drawn menu would be clipped by the small window | A menu drawn in the page |
| The page measures cells, scroll bar and scale | Only the page knows its font, its engine's bar and the ratio it is drawn at | Widths and thicknesses written into Go |
| The grip's cursor read from the desktop on Windows | The page's pointer events jumped backwards while the window resized under them | The page's `screenX`, `screenY` |
| The web view's data inside the settings folder | Wails' default sat beside it, out of reach of forgetting the settings | Deleting Wails' folder by name |
| Linux on X11; DMABUF and acceleration off | Wayland forbids choosing a position; NVIDIA's own driver drew blank | Wayland; per-driver detection |
| A StatusNotifierItem of TimeRibbon's own | `fyne.io/systray` could not rebuild its menu as it opens and kept global state (v1.12.2) | `fyne.io/systray` |
| Toolkit-free code shared in `_unix.go` | Linux and macOS cannot drift apart | A copy per platform |

See also [TESTING.md](TESTING.md) and [DEVELOPMENT.md](DEVELOPMENT.md).
