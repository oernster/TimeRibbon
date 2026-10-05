# Testing

What is tested, what is not and why the line falls where it does. Every figure was measured by the
commands in [Running it](#running-it) when it was written; every shortfall is named with its reason.
The ribbon itself is [ribbonkit](https://github.com/oernster/ribbonkit), tested by its own gate and
its own TESTING.md; this document covers what TimeRibbon adds.

## Before the first run on Windows

Some anti-virus programs quarantine a freshly built test binary in Go's scratch folder, which stops
the suite. If a run fails with access denied or a missing file on a test binary, allow the folder
`go env GOTMPDIR` names (the system temporary folder where it prints nothing). Install the front end's
packages once, which fetches the kit's page half from GitHub at the tag `package.json` names. Build
the page once too, since the application embeds `frontend/dist`, which git does not hold:

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
| `internal/domain/clock`, `settings`, `sun` | 100% | 100% |
| `internal/application` | 100% | 100% |
| `internal/infrastructure/store`, `zones` | 100% | 100% |
| `tools/versioninfo` | 86.7% | 86% |
| `tools/payload` | 82.8% | 82% |
| `tools/identity` | 75% | 75% |
| `tools/linuxicons` | 67.7% | 67% |
| the root package (the Wails facade) | 66.2% | 66% |
| `tools/genplaces` | 58.6% | 58% |
| `internal/product` | 100% | not gated |
| `installer` | 0% | not gated |

`installer` is the setup program's composition root; its one test reads the pictures it carries,
which runs no statement.

Every figure is the Windows build's, which `test.ps1` measures. That build compiles 138 Go test
functions, counted from the test files `go list` selects, each running once with no subtests.
Twenty-five are the structural tests, which read the source and are the same on every platform;
[ARCHITECTURE.md](ARCHITECTURE.md) lists each against its rule. `TestA1Point0SettingsFileIsReadWhole`
in `store` holds the settings file's promise (NFR-C-1); `contrast_test.go` holds NFR-U-1 in Go because
Vitest hands a CSS import back empty. The macOS and Linux builds compile 134
([On macOS and Linux](#on-macos-and-linux)).

### The front end

53 tests in 10 files under Vitest with jsdom, run from `frontend`: the app around the clocks
(`app.test.tsx`), the clocks in the kit's band (`ribbon.test.tsx`), Settings (`settings.test.tsx`), an
accessible name and tooltip on every icon-only control (`a11y.test.tsx`, NFR-U-4), measuring a cell's
widest text (`measure.test.ts`, FR-620), the corner grip's use on the ribbon (`scaleGrip.test.tsx`,
FR-623), the sun map, its blend and its labels (`surface.test.tsx`, `sunLight.test.ts`,
`labels.test.ts`, FR-914) and every timer the page schedules, the kit's included, against a reasoned
allow-list (`timers.test.ts`, NFR-P-4). No coverage provider is installed, so no figure is claimed.

## How each layer is tested

| Layer | Kind of test | Touches |
|---|---|---|
| `internal/domain` | pure unit over fixed instants and zones loaded with the tz database embedded | nothing |
| `internal/application` | unit over hand-written fakes of its ports | nothing |
| `internal/infrastructure` | integration over temporary folders | the filesystem |
| the root package | unit over a scripted service and a stand-in window | reads `frontend/src/api.ts` and the kit's `bridge.ts` as npm installed it |
| `tests/structural` | source and AST scans, a `go list` per platform, one `git ls-files`, the kit's files read through `go list -m` | reads files |
| the front end | component tests under jsdom over `fakeBridge.ts`, which records every call and builds on the kit's stand-in (`@oernster/ribbonkit/testing`) | nothing |

No Go test uses a mocking library. **No test writes to the user's own settings, sign-in entry or
Apps list** and **no test reaches the network**.

## What is not tested and why

- **The root package (66.2%).** TimeRibbon's own half of the facade is tested over a scripted service
  and a stand-in window: every change to the clocks fits the ribbon once, the Snapshot carries the
  window's reading, TimeRibbon's menu actions, the measurements, the product handed to the window, the
  adapter reading the ribbon's choices out of the settings; every method the page's `Bridge` calls is
  bound, with nothing of the `Control`. Not reached: the composition root (`main`, `keepLog`,
  `settingsDir`, `run`).
- **`installer` (0%).** Only the composition root; the setup window and its policy are tested in the
  kit.
- **The tools:** each `main` handing `run` its real arguments; folders and archives refusing to be
  made or closed. `genplaces` (58.6%) also reads the tz database's own files, which a test machine need
  not have; its parsing and writing are tested. What `versioninfo` writes was read back from both
  released executables through Windows' version API.

What the window, the tray, focus, paint and the install do on a real desktop is checked by hand in a
real build of each platform.

## On macOS and Linux

The root package and the store compile their own halves there; the root package also links the
kit's cgo half. Check them on a machine of that platform, set up as [DEVELOPMENT.md](DEVELOPMENT.md)
says, with the page built. The kit's own macOS and Linux checks are in its TESTING.md.

| What | macOS | Linux |
|---|---|---|
| Tags | `desktop,production` | `desktop,production,webkit2_41` |
| Go test functions | 134 | 134 |

With the platform's tags in `TAGS`, run each and read its exit code:

```bash
test -z "$(gofmt -l . | grep -v node_modules)"
```

```bash
go vet -tags "$TAGS" $(go list ./... | grep -v node_modules)
```

```bash
go run honnef.co/go/tools/cmd/staticcheck@v0.8.1 -tags "$TAGS" $(go list ./... | grep -v node_modules)
```

```bash
go test -count=1 -tags "$TAGS" $(go list ./... | grep -v node_modules)
```

```bash
go build -tags "$TAGS" -o /tmp/timeribbon .
```

staticcheck is the version `test.ps1` pins. To run from source with a scratch settings folder (`HOME`
on macOS, `XDG_CONFIG_HOME` on Linux):

```bash
go run -tags "$TAGS" .
```

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
`finally`; confirm first that the clean tree passes. **Re-measure before quoting:** a figure copied
forward describes a repository that no longer exists. **Read the exit code, never the last line.**

See also [DEVELOPMENT.md](DEVELOPMENT.md), [ARCHITECTURE.md](ARCHITECTURE.md) and
[TECH_DEBT.md](TECH_DEBT.md).
