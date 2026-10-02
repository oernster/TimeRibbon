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

## 1. A Linux and macOS test writes the panel's size out again

`TestTheRibbonReturnsFromAPanelToWhereItIsPlaced` in `internal/infrastructure/desktop/ribbon_unix_test.go`
grows the window to a panel and back, with the panel's size written in as 560 by 760. The one home of
the panel sizes is `panels` in `main.go`, which the desktop package cannot import, so the test holds a
second copy that nothing keeps in step. It has already drifted in meaning: 560 by 760 is now About's
size, while Settings opens at 900 by 760 (FR-625), so the larger of the two jumps is not the one this
test makes.

Cost of leaving it: low. The test still proves what it was written for, that the window is put back
where it was placed after any larger window; a panel size changed in `main.go` breaks nothing here.
It only stops describing the sizes the application really uses. Resolving it means giving the panel
sizes a home the desktop tests can read (the `internal/product` package, say) or making the test
exercise both sizes read from there. Blocked on nothing but a machine to run it on, since the test
builds for Linux and macOS only.

## Looks like debt, not worth touching

**The drag sends Wails an internal message.** `startDrag` calls `window.WailsInvoke('drag')`, the
message Wails' own drag regions send, rather than a documented call. It is the one way to hand a
press to the platform's own move loop without writing that loop again; on macOS Wails answers it
with `performWindowDragWithEvent` on the press it kept. The Wails version in `go.mod` pins it.
Check dragging by hand on every platform on any Wails upgrade.

**macOS hides the Dock icon after Wails shows it.** Wails 2.12.0 has its activation policy option
commented out and makes TimeRibbon a regular application as it finishes launching, so the switch to
an accessory is made afterwards, through the main queue AppKit serves only once launching is done.
It rests on that order inside Wails. Check `lsappinfo` reports `UIElement` on any Wails upgrade.

**macOS borrows Windows' drag distance.** macOS publishes no distance a press must move before it
becomes a drag, so `ribbon_darwin.go` uses Windows' 4 DIP as a named value (ruled by Oliver,
2026-09-28). There is nothing to read it from.

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
to refuse. `setup` builds for Windows only and is unit tested there; `installer` has no tests, since
every method on it acts on the machine.

**The browser opens through the desktop, not through Wails.** Wails' `BrowserOpenURL` answers no
error, so a desktop with no browser left the donation button doing nothing with nothing said.
`desktop.OpenInBrowser` calls `ShellExecute` on Windows, `open` on macOS and `xdg-open` on Linux,
each of which reports a refusal. The donation page and an offered update (FR-509) both open through
it; Settings or Help shows the refusal with the address. Moving it back to Wails would bring the
silence back.

**Nothing holds the ribbon on a display during a drag on macOS and Linux.** `KeepOnDisplays` does
nothing there, since neither AppKit nor the window manager offers a say while a drag lasts; a ribbon
left partly off every display is put back when the move ends, through the same `Service.Moved` as on
Windows. Writing a move loop of their own to match Windows would fight the desktop.

**The macOS build names a framework Wails needs.** `platform_darwin.go` links UniformTypeIdentifiers
because Wails' macOS half uses it and only the `wails` command adds it, which TimeRibbon does not
build with. Removing the line breaks the link, measured 2026-09-28.

**Linux has a tray icon of its own rather than a library's.** The StatusNotifierItem and its menu in
`desktop/tray*_linux.go` and `desktop/dbusmenu_linux.go` look like something a library would do.
`fyne.io/systray` was measured and rejected (ARCHITECTURE.md, Design decisions); going back to it
would lose a menu rebuilt as it opens.

**WebKit's DMABUF renderer is off on every Linux machine, not only NVIDIA's.** On NVIDIA's own driver
it drew a blank window (measured 2026-10-02), while on Mesa it works. Telling the drivers apart
would mean reading the GPU before GTK opens, to save a faster path the clocks never need; a value the
user sets in `WEBKIT_DISABLE_DMABUF_RENDERER` is kept, so the choice stays theirs.

**The scroll bar's thickness comes from the page.** It is the web engine's bar, which Windows' own
scroll bar metric does not describe, so the page is the only place that can measure it.
