# Testing

What is tested, what is not and why the line falls where it does.

A coverage figure on its own is a number without a claim behind it. Every figure here was measured
at the moment of writing by the commands in [Running it](#running-it); every shortfall is named with
the reason it is one.

## Before the first run on Windows

Go builds each test binary in its scratch directory. Some anti-virus programs quarantine a freshly
built binary there, which stops the suite before a test runs. If a run fails with an
access-denied or missing-file error on a test binary, allow the folder `go env GOTMPDIR` names (the
system temporary folder where that prints nothing) in the anti-virus. The front end's checks need
their packages installed once. The application embeds the built page (`frontend/dist`, which git
does not hold), so on a fresh checkout build it once too; without it `go list` stops the gate at
its first step with `pattern all:frontend/dist: no matching files found`:

```powershell
npm --prefix frontend install
```

```powershell
npm --prefix frontend run build
```

Every text file is checked out with LF endings (`.gitattributes`), since gofmt refuses CRLF. A
checkout made before that file existed may hold CRLF; checking it out afresh brings it back to LF.

## The standard

**A floor is a measurement, never an aspiration.** Every floor in `test.ps1` sits at or just below
what that package measured, so it fails once cover is lost, which is the only moment worth being
told. A floor picked from an aspiration only teaches people to lower it.

**A gap is named or it is closed.** Where something cannot be tested, this document says what it is
and what stops it. An unexplained shortfall cannot be told from an oversight.

## What the numbers are

### Go

| Package | Coverage | Floor | Gated by |
|---|---|---|---|
| `internal/domain/clock` | 100% | 100% | `test.ps1`, with the application |
| `internal/domain/hover` | 100% | 100% | `test.ps1`, with the application |
| `internal/domain/placement` | 100% | 100% | `test.ps1`, with the application |
| `internal/domain/settings` | 100% | 100% | `test.ps1`, with the application |
| `internal/application` | 100% | 100% | `test.ps1`, with the domain |
| `internal/infrastructure/system` | 100% | 100% | `test.ps1` |
| `internal/infrastructure/update` | 100% | 100% | `test.ps1` |
| `internal/infrastructure/zones` | 100% | 100% | `test.ps1` |
| `internal/infrastructure/iconscale` | 100% | 100% | `test.ps1` |
| `internal/infrastructure/store` | 92.9% | 92% | `test.ps1` |
| `internal/infrastructure/setup` | 84.0% | 84% | `test.ps1` |
| `tools/versioninfo` | 86.7% | 86% | `test.ps1` |
| `tools/payload` | 82.8% | 82% | `test.ps1` |
| `internal/infrastructure/monitors` | 82.6% | 82% | `test.ps1` |
| `internal/infrastructure/startup` | 80.6% | 80% | `test.ps1` |
| `tools/identity` | 75% | 75% | `test.ps1` |
| `tools/linuxicons` | 67.7% | 67% | `test.ps1` |
| `tools/genplaces` | 38.8% | 38% | `test.ps1` |
| `internal/infrastructure/appdata` | 100% | 100% | `test.ps1` |
| `internal/infrastructure/runlog` | 76.5% | 76% | `test.ps1` |
| the root package (the Wails facade) | 81.8% | 76% | `test.ps1` |
| `internal/infrastructure/desktop` | 32.8% | 14% | `test.ps1` |
| `internal/product` | 100% | none | not gated |
| `installer` | 0%, no tests | none | not gated |

Every figure is the Windows build's, which is what `test.ps1` measures. The Windows build compiles
260 Go test functions, counted from the test files `go list` selects for it; each runs once with no
subtests, plus one `TestMain` in `internal/infrastructure/setup`. Twenty-four of them are the
structural tests in `tests/structural`, which read the source rather than run it and are the same
on every platform; [ARCHITECTURE.md](ARCHITECTURE.md) lists each against the rule it holds. One
test in `store` holds a promise rather than a rule of structure:
`TestA1Point0SettingsFileIsReadWhole` reads a frozen settings file of the first release (NFR-C-1);
see ARCHITECTURE.md, The settings file. The macOS and Linux builds compile 241 each
([On macOS and Linux](#on-macos-and-linux)).

### The front end

82 tests across 8 files, under Vitest with jsdom: the ribbon, Settings, About, Licence and the
update panel; the self-reading cycle; the watch on the page's `devicePixelRatio`; then the setup
page's screens, keyboard ring and unreachable-program cases. The front end has no coverage figure:
no coverage provider is installed, so none is measured or claimed.

## How each layer is tested

| Layer | Kind of test | Touches |
|---|---|---|
| `internal/domain` | pure unit, over fixed instants and zones loaded through `time.LoadLocation` with the tz database embedded (the only source on Windows; the system's zone files come first on macOS and Linux) | nothing |
| `internal/application` | unit, over hand-written fakes of the seven ports | nothing |
| `internal/infrastructure` | integration, over temporary folders and scratch registry keys | the filesystem, `HKCU` under a scratch key, child processes, the real displays |
| the root package | unit, over a scripted service with Wails and the desktop stood in for by the facade's own fields | nothing |
| `tests/structural` | source and AST scans, plus a `go list` for each platform and one `git ls-files` | reads files |
| the front end | component tests under jsdom | nothing |

No Go test uses a mocking library; every double is a hand-written fake with the real interface
behind it. The front end stands in for Go through `src/fakeBridge.ts`, which records every facade
call the page makes and gives each a canned answer of the real one's shape. **No test writes to the
user's own settings, sign-in entry or Apps list**: the store and the log are tested in temporary
folders; `startup` under a scratch key beneath `HKCU` on Windows and in a temporary folder on macOS
and Linux; the setup record under a scratch key.
**No test reaches the network**: the update check's adapter is tested over a stand-in HTTP client, so
GitHub is asked only by the running application (M-13).

## What is not tested and why

### The platform owns it

- **`internal/infrastructure/desktop` (14.7%).** The tray icon, the native menus, the move fence and
  the desktop's broadcasts all run on a hidden window's message loop; the ribbon functions act on the
  real ribbon window. `PixelsPerDIP`, which on Windows hands the page's ratio straight back, is
  called only by the root package's tests, which this figure does not count. The tests cover the
  menu identifier numbering (a submenu included), the fence's rectangle arithmetic, a work area read
  at a point, Windows' drag distance, an address Windows cannot open being refused and the clock
  watch seeing a jump of the wall clock, then stopping. The loop itself, the menus as drawn (with
  `separatedBefore`, which only drawing calls), the broadcasts arriving and a browser actually
  opening (M-11) are checks for a person.
- **`internal/infrastructure/monitors` (82.6%).** The displays are read for real; what is not reached
  is Windows refusing to enumerate them or to describe one.
- **`internal/infrastructure/runlog` (76.5%).** Opening the log and pointing standard error at it are
  tested, as is the folder refusing to be made; the log file refusing to open, the start line failing
  to write and `SetStdHandle` refusing only fail inside the system.
- **The root package (76.6%).** The facade's tests are `facade_test.go`, `window_life_test.go`,
  `updates_test.go` and `quit_signal_test.go`, over the scripted service in `fakes_test.go`. The
  facade's decisions are tested: which calls fit the ribbon, that a drag whose save failed is still
  fitted, the panel state, the menu actions, the close, a signal from outside ending the application
  even with the tray up, the recover round each desktop event and each update check, the update
  watch's timing and what Download and Skip act on. Not reached: the composition root (`main.go`,
  `launch.go`), `startup`, `listen` and `shutdown`, which need the real ribbon window and the tray's
  message loop. Nor are the one-line calls in `wails_calls.go` and `window_life.go` that hand a
  request to Wails or Win32 and do nothing else, nor `preparePlatform` in `platform_windows.go`,
  which does nothing on Windows.

### It would change the machine

- **`installer` (0%).** The setup program's facade. Its methods read the machine or act on it: they
  read what is installed and whether the application runs; they write the install folder, the
  shortcuts, the record or Start with Windows; they close or start the application; they drive the
  window. `Licence` alone reads only the terms the setup program carries. The policy beneath it is
  tested in `internal/infrastructure/setup`. The page is tested in `setupScreens.test.ts`,
  `setupRing.test.ts` and `setupUnreachable.test.ts`.
- **`internal/infrastructure/setup` (84.0%).** Tested over temporary folders and a scratch registry
  key. A running copy still there when the wait runs out (FR-807) is tested against a real stand-in
  process whose ending is refused, as it is for a copy setup cannot open. Not reached: the real Apps
  list record (`AppsList`), deleting the install folder after the real setup exits
  (`DeleteAfterExit`; the PowerShell hand-off itself is tested against a stand-in process), COM
  refusing to start or a shortcut refusing to save, a copy or removal failing part way, giving the
  setup window the keyboard (`TakeFocus`), finding the ribbon's own window after a launch and the
  `Places` accessor, which only the setup program's facade reads.
- **`internal/infrastructure/startup` (80.6%).** Written, read and removed under a scratch key; the
  registry refusing to open the key or to read, write or delete its value is not reached.
- **`internal/infrastructure/store` (92.9%).** The folder refusing to be made and the rename over the
  old file failing are tested. Not reached: the temporary file refusing to be made, written, flushed
  or closed, which only a failing disk produces; the error returns in `encode` and `extrasOf`, which
  guard values and a file already known to be well formed.
- **`tools/versioninfo` (86.7%).** `main` hands `run` the real arguments; an output folder that
  cannot be made is not reached. What it writes was read back from both released executables through
  Windows' own version API on 2026-09-27.
- **`tools/payload` (82.8%).** `main` hands `run` the real arguments; the archive failing to close
  or to move into place is not reached.
- **`tools/genplaces` (38.8%).** `main`, `run`, `readTable` and `readVersion` read the tz database's
  own files, which a test machine need not have. The parsing and the writing of the catalogue are
  tested; a test in `zones` holds every zone in the committed catalogue to resolving.

## On macOS and Linux

The macOS and Linux halves of infrastructure compile only for their own platform; the parts that
face the desktop (`cocoamain`, `gtkmain`, `monitors`, `desktop`) also need cgo against AppKit or
GTK. `test.ps1` reaches none of them. They are checked on a machine of their own
platform, set up as [DEVELOPMENT.md](DEVELOPMENT.md) says, from a checkout with the page built. Each
build compiles 241 Go test functions: the shared ones, the structural tests and its own.

| What | macOS | Linux |
|---|---|---|
| Tags | `desktop,production` | `desktop,production,webkit2_41` |
| Tests that need cgo | `cocoamain` 3, `monitors` 3, `desktop` 16 (with the shared `_unix` tests) | `gtkmain` 3, `monitors` 2, `desktop` 17 (with the shared `_unix` tests) |
| Needs | a signed-in desktop | a signed-in desktop session with a display and a tray host |

With the tags for the platform in `TAGS`, run each check and read its exit code:

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
go test -count=1 -tags "$TAGS" ./internal/...
```

```bash
go build -tags "$TAGS" -o /tmp/timeribbon .
```

staticcheck is the version `test.ps1` pins. The test line covers `./internal/...`, where every
macOS and Linux test lives; the root package, the tools and the structural tests compile there only
tests that `test.ps1` already runs on Windows (`tools/payload` builds for Windows alone).

To run the application from source, with a scratch settings folder so a real one is not touched
(point `HOME` at one on macOS; `XDG_CONFIG_HOME` on Linux):

```bash
go run -tags "$TAGS" .
```

**What those tests do on the desktop.** The desktop tests open real windows through the toolkit, put
the ribbon where they place it and read back where it stands; on macOS one puts the icon in the menu
bar and takes it out; on Linux one registers the tray icon with the session's real tray host. So
they need a desktop; a person at it sees windows open and close. Each Linux `TestMain` runs its
package under `gtkmain.ServeTests`, which fails at once, saying so, where no display can be opened.
What they proved when they were written, measured 2026-09-28: a ribbon stands exactly where it is
placed; it returns exactly from a panel's size (on Linux only once the size is awaited); a position
`Place` chose is never taken for a drag.

**What has no figure.** These packages have no coverage gate: the parts that matter act on a real
desktop, which a coverage run on another machine cannot reach; a floor measured on one person's
desktop would not hold on another's. The practice, which nothing enforces: the checks above are run
before each release.

## Checks a person settles

These need a real desktop, real input or a real install; no harness here reaches them. The M numbers
are REQUIREMENTS.md's section 12. Each is checked on every platform, with these differences: M-1's
taskbar is the Dock on macOS; M-4's left click is a double click on Ubuntu's tray and opens the menu
on macOS; M-5 and M-7 use the platform's own clock, time zone and theme settings; M-9 is setup on
Windows, installing and removing the DMG on macOS and the Flatpak on Linux.

| Check | What to do |
|---|---|
| M-1 | The ribbon shows with no title bar, border or taskbar button |
| M-2 | Dragging empty ribbon area moves it; pressing a control does not; a small wobble does not |
| M-3 | Dragged onto a display at other scaling, the ribbon keeps its size and stays sharp; unplugging that display brings it back onto a visible one |
| M-4 | The tray icon, its menu, a left click, Always on top and Exit behave as FR-501 to FR-505 say |
| M-5 | Changing the Windows clock, changing the time zone and sleeping then waking the machine each leave every clock right within 2 seconds; each writes a line to `TimeRibbon.log` |
| M-6 | Launching a second copy leaves one tray icon and hides a shown ribbon; launching again shows it; a Stream Deck Open action pointed at TimeRibbon does the same on each press |
| M-7 | Switching the Windows theme while TimeRibbon follows it recolours the ribbon |
| M-8 | Settings and the place search can be driven entirely from the keyboard |
| M-9 | Setup installs, updates, repairs and uninstalls on a real machine without asking for administrator rights, closing a running copy first |
| M-10 | Both menus open a Help submenu whose About and Licence each show their panel; the licence reads itself down after 5 seconds, a wheel stops it and it resumes; setup's Licence screen does the same |
| M-11 | The donate button at the foot of Settings opens the default browser on the donation page |
| M-12 | Each Position item puts the ribbon flush against its edge and centred along it on the display it is on; it opens there next time; choosing Horizontal or Vertical from either menu sends it to the top or right edge; small clocks show their whole date and time in both styles; the Settings title and Close stay put while the panel scrolls; each colour scheme looks right and unmistakably its own in Light, Dark and System, Neon glowing on its dark side only; every date format shows its whole date in large and small cells, a Wednesday in September the widest |
| M-13 | Help's Check for updates says this is the latest version with the network on and that GitHub could not be reached with it off; a build older than the latest release shows the update panel a few seconds after it starts; Download opens this platform's download in the browser; after Skip this version the next start shows nothing; the Flatpak build reaches GitHub too |
| M-14 | Unticking Pin ribbon in either menu shrinks the ribbon to an accent tab 8 wide on its side nearer the display's edge a second after the pointer leaves; resting the pointer on the tab for 0.3 s opens it while the window being typed in keeps focus; crossing the tab quickly does not; the tab stays above a maximised window with Always on top off; the right-click menu and a panel keep it open; dragging the open ribbon does not collapse it; a Stream Deck press hides the tab and the next brings it back; ticking Pin ribbon ends all of it. On Linux, dragging and the menu are the checks that matter most: crossings made by a grab are ignored, which no test can reach |
| Wheel at 250% | On a display at 250%, one notch of a plain wheel over a scrolling horizontal ribbon moves it as far as a native notch moves a vertical one; only a physical wheel settles it |

The ribbon's layout at its full size was measured in headless Edge 154.0.4258.37, the version of the
installed WebView2 runtime, with the application's own stylesheets. At 100% and 250% a scrolling
ribbon shows one scroll bar with no clock cut off; every clock is reachable beside a notice.

## Running it

Every command is PowerShell, run from the repository root.

The whole gate, which `build.ps1` runs before it builds and cannot be told to skip:

```powershell
./test.ps1
```

It checks formatting, runs `go vet`, runs staticcheck at the version it pins, runs every Go test,
runs the front end's `lint`, `typecheck` and `test` scripts, holds the domain and the application to
100%, then holds each other gated package at its floor. A front end without its packages installed
stops the gate rather than being skipped. Read the exit code, never the last line: `0` means every
check passed and every floor held.

To hold the domain and the application to another floor, for a deliberate check:

```powershell
./test.ps1 -Floor 95
```

The front end on its own, from `frontend`:

```powershell
npx vitest run
```

One package's coverage in detail, when a figure needs explaining:

```powershell
go test -coverprofile=cover.out ./internal/infrastructure/store
```

```powershell
go tool cover -func=cover.out
```

The profile's total need not equal the `-cover` figure that the table and `test.ps1` use: for
`monitors` it reads 81.0% against 82.6%. The floor holds the `-cover` figure.

## Keeping this honest

**Prove a new guard bites.** A test that has never been seen to fail is not yet a guard: plant the
violation, read the exit code, then restore the file in a `finally` so a failed run cannot leave the
plant behind. The structural package's own header records that each of its assertions was proved
this way.

**Re-measure before quoting.** Every figure above was measured when it was written; copying one
forward is how a document starts describing a repository that no longer exists.

**Read the exit code, never the last line.**

## See also

- [DEVELOPMENT.md](DEVELOPMENT.md) for building and running from source.
- [ARCHITECTURE.md](ARCHITECTURE.md) for the invariants the structural tests enforce.
- [TECH_DEBT.md](TECH_DEBT.md) for what is open, what is deliberately left and what only looks like
  debt.
