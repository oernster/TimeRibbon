# Technical debt

What is still open, what is deliberately left and what only looks like debt.

Every item is a behaviour-preserving internal concern; nothing here reverts a feature or changes what
the user sees. `ARCHITECTURE.md` and the structural tests are the authority on the invariants.

Open items are numbered sections, so a scan for `## <number>.` tells whether the file is clear. The
two standing sections at the end are unnumbered. A resolved item is deleted outright; history belongs
in the release notes. Debt in the ribbon's own code is recorded where that code lives, in
ribbonkit's TECH_DEBT.md.

There is no open technical debt.

## Looks like debt, not worth touching

Nothing of TimeRibbon's own; the ribbon's are in ribbonkit's TECH_DEBT.md.

## Not debt (do not "fix" these)

**The setup program holds no install logic.** TimeRibbon's `installer` is only its composition root:
it carries the payload and the pictures, then hands them to the kit's setup window, which hands every
act to the kit's install policy.

**The macOS build names no framework of its own.** Wails' macOS half needs UniformTypeIdentifiers,
which only the `wails` command adds; the kit's window names it, so TimeRibbon's `go build` links it
without a line here.
