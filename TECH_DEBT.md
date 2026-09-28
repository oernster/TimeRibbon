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

There is no open technical debt.

## Looks like debt, not worth touching

**The drag sends Wails an internal message.** `startDrag` calls `window.WailsInvoke('drag')`, the
message Wails' own drag regions send, rather than a documented call. It is the one way to hand a
press to the platform's own move loop without writing that loop again; on macOS Wails answers it
with `performWindowDragWithEvent` on the press it kept. The Wails version in `go.mod` pins it.
Check it by hand on every platform on any Wails upgrade (TESTING.md, M-2).

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
to refuse. The portable half of `setup` is unit tested; `installer` has no tests, since every method
on it acts on the machine.

**The donation page opens through the desktop, not through Wails.** Wails' `BrowserOpenURL` answers
no error, so a desktop with no browser left the button doing nothing with nothing said.
`desktop.OpenInBrowser` calls `ShellExecute` on Windows, `open` on macOS and `xdg-open` on Linux,
each of which reports a refusal; Settings shows it with the address. Moving it back to Wails would
bring the silence back.

**Nothing holds the ribbon on a display during a drag on macOS and Linux.** `KeepOnDisplays` does
nothing there, since neither AppKit nor the window manager offers a say while a drag lasts; a ribbon
left partly off every display is put back when the move ends, through the same `Service.Moved` as on
Windows. Writing a move loop of their own to match Windows would fight the desktop.

**The macOS build names a framework Wails needs.** `platform_darwin.go` links UniformTypeIdentifiers
because Wails' macOS half uses it and only the `wails` command adds it, which TimeRibbon does not
build with. Removing the line breaks the link, measured 2026-09-28.

**Linux has a tray icon of its own rather than a library's.** The StatusNotifierItem and its menu in
`desktop/tray*_linux.go` look like something a library would do. `fyne.io/systray` was measured and
rejected (ARCHITECTURE.md, Design decisions); going back to it would lose a menu rebuilt as it opens.

**The scroll bar's thickness comes from the page.** It is the web engine's bar, which Windows' own
scroll bar metric does not describe, so the page is the only place that can measure it.
