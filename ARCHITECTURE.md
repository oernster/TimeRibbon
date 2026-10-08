# TimeRibbon Architecture

A desktop application for Windows, macOS and Linux showing a ribbon of clocks, one per chosen place.
It reads the system clock, its settings file, the desktop (displays, tray, sign-in entry; on Linux
logind's notice of a shutdown) and GitHub's latest release. The time zone rules are built in; macOS and Linux read their own zone files first
([Time](#time)). The domain and application are the same code everywhere; each platform's own half
sits in files its build tags or names select.

The ribbon itself (placing, dragging, the tab, the pull out, the grip, opacity, the tray and menus,
sign-in start, the update check, setup) is [ribbonkit](https://github.com/oernster/ribbonkit), a Go
module and npm package TimeRibbon depends on at one tag, shared with WeatherRibbon. Its own
ARCHITECTURE.md describes how a ribbon works; this document describes what TimeRibbon puts in it and
how the two are joined.

Its one network request is the update check (FR-509), which is the kit's; TimeRibbon hands it
`product.Repository`. No file of TimeRibbon's imports a network package, starts a program or asks a
network from its page. The donation page and a release's download go to the desktop's browser.

FR, NFR and CON numbers are those of [REQUIREMENTS.md](REQUIREMENTS.md).

## Invariant

`UI -> Application -> Domain <- Infrastructure`

Dependencies point inward. Every rule below is a test; a guard not listed here does not exist. The
structural tests run on the kit's `structure` package, the code that holds the kit to the same rules;
the kit's own invariants are listed in its ARCHITECTURE.md.

| Invariant | Enforcing test | File |
|---|---|---|
| Domain imports nothing of TimeRibbon or the kit outside a domain | `TestDomainHasNoOutwardImports` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Domain is pure: no network, filesystem, process, random or tz package; no wall clock read, no zone loaded (FR-207, CON-5) | `TestDomainIsPure` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Application never imports infrastructure (TimeRibbon's or the kit's) or Wails | `TestApplicationDoesNotImportInfrastructure` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Neither application nor infrastructure imports a UI package | `TestNothingBelowTheUIImportsIt` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Infrastructure never imports Wails | `TestWailsStaysOutOfInfrastructure` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Only `main.go` imports both application and infrastructure, the kit's included | `TestCompositionRootIsWhitelisted` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| No Go file and no TypeScript or CSS file under `frontend/src` exceeds 400 lines | `TestNoFileExceedsLineLimit` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| No such file sits in the danger band of 381 to 400 lines | `TestNoFileInDangerBand` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Every exported type carries a doc comment | `TestEveryExportedTypeIsDocumented` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| `go.mod` and the front end's `package.json` name the same kit tag | `TestBothHalvesOfTheKitNameOneTag` | [`kittag_test.go`](tests/structural/kittag_test.go) |
| No Go file of TimeRibbon's imports `net`, `crypto/tls` or `golang.org/x/net` (NFR-S-1) | `TestNothingOfTimeRibbonsImportsANetworkPackage` | [`network_test.go`](tests/structural/network_test.go) |
| No Go file of TimeRibbon's starts a program or names a Windows library outside the kit's `SystemLibraries` (NFR-S-1) | `TestNothingOfTimeRibbonsStartsAProcess` | [`network_test.go`](tests/structural/network_test.go) |
| The page uses no request API and names no web address, the SVG namespace aside (NFR-S-1) | `TestThePageMakesNoRequest` | [`network_test.go`](tests/structural/network_test.go) |
| The kit's setup page never spells the product's name | `TestTheSetupPageNeverWritesTheProductsName` | [`setup_test.go`](tests/structural/setup_test.go) |
| Each platform's About credits exactly the third-party modules its build links (FR-607) | `TestEveryLinkedModuleIsCredited` | [`credits_test.go`](tests/structural/credits_test.go) |
| No platform credits a module twice | `TestAModuleIsCreditedOncePerPlatform` | [`credits_test.go`](tests/structural/credits_test.go) |
| TimeRibbon's half of the wire is stated alike in `dto.go` and `frontend/src/wire.ts` | `TestTheWireIsStatedAlikeOnBothSides` | [`wire_test.go`](tests/structural/wire_test.go) |
| The page names the word `app.go` sends to open Add clock | `TestThePageNamesEveryEventAppEmits` | [`wire_test.go`](tests/structural/wire_test.go) |
| Every method the page's `Bridge` calls is bound on `App`; none of the window's `Control` is | `TestEveryMethodThePageCallsIsBound`, `TestNothingOfTheControlIsBound` | [`page_api_test.go`](page_api_test.go) |
| TimeRibbon's setup carries every picture the kit's setup page shows | `TestTimeRibbonCarriesEveryPictureTheSetupPageShows` | [`main_test.go`](installer/main_test.go) |
| Each `wails.json` names its executable as `internal/product` does | `TestEachWailsConfigNamesItsExecutableAsTheProductDoes` | [`names_test.go`](tests/structural/names_test.go) |
| In `dials.css` every offered scheme has a block stating each of Classic's tokens (the problem colour aside); every block is offered (FR-611) | `TestEveryOfferedSchemeHasItsOwnCompleteBlock` | [`colours_test.go`](tests/structural/colours_test.go) |
| `dials.css` states Classic's dark the same under the system's dark mode as under a chosen dark theme | `TestClassicDarkIsTheSameUnderTheSystemAsWhenChosen` | [`colours_test.go`](tests/structural/colours_test.go) |
| Text, muted text and problem text meet 4.5:1 on the cell and the surface, every scheme, both themes, over the kit's palette and `dials.css` together (NFR-U-1) | `TestTextMeetsTheContrastFloorOnEverySchemeAndTheme` | [`contrast_test.go`](tests/structural/contrast_test.go) |
| The Licence panel is sized for the LICENSE's widest line | `TestTheLicencePanelIsSizedForTheLicencesWidestLine` | [`licence_test.go`](tests/structural/licence_test.go) |
| The settings file of the first release is read whole (NFR-C-1) | `TestA1Point0SettingsFileIsReadWhole` | [`contract_test.go`](internal/infrastructure/store/contract_test.go) |
| No tracked or new file holds the product's former name or the word its window went by before the ribbon, the npm lock file and Python's string method of that name aside | `TestNoTrackedFileHoldsTheRetiredWord` | [`retired_test.go`](tests/structural/retired_test.go) |
| That word is found in any case and inside names while current names pass | `TestTheRetiredWordIsFoundInAnyCaseAndInsideNames` | [`retired_test.go`](tests/structural/retired_test.go) |

The structural tests that read the kit's files (the palette, the Licence panel's width, the setup
page) read the kit Go builds against, through `go list -m`; `page_api_test.go` reads the kit's
`bridge.ts` as npm installed it, since that is what the page compiles.

## Layers

- **Domain** (`internal/domain`), pure Go: time arrives as an argument and a zone already resolved.
  - `clock`: an instant and a zone become what a cell shows (the time and the zone mark as the
    kit's `localtime` writes them, the date in the chosen `DateFormat`, the hand angles); a label's
    default from its zone, capped by the kit's `ribbon.Label`; `Samples` writes every time and date a
    cell can show, for the page to measure (FR-620). The next minute boundary is `localtime`'s too.
  - `settings`: the user's choices as one value, every operation answering a new one: the kit's
    `ribbon.Choices` embedded, so they read as its own fields, then the clocks' style, size, formats,
    sun map, pull out and the clocks themselves. The settings file is written exactly as before the
    split (NFR-C-1).
  - `sun`: the subsolar point from NOAA's equations (FR-906), checked against NOAA's values in
    `testdata`; `Meridian`, the longitude a UTC offset keeps, for ordering a zone with no city
    (FR-102).
- **Application** (`internal/application`): one `Service` over its ports in `ports.go` (`Store`,
  `Zones`, `Clock`, `IDs`, the kit's `arranger.Monitors` and `arranger.Neighbours`, `StartupEntry`
  and the kit's `release.Source`). It builds the snapshot, edits clocks, searches places
  (`SearchPlaces`, FR-302), changes settings, takes the page's measurements, checks for updates
  through the kit's `release.Check` with the release the user skipped and answers the menus. It
  embeds the kit's `arranger.Arranger`, so arranging the ribbon is the service's own method set. The
  arranger asks its `Host` for the ribbon's choices and content read together; the service answers
  through `host.go` (a cell per notice and per clock, the style's or the prompt's size, the sun map's
  handle lane) and saves the arranger's changes through its one save path, so they raise the same
  notice (FR-707). The snapshot orders cells west to east (`westToEast`) by the longitude of each
  clock's city in the catalogue, the same city the sun map marks; a zone with no city stands at the
  meridian its offset keeps (`sun.Meridian`). Ties keep their stored order and an unshowable clock
  goes last (FR-102). A change that cannot be saved stays in effect with a notice (FR-707).
- **Infrastructure** (`internal/infrastructure`): `store` (the settings file) and `zones` (the place
  catalogue and the zone rules). Everything else it reaches (displays, sign-in, the log, the data
  folder, the desktop, the shared ribbon folder, the update check) is the kit's.
- **UI**: the React front end in `frontend/src` over the kit's page half; the Wails facade in
  package `main`, which embeds the kit's window and maps the service's answers about the clocks into
  `dto.go`.
- **Outside the layers**: `internal/product` holds the name, app id, repository, setup program's name,
  window class, donation address, version, author, copyright line, sign-in label and credits. The
  domain and application never read it. `installer/main.go` is the setup program's composition root
  over the kit's setup window.
- **Tools**, never shipped: `genplaces` (the place catalogue) and `genicons.py` (every committed
  icon); `payload` (the setup program's payload), `versioninfo` (each executable's version resource),
  `identity` (names for the Linux and macOS scripts) and `linuxicons` (the Flatpak's icons) are mains
  handing `internal/product` to the kit's `delivery`, which does the work.

## Composition root

`main.go` points standard error at the run log before anything can fail, builds the adapters (the
kit's among them), injects them into the service, prepares the platform, starts the tray and hands
the facade to Wails. What it does for the platform is the kit's `platform` package: on Linux and
macOS `platform.Prepare` hands the desktop the icon and ends the run on SIGTERM or SIGINT; on Linux
it also ends the run when logind announces a shutdown or restart (FR-507). On Linux importing it
sends GTK through X11 and turns off the DMABUF renderer. TimeRibbon says only where its
tray icon lies (`trayicon_unix.go`; none on Windows, whose tray reads the executable's). The cell
sizes (`layouts`) and panel sizes (`panels`) live there.
No service is held in a global.

The facade Wails binds is `App` in `app.go`, in two halves. The window is the kit's `window.Window`,
embedded, so every exported method of the window is page API. What is TimeRibbon's own stays in
`app.go` (the clocks, their style, size and formats, the sun map, the Snapshot) and `measure.go`,
over the `ribbonService` interface. TimeRibbon reaches its window through the kit's `window.Control`,
a named field that is never embedded and so never bound; menu actions the kit does not know reach
TimeRibbon through the `Act` hook it hands the window. `kit.go` embeds the page and the LICENSE and
adapts the service to the window's port; `dto.go` is TimeRibbon's half of the wire.
The kit's `platform.GeneratingBindings` keeps the binding-generation run from writing the log or
showing a tray icon.

```
             +-----------------------------------+
   Wails/UI  | app.go (App) embeds window.Window |
             +--------------+--------------------+
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

The page half joins the same way: `frontend/src/api.ts` extends the kit's bridge with TimeRibbon's
own calls over the kit's guarded call; `App.tsx` is the kit's shell (`useShell`) around the clocks;
`Ribbon.tsx` draws them in the kit's `Band`; `Surface.tsx` pulls the sun map out beside it through
the kit's `PullOut`. The front end extends the kit's `tsconfig.json`, uses its eslint rules and runs
its suites under its test set-up.

## The window and the ribbon

How the window opens hidden, places itself, collapses to its tab, scales under the grip, fades with
opacity and shares itself with panels and the pull out is the kit's (its ARCHITECTURE.md, "One
window" and "Place and drag"). What TimeRibbon decides:

- **Panels.** Settings opens 900 DIP wide (FR-625), the others 560 (`panels` in `main.go`).
- **Opacity (FR-622).** `app.css` mixes `--window-opacity` into the surface, the dial's face and the
  tab's accent, so only backgrounds fade and the clocks stay solid.
- **Size (FR-105, FR-106, FR-620).** The ribbon is sized from its notices and clocks (the Add clock
  prompt when there are none): cells plus padding along the orientation up to the work area, beyond
  which they scroll; one cell across, plus the scroll bar when scrolling and the handle's lane while
  the sun map is on. The page measures the widest time and date its font draws (`measure.ts`,
  `SetMeasured`); `layoutFor` widens the size's cell to that width, so the cells drawn and the window
  sized cannot disagree.
- **Refits.** Every change that can alter the cells refits the ribbon where it stands; where the
  length changed, the ribbon is centred along it with its position across kept and that place stored
  (FR-104).
- **The sun map** is the pull out. The page lays it out (`Surface.tsx`) at the boxes Go sends, blends
  day and night by solar altitude (`sunLight.ts`) and stands labels clear (`labels.ts`, FR-914).

## Time

Zones resolve through the kit's `zones.Resolver`: `time.LoadLocation` with `time/tzdata` built in
(CON-5), each zone loaded once; an empty id and `Local` are refused.
Windows has no zone files Go reads, so there the built-in rules are used; macOS and Linux read their
own first. The place catalogue, `internal/infrastructure/zones/places.tsv`, comes from tz 2025b's
`zone.tab` and `iso3166.tab` through `tools/genplaces` and holds 418 zones, each held to resolving by
`TestEveryPlaceResolvesInTheEmbeddedDatabase`.

Each snapshot carries the time to the next minute and the page takes the next one then (FR-208). A
change of the system clock or a resume reaches the page through the kit's desktop (FR-209).

## The settings file

`settings.json` in the settings folder ([Data locations](#data-locations)) is indented JSON with a
format version (FR-701); derived values are never stored. `store` says what the file holds (its keys
in writing order and its clocks, `store/codec.go`); what every ribbon's file does alike is the kit's
`settingsfile`, proved by its own tests. A save writes a temporary file in the same folder, flushes it
and renames it over the old one (FR-702). Reading is tolerant:

- No file means the defaults and no notice (FR-703).
- A file that is there but cannot be read raises a notice and refuses every save that run.
- A file that is not JSON is renamed to `settings.unreadable.json` with a notice (FR-704); where that
  name is taken, to the first free of `settings.unreadable-2.json` onwards, up to the kit's
  `KeptAsideLimit`.
  Where the rename fails, saving is refused from then on.
- A UTF-8 byte order mark is passed over.
- A clock that cannot be read or names an unknown zone is kept as it was and shown in words as invalid
  (FR-705, FR-706); two clocks sharing an id are both kept, the later renumbered (`same-2`).
- An unknown top-level key is written back as found.

**The file is a contract from the first release (NFR-C-1).** No key the first release writes is
renamed, dropped or given another meaning; no stored word changes. Later keys (`size`, `colour`,
`skippedUpdate`, `dateFormat`, `pinned`, `lastEdge`, `sunMap`, `pullOut`, `opacity`, `scale`,
`pullOutSide`) are written after the first release's (`store/codec.go`); a file without them reads
as the defaults. `TestA1Point0SettingsFileIsReadWhole` reads the frozen fixture
`store/testdata/settings-1.0.0.json`, every key set away from its default; it was proved by renaming
a key and by changing a stored word. The fixture is never regenerated.

## Colour

Every colour has one home per scheme, in two halves keyed off the same theme and `data-colour`
(FR-611). The kit's half holds the ribbon's own tokens (surface, cell, divider, the texts, accent,
problem, focus and Neon's glow): Classic in its `theme.css`, the others in its `colours.css`.
TimeRibbon's half, `frontend/src/dials.css`, holds its content's: the analogue dials for every
scheme, the sun map's marks and Neon's glow on the digits and hands. Each half states a scheme's
token once as `light-dark(light, dark)`; no token is stated in both.

## Menus

Both menus are the kit's native popups. Their items have one home,
`internal/application/menus.go` and `menu_choices.go`, built from the kit's `menus` package (the
`Item` type and the actions every ribbon offers). Settings offers every menu choice from the same
items: `Service.SettingsChoices` answers them; `TestEveryMenuChoiceIsOfferedBySettings` fails for a
menu choice Settings lacks. The tray menu (in `main.go`) and the Settings choices (in the snapshot)
pass through the kit's `Control.Offered` as the right-click menu does, so a Position item that would
leave the ribbon where it stands is greyed in all three (FR-408). The window carries out every
ribbon's actions itself and hands any other, such as Add clock, Style or Sun map, to TimeRibbon's
`actOn`.

## Help, About and Licence

The three panels are the kit's. About shows TimeRibbon's picture, name and version, author,
copyright line and a credit for every third-party component this platform's build ships (FR-607),
from one table in `internal/product/credits.go`; the kit is TimeRibbon's author's own and is not
credited. Licence shows the embedded `LICENSE` exactly as written (FR-608).

## Delivery

**macOS and Linux** build with `go build` and Wails' `desktop,production` tags, the version from
`VERSION` through `-ldflags`, names from `tools/identity`. Every build script sets `GOWORK=off`, so
what ships is the kit tag `go.mod` requires.

- `builddmg.sh` builds for Apple Silicon, assembles and signs `TimeRibbon.app` with the hardened
  runtime, notarises and staples it, then the DMG. The minimum macOS is read from the Go toolchain and
  passed through the cgo flags; a link of code built for a newer macOS is refused. The bundle's
  `Info.plist` declares `LSUIElement`, so the application checks in with macOS as an agent and the
  Dock never records it as a recent app (FR-101); the kit refuses Wails' later switch to a regular
  application, so both are needed.
- `build_flatpak.sh` builds in the GNOME 50 SDK against WebKitGTK 4.1 (`-tags webkit2_41`). The
  sandbox gets X11 with IPC, the GPU, the tray host's bus name, the single-instance lock's bus name,
  logind on the system bus (`org.freedesktop.login1`, so the ribbon leaves before a shutdown or
  restart, FR-507), the autostart folder, the shared ribbon folder (`xdg-run/ribbonkit`) and the
  network for the update check alone. Both scripts first stop a copy left running, which holds the single-instance lock
  (FR-506).

**The setup program** (Windows) is a second Wails application in `installer/`, embedding the built
application as a zip. `build.ps1` packs it through `tools/payload`, builds setup, then writes the
empty placeholder back whatever happened. The install policy and the setup window are the kit's;
`installer/main.go` is the composition root, its wiring the kit's `installer.Main`: it carries the
payload, names the product (a `setup.Product` from `internal/product`), binds the kit's facade by
embedding it in its own `App` and hands in the page's pictures, which `tools/genicons.py` writes to
`installer/frontend/dist`.

## Data locations

| What | Where |
|---|---|
| Settings | `settings.json`: `%APPDATA%\TimeRibbon` on Windows, `~/Library/Application Support/TimeRibbon` on macOS, `~/.var/app/uk.codecrafter.TimeRibbon/config/TimeRibbon` for the Flatpak, `$XDG_CONFIG_HOME/TimeRibbon` else `~/.config/TimeRibbon` outside it; kept-aside copies beside it |
| Run log | `TimeRibbon.log` beside the settings, started afresh over 1 MiB |
| Web view data on Windows | `%APPDATA%\TimeRibbon\WebView2`, so forgetting the settings removes it |
| Ribbons on this desktop | `uk.codecrafter.TimeRibbon.json` and `uk.codecrafter.TimeRibbon.lock` in `%LOCALAPPDATA%\ribbonkit`, `~/Library/Application Support/ribbonkit` or `$XDG_RUNTIME_DIR/ribbonkit` (FR-412) |
| Time zone rules, place catalogue | built in; macOS and Linux read their own zone files first |
| Installed files | Windows: `%LOCALAPPDATA%\Programs\TimeRibbon` with `uninstall.exe`; macOS: where the user drags the app; Linux: the user's Flatpak installation |
| Start at sign-in | Windows: `TimeRibbon` under `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, the quoted path and no arguments. macOS: `~/Library/LaunchAgents/uk.codecrafter.TimeRibbon.plist`. Linux: `~/.config/autostart/uk.codecrafter.TimeRibbon.desktop` |
| Apps list record | `HKCU\...\Uninstall\TimeRibbon` |
| Setup's log and web view data | `TimeRibbonSetup.log` and `TimeRibbonSetup` in the temporary folder |

## Errors

Errors are wrapped with `%w` at each boundary, so `errors.Is` finds sentinels such as
`ErrNoSuchClock` beneath.

- **Before the window** only a failure to run the window or to read the embedded catalogue ends the
  run. Standard error goes to the log first (the kit's `runlog.Keep`), so even a runtime panic is
  kept. A missing settings folder falls back to the temporary folder; unreadable settings, a failed
  tray icon and a missing executable path are logged and the ribbon still opens.
- **On the ribbon:** a kept-aside file and a failed save as notices; an invalid clock in words.
- **Beneath the control pressed:** every page call Go can refuse takes a refusal handler and answers
  null rather than rejecting, so a call without one does not compile.

## Quality enforcement

- The structural tests above run with the suite; the kit's run in the kit's own gate.
- `test.ps1` checks formatting, vet and staticcheck, runs the Go suite and the front end's lint, type
  check and tests, holds the domain and application to 100% and every other gated package to its
  measured floor ([TESTING.md](TESTING.md)).
- `build.ps1` runs `test.ps1` first with no switch to skip it, cgo off and `GOWORK=off`.
- The Linux and macOS code is checked on its own platform ([TESTING.md](TESTING.md#on-macos-and-linux)).

## Design decisions

The architectural ones; the full set with their costs is in
[DECISIONS-TRADEOFFS.md](DECISIONS-TRADEOFFS.md). The ribbon's own (displays through each desktop's
calls, one window, native menus, the cursor read from the desktop, the tray of its own, Linux on X11)
are recorded in the kit's ARCHITECTURE.md.

| Decision | Why | Rejected alternative |
|---|---|---|
| The ribbon's desktop half in ribbonkit, a module of its own at one tag | WeatherRibbon needs the same ribbon; a fix lands once and reaches both | A copy of the desktop code per application |
| One tag for both halves of the kit, held by a test | A page built from one release against a window from another breaks only at run time | Versioning the Go and npm halves apart |
| The page measures cells and scale | Only the page knows its font and the ratio it is drawn at | Widths written into Go |
| The web view's data inside the settings folder | Wails' default sat beside it, out of reach of forgetting the settings | Deleting Wails' folder by name |

See also [TESTING.md](TESTING.md) and [DEVELOPMENT.md](DEVELOPMENT.md).
