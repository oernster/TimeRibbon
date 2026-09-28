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
| `internal/domain/placement` | 100% | 100% | `test.ps1`, with the application |
| `internal/domain/settings` | 100% | 100% | `test.ps1`, with the application |
| `internal/application` | 100% | 100% | `test.ps1`, with the domain |
| `internal/infrastructure/system` | 100% | 100% | `test.ps1` |
| `internal/infrastructure/zones` | 100% | 100% | `test.ps1` |
| `internal/infrastructure/store` | 92.7% | 92% | `test.ps1` |
| `internal/infrastructure/setup` | 84.0% | 84% | `test.ps1` |
| `tools/versioninfo` | 86.7% | 86% | `test.ps1` |
| `tools/payload` | 82.8% | 82% | `test.ps1` |
| `internal/infrastructure/monitors` | 82.6% | 82% | `test.ps1` |
| `internal/infrastructure/startup` | 80.6% | 80% | `test.ps1` |
| `tools/genplaces` | 38.8% | 38% | `test.ps1` |
| `internal/infrastructure/appdata` | 100% | 100% | `test.ps1` |
| `internal/infrastructure/runlog` | 76.5% | 76% | `test.ps1` |
| the root package (the Wails facade) | 72.5% | 72% | `test.ps1` |
| `internal/infrastructure/desktop` | 14.8% | 14% | `test.ps1` |
| `installer` | 0%, no tests | none | not gated |
| `internal/product` | no statements | none | not gated |

217 Go test functions, each run once with no subtests (an uncached `go test -count=1 -json` over the
packages `go list ./...` gives outside `node_modules`), plus one `TestMain` in
`internal/infrastructure/setup`. Twenty-one of them are the structural tests in `tests/structural`,
which read the source rather than run it; [ARCHITECTURE.md](ARCHITECTURE.md) lists each against the
rule it holds. One more holds a promise rather than a rule of structure:
`TestA1Point0SettingsFileIsReadWhole` reads a frozen settings file of the first release (NFR-C-1); see
ARCHITECTURE.md, The settings file.

### The front end

74 tests across 7 files, under Vitest with jsdom: the ribbon, Settings, About and Licence, the
self-reading cycle, then the setup page's screens, keyboard ring and unreachable-program cases. The
front end has no coverage figure: no coverage provider is installed, so none is measured or claimed.

## How each layer is tested

| Layer | Kind of test | Touches |
|---|---|---|
| `internal/domain` | pure unit, over fixed instants and zones from the embedded tz database | nothing |
| `internal/application` | unit, over hand-written fakes of the six ports | nothing |
| `internal/infrastructure` | integration, over temporary folders and scratch registry keys | the filesystem, `HKCU` under a scratch key, child processes |
| the root package | unit, over a scripted service with Wails and the desktop stood in for by the facade's own fields | nothing |
| `tests/structural` | source and AST scans, plus one `go list` and one `git ls-files` | reads files |
| the front end | component tests under jsdom | nothing |

No Go test uses a mocking library; every double is a hand-written fake with the real interface
behind it. The front end stands in for Go through `src/fakeBridge.ts`, which answers every facade
call the way the real one does. **No test writes to the user's own settings, Start with Windows value
or Apps list**: the store and the log are tested in temporary folders, `startup` and the setup record
under scratch keys beneath `HKCU`.

## What is not tested and why

### The platform owns it

- **`internal/infrastructure/desktop` (14.8%).** The tray icon, the native menus, the move fence and
  the desktop's broadcasts all run on a hidden window's message loop; the ribbon functions act on the
  real ribbon window. The tests cover what is portable: the menu identifier numbering (a submenu
  included), the fence's rectangle arithmetic, a work area read at a point, Windows' drag distance
  and an address Windows cannot open being refused. The loop itself, the menus as drawn, the
  broadcasts arriving and a browser actually opening (M-11) are checks for a person.
- **`internal/infrastructure/monitors` (82.6%).** The displays are read for real; what is not reached
  is Windows refusing to enumerate them or to describe one.
- **`internal/infrastructure/runlog` (76.5%).** Opening the log and pointing standard error at it are
  tested; making the folder failing, the start line failing to write and `SetStdHandle` refusing only
  fail inside the system.
- **The root package (72.5%).** The facade's tests are `facade_test.go` and `window_life_test.go`,
  over the scripted service in `fakes_test.go`. The facade's decisions are tested: which calls fit
  the ribbon, that a drag whose save failed is still fitted, the panel state, the menu actions, the close and the
  recover round each desktop event. Not reached: the composition root (`main.go`, `launch.go`),
  `startup`, `listen` and `shutdown`, which need the real ribbon window and the tray's message loop.
  Nor are the one-line calls in `wails_calls.go` and `window_life.go` that hand a request to Wails
  or Win32 and do nothing else.

