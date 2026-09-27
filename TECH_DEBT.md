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

**The donation page opens through the shell, not through Wails.** Wails' `BrowserOpenURL` answers
no error, so a desktop with no browser left the button doing nothing with nothing said.
`desktop.OpenInBrowser` calls `ShellExecute`, which reports the refusal; Settings shows it with the
address. Moving it back to Wails would bring the silence back.

**The scroll bar's thickness comes from the page.** It is the web engine's bar, which Windows' own
scroll bar metric does not describe, so the page is the only place that can measure it.
