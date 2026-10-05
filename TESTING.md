# Testing

What is tested, what is not and why the line falls where it does. Every figure was measured by the
commands in [Running it](#running-it) when it was written; every shortfall is named with its reason.

## Before the first run on Windows

Some anti-virus programs quarantine a freshly built test binary in Go's scratch folder, which stops
the suite. If a run fails with access denied or a missing file on a test binary, allow the folder
`go env GOTMPDIR` names (the system temporary folder where it prints nothing). Install the front end's
packages once and build the page once, since the application embeds `frontend/dist`, which git does
not hold:

```powershell
npm --prefix frontend install
```

```powershell
npm --prefix frontend run build
```

Text files check out with LF endings (`.gitattributes`), since gofmt refuses CRLF; an older checkout
holding CRLF is fixed by checking it out afresh.

## The standard

**A floor is a measurement, never an aspiration.** Each floor in `test.ps1` is its package's measured
figure with the fraction dropped, so it fails once cover is lost.

**A gap is named or it is closed.** An unexplained shortfall cannot be told from an oversight.

## What the numbers are

### Go

| Package | Coverage | Floor |
|---|---|---|
| `internal/domain/clock`, `settings`, `sun`; `ribbonkit/domain/hover`, `placement`, `ribbon` | 100% | 100% |
| `internal/application`; `ribbonkit/application/arranger`, `menus`, `release` | 100% | 100% |
| `internal/infrastructure/zones`; `ribbonkit/infrastructure/appdata`, `iconscale`, `system`, `update` | 100% | 100% |
| `internal/infrastructure/store` | 94.0% | 94% |
| `ribbonkit/ui/window` | 93.8% | 93% |
| `tools/versioninfo` | 86.7% | 86% |
| `internal/infrastructure/setup` | 84.0% | 84% |
| `tools/payload` | 82.8% | 82% |
| `ribbonkit/infrastructure/monitors` | 82.6% | 82% |
| `ribbonkit/infrastructure/startup` | 80.6% | 80% |
| `ribbonkit/infrastructure/runlog` | 77.8% | 77% |
| `tools/identity` | 75% | 75% |
| `tools/linuxicons` | 67.7% | 67% |
| the root package (the Wails facade) | 66.1% | 66% |
| `tools/genplaces` | 58.6% | 58% |
| `ribbonkit/infrastructure/desktop` | 46.8% | 46% |
| `internal/product` | 100% | not gated |
| `installer` | 0%, no tests | not gated |

Every figure is the Windows build's, which `test.ps1` measures. That build compiles 419 Go test
functions, counted from the test files `go list` selects, each running once with no subtests, plus
one `TestMain` in `internal/infrastructure/setup`. Thirty-five are the structural tests, which read
the source and are the same on every platform; [ARCHITECTURE.md](ARCHITECTURE.md) lists each against
its rule. `TestA1Point0SettingsFileIsReadWhole` in `store` holds the settings file's promise
(NFR-C-1); `contrast_test.go` holds NFR-U-1 in Go because Vitest hands a CSS import back empty. The
macOS build compiles 392 and the Linux build 396 ([On macOS and Linux](#on-macos-and-linux)).

### The front end

149 tests in 24 files under Vitest with jsdom, run from `frontend`; nine of the files are ribbonkit's
own (`ribbonkit/web`), reached through the front end's link to the kit's package. They cover the
ribbon's band, its tab and the report that it has been drawn (`Band.test.tsx`, FR-614, FR-615); the
handle that pulls the sun map out (`PullOut.test.tsx`, FR-903); the shell's panels, refreshes, theme
and colour scheme (`shell.test.tsx`); the clocks, Settings, About, Licence and the update panel; an
accessible name and tooltip on every icon-only control (`a11y.test.tsx`, NFR-U-4); the sun map, its
blend and its labels (FR-914); the page's background colour; the self-reading cycle; the
`devicePixelRatio` watch; measuring a cell's widest text (FR-620); Settings fitting its content
(`panelFit.test.tsx`, FR-621); the opacity and its slider and opaque panels (`opacity.test.ts`,
`OpacitySlider.test.tsx`, `shell.test.tsx`, FR-622); the corner grip (`ScaleGrip.test.tsx`,
`scaleGrip.test.tsx`, FR-623); every timer the page schedules, the kit's included, against a reasoned
allow-list (`timers.test.ts`, NFR-P-4); the setup page's screens,
keyboard ring and unreachable-program cases. No coverage provider is installed, so no figure is
claimed.

## How each layer is tested

| Layer | Kind of test | Touches |
|---|---|---|
| `internal/domain` | pure unit over fixed instants and zones loaded with the tz database embedded | nothing |
| `internal/application` | unit over hand-written fakes of the seven ports | nothing |
| `internal/infrastructure` | integration over temporary folders and scratch registry keys | the filesystem, `HKCU` under a scratch key, child processes, the real displays |
| `ribbonkit/ui/window` | unit over a scripted service, with Wails and the desktop stood in for | nothing |
| the root package | unit over a scripted service and a stand-in window | reads `frontend/src/api.ts` and `ribbonkit/web/bridge.ts` |
| `tests/structural` | source and AST scans, a `go list` per platform, one `git ls-files` | reads files |
| the front end | component tests under jsdom over `fakeBridge.ts`, which records every call and builds on ribbonkit's own stand-in (`@oernster/ribbonkit/testing`) | nothing |

No Go test uses a mocking library. **No test writes to the user's own settings, sign-in entry or
Apps list** and **no test reaches the network**: the update adapter runs over a stand-in HTTP client,
so GitHub is asked only by the running application, which is checked by hand.

## What is not tested and why

### The platform owns it

- **`desktop` (46.8%).** The tray, native menus, move fence and broadcasts run on a hidden window's
  message loop and act on the real ribbon window. Tested: menu identifier numbering, the fence's
  arithmetic, a work area at a point, the drag distance, an address Windows refuses, the clock watch,
  the window's cut (FR-913) with the pointer read against it; the cursor read for the grip; every
  operation of the `shell.Desktop` port answering as the package's own function does, on a hidden
  window that is never shown. The loop itself, the menus as drawn, the broadcasts, a browser opening and the desktop
  showing through are checked by hand in a real build.
- **`monitors` (82.6%):** Windows refusing to enumerate or describe a display.
- **`runlog` (77.8%):** the log refusing to open, its first line failing and `SetStdHandle` refusing.
- **`ribbonkit/ui/window` (93.8%).** The window's decisions are tested across its seventeen test
  files over a scripted service: which calls refit the ribbon, panels and Settings' fit, the tab, the
  window holding and cut to the map, the menu actions and the hand-over of those the kit does not know,
  closing, a signal ending the run, the recover round each event and update check, the update watch's
  timing, the grip's drag (following the desktop's pointer where it can read it, else the page's),
  the window's paint below full opacity, the first showing, Help, the Control's calls and every call
  into the desktop going through the `shell.Desktop` port with the ribbon's window. Not reached:
  `run.go`, the one-line calls into Wails, `startup`, `listen` and `shutdown`, which only Wails runs.
- **The root package (66.1%).** TimeRibbon's own half of the facade is tested over a scripted service
  and a stand-in window: every change to the clocks fits the ribbon once, the Snapshot carries the
  window's reading, TimeRibbon's menu actions, the measurements, the product handed to the window, the
  adapter reading the ribbon's choices out of the settings; every method the page's `Bridge` calls is
  bound, with nothing of the `Control`. Not reached: the composition root (`main`, `keepLog`,
  `settingsDir`, `run`) and `preparePlatform` on Windows, which does nothing.

The ribbon's layout at full size was measured in headless Edge 154.0.4258.37 with the application's
stylesheets: at 100% and 250% a scrolling ribbon shows one scroll bar with no clock cut off.

### It would change the machine

- **`installer` (0%).** Every method reads or acts on the machine; the policy beneath it is tested in
  `setup` and the page in `setupScreens.test.ts`, `setupRing.test.ts` and `setupUnreachable.test.ts`.
- **`setup` (84.0%).** Tested over temporary folders, a scratch key and real stand-in processes. Not
  reached: the real Apps list record, deleting the install folder after setup exits, COM or a shortcut
  refusing, a copy failing part way, `TakeFocus`, finding the launched ribbon and `Places`.
- **`startup` (80.6%):** the registry refusing to open, read, write or delete.
- **`store` (94.0%):** the temporary file refusing to be made, written, flushed or closed (a failing
  disk); the kept-aside check failing straight after a read; the guarded error returns in `encode`
  and `extrasOf`. A file held open with no sharing is tested on Windows (`lock_windows_test.go`).
- **The tools:** each `main` handing `run` its real arguments; folders and archives refusing to be
  made or closed. `genplaces` (58.6%) also reads the tz database's own files, which a test machine need
  not have; its parsing and writing are tested. What `versioninfo` writes was read back from both
  released executables through Windows' version API.

## On macOS and Linux

Their halves of infrastructure compile only there; `cocoamain`, `gtkmain`, `monitors` and `desktop`
also need cgo against AppKit or GTK, so `test.ps1` reaches none of them. Check them on a machine of
that platform, set up as [DEVELOPMENT.md](DEVELOPMENT.md) says, with the page built.

| What | macOS | Linux |
|---|---|---|
| Tags | `desktop,production` | `desktop,production,webkit2_41` |
| Go test functions | 391, plus 3 `TestMain` | 395, plus 3 `TestMain` |
| Tests that need cgo | `cocoamain` 3, `monitors` 3, `desktop` 17 | `gtkmain` 5, `monitors` 2, `desktop` 20 |
| Needs | a signed-in desktop | a signed-in desktop with a display and a tray host |

With the platform's tags in `TAGS`, run each and read its exit code:

```bash
test -z "$(gofmt -l . | grep -v node_modules)"
```

```bash
go vet -tags "$TAGS" ./...
```

```bash
go run honnef.co/go/tools/cmd/staticcheck@v0.8.1 -tags "$TAGS" ./...
```

```bash
go test -count=1 -tags "$TAGS" ./internal/... ./ribbonkit/...
```

```bash
go build -tags "$TAGS" -o /tmp/timeribbon .
```

staticcheck is the version `test.ps1` pins. Every macOS and Linux test lives under `./internal/...`
or `./ribbonkit/...`; the rest already runs on Windows. To run from source with a scratch settings folder (`HOME` on macOS,
`XDG_CONFIG_HOME` on Linux):

```bash
go run -tags "$TAGS" .
```

The desktop tests open real windows, place the ribbon and read back where it stands; one registers a
real tray or menu bar icon, so a person at the desktop sees windows come and go. Each Linux `TestMain`
fails at once, saying so, where no display opens. These packages carry no coverage gate: a floor
measured on one desktop would not hold on another. Running these checks before each release is a
practice nothing enforces.

## Running it

The whole gate, which `build.ps1` runs first and cannot skip:

```powershell
./test.ps1
```

It checks formatting, vet and staticcheck, runs every Go test, runs the front end's `lint`,
`typecheck` and `test`, holds the domain and application to 100% and every other gated package to its
floor. A front end without its packages stops the gate. Read the exit code: `0` means every check
passed and every floor held.

Another floor for the domain and application, for a deliberate check:

```powershell
./test.ps1 -Floor 95
```

The front end alone, from `frontend`:

```powershell
npx vitest run
```

One package's coverage in detail; the profile's total can differ slightly from the `-cover` figure
the floors hold:

```powershell
go test -coverprofile=cover.out ./internal/infrastructure/store
```

```powershell
go tool cover -func=cover.out
```

## Keeping this honest

**Prove a new guard bites:** plant the violation, read the exit code, restore the file in a
`finally`. **Re-measure before quoting:** a figure copied forward describes a repository that no
longer exists. **Read the exit code, never the last line.**

See also [DEVELOPMENT.md](DEVELOPMENT.md), [ARCHITECTURE.md](ARCHITECTURE.md) and
[TECH_DEBT.md](TECH_DEBT.md).