### It would change the machine

- **`installer` (0%).** The setup program's facade. Every method acts on the machine: it writes the
  install folder, the shortcuts, the record or Start with Windows; it closes or starts the
  application; it drives the window. The policy beneath it is tested in
  `internal/infrastructure/setup`. The page is tested in `setupScreens.test.ts`, `setupRing.test.ts`
  and `setupUnreachable.test.ts`.
- **`internal/infrastructure/setup` (84.0%).** Tested over temporary folders and a scratch registry
  key. A running copy still there when the wait runs out (FR-807) is tested against a real stand-in
  process whose ending is refused, as it is for a copy setup cannot open. Not reached: the real Apps
  list record (`AppsList`), deleting the install folder after the real setup exits
  (`DeleteAfterExit`; the PowerShell hand-off itself is tested against a stand-in process), COM
  refusing to start or a shortcut refusing to save, a copy or removal failing part way, giving the
  setup window the keyboard (`TakeFocus`) and finding the ribbon's own window after a launch.
- **`internal/infrastructure/startup` (80.6%).** Written, read and removed under a scratch key; the
  registry refusing to open or write the key is not reached.
- **`internal/infrastructure/store` (92.7%).** Not reached: the folder or temporary file refusing to
  be made or flushed; the rename over the old file failing. Only a failing disk produces either.
- **`tools/versioninfo` (86.7%).** `main` hands `run` the real arguments; an output folder that
  cannot be made is not reached. What it writes was read back from both released executables through
  Windows' own version API on 2026-09-27.
- **`tools/payload` (82.8%).** `main` hands `run` the real arguments; the archive failing to write
  part way is not reached.
- **`tools/genplaces` (38.8%).** `main`, `run`, `readTable` and `readVersion` read the tz database's
  own files, which a test machine need not have. The parsing and the writing of the catalogue are
  tested; a test in `zones` holds every zone in the committed catalogue to resolving.

## Checks a person settles

These need a real desktop, real input or a real install; no harness here reaches them. The M numbers
are REQUIREMENTS.md's section 12.

| Check | What to do |
|---|---|
| M-1 | The ribbon shows with no title bar, border or taskbar button |
| M-2 | Dragging empty ribbon area moves it; pressing a control does not; a small wobble does not |
| M-3 | Dragged onto a display at other scaling, the ribbon keeps its size and stays sharp; unplugging that display brings it back onto a visible one |
| M-4 | The tray icon, its menu, a left click, Always on top and Exit behave as FR-501 to FR-505 say |
| M-5 | Changing the Windows clock, changing the time zone and sleeping then waking the machine each leave every clock right within 2 seconds; each writes a line to `TimeRibbon.log` |
| M-6 | Launching a second copy shows the first and leaves one tray icon |
| M-7 | Switching the Windows theme while TimeRibbon follows it recolours the ribbon |
| M-8 | Settings and the place search can be driven entirely from the keyboard |
| M-9 | Setup installs, updates, repairs and uninstalls on a real machine without asking for administrator rights, closing a running copy first |
| M-10 | Both menus open a Help submenu whose About and Licence each show their panel; the licence reads itself down after 5 seconds, a wheel stops it and it resumes; setup's Licence screen does the same |
| M-11 | The donate button at the foot of Settings opens the default browser on the donation page |
| M-12 | Each Position item puts the ribbon flush against its edge and centred along it on the display it is on; it opens there next time; choosing Horizontal or Vertical from either menu sends it to the top or right edge; small clocks show their whole date and time in both styles; the Settings title and Close stay put while the panel scrolls; each colour scheme looks right in Light, Dark and System, Neon glowing and dark in all three |
| Wheel at 250% | On a display at 250%, one notch of a plain wheel over a scrolling horizontal ribbon moves it as far as a native notch moves a vertical one. In headless Edge through the DevTools protocol it moved 48 against 120; whether a physical wheel does the same is not known |

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

## Keeping this honest

**Prove a new guard bites.** A test that has never been seen to fail is not yet a guard: plant the
violation, read the exit code, then restore the file in a `finally` so a failed run cannot leave the
plant behind. The structural package's own header records that each of its assertions was proved
this way; the credits test was proved on 2026-09-27 in both directions.

**Re-measure before quoting.** Every figure above was measured when it was written; copying one
forward is how a document starts describing a repository that no longer exists.

**Read the exit code, never the last line.**

## See also

- [DEVELOPMENT.md](DEVELOPMENT.md) for building and running from source.
- [ARCHITECTURE.md](ARCHITECTURE.md) for the invariants the structural tests enforce.
- [TECH_DEBT.md](TECH_DEBT.md) for what is open, what is deliberately left and what only looks like
  debt.
