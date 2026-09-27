# Technical debt

What is still open, what is deliberately left and what only looks like debt.

Every item in this file is a behaviour-preserving internal concern. Nothing here reverts a feature or
changes what the user sees. Read it against `ARCHITECTURE.md` and the structural tests, which are the
authority on the invariants an item might threaten.

Open items are numbered sections. A numbered heading is the definition of an open item, so a scan for
`## <number>.` is the machine check for whether this file is clear. The two standing sections below
are deliberately unnumbered and are not open items.

History is not recorded here. A resolved item is deleted outright, never rewritten as done and never
archived. A resolution worth remembering belongs in the release notes.

## 1. The executables' version resource is empty

Both executables' Windows version resource comes from Wails' template, `build/windows/info.json`,
which fills its fields from an `info` block that neither `wails.json` has. Measured on the build of
2026-09-27: `TimeStrip.exe` and `TimeStripSetup.exe` each carry an empty product version, file
version and description. `VERSION` reaches the running code through `-ldflags` but not the file's
properties, so Explorer's Details tab shows none of them. The fix is to
write the version and the product name into the resource from `VERSION` and `internal/product` at
build time, keeping each in its one home. Blocked on nothing.

## 2. The product's name has homes outside `internal/product`

`wails.json` and `installer/wails.json` spell `TimeStrip` and `TimeStripSetup` as each executable's
name, which `build.ps1` then reads back. A rename would have to reach both files by hand; the
structural test holding the setup page to naming nothing does not look at them. The cost is small
while the name is settled, which is why this is open rather than urgent. Blocked on nothing.

## 3. The setup program's progress event word is stated twice

`installer/app.go` emits `progress` and `setup-routes.js` listens for `'progress'`, with no test
pairing the two as `TestThePageNamesEveryEventGoEmits` pairs the application's words with its page. A
rename on one side alone would leave the progress bar standing still with nothing failing. Blocked on
nothing.

## 4. The Wails facade has no tests

The root package measures 0%: `app.go`, `window_life.go` and `identity.go` have no tests of their
own. What they call is tested in the application layer and what the page does with the answers in the
front-end suites. The facade's own decisions are not: which calls refit the strip, that a drag
whose save failed is still fitted, the panel state and the menu actions. Each reaches Wails and Win32
directly rather than through a field a test could replace, which is what has to change first.
Blocked on nothing.

## 5. Three measured packages carry no floor

`internal/infrastructure/appdata` (100%), `internal/infrastructure/runlog` (76.5%) and
`internal/infrastructure/desktop` (12.0%) are measured by any coverage run but are missing from the
floors in `test.ps1`, so losing their cover fails nothing. Each needs a floor at what it reaches.
Blocked on nothing.

## 6. The running copy that will not close is untested

FR-807 says that a copy still running 5 seconds after setup asked it to close is reported, asking for
it to be closed by hand. `setup.AppProcesses().Close` holds that deadline; no test reaches it.
Closing is forced, so a stand-in that refuses to go is needed. Blocked on nothing.

## 7. A donation page the desktop refused to open cannot be reported

`OpenDonation` hands the address to Wails' `BrowserOpenURL`, which answers no error. Where no browser
is registered, nothing happens and nothing says why. Blocked on Wails: the call would have to answer
a failure. The alternative, opening the address through the shell directly, would give the facade a
second way to reach the desktop.

## Looks like debt, not worth touching

**The drag sends Wails an internal message.** `startDrag` calls `window.WailsInvoke('drag')`, the
message Wails' own drag regions send, rather than a documented call. It is the one way to hand a
press to Windows' own move loop without writing that loop again. The Wails version in `go.mod` pins
it. Check it by hand on any Wails upgrade (TESTING.md, M-2).

**`placement.Fit` still makes room for one cell when handed none.** The application now always
counts at least one cell (the prompt), so that branch is never taken from there. It is the domain's
own contract (tested at the domain) and costs one line.

## Not debt (do not "fix" these)

**The setup page's keyboard ring is the window's model written again.** The setup page has no build
step and can import nothing, so its ring is its own script; `setupRing.test.ts` loads the shipped
script and holds it to the same behaviour. The self-reading cycle went the other way because the
window's build can import a script from the setup page's folder, so that one has a single home.

**The setup program holds no install logic of its own.** Every act the setup window performs goes
through `internal/infrastructure/setup`; `installer/app.go` decides only which screen to open and when
to refuse. The portable half of `setup` is unit tested; `installer` has no tests, since every method
on it acts on the machine.

**The scroll bar's thickness comes from the page.** It is the web engine's bar, which Windows' own
scroll bar metric does not describe, so the page is the only place that can measure it.
