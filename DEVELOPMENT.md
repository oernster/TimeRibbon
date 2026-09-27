# Development

How to build and run TimeStrip on Windows, from a machine with nothing installed to a setup program.

Every command here is PowerShell, one command per block, run from the repository root unless it says
otherwise. `README.md` is for somebody using the application; this is for somebody building it.
Testing has a document of its own, [TESTING.md](TESTING.md).

## What the machine needs

| Tool | Version | What for | Where from |
|---|---|---|---|
| Go | 1.26.3, which `go.mod` declares | the application, the setup program and the tools | [go.dev/dl](https://go.dev/dl/) |
| Node.js with npm | a current LTS release; `package.json` pins none | the React front end, its lint, type check and tests | [nodejs.org](https://nodejs.org/) or `winget install OpenJS.NodeJS.LTS` |
| Wails CLI | v2.12.0, the version of the Wails module `go.mod` requires | builds both executables | `go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0` |
| WebView2 runtime | any current | the window the front end is drawn in; Windows 11 ships it | Microsoft's WebView2 page |
| Python 3 with Pillow | any current | only to regenerate the icons, which are committed | [python.org](https://www.python.org/), then `python -m pip install pillow` |

The gate runs staticcheck at the version `test.ps1` pins through `go run`, so the first run on a
machine fetches it and needs the network. Nothing needs a C compiler: `build.ps1` pins cgo off.

`wails.exe` lands in `%USERPROFILE%\go\bin`. Where the next command is not found, that folder is not
on the path:

```powershell
wails doctor
```

Add it for the current session:

```powershell
$env:PATH = "$env:USERPROFILE\go\bin;$env:PATH"
```

## Getting the source

```powershell
git clone https://github.com/oernster/TimeStrip.git
```

```powershell
cd TimeStrip
```

Fetch the Go modules and the front end's packages once:

```powershell
go mod download
```

```powershell
npm --prefix frontend install
```

The gate stops at its front-end step without `frontend/node_modules`, so run the second command once
before `test.ps1` or `build.ps1`.

## Building

```powershell
./build.ps1
```

It does these things in order and stops at the first failure:

1. Reads the version from `VERSION` and refuses one that is not `major.minor.patch`. Reads the module
   path from `go.mod` and each executable's name from its `wails.json`.
2. Pins `CGO_ENABLED=0` for everything that follows, so the tests exercise what ships.
3. Runs `test.ps1`, the gate [TESTING.md](TESTING.md#running-it) describes. There is no switch to
   skip it: a gate that can be skipped is skipped on the day it would have caught something.
4. Refuses to go on without `build/windows/icon.ico` and `build/appicon.png`, which
   `tools/genicons.py` makes and which are committed.
5. Runs `wails build` for the application with the version passed in through `-ldflags`. That runs
   the front end's `npm run build`, which runs `eslint` and `tsc --noEmit` before bundling.
6. Packs the built application and `LICENSE` into `installer/payload.zip` through
   `go run ./tools/payload`.
7. Copies the icon into the setup program's build folder, builds the setup program with the same
   `-ldflags` and copies it to `dist-installer`.
8. Writes the empty placeholder back over `installer/payload.zip` whether or not step 7 succeeded,
   so `go build ./...` and the tests keep working and the real payload never reaches a commit.

| Output | What it is |
|---|---|
| `build/bin/TimeStrip.exe` | the application |
| `dist-installer/TimeStripSetup.exe` | the setup program, application included |

To stop after the application, the faster loop when only the application has changed:

```powershell
./build.ps1 -SkipInstaller
```

If a build is interrupted between steps 6 and 8, run `./build.ps1` again rather than committing
`installer/payload.zip`.

### A note on `-ldflags`

The version reaches both executables through `-X github.com/oernster/timestrip/internal/product.Version=<version>`.
`-X` writes only to a `var`; against a `const` it silently does nothing, which is why `Version` is a
var holding `0.0.0-dev` until the flag replaces it. An executable reporting `0.0.0-dev` was built
without `build.ps1`.

## Running it while working

The development loop, with the front end hot-reloading:

```powershell
wails dev
```

It serves the front end from Vite and rebuilds the Go side on change. The Vite configuration allows
the development server to read `installer/frontend/dist`, where the self-reading cycle's script
lives.

Each run appends to `%APPDATA%\TimeStrip\TimeStrip.log`, which is where a fault in a windowed run
goes, the Go runtime's own panic report included. The settings are in `settings.json` beside it.
Only one copy runs per Windows user: a second launch shows the first and exits.

## Installing what you built

```powershell
./dist-installer/TimeStripSetup.exe
```

Everything it writes is per user, so Windows never asks for administrator rights. The files go to
`%LOCALAPPDATA%\Programs\TimeStrip` with a copy of setup as `uninstall.exe`, recorded in the Apps
list; the Start Menu entry, the Desktop shortcut and Start with Windows are the boxes on its first
screen. Over the same version it opens on Repair, Reinstall and Uninstall. Its step log is
`TimeStripSetup.log` in the temporary folder. Neither executable is signed: `build.ps1` has no
signing step.

## Generated files

All of these are committed, so a clone builds without regenerating any of them.

**The icons**, after changing an image in `assets/`:

```powershell
python tools/genicons.py
```

It writes `build/windows/icon.ico` (the one icon both executables, their shortcuts and the tray
wear), `build/appicon.png`, the setup page's header mark and theme toggles, then the page's
artwork: the donate mark, the Add clock button and the icon at the head of About. Each is rendered at about
four times the size it is drawn at, so it stays sharp under display scaling.

**The place catalogue**, `internal/infrastructure/zones/places.tsv`, after a new tz database release,
from a folder holding its `zone.tab`, `iso3166.tab` and `tzdata.zi`:

```powershell
go run ./tools/genplaces -tzdir "C:\Program Files\Git\mingw64\share\zoneinfo"
```

The catalogue records the tz release it came from in its first line. The rules the clocks follow are
the ones Go embeds through `time/tzdata`, so a new release reaches the application only through a
newer Go; `zones_test.go` fails where a catalogue zone does not resolve in them.

## Versioning

`VERSION` holds the one version string. `build.ps1` passes it into both executables; nothing in the
source holds a version literal. The setup program compares the version it carries with the one the
Apps list records to choose between Install, Update, Go back and the Installed screen.

## Cutting a release

1. Set `VERSION`.
2. Run `./build.ps1` and read its exit code.
3. Run the checks a person settles in [TESTING.md](TESTING.md#checks-a-person-settles) against
   `dist-installer/TimeStripSetup.exe`.
4. Tag the commit and attach `TimeStripSetup.exe` to the release.

## Where things live

| Path | What it holds |
|---|---|
| `main.go` | the composition root and the cell and panel sizes |
| `app.go`, `window_life.go` | the facade: the calls the page makes; the window's own life with the desktop's events |
| `identity.go`, `dto.go`, `launch.go` | About and Licence, the wire, the window's options |
| `bindings_on.go`, `bindings_off.go` | keep the binding-generation run from writing the log or showing a tray icon |
| `internal/domain` | clock readings, placement and the settings value; no I/O |
| `internal/application` | the use cases over their ports |
| `internal/infrastructure` | appdata, desktop, monitors, runlog, setup, startup, store, system, zones |
| `internal/product` | the name, the window class, the donation address, the version, the author and the credits |
| `frontend/src` | the React front end |
| `installer/` | the setup program, a Wails application of its own; its page in `installer/frontend/dist` has no build step |
| `tests/structural` | the tests that hold the architecture in place |
| `tools/` | the icons, the place catalogue and the payload |
| `assets/` | the master artwork `tools/genicons.py` reads |

## House rules worth knowing before a first change

- **The layer direction is enforced, not suggested.** The domain imports nothing outside the domain,
  performs no I/O and reads no clock. The application imports neither infrastructure nor Wails. Only
  `main.go`, `app.go` and `window_life.go` wire the two together.
- **No file over 400 lines**: the Go, the front end's TypeScript and CSS and the setup page, tests
  included. A file landing between 381 and 400 lines is reduced to 350 or fewer. Build scripts are
  outside the rule.
- **No magic numbers.** A literal that needs a comment to say what it represents is a named constant
  or is derived from data.
- **The product is named once**, in `internal/product/product.go`. The setup page must never write
  it: it is handed the name.
- **The wire is written twice**, in `dto.go` and `frontend/src/wire.ts`. Change both; the
  structural test fails otherwise.
- **A call the page makes that Go can refuse takes a refusal handler.** It answers null rather than
  rejecting; a call without one does not compile.
- **Every new guard is proved by planting a violation**; [TESTING.md](TESTING.md#keeping-this-honest)
  says how.

## See also

- [README.md](README.md) for what the application is and how to use it.
- [TESTING.md](TESTING.md) for the test suite in full.
- [ARCHITECTURE.md](ARCHITECTURE.md) for the invariants and the design decisions.
- [TECH_DEBT.md](TECH_DEBT.md) for what is open and what only looks like debt.
