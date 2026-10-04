# TimeRibbon Architecture

A small desktop application for Windows, macOS and Linux showing a ribbon of clocks, one per chosen
place. What it reads from outside itself is the system clock, its settings file, the desktop (the
displays, the tray, the sign-in entry) and GitHub's latest release for the update check. The time
zone rules are built into the executable; on macOS and Linux the system's own zone files are read
before them ([Time](#time)). The domain and application are the same code on every
platform; each platform's own half sits in files its build tags or file names select, in
infrastructure and in the root package's `platform_*.go` files
([The desktop on Linux and macOS](#the-desktop-on-linux-and-macos)). Its one network request
is the update check (FR-509): of this module's Go files only those in `internal/infrastructure/update`
import a network package, which `TestOnlyTheUpdateCheckImportsANetworkPackage` holds. Each of those
guards is a list of what is known, not a proof that nothing else asks: `requests_test.go` holds the
other ways out (a program started, a Windows library loaded by name, a request from either page) to
the ones named there. The donation page and a release's download are handed to the desktop's browser
rather than fetched; a release's addresses are taken only as `https` on `github.com`.

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
| Domain imports nothing from this module outside `internal/domain` | `TestDomainHasNoOutwardImports` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Domain is pure: no network, filesystem, process, random or tz database package; no wall clock read and no zone loaded (FR-207, CON-5) | `TestDomainIsPure` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Application never imports infrastructure or Wails | `TestApplicationDoesNotImportInfrastructure` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Infrastructure never imports Wails | `TestWailsStaysOutOfInfrastructure` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Only `main.go`, `app.go` and `window_life.go` import both the application and infrastructure | `TestCompositionRootIsWhitelisted` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| No source file exceeds 400 lines: the Go, the front end's TypeScript and CSS, the setup page | `TestNoFileExceedsLineLimit` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| No source file sits in the danger band of 381 to 400 lines | `TestNoFileInDangerBand` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Every exported type carries a doc comment | `TestEveryExportedTypeIsDocumented` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| A file's lines are counted as an editor numbers them | `TestLineCountCountsTheLinesAnEditorShows` | [`linecount_test.go`](tests/structural/linecount_test.go) |
| No Go file of this module outside `internal/infrastructure/update` imports `net`, `crypto/tls` or `golang.org/x/net` (NFR-S-1; a denylist of imports, so the three rows after the next hold the other ways out) | `TestOnlyTheUpdateCheckImportsANetworkPackage` | [`network_test.go`](tests/structural/network_test.go) |
| Only the files named in `processStarters` start a program or hand an address to the desktop; no shipped Go file names a Windows library outside `systemLibraries` (NFR-S-1) | `TestOnlyNamedFilesStartAProcess` | [`requests_test.go`](tests/structural/requests_test.go) |
| Every file named in `processStarters` exists | `TestEveryNamedProcessStarterExists` | [`requests_test.go`](tests/structural/requests_test.go) |
| No file of the page or the setup page uses a request API or names a web address, the SVG namespace aside (NFR-S-1) | `TestThePageMakesNoRequest` | [`requests_test.go`](tests/structural/requests_test.go) |
| Those checks recognise `os/exec`, `ShellExecute`, a library such as `WinHTTP.DLL`, `fetch`, `WebSocket` and an address while passing look-alikes such as `prefetch` | `TestRequestRecognitionIsExact` | [`requests_test.go`](tests/structural/requests_test.go) |
| The network exemption names a directory that exists, so a moved update package cannot leave it pointing at nothing | `TestTheNetworkExemptionNamesTheUpdatePackage` | [`network_test.go`](tests/structural/network_test.go) |
| The network package check recognises `net`, `crypto/tls` and `golang.org/x/net` paths and passes look-alikes | `TestNetworkPackageRecognitionIsExact` | [`network_test.go`](tests/structural/network_test.go) |
| The setup page loads every script beside it | `TestTheSetupPageLoadsEveryScript` | [`setup_test.go`](tests/structural/setup_test.go) |
| No setup page file spells the product's name, which the setup program hands it | `TestTheSetupPageNeverWritesTheProductsName` | [`setup_test.go`](tests/structural/setup_test.go) |
| Each platform's About credits exactly the modules that platform's build links (FR-607) | `TestEveryLinkedModuleIsCredited` | [`credits_test.go`](tests/structural/credits_test.go) |
| No platform's About credits one module twice | `TestAModuleIsCreditedOncePerPlatform` | [`credits_test.go`](tests/structural/credits_test.go) |
| The wire is stated identically in `dto.go` and `frontend/src/wire.ts` | `TestTheWireIsStatedAlikeOnBothSides` | [`wire_test.go`](tests/structural/wire_test.go) |
| The page listens for every event `app.go` emits and keys every panel it names | `TestThePageNamesEveryEventGoEmits` | [`wire_test.go`](tests/structural/wire_test.go) |
| The setup page listens for every event `installer/app.go` emits | `TestTheSetupPageNamesEveryEventSetupEmits` | [`wire_test.go`](tests/structural/wire_test.go) |
| Each `wails.json` names its executable as `internal/product` does | `TestEachWailsConfigNamesItsExecutableAsTheProductDoes` | [`names_test.go`](tests/structural/names_test.go) |
| Every scheme the menus offer has its own block in `colours.css` stating each of Classic's tokens (the problem colour aside); every block is offered (FR-611) | `TestEveryOfferedSchemeHasItsOwnCompleteBlock` | [`colours_test.go`](tests/structural/colours_test.go) |
| Text, muted text and problem text meet 4.5:1 against the cell and the surface on every offered scheme in both themes, a token a scheme leaves out read as Classic's (NFR-U-1) | `TestTextMeetsTheContrastFloorOnEverySchemeAndTheme` | [`contrast_test.go`](tests/structural/contrast_test.go) |
| Classic's dark colours are the same under the system's dark mode as under a chosen dark theme, so the one the contrast test reads is the one shown | `TestClassicDarkIsTheSameUnderTheSystemAsWhenChosen` | [`contrast_test.go`](tests/structural/contrast_test.go) |
| The contrast ratio is computed as WCAG 2.x states it and a colour form it cannot read is refused | `TestContrastIsComputedAsTheStandardStatesIt` | [`contrast_test.go`](tests/structural/contrast_test.go) |
| The Licence panel is sized for the LICENSE's widest line, so it shows unwrapped | `TestTheLicencePanelIsSizedForTheLicencesWidestLine` | [`licence_test.go`](tests/structural/licence_test.go) |
| No tracked or new file holds the product's former name or the word its window went by before the ribbon; the npm lock file and calls of Python's string method of that name aside | `TestNoTrackedFileHoldsTheRetiredWord` | [`retired_test.go`](tests/structural/retired_test.go) |
| That word is recognised in any case and inside names while the current names pass | `TestTheRetiredWordIsFoundInAnyCaseAndInsideNames` | [`retired_test.go`](tests/structural/retired_test.go) |

## Layers

- **Domain** (`internal/domain`: `clock`, `hover`, `placement`, `settings`, `sun`): pure Go. Time
  arrives as an argument and a zone arrives already resolved, so the domain holds no tz database and
  reads no clock. `clock` turns an instant and a zone into what a cell shows: the local time in either
  format, the date in the chosen date format (`DateFormat`, FR-612), the zone mark (the
  abbreviation where the tz database gives one beginning with a letter, else `UTC` and the signed
  offset) and the hand angles; `NextRefresh` names the next minute boundary. It also derives a
  zone's default label and writes every time and date a cell can show for the page to measure
  (`Samples`, FR-620). `placement` decides where the ribbon goes, in physical pixels: the default
  place, a stored placement restored on its monitor at that monitor's DPI, the least move that
  brings a ribbon wholly inside a work area (`Clamp`, `Recover`), the ribbon's length along its
  orientation (`Fit`) plus a ribbon centred along its length on a work area with its position
  across kept (`CentredAlong`) or flush against one of its edges and centred along it
  (`AgainstEdge`), plus the band an unpinned ribbon shrinks to on the side flush against its edge
  (`Tab`, FR-614). Which edge a ribbon stands flush against, counting only the edges along its
  orientation of each display's own work area (`FlushAgainst`, `Along`) lives in `edge.go`
  beside the snap of a drop within `SnapReach` of one (`Snapped`, FR-410). `settings` is the user's choices
  as one value; every operation answers a new value and leaves the old one as it was. It holds the
  pin in effect (`PinnedInEffect`: pinned or flush against no edge, FR-619) and the one rule for
  staying on top built on it (`OnTop`: Always on top or unpinned in effect, FR-617), plus the edge
  last stood against (`LastEdge`, FR-411). It bounds the opacity (`MinOpacity` 20 to `MaxOpacity` 100
  percent, FR-622) and the scale (`MinScale` 75 to `MaxScale` 200 percent, FR-623). Every arrangement names its flush edge
  (`Arrangement.Edge`), which the facade reads the pin in effect from. `sun` answers the subsolar
  point for an instant from NOAA's equations (FR-906), checked against NOAA's own values in
  `testdata`; `placement/sunmap.go` puts the sun map beside the ribbon on the side away from its edge
  (`InnerSide`) at its size (`MapBeside`, FR-902 to FR-904). The map shares the ribbon's window:
  every placement is decided for the ribbon alone, the facade makes the window the ribbon with its
  map (`windowOf`) and reads a dragged window back to the ribbon's corner (`ribbonFromWindow`). On
  Windows that window is then cut to the two (`placement.Shape`, `desktop.Shape` over
  `SetWindowRgn`, FR-913), so the desktop shows round a map shorter or longer than the ribbon; the
  cut is made before every placing, in the new window's pixels; the tab and panels keep the whole
  window. The page lays the two out (`Surface.tsx`) at boxes Go sends already in the page's units,
  divided by the pixels to each unit that windows are sized with (`boxOf`), so an opening ribbon drawn
  while the window is still its tab is drawn at its full size; the map counts as shown while that
  drawing is under way (`mapLayout`). It blends the day and night pictures by
  solar altitude (`sunLight.ts`); each label is measured once drawn, then stood clear of the other
  labels and dots (`labels.ts`, FR-914), the dot size and gap read from the page's style. The zone
  cities come from the tz database's `zone.tab` through `tools/genplaces`.
  `hover` decides when an unpinned ribbon opens from its tab and collapses back (FR-615, FR-616). It
  is told the pointer arrived or left and the time; it answers whether the ribbon is open and when
  to ask again. The facade owns its timer and carries the answer out (`unpinned.go`). Opening tells
  the page first and grows the window once the page says it has drawn the full ribbon
  (`RibbonDrawn`), with a second timer (`drawWait`) growing it regardless should the page never say;
  an unpinned ribbon wears the tab's frame when full too; the page hands Go its background
  colour (`frontend/src/background.ts`, `SetBackground`). Each of the three removed a flicker
  measured on Windows (REQUIREMENTS section 2.3). Where the
  pointer is comes from the desktop, differently on each system because each was measured to need
  it: read every 50 ms on Windows and macOS (on Windows against the window's cut shape, so the
  desktop showing beside a vertical ribbon's map counts as off it), told by GTK's crossing events on Linux, since under
  XWayland neither the page nor the X server sees it leave (REQUIREMENTS section 2.3).
- **Application** (`internal/application`): one `Service` holding every use case over seven ports
  (`Store`, `Zones`, `Clock`, `IDs`, `Monitors`, `StartupEntry` in `ports.go`; `ReleaseSource` in
  `updates.go`). It builds the snapshot the ribbon draws, adds, edits and removes clocks, searches
  places (what is typed beginning a word, the best matches first: `SearchPlaces` in `clocks.go`,
  FR-302), changes settings, arranges the ribbon (`Launch`, `Rearrange`, `Moved`, `ToEdge`,
  `ToLastEdge`, `Centred`), takes the cell width the page measured (`SetMeasured`) and the scale the
  grip previews or keeps (`PreviewScale`, `SetScale` in `scale.go`), checks for an update
  (`CheckForUpdate`, `SkipUpdate`) and answers the tray and context menus. The snapshot orders its cells east from Greenwich (`eastFromGreenwich` in `snapshot.go`):
  places level with or ahead of UTC by ascending offset, then the places behind UTC, since going
  east reaches them last. Offsets are read at the snapshot's instant, so the order is worked out
  afresh each time and daylight saving can move it; clocks keeping the same time keep their stored
  order and a clock that cannot be shown goes last. There is no ordering by hand. A change that
  cannot be saved stays in effect and raises a notice until a later save succeeds (FR-707). It
  never imports infrastructure or Wails.
- **Infrastructure** (`internal/infrastructure`): the adapters behind the ports and the desktop
  integration. The same on every platform: `store` (the settings file), `zones` (resolution through
  the embedded tz database and the place catalogue), `system` (the wall clock and new clock ids),
  `update` (the latest release, asked of GitHub: the one network request) and `iconscale` (decoding
  and scaling the icon for the Linux tray and icons). One file or more per
  platform: `monitors` (the displays), `startup` (the sign-in entry), `appdata` (the settings
  folder), `runlog` (the run's log), `desktop` (the tray icon, the native menus, the ribbon's window,
  the end of a move, the desktop's broadcasts and handing an address to the default browser, which
  reports a refusal). Windows only: `setup` (the install policy behind the setup program). Linux
  only: `gtkmain` (running work on GTK's loop). macOS only: `cocoamain` (running work on AppKit's
  main thread).
- **UI**: the React front end plus the Wails facade in package `main`, which calls the service and
  maps what it answers into the shapes in `dto.go`.
- **Outside the layers**: `internal/product` holds the product's name, its app id
  (`uk.codecrafter.TimeRibbon`), the setup program's name, the window class, the donation address,
  the version each build script stamps, the author, the copyright line, the sign-in label in each
  platform's words and the credits for each platform. Infrastructure, the facade, the setup program
  and the tools read it; the domain and application never do. It belongs to no layer.
- **Tools**, never shipped: `tools/genplaces` writes the place catalogue from the tz database's
  `zone.tab` and `iso3166.tab`; `tools/payload` packs the built application for the setup program;
  `tools/versioninfo` writes each executable's Windows version resource from `VERSION` and
  `internal/product` before its build, in place of Wails' template, which carried its fallback
  version and a placeholder copyright; `tools/identity` prints the names the Linux and macOS build
  scripts need, read from `internal/product`, so neither script keeps a second copy;
  `tools/linuxicons` writes the Flatpak's icon sizes from `build/appicon.png`;
  `tools/genicons.py` writes every committed icon from the masters in `assets/`.

## Composition root

`main.go` is the composition root. It opens the run log and points standard error at it before
anything can fail, builds the adapters, injects them into the service by constructor, prepares the
platform, starts the tray and hands the facade to Wails. `preparePlatform` does nothing on Windows;
on Linux and macOS (`platform_unix.go`) it hands the desktop the icon, which there is an image rather
than a resource in the executable, then ends the run on SIGTERM or SIGINT through Exit
(`exitWhen` in `quit_signal.go`). `platform_linux.go` sends GTK through X11 and turns off the web view's DMABUF renderer before Wails
opens it;
`platform_darwin.go` links the UniformTypeIdentifiers framework, which Wails' macOS half uses and
which the `wails` command would otherwise have added. The cell sizes with the padding and the
handle's lane (`layouts`) and the panel sizes (`panels`) have their one home there. No service is held in a package-level variable and there
is no service locator.

The facade is `app.go` (the calls the page makes) and `window_life.go` (startup, showing, hiding,
closing and the desktop's events), split only to keep each file small; the structural whitelist
names those two with `main.go`. The facade holds the service through `ribbonService`, an interface
in `app.go`. It holds each call into Wails and the desktop as a field, pointed by `newApp` at the
real calls: Wails' in `wails_calls.go`, the ribbon's position and placing in `window_life.go` and
the desktop package's `OpenInBrowser` and `ShowMenu`. That is what lets the facade's tests stand in
for the service, Wails and the desktop and read what it decided. `identity.go` answers About and
Licence; `updates.go` runs the update check ([The update check](#the-update-check)); `measure.go`,
`clockscale.go` and `opacity.go` carry the page's measurement, the grip's scale and the opacity to
the service ([The ribbon's size and place](#the-ribbons-size-and-place), [One window](#one-window));
`dto.go` holds the wire; `launch.go` holds the window's options; `launch_show.go` decides when the
launched ribbon is first shown ([One window](#one-window)); `bindings_on.go` and `bindings_off.go` tell the
run `wails build` makes to generate bindings, which carries the `bindings` build tag, not to write
the log, read the settings or show a tray icon.

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
                     | runlog, desktop, setup, update,|
                     | iconscale, gtkmain, cocoamain  |
                     +--------------------------------+
```

## One window

Wails v2 offers one window, so the ribbon, Settings, About and Licence share it (CON-6). The ribbon is
the window at the size its clocks need. Opening a panel resizes the window to its size in `panels`
(`panel.go`: Settings 900 DIP wide for its columns of choices, FR-625; every other panel 560), centred on
the ribbon's display and never larger than its work area (`Service.Centred`); closing one returns the
window to where the ribbon was last left. While a panel is open, a move of the window is not recorded
as the ribbon's and a change of content is fitted when the panel closes.

Settings then grows to its content (FR-621). Whenever anything inside it changes, `panelFit.ts`
measures the panel laid out with no height of its own and hands that height to `FitPanel`, which centres
the panel again at it through the same `Service.Centred`; so on a display with room nothing scrolls,
while a shorter one still caps it at the work area; the fit keeps the width the panel opened at.
About and the update panel keep their size in `panels`; Licence keeps it too, since it reads itself
down its own scroller.

**Opacity (FR-622).** The web view is transparent on Windows and macOS and the window translucent on
Linux (`launch.go`); the desktop showing through is checked by hand in a real build. Everything is
drawn inside `#root`; `html` and `body` are clear. Only the backgrounds take the chosen opacity:
`app.css` mixes `--window-opacity` into the surface, the dial's face and the tab's accent with
`color-mix`, so the clocks drawn on them stay wholly opaque. `App` sets `--window-opacity` only while
the window is the ribbon and to 1 while it is a panel, so Settings is always opaque; the slider
saves once it is let go. The window's own paint shows behind any part of the page drawn less than
opaque, so `opacity.go` paints it in the page's colour only at full opacity (a window catching up
with a new size then shows that colour rather than white) and clear below it.

**Scale (FR-623).** A size's layout stays in unscaled DIP everywhere Go keeps it, the measured cell
widths included. `ribbonSize` alone applies the scale, by giving each unscaled DIP scale percent of
the pixels it otherwise takes; only the scroll bar is added after, since zoom leaves the web
engine's bar at its own thickness (measured in Edge's engine, 2026-09-29). The page draws the
ribbon at its unscaled sizes under CSS `zoom`, which multiplies every length inside while the
ribbon's own 100 percent box still fills the window; `Surface` sizes the pull out's handle by the
same scale. The corner grip (`ScaleGrip.tsx`) sends `PreviewScale` while it moves, held in memory
and never saved, then `SetScale` once let go; each refits the window and tells the page to draw
again, since the page cannot reload in the middle of a drag. The refit keeps the window's top-left
corner, as a resized window does: `lengthChanged` remembers the scale beside the length and counts
only a change of length at the same scale, so a change of clocks re-centres the ribbon (FR-104) and
a change of scale does not.

The window opens hidden. `startup` finds it, takes it off the taskbar, fences its moves and places
it, all before the page is shown, so it never appears blank or in the wrong place. On Windows it is
found by the class `TimeRibbonWindow`; Wails always marks its window as an application window, which
forces a taskbar button, so `HideFromTaskbar` takes that style off and marks it a tool window once,
before it is shown. Linux and macOS find it by its title and keep it off the taskbar or Dock their
own way ([below](#the-desktop-on-linux-and-macos)).

**The first showing (`launch_show.go`, every platform).** A launched ribbon is shown once the page
is ready and has reported both its scale and its widest text, the two reports that resize the
window after it loads; a page that has not within a second (`sizeWait`) is shown anyway. Shown at
the page's readiness alone, the window grew while visible. At a fractional KDE scale WebKitGTK then
often kept painting the size it was first shown at, the ribbon cut off until the page next changed
(measured 2026-10-02 at 150 percent: 8 of 16 launches cut off; shown once sized, 16 of 16 whole).

## The ribbon's size and place

**Size (FR-105, FR-106).** `ribbonSize` in `arrange.go` sizes the ribbon from the cells the page
draws: each notice, then each clock (the Add clock prompt standing in for them when there are
none). Along the orientation the ribbon is that many cells plus padding, while that fits the work
area of its display; beyond that it is the work area's length and its cells scroll. Across, it is
one cell plus padding, plus the thickness of the scroll bar when the cells scroll, so the bar never
covers them. The bar is the web engine's, not one Windows reports, so
the page measures it once it has loaded and hands it to Go through `SetScrollbar`. The ribbon hides
overflow on both axes and scrolls only along its own; hiding one axis alone let the browser turn the
other into a second scroll bar, measured in Edge on 2026-09-27. A plain wheel moves a scrolling
horizontal ribbon along. While the sun map is on, the ribbon is deeper by the handle's lane, so
the handle never covers a cell either (FR-903).

**Cell width (FR-620).** A size's widths in `main.go` are the least a clock cell is drawn at. Only
the page knows the font it really draws with, so it measures: `clock.Samples` writes every time of
the day and every date of a 28 year span in the chosen formats; `measure.ts` lays them all out in
one pass with a cell's own classes and hands the widest, padding and divider included, to Go
through `SetMeasured`. `layoutFor` in `measure.go` widens the style's cell to it when it was taken
under the choices now in force; the snapshot and `ribbonSize` both read `layoutFor`, so the cells
the page draws and the window Go sizes cannot disagree. The page measures again whenever the size,
style or either format changes.

Every change that can alter the cells (a clock added or removed, the style, size or scale changed,
the sun map turned on or off, a notice raised by a failed save or dismissed, the scroll bar or a cell
width reported) refits the ribbon where it stands; a
change of orientation sends it to that orientation's home edge instead (FR-409, below). Where the
refit changes the ribbon's length, it is centred along that length on its display with its
position across kept (`placement.CentredAlong`); `recentredKept` in `arrange.go` stores
that place (FR-104). The service remembers the length it last arranged, with the scale it was
drawn at, to tell a change; the first arrangement of a run never counts as one, nor does a length
changed by the scale, which keeps the top-left corner (FR-623, [above](#one-window)). A length that changed while a panel was open is centred
as the panel closes. Should the save fail, its notice is one more cell, so the ribbon is arranged once
more to fit it and that arrangement is not saved again. Apart from Position and a change of
orientation, nothing else re-centres the ribbon, so a drag holds until the length next changes. Sizes
are computed in DIP, so a ribbon moved between displays at different scaling keeps its size in DIP
(FR-407). They are turned into window pixels at the scale the page is really drawn at: the page
reports its `devicePixelRatio` once it has loaded and again whenever it changes
(`frontend/src/pixelRatio.ts`, `SetPixelRatio`); the display's DPI stands in only until it has.
On Windows that ratio includes the user's text size, which enlarges the page without changing the
display's DPI, so a window sized by the DPI alone cut the page off above 100%. On macOS AppKit
sizes the window in points, so one CSS pixel is one unit. On Linux GTK sizes the window in its own
units, device pixels over its whole window scale, while WebKitGTK draws the page at a ratio that
also carries the font DPI the desktop sets: KDE hands an X11 program a fractional display scale as
font DPI alone, so at 150% the page reports 1.5 with GTK's scale still 1. A window takes the ratio
over GTK's scale (`desktop.PixelsPerDIP`, with the scale read by `desktop.ToolkitScale`); one unit
for every CSS pixel left the ribbon cut off and Settings squeezed into a narrower layout.

The cell sizes live in one table in `main.go`, one layout per size setting (FR-610): large and
small, each giving a digital, an analogue and a prompt cell plus the padding and the handle's lane.
The service picks the layout for the current size (`Layouts.For`, widened by `layoutFor` as above)
and hands it to the page in the snapshot along with the size itself; the page marks the ribbon
`small` so `app.css` reduces the text and the dial to fit. The widths in the table are floors, so a
date format or a font wider than they allow still shows whole once the page has measured it; each date
format on screen is checked by hand in a real build.

**Centred on an edge (FR-408).** The Position submenu's items name an edge each (`EdgeOf` in
`menus.go`); `ToEdge` puts the ribbon flush against that edge of the work area it overlaps most,
centred along it (`placement.AgainstEdge`); it stores the place through `recentredKept`, so a
failed save fits the ribbon to its notice and keeps it flush. A later change of length re-centres it
along that edge, since re-centring keeps the position across. A ribbon is placed by its top-left
corner, so against the right or bottom edge a change of thickness (a change of size or style, a
scroll bar) would pull it off: the service remembers where it last arranged the ribbon. When it
places the ribbon again on the same display with that corner unmoved it keeps the far edge flush
(`placement.KeptFlush`), measured to fail without it by `TestShrinkingKeepsTheRibbonAgainstItsEdge`.

**An orientation's home edge (FR-409).** Choosing an orientation sends the ribbon to that
orientation's home edge (`settings.HomeEdge`, a domain rule since the default place uses it too):
the top for horizontal, the right for vertical. The facade's `SetOrientation` asks the service to
choose, then reads the settings back: where the choice took, even with its save failed, it puts the
ribbon against the home edge through `ToEdge`; where it was refused, it fits the ribbon where it
stands.

**Place (FR-403 to FR-406).** On Windows, coordinates are physical pixels on the virtual desktop;
Linux and macOS use DIP, as their section below says. Wails' `WindowSetPosition` places a window
relative to the work area of the monitor it is on while `WindowGetPosition` answers absolute
coordinates. Its screen list carries no origin, device name or work area either. So displays are
read through `EnumDisplayMonitors` and `GetMonitorInfoW` (`monitors`) and the window is placed with
`SetWindowPos` (`desktop.Place`). With nothing stored the ribbon goes flush against its
orientation's home edge on the primary work area (the right for vertical, the top for horizontal),
centred along it; a ribbon whose monitor has gone or which was left off every display goes there
too. The end of a drag is heard through a WinEvent hook on `EVENT_SYSTEM_MOVESIZEEND`; the placement
is stored as the monitor's device name, its work area, its DPI and the ribbon's offset from the work
area's corner. At launch it is restored on that monitor, the offset scaled by any change of DPI;
where that monitor is gone it goes to the default place on the primary. A display change refits the
ribbon where it is.

**The drag (FR-401, FR-402).** A press on empty ribbon area that moves past the desktop's drag
distance (Windows' `SM_CXDRAG` and `SM_CYDRAG`, GTK's `gtk-dnd-drag-threshold`; macOS publishes
none, so it uses Windows' 4 DIP) hands the press to the platform's own move loop through
`window.WailsInvoke('drag')`, the message Wails' own drag regions send. That message is internal to
Wails v2 rather than a documented call; a press on a control never starts one. On Windows, while
the window moves, a window procedure placed in front of Wails' own (`desktop.KeepOnDisplays`)
answers each `WM_MOVING` by moving the proposed rectangle the least distance that keeps it inside
the work area of the display under the pointer, so the ribbon can be carried onto another display
but never left half off one.

## Time

Zones resolve through `time.LoadLocation` with `time/tzdata` built in (CON-5); an empty zone id is
refused rather than read as UTC. `LoadLocation` reads a directory named by `ZONEINFO` first, then
the platform's own zone files, then the built-in copy. Windows has no zone files Go reads, so there
the built-in rules are the ones used; macOS and Linux use their own zone files (`/usr/share/zoneinfo`
first) and fall back on the built-in rules only for a zone those lack (read in Go 1.26's
`time/zoneinfo.go` and `zoneinfo_unix.go`). The place catalogue,
`internal/infrastructure/zones/places.tsv`, is written by `tools/genplaces` from tz 2025b's
`zone.tab` and `iso3166.tab` and holds 418 zones; a test holds every one of them to resolving
(`TestEveryPlaceResolvesInTheEmbeddedDatabase`).

Each snapshot carries the milliseconds to the next minute boundary; the page takes the next snapshot
then, so each refresh is scheduled from the current time rather than from the last one
(FR-208). On Windows the hidden tray window hears `WM_TIMECHANGE` and the resume broadcasts of
`WM_POWERBROADCAST`, on which the page takes a fresh snapshot at once (FR-209). Linux and macOS
broadcast neither, so `desktop/clockwatch.go` compares the wall clock with Go's monotonic clock every
2 seconds. Setting the time moves only the first; the second does not advance while the machine
sleeps. Either shows as the two drifting apart by more than 2 seconds.

## The settings file

`settings.json` in the settings folder (`internal/infrastructure/appdata`: `%APPDATA%\TimeRibbon` on
Windows, `~/Library/Application Support/TimeRibbon` on macOS, `$XDG_CONFIG_HOME/TimeRibbon` else
`~/.config/TimeRibbon` on Linux), indented JSON a person can read, carrying a format version
(FR-701). Derived values (offsets, abbreviations, times) are never stored. It is written to a
temporary file in the same folder, flushed and renamed over the old one, so a failure part way leaves
the previous file whole (FR-702). Reading is tolerant: no file means the defaults and no notice
(FR-703); a file that is there but cannot be read (held open by another program at sign-in, say)
raises a notice and refuses every save for the rest of the run, since the defaults standing in for
it are not the user's. A file that is not JSON is renamed to `settings.unreadable.json` and a notice
says so (FR-704); where that name is taken by an earlier copy, the first free of
`settings.unreadable-2.json` onwards (up to `keptAsideLimit`) is used instead, so no kept copy is
ever replaced. Should the rename fail or every name be taken, saving is refused from then on so the
only copy is never overwritten. A UTF-8 byte order mark before the text, which Notepad and
PowerShell 5 can write, is passed over. One clock that cannot be read or names an unknown zone is
kept in the file as it was and shown in words as an invalid clock while the others work (FR-705,
FR-706). Two clocks sharing an id, as a hand-edited file can hold, are both kept: the later is given
that id numbered (`same-2`), which the next save writes, so a change to one never reaches the other.
A top-level key this version does not know is written back as it was found.

**The file is a contract from the first release (NFR-C-1).** Every later release, the next major
version included (Amendment 11), reads every file the first release writes to the same settings. No key it writes may be
renamed, dropped or given another meaning. No stored word (such as `12h` or `analogue`) may change.
A later release may add keys. `size` (FR-610), `colour` (FR-611), `skippedUpdate` (FR-509, the
release the user chose to skip), `dateFormat` (FR-612), `pinned` (FR-613), `lastEdge` (FR-411),
`sunMap` (FR-901), `pullOut` (FR-903), `opacity` (FR-622) and `scale` (FR-623) came after the first
release and are written after the first release's keys (`internal/infrastructure/store/decode.go`);
a file without them reads as the large size, Classic, nothing skipped, the date in words day first,
pinned, no edge stood against yet, the sun map off and not pulled out, wholly opaque and at 100
percent.
The guard is `TestA1Point0SettingsFileIsReadWhole`, which reads the frozen fixture
`internal/infrastructure/store/testdata/settings-1.0.0.json` (every key set away from its default)
and requires every key to be read rather than merely carried. It was proved by renaming a key and by
changing a stored word: each failed it. The fixture is never regenerated from a later writer, since
what it proves is that the old shape still reads.

## Colour

Every colour has one home per scheme. `frontend/src/theme.css` holds Classic, light and dark;
`frontend/src/colours.css` holds the other schemes (FR-611), keyed off the `data-colour` attribute
the page sets from the snapshot. Each of those schemes' tokens is stated once as
`light-dark(light, dark)`, so the `color-scheme` the theme sets picks the side and no dark value is
written twice; Neon's glow is itself a `light-dark()` token, transparent on the light side. A
scheme's hue lives in the tokens the ribbon paints (surface, cell, divider and both texts), because
the accent reaches only Settings: an Ocean whose only sea colour was its accent measured barely
apart from Classic. The side each scheme
resolves to was measured in Edge under Light, Dark and System on 2026-09-28, before Amendment 14.

## The desktop

On Windows, `desktop` owns a hidden top-level window on its own locked thread: the notification-area
icon, the native menus and the desktop's broadcasts. A message-only window would not hear the
broadcasts. The icon is read out of the executable itself. When Explorer restarts it re-adds the
icon on the `TaskbarCreated` message. Nothing crosses the thread boundary by callback: the desktop
reports on a buffered channel, dropping an event with a line in the log rather than blocking the
thread Windows called in on; the facade's `listen` loop acts on it. Both the window procedure and
the listen loop recover a panic and log it, so one fault cannot leave a ribbon that reacts to
nothing.

Both menus are native popup menus, so the ribbon's small window never clips them. Their items and
words have one home, `internal/application/menus.go` with its submenus of choices in
`menu_choices.go` beside it. The tray menu offers Show ribbon or Hide ribbon (whichever applies),
Add clock, Settings, Style, Colour, Orientation, Position, Always on top, Pin ribbon, Sun map, Help
and Exit; the ribbon's right-click menu offers Add clock, Settings, Style, Colour, Orientation,
Position, Always on top, Pin ribbon, Sun map, Help, Hide ribbon and Exit. Style, Colour and
Orientation are submenus ticking the current choice, whose items reach the facade's own `SetStyle`,
`SetColour` and `SetOrientation` (FR-502). Settings offers every one of those choices too, from the
same items: `Service.SettingsChoices` answers them, the snapshot carries them and the page hands the
chosen item's action to the facade's `Choose` (`choices.go`), which refuses any action not among
them and otherwise does exactly what the menu item does (FR-624). What stays on the menus alone is
commands. `TestEveryMenuChoiceIsOfferedBySettings` fails for a menu item that is neither a command
nor offered by Settings, so a choice added to the menus cannot be missed. Position is a submenu holding the two
edges the ribbon runs along (FR-408); Help is a submenu holding About, Licence and Check for updates
in both. On Windows a left click on the tray icon shows or hides the ribbon; on Linux the tray
host's activation does the same (a double click on Ubuntu); on macOS a click opens the menu, as
every menu bar icon does. A menu item may hold children, which become a submenu (Style, Colour,
Orientation, Position, then the Help submenu of FR-508); identifiers are numbered depth first
(`desktop/menu.go`, shared by every platform), so a choice inside a submenu still names its action.
A tray icon that cannot be created is not fatal: the ribbon still runs. Closing it then quits,
since nothing would bring it back.

## The desktop on Linux and macOS

Both reach the desktop through cgo: GTK 3 on Linux, AppKit on macOS. What does not depend on the
toolkit is written once, in `_unix.go` files and in `desktop/clockwatch.go`, built everywhere but
Windows: the `Desktop` itself (its events, the move-end settling, the clock watch, starting and
stopping the tray); the registry of native windows handed
out as `desktop.Window`; the callbacks the C and Objective-C halves reach; the browser opener; the
file behind the sign-in entry; the root package's icon and signal handling. Each toolkit supplies
the rest in its own files.

**One thread.** Each toolkit may only be called from its own loop, which Wails runs on the main
thread while the facade's calls arrive on other goroutines. `gtkmain.Do` (through
`g_main_context_invoke`) and `cocoamain.Do` (through the main dispatch queue) run a function there
and wait, running it at once when already there; a panic in it is raised again on the caller.
Their tests run the loop themselves (`gtkmain.ServeTests`, `cocoamain.Serve`), as Wails does.

**Finding and hiding the ribbon (FR-101).** There is no window class, so the ribbon is the process's
top-level window titled with the product's name, as Wails creates it. On Linux the window is marked
to skip the taskbar and the workspace switcher. On macOS the application becomes an accessory, with
no Dock icon and no place in the application switcher; Wails makes it a regular application as it
finishes launching, so the change is made afterwards, from `startup` through the main queue, which
AppKit serves only once launching has finished (read in Wails v2.12.0; measured by `lsappinfo`
reporting `UIElement`).

**Coordinates (FR-403 to FR-407).** Both platforms count in DIP: GTK's logical units on Linux,
points on macOS, which the toolkit scales for the display. Every display is therefore reported at
`placement.BaseDPI`. AppKit counts upward from the bottom-left corner of the menu-bar display, so
`monitors` and `desktop` turn every rectangle over against that display's height into the top-left
reckoning the domain uses. Wails' own position calls are not used on either, for the reasons of
CON-7.

**Placing the ribbon.** On macOS one `setFrame` moves and sizes it. On Linux the size is set and
awaited before the move (`awaitSize`, up to 500 ms): the window manager keeps a window on screen by
the size it has when the move arrives, so a move sent before a shrink landed was clamped as if the
window were still large (measured 2026-09-28, `TestTheRibbonReturnsFromAPanelToWhereItIsPlaced`).
On Linux GTK is sent through X11 (`gtkmain.ForceX11`), since a window on Wayland may not choose where
it stands. The web view's DMABUF renderer is turned off before it starts (`gtkmain.AvoidDMABUF`),
since on NVIDIA's own driver it draws nothing but the window's background (measured 2026-10-02); a
value already in `WEBKIT_DISABLE_DMABUF_RENDERER` is kept (`TestAChosenDMABUFSettingIsKept`).
WebKit's hardware acceleration is off too (`WebviewGpuPolicyNever` in `launch.go`). Wails turns it
off unless Linux options are given, as the workaround for its blank windows; giving them left the
policy at its zero value, Always. The ribbon still draws see-through at a lowered opacity
without it (checked on Plasma, 2026-10-02).

**The end of a move (FR-404).** Neither platform says when the button is let go, so a move ends when
the ribbon has stood still for 300 ms (`moveSettle`), heard through GTK's `configure-event` or
AppKit's `NSWindowDidMoveNotification`. A position `Place` put the ribbon at is never a move and
cancels a wait already begun, since the window may stand elsewhere for a moment on its way there
(`TestAPassingPositionOnTheWayToAPlacementIsNotAMove`). Nothing holds the ribbon on a display during
the drag (`KeepOnDisplays` does nothing): the end of the move takes the same path as on Windows,
where `Service.Moved` brings a ribbon left partly off every display back. A change of displays is
heard through GDK's `monitors-changed` or AppKit's
`NSApplicationDidChangeScreenParametersNotification`.

**The menus (FR-108, FR-502).** The right-click menu is a native popup at the pointer: a GTK menu
handed a made-up button press stamped with the X server's own time, without which GTK closed it on
the release of the click that opened it (measured 2026-09-28); an `NSMenu` started once the current
work returns, since it runs its own loop until it closes. The Linux tray icon is TimeRibbon's own
StatusNotifierItem with a `com.canonical.dbusmenu` menu over godbus, registered with the tray host by
object path, which needs no bus name of its own and so no sandbox permission; it registers again
whenever a tray host takes over. The macOS icon is an `NSStatusItem` whose menu is rebuilt each
time it opens. Both menus come from the same `internal/application/menus.go` as on Windows.

**Starting at sign-in (FR-605).** Linux writes an XDG autostart entry named for the app id; under a
Flatpak it goes to the real `~/.config/autostart`, since the session never reads the sandbox's own
`XDG_CONFIG_HOME`; its command is then `flatpak run` with the app id. macOS writes a launchd agent to
`~/Library/LaunchAgents`, limited to a desktop sign-in. Each is off until turned on and is removed,
never switched off, when turned off.

## Help, About and Licence

About and Licence are panels of the one window (FR-607, FR-608). About shows the application icon,
the name and version, the author, the copyright line and a credit for every component this
platform's build ships, each naming its licence and what it does here. The credits are one table in
`internal/product/credits.go`, each entry naming the platforms that ship it; `CreditsFor` reads it
for a platform. `TestEveryLinkedModuleIsCredited` asks the Go tool which modules each platform's
build links (the application and the setup program on Windows, the application alone on Linux and
macOS) and holds that platform's credits to that list in both directions. Licence shows the
`LICENSE` file embedded in the binary exactly as written: its own line breaks are kept and nothing
wraps it again. Its type is sized so the widest
line fits the panel, 13px at most (`frontend/src/help.css`); the width `help.css` sizes for is held
to the file's widest line by `TestTheLicencePanelIsSizedForTheLicencesWidestLine`.

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

## The update check

The house update check, ported from PigeonPost (FR-509). `internal/infrastructure/update` asks
GitHub's `releases/latest` endpoint, which answers only a published release that is neither a draft
nor a prerelease, so a tag pushed during development can never prompt; the guard is the endpoint's
own contract. It is unauthenticated, bounded by a 5 second timeout, never retried and never reads
more than a megabyte of the answer. The service compares the release's tag with the version the
build stamped into `internal/product`, as dotted integers; anything else is never newer. It picks
this platform's asset by its ending and reads the skipped release from the settings, which a manual
check ignores. Both addresses are handed to the desktop to open, so the adapter takes only `https`
addresses on `github.com` with no user or port (`onGitHub`): a page elsewhere refuses the release and
a download elsewhere is left out, the page standing in for it.

`updates.go` in the facade owns the timing: a goroutine started with the window checks 3 seconds in,
then every 24 hours, until the run ends; Help's `Check for updates` runs one more on a goroutine of
its own. Each check recovers from a panic into the log. A check with something to say shows the
ribbon and sends `open-panel` with the word `update` and the outcome, which the page draws as a
fourth panel in the frame About and Licence share. The addresses and the version stay in Go: the
page's Download and Skip ask Go to act on what it offered, so no address crosses from the page.
`TestOnlyTheUpdateCheckImportsANetworkPackage` holds Go's network imports to this one package and
`requests_test.go` the other ways out (see [Invariant](#invariant)).

## Delivery on macOS and Linux

Neither builds with the `wails` command: each runs `go build` with Wails' `desktop,production` tags,
passing the version from `VERSION` through `-ldflags` as `build.ps1` passes it to `wails build`,
then packages the result. `tools/identity` hands both scripts the product's names.

- **macOS, `builddmg.sh`** (ported from PigeonPost's): builds the page and the executable for Apple
  Silicon, makes the icon with `sips` and `iconutil`, assembles `TimeRibbon.app` with its
  `Info.plist`, signs it with the hardened runtime, notarises and staples it, then does the same for
  the DMG `create-dmg` makes. The oldest macOS it claims is the one the Go toolchain needs, read from
  a pure Go program it builds, then handed to the compiler through the cgo flags; a build that links
  code made for a newer macOS is refused (measured 2026-09-28: with the target set only through
  `MACOSX_DEPLOYMENT_TARGET`, objects built for macOS 26 were linked into an executable claiming 12).
- **Linux, `build_flatpak.sh`** (ported from PigeonPost's): writes the desktop entry, metainfo and
  manifest, then builds inside the GNOME 50 runtime's sandbox with the golang and node22 SDK
  extensions, against WebKitGTK 4.1 (`-tags webkit2_41`). The sandbox is granted these alone: X11
  with its shared memory (`--share=ipc`) and not Wayland; the GPU (`--device=dri`); the tray host's
  bus name; the bus name of Wails' single-instance lock; the session's autostart folder; the
  network, for the update check alone (FR-509), without which every check would report GitHub out
  of reach. No other part of the file system. `cleanup_flatpak.sh` uninstalls it and
  removes its sign-in entry and build outputs, leaving the settings alone. Each script first stops
  a copy left running, which holds the single-instance lock: a new build's first launch would
  otherwise only show or hide the old ribbon (FR-506).

## The setup program

Windows only. Delivery is a second Wails application. `installer/` is its own `main` package in the
same module, embedding the built application as a zip and the setup page as assets, so one file is
the whole distribution. `build.ps1` packs the built application and `LICENSE` into
`installer/payload.zip` through `tools/payload`, builds the setup program with the version from
`VERSION`, then writes the empty placeholder zip back whether or not that build succeeded, so the
real payload never reaches a commit. The payload is embedded as a string rather than a byte slice,
which Go keeps in the read-only image rather than charging to the process.

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
Every place written is per user: the files under `%LOCALAPPDATA%\Programs\TimeRibbon`, the Start Menu
shortcut under `%APPDATA%`, the Desktop shortcut on the user's own Desktop, the record and Start with
Windows under `HKCU`. Windows never asks for administrator rights (FR-810). Uninstall removes the
shortcuts, Start with Windows and the record. It deletes `%APPDATA%\TimeRibbon` (the settings, the log
and the web view's data) only when **Also forget my settings** is ticked, then hands the install
folder to a hidden PowerShell that deletes it once setup has exited.

Setup refuses to write while the application runs and offers to close it, waiting up to 5 seconds
for it to go (FR-807). **The setup page names nothing**: it has no build step, so nothing compiles or
type checks it. The product's name arrives on the state it is handed;
`TestTheSetupPageNeverWritesTheProductsName` holds that. Its keyboard ring (`setup-ring.js`) is the
window's model written again for a page that cannot import it; `frontend/src/setupRing.test.ts`
loads the shipped script and presses keys against it. Each screen opens with nothing focused
(`startNeutral` in `setup-shell.js`, FR-809): the first Tab or Right enters the ring at its first
control and the first Shift+Tab or Left at its last. The window is painted the page's own ground
before the page loads, read from the page's stylesheet, so it never flashes the wrong colour. Setup's
own web view data and step log sit under the temporary folder.

## Data locations

| What | Where |
|---|---|
| Settings | `settings.json` in the settings folder: `%APPDATA%\TimeRibbon` on Windows, `~/Library/Application Support/TimeRibbon` on macOS, `~/.var/app/uk.codecrafter.TimeRibbon/config/TimeRibbon` for the Flatpak (measured 2026-09-28) and `~/.config/TimeRibbon` for a Linux build run outside it; `settings.unreadable.json` beside it when a damaged file was kept aside, then `settings.unreadable-2.json` and onwards for each later one |
| Run log | `TimeRibbon.log` in the settings folder, started afresh by a run that finds it over 1 MiB |
| The window's web view data on Windows | `%APPDATA%\TimeRibbon\WebView2`, named in `launch.go` inside the settings folder so uninstalling with **Also forget my settings** removes it; nothing of TimeRibbon's own is kept there |
| Time zone rules and the place catalogue | built into the executable; macOS and Linux read their own zone files first |
| Installed files | Windows: `%LOCALAPPDATA%\Programs\TimeRibbon`, with `uninstall.exe`. macOS: wherever the user drags `TimeRibbon.app`. Linux: the user's Flatpak installation |
| Shortcuts on Windows | the user's Start Menu Programs folder and Desktop |
| Start at sign-in | Windows: the value `TimeRibbon` under `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, holding the quoted path and no arguments. macOS: `~/Library/LaunchAgents/uk.codecrafter.TimeRibbon.plist`. Linux: `~/.config/autostart/uk.codecrafter.TimeRibbon.desktop`; `$XDG_CONFIG_HOME/autostart` holds it instead for a build run outside the Flatpak with that set |
| Apps list record on Windows | `HKCU\...\Uninstall\TimeRibbon` |
| Setup's step log and web view data on Windows | `TimeRibbonSetup.log` and `TimeRibbonSetup` in the temporary folder |

## Errors

Errors are wrapped with context at each boundary using `%w`, so `errors.Is` still finds a sentinel
beneath. The adapters read a missing file or registry value as absence that way; the tests tell the
module's own refusals (such as `ErrNoSuchClock` or `ErrNoMonitors`) apart by it.

- **Before the window, a run is ended only by a failure to run the window at all or to read the
  embedded place catalogue.** Standard error is pointed at the log as the first act of the run
  (`runlog.Keep`), so even the Go runtime's own panic report is kept. A settings folder that cannot
  be found falls back to a folder in the temporary folder; settings that cannot be read, a tray icon
  that cannot be made and a missing executable path are logged and the ribbon still opens. Settings
  that are there but cannot be read are never saved over that run ([The settings file](#the-settings-file)). A
  catalogue that fails to parse is a build defect, which a test holds against.
- **Shown on the ribbon, which keeps working:** a settings file kept aside and a save that failed, as
  notices with OK; an invalid clock, in words in its own cell.
- **Refused beneath the control that was pressed:** every page call that Go can refuse. Each `api`
  wrapper takes a refusal handler as its last argument and answers null rather than rejecting, so a
  call written without a handler does not compile and no refusal is dropped on the page.
- **Logged and carried on:** a desktop event nobody was reading, a panic in the desktop's thread or in
  handling one of its events, a failure to place, hide from the taskbar or fence the ribbon.

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
- The Linux and macOS code builds only with cgo against its toolkit, so it is checked on a machine of
  its own platform (gofmt, vet, staticcheck, the internal tests with the build's tags, a build of the
  executable) rather than by `test.ps1`; [TESTING.md](TESTING.md) says how. The structural tests run
  on every platform; the credits test lists every platform's modules from any one of them.

## Design decisions

The architectural decisions, with the evidence behind each. The whole set the product rests on,
with what each one costs, is in [DECISIONS-TRADEOFFS.md](DECISIONS-TRADEOFFS.md).

| Decision | Why | Rejected alternative |
|---|---|---|
| Go with Wails and a web front end | One executable with no runtime to install; the same stack draws the setup program | A Python and Qt desktop stack |
| The tz database built into the executable | Windows has no zone files Go reads, so without it Windows would find no rules outside a Go installation; macOS and Linux fall back on it for a zone their own files lack; no DST rule is written by hand | Relying on the machine's zone files alone; offsets written by hand |
| Displays and placement through Win32 | Wails' screen list has no origin, device name or work area; its position calls mix relative and absolute coordinates | Wails' own position calls |
| One window for the ribbon and every panel | Wails v2 offers one window | A second window per panel |
| Native popup menus | The ribbon's window is small; a menu drawn in the page would be clipped by it | A menu drawn in the page |
| The scroll bar measured by the page | It is the web engine's bar, which Windows' scroll bar metric does not describe | A thickness written into the Go code |
| The self-reading cycle in one script beside the setup page | The setup page cannot import; the window's build can import the file, so both run the same code | The cycle written twice, once in TypeScript and once for the setup page |
| The ribbon hides from the taskbar by swapping its window style | Wails always marks its window as an application window | Accepting a taskbar button |
| The web view's data folder named inside the settings folder | Left to Wails it was `%APPDATA%\TimeRibbon.exe`, beside the settings folder, which forgetting the settings on uninstall did not reach | Deleting Wails' default folder by name at uninstall, which hangs on a rule Wails does not promise |
| Settings in one JSON file, written whole | A person can read and repair it; a crash mid-write cannot damage it | A database |
| Everything per user | Nothing needs administrator rights, so nothing asks for them | A machine-wide install |
| GTK and AppKit reached directly through cgo | Wails' screen list and position calls fall short on every platform (CON-7); the desktop's own calls do not | Wails' position calls; a cross-platform window library over them |
| Linux forced onto X11 | A window on Wayland may not choose where it stands, which the ribbon must (ruled 2026-09-28) | Wayland, with the compositor placing the ribbon |
| WebKit's DMABUF renderer off on Linux | On NVIDIA's own driver it drew only the window's background (measured on an RTX 3080 Ti, 2026-10-02); the clocks redraw once a minute, so the faster path buys nothing here | Leaving it on and telling NVIDIA users to set the variable; setting it in the Flatpak manifest alone, which a build run outside it would miss |
| WebKit's hardware acceleration off on Linux | It is what Wails itself chooses as the workaround for its blank windows; giving Linux options had left the policy at Always by accident | The zero value, Always, on every Linux machine |
| The first showing waits for the page to size the window | Shown before it was sized, the window grew while visible and at a fractional KDE scale often stayed cut off (8 of 16 launches, 2026-10-02) | Showing it as soon as the page is ready |
| TimeRibbon's own StatusNotifierItem on Linux | `fyne.io/systray` offered no way to rebuild the menu as it opens and kept its state global (measured in v1.12.2) | `fyne.io/systray`; no tray icon off Windows |
| One code path for Linux and macOS where the toolkit does not matter | Written once in `_unix.go` files, the move-end settling, the events and the sign-in file cannot drift apart | A copy per platform |

See also [TESTING.md](TESTING.md) for the test suite and [DEVELOPMENT.md](DEVELOPMENT.md) for
building from source.
