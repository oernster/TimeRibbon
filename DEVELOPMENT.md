# Development

How to build and run TimeRibbon, from a machine with nothing installed to what ships: the setup
program on Windows, the DMG on macOS ([Building on macOS](#building-on-macos)) and the Flatpak on
Linux ([Building on Linux](#building-on-linux)). Each platform's build runs on a machine of that
platform.

Every command is one per block, run from the repository root unless it says otherwise: PowerShell on
Windows, the Terminal's own shell on macOS and Linux. `README.md` is for somebody using the
application; this is for somebody building it. Testing has a document of its own,
[TESTING.md](TESTING.md).

## What a Windows machine needs

| Tool | Version | What for | Where from |
|---|---|---|---|
| Go | 1.26.3, which `go.mod` declares | the application, the setup program and the tools | [go.dev/dl](https://go.dev/dl/) |
| Node.js with npm | a current LTS release; `package.json` pins none | the React front end, its lint, type check and tests | [nodejs.org](https://nodejs.org/) or `winget install OpenJS.NodeJS.LTS` |
| Wails CLI | v2.12.0, the version of the Wails module `go.mod` requires | builds both executables | `go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0` |
| WebView2 runtime | any current | the window the front end is drawn in; Windows 11 ships it | Microsoft's WebView2 page |
| Python 3 with Pillow | any current | Python for every build, which stamps the site's version; Pillow only to regenerate the icons, which are committed | [python.org](https://www.python.org/), then `python -m pip install pillow` |

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
git clone https://github.com/oernster/TimeRibbon.git
```

```powershell
cd TimeRibbon
```

Fetch the Go modules and the front end's packages once:

```powershell
go mod download
```

```powershell
npm --prefix frontend install
```

The application embeds the built page, which git does not hold; `build.ps1` runs the gate before
it builds the page. So on a fresh clone build the page once too:

```powershell
npm --prefix frontend run build
```

Without the packages the gate stops at its front-end step; without the page `go list` stops it at
its first, with `pattern all:frontend/dist: no matching files found`.

## Building on Windows

```powershell
./build.ps1
```

It does these things in order and stops at the first failure:

1. Reads the version from `VERSION` and refuses one that is not `major.minor.patch`; stamps it into
   the site's version tokens under `docs/` through `python stamp_version.py`, which touches nothing
   when they already match. Reads the module path from `go.mod` and each executable's name from its
   `wails.json`.
2. Pins `CGO_ENABLED=0` for everything that follows, so the tests exercise what ships.
3. Runs `test.ps1`, the gate [TESTING.md](TESTING.md#running-it) describes. There is no switch to
   skip it: a gate that can be skipped is skipped on the day it would have caught something.
4. Refuses to go on without `build/windows/icon.ico` and `build/appicon.png`, which
   `tools/genicons.py` makes and which are committed.
5. Writes the application's Windows version resource, `build/windows/info.json`, from `VERSION` and
   `internal/product` through `go run ./tools/versioninfo`, then runs `wails build` for the
   application with the version passed in through `-ldflags`. That runs the front end's
   `npm run build`, which runs `eslint` and `tsc --noEmit` before bundling.
6. Packs the built application and `LICENSE` into `installer/payload.zip` through
   `go run ./tools/payload`.
7. Copies the two icons into the setup program's build folder, writes its version resource the same way
   (described as the setup program), builds it with the same `-ldflags` and copies it to
   `dist-installer`.
8. Writes the empty placeholder back over `installer/payload.zip` whether or not step 7 succeeded,
   so `go build ./...` and the tests keep working and the real payload never reaches a commit.

| Output | What it is |
|---|---|
| `build/bin/TimeRibbon.exe` | the application |
| `dist-installer/TimeRibbonSetup.exe` | the setup program, application included |

To stop after the application, the faster loop when only the application has changed:

```powershell
./build.ps1 -SkipInstaller
```

If a build is interrupted between steps 6 and 8, run `./build.ps1` again rather than committing
`installer/payload.zip`.

### A note on `-ldflags`

The version reaches both executables through `-X github.com/oernster/timeribbon/internal/product.Version=<version>`.
`-X` writes only to a `var`; against a `const` it silently does nothing, which is why `Version` in
`internal/product/product.go` is a var holding a development placeholder until the flag replaces it.
An executable reporting that placeholder was built without `build.ps1`.

## Running it while working

The development loop, with the front end hot-reloading:

```powershell
wails dev
```

It serves the front end from Vite and rebuilds the Go side on change. The Vite configuration allows
the development server to read `installer/frontend/dist`, where the self-reading cycle's script
lives.

Each run appends to `TimeRibbon.log` in the settings folder (`%APPDATA%\TimeRibbon` on Windows;
the other platforms' folders are in [ARCHITECTURE.md](ARCHITECTURE.md#data-locations)), which is
where a fault in a windowed run goes, the Go runtime's own panic report included. A run that finds
the log over 1 MiB starts it afresh. The settings are
in `settings.json` beside it; on Windows the web view keeps its data in `WebView2` in the same
folder. Only one copy runs per user: a second launch shows or hides the first ribbon (as its tray
icon's click does) then exits (FR-506).

`wails dev` is the Windows loop. On macOS and Linux build the page, then run the application with
`go run` and the build's tags, as [TESTING.md](TESTING.md#on-macos-and-linux) describes.

## Installing what you built

```powershell
./dist-installer/TimeRibbonSetup.exe
```

Everything it writes is per user, so Windows never asks for administrator rights. The files go to
`%LOCALAPPDATA%\Programs\TimeRibbon` with a copy of setup as `uninstall.exe`, recorded in the Apps
list; the Start Menu entry, the Desktop shortcut and Start with Windows are the boxes on its first
screen. Over the same version it opens on Repair, Reinstall and Uninstall. Its step log is
`TimeRibbonSetup.log` in the temporary folder. Neither executable is signed: `build.ps1` has no
signing step.

## Building on macOS

An Apple Silicon Mac, with:

| Tool | What for | Where from |
|---|---|---|
| Xcode | the C and Objective-C compiler cgo uses, `codesign`, `notarytool`, `stapler`, `vtool` | the App Store |
| Go, the version `go.mod` declares | the application | [go.dev/dl](https://go.dev/dl/) |
| Node.js with npm | the front end | [nodejs.org](https://nodejs.org/) or `brew install node`; the script installs it where missing |
| create-dmg | the DMG | `brew install create-dmg`; the script installs it where missing |
| A Developer ID Application certificate | signing | the Apple Developer account, in the login keychain |

No `wails` command is needed. Store the notary credential once, in a Terminal at the Mac: the script
reads it from the keychain as the profile `TimeRibbon` and asks for nothing else.
`APPLE_KEYCHAIN_PROFILE` names another profile; with `APPLE_ID` and `APPLE_APP_PASSWORD` both set it
notarises as that Apple ID instead. The command asks for an app-specific password from
appleid.apple.com:

```bash
xcrun notarytool store-credentials TimeRibbon --apple-id <Apple ID> --team-id W7K465GKFJ
```

Then build, in a Terminal at the Mac itself. Signing and notarising read the login keychain, which
refuses a remote shell (measured 2026-09-28 over SSH: `codesign` failed with
`errSecInternalComponent`):

```bash
bash builddmg.sh
```

It does these things in order and stops at the first failure:

1. Refuses to run anywhere but an Apple Silicon Mac; reads the names from `tools/identity` and the
   version from `VERSION`; where `APPLE_ID` and `APPLE_APP_PASSWORD` are set, refuses a password not
   shaped like an app-specific one before building anything.
2. Builds the page with npm.
3. Builds an empty pure Go program and reads from it the oldest macOS the Go toolchain supports;
   hands that to the compiler through `CGO_CFLAGS` and `CGO_LDFLAGS`.
4. Builds the executable with `go build -tags desktop,production` and the version through
   `-ldflags`, refusing it if the linker reports code built for a newer macOS or if the executable
   claims another one.
5. Makes `iconfile.icns` from `build/appicon.png` with `sips` and `iconutil`.
6. Assembles `build/bin/TimeRibbon.app`: the executable, the icon and an `Info.plist` naming the app
   id, the version, the minimum macOS and the copyright line.
7. Signs the bundle with the hardened runtime, notarises it through the profile and staples the
   ticket to it.
8. Makes the DMG with `create-dmg`, stamps its file icon, signs it, notarises and staples it, then
   runs `stapler validate` and `spctl --assess` as a user's machine would.

The output is `TimeRibbon.dmg` in the repository root. For a local trial with no certificate,
`DEVELOPER_ID_APPLICATION=-` signs ad hoc and `ALLOW_UNNOTARIZED=1` skips notarising; such a DMG
opens only on the Mac that built it and is never released.

## Building on Linux

The packages as Ubuntu and Debian name them; another distribution needs the same pieces under its
own names:

```bash
sudo apt-get install -y build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev nodejs npm flatpak flatpak-builder
```

Go from the distribution or [go.dev/dl](https://go.dev/dl/); an older Go 1.26 fetches the version
`go.mod` declares on first use. The packages above are for running and testing from source; the
Flatpak build itself compiles inside the GNOME SDK, which `build_flatpak.sh` installs from Flathub
with its golang and node22 extensions. The script still needs Go on the machine: it reads the names
through `go run ./tools/identity` before the sandbox starts.

```bash
bash build_flatpak.sh
```

It writes the desktop entry, the metainfo and the manifest (named for the app id, all gitignored),
builds the page, the icon sizes (`tools/linuxicons`) and the executable
(`-tags desktop,production,webkit2_41`) inside the sandbox, installs the result for the current user
and exports `timeribbon.flatpak`. Run what it installed:

```bash
flatpak run uk.codecrafter.TimeRibbon
```

`cleanup_flatpak.sh` uninstalls it, removes its sign-in entry and deletes the build outputs; the
settings under `~/.var/app` are left alone.

## Generated files

All of these are committed, so a clone builds without regenerating any of them.

**The icons**, after changing an image in `assets/`:

```powershell
python tools/genicons.py
```

It writes `build/windows/icon.ico` (the one icon both executables, their shortcuts and the tray
wear), `build/appicon.png`, the setup page's header mark and theme toggles, then the page's
artwork: the donate mark, the Add clock button and the icon at the head of About. The page's
artwork is rendered at four times the size the page draws it at and the setup page's header mark
and toggles at about twice theirs, so each stays sharp under display scaling.

**The place catalogue**, `internal/infrastructure/zones/places.tsv`, after a new tz database release,
from a folder holding its `zone.tab`, `iso3166.tab` and `tzdata.zi`:

```powershell
go run ./tools/genplaces -tzdir "C:\Program Files\Git\mingw64\share\zoneinfo"
```

The catalogue records the tz release it came from in its first line. The rules the clocks follow
come through Go's `time.LoadLocation`, which reads the system's tz database first where there is
one, as on macOS and Linux. Windows has none, so there the rules are the ones Go embeds through
`time/tzdata` and a new release reaches the application only through a newer Go; elsewhere the
embedded rules stand in only for a zone the system lacks. `zones_test.go` fails where a catalogue
zone does not resolve.

## Versioning

`VERSION` holds the one version string. `build.ps1`, `builddmg.sh` and `build_flatpak.sh` each pass
it into what they build through the same `-X` flag in `-ldflags`; the Windows version resources, the
DMG's `Info.plist`, the Flatpak's metainfo and the site's version tokens under `docs/` are written
from it too. Nothing in the source holds the release's version, only the development placeholder
above.

The setup program compares the version it carries with the one the Apps list records to choose its
first screen: Install where nothing is recorded, Update over an older version, Go back over a newer
one; Repair, Reinstall or Uninstall over the same one.

The update check (FR-509) compares the running build's version with the tag of the latest release
GitHub publishes, a leading `v` allowed. A version that is not dotted whole numbers is never newer,
so a build carrying the development placeholder is never offered a release.

## Cutting a release

1. Set `VERSION`.
2. Run `./build.ps1` and read its exit code. It stamps the new version into the site under `docs/`;
   commit what it changed there with `VERSION`.
3. On the Mac, run `bash builddmg.sh`; on a Linux machine, `bash build_flatpak.sh`. Run the macOS
   and Linux checks in [TESTING.md](TESTING.md#on-macos-and-linux) on each.
4. Run the checks a person settles in [TESTING.md](TESTING.md#checks-a-person-settles) against
   `dist-installer/TimeRibbonSetup.exe`, `TimeRibbon.dmg` and `timeribbon.flatpak`.
5. Tag the commit as `v` and the version, then publish a release on it with the three attached.
   The update check reads only GitHub's latest published release, never a draft or a prerelease.
   It offers each platform the first asset whose name ends in `.exe`, `.dmg` or `.flatpak`; the setup
   program must therefore be the release's only `.exe`.

## Where things live

| Path | What it holds |
|---|---|
| `main.go` | the composition root; each cell's least size (the page's measure of its widest time and date can widen it), the pull out handle's lane and the panel's size (Settings then grows or shrinks to its content) |
| `app.go`, `window_life.go` | the facade: the calls the page makes; the window's own life with the desktop's events |
| `unpinned.go`, `sunmap.go` | the facade's side of the unpinned ribbon and its tab; of the sun map sharing the ribbon's window |
| `measure.go`, `clockscale.go`, `opacity.go` | the facade's calls for a cell's measured width (FR-620), the corner grip's scale (FR-623) and the window's opacity (FR-622) |
| `wails_calls.go` | the facade's calls into Wails (show, hide, quit, always on top, events), held as fields so its tests can stand in for them |
| `updates.go` | the facade's side of the update check (FR-509): its timing, the check itself whether automatic or asked for from Help and the calls the update panel makes |
| `quit_signal.go` | ending the run when a signal from outside asks, which a close would only turn into hiding while the tray is up |
| `*_test.go` in the root | the facade's tests, over a scripted service and a stand-in window (`fakes_test.go`) |
| `identity.go`, `dto.go`, `launch.go` | About and Licence, the wire, the window's options |
| `platform_windows.go`, `platform_unix.go`, `platform_linux.go`, `platform_darwin.go` | what each platform's run needs before Wails opens: the tray's image and ending on a signal off Windows, X11 on Linux, a framework to link on macOS |
| `bindings_on.go`, `bindings_off.go` | keep the binding-generation run from writing the log or showing a tray icon |
| `internal/domain` | clock readings, placement and the settings value; no I/O |
| `internal/application` | the use cases over their ports |
| `internal/infrastructure` | appdata, cocoamain (macOS), desktop, gtkmain (Linux), iconscale, monitors, runlog, setup (Windows), startup, store, system, update, zones; a file's platform is in its name (`_windows`, `_linux`, `_darwin`, `_unix` for Linux and macOS together) |
| `internal/product` | the name, the app id, the setup program's name, the window class, the donation address, the version, the author, the sign-in label and the credits for each platform |
| `build.ps1`, `test.ps1` | the Windows build; the gate it runs first |
| `VERSION`, `stamp_version.py` | the one version string; stamping it into the site |
| `builddmg.sh` | the macOS DMG |
| `build_flatpak.sh`, `cleanup_flatpak.sh` | building the Linux Flatpak; taking it away again |
| `frontend/src` | the React front end |
| `installer/` | the setup program, a Wails application of its own; its page in `installer/frontend/dist` has no build step |
| `tests/structural` | the tests that hold the architecture in place |
| `tools/` | the icons (committed and the Flatpak's), the place catalogue, the payload, the version resources and the names the Linux and macOS scripts read |
| `assets/` | the master artwork `tools/genicons.py` reads |
| `docs/` | the GitHub Pages site |

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
  it: it is handed the name. The two `wails.json` files must spell it, since Wails reads the
  executable's name from there; a structural test holds them to `internal/product`.
- **The old names stay retired.** Neither the product's former name nor the word its window went by
  before the ribbon may appear in any tracked or new file, the npm lock file aside;
  `tests/structural/retired_test.go` holds that.
- **The wire is written twice**, in `dto.go` and `frontend/src/wire.ts`. Change both; the
  structural test fails otherwise.
- **The update check is the one network request.** No Go file outside
  `internal/infrastructure/update` may import a network package;
  `tests/structural/network_test.go` holds that.
- **A call the page makes that Go can refuse takes a refusal handler.** It answers null rather than
  rejecting; a call without one does not compile.
- **Every new guard is proved by planting a violation**; [TESTING.md](TESTING.md#keeping-this-honest)
  says how.

## See also

- [README.md](README.md) for what the application is and how to use it.
- [TESTING.md](TESTING.md) for the test suite in full.
- [ARCHITECTURE.md](ARCHITECTURE.md) for the invariants and the design decisions.
- [TECH_DEBT.md](TECH_DEBT.md) for what is open and what only looks like debt.
