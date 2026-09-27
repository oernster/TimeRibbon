# <img width="128" height="128" alt="application-icon" src="https://github.com/user-attachments/assets/fc124b11-f3a8-467c-9922-0abb40e9871e" /> TimeStrip

A simple strip of configurable world clocks, horizontal or vertical.

TimeStrip answers one question at a glance: what time and what day is it where the people you talk
to are? It shows one clock per place you choose, each with its own local time, weekday and date,
in a small frameless strip you can put anywhere on any monitor.

## Who it is for

- Anyone who talks with friends, family or colleagues in other time zones and wants their time and
  day in view without opening anything.
- Anyone with several monitors who wants the clocks on one of them, left where they put it.

## Who it is not for

- Anyone wanting a calendar, a meeting planner, reminders or alarms. It shows the time; it plans
  nothing.
- Anyone wanting to track people. A clock is a place, never a person.
- Anyone on macOS or Linux. It is built for Windows 10 and 11 only.
- Anyone wanting seconds.
- Anyone wanting a 12-hour clock beside a 24-hour one: the format is one choice for every clock.

## What it does

- **Shows each place's own time and date.** Every clock shows its place, its local time, its
  weekday and date there, plus its zone's abbreviation (such as EDT) or its offset from UTC where
  the zone has no letters (such as UTC+5:45). Daylight saving follows the time zone rules
  themselves; nothing is set by hand.
- **Finds places by city, zone or country.** Search the 418 zones of the tz database by name, zone
  id or country; label a clock whatever you like (a city with no zone of its own, such as
  Manchester, takes its zone's clock and your label).
- **Stays out of the way.** The strip has no title bar, no border and no taskbar button. Its icon in
  the notification area shows or hides it, adds a clock, opens Settings, turns Always on top on or
  off and quits. Closing the strip (Alt+F4) hides it; only Exit ends it.
- **Goes where you put it.** Drag the strip by any empty part onto any monitor; it stays wholly on
  that display. It opens there next time. When that monitor is gone, it opens on the primary one.
- **Fits its clocks, then scrolls.** The strip is as long as its clocks until it reaches the edge
  of the display, then its clocks scroll rather than shrink or wrap. A plain mouse wheel moves a
  horizontal strip along.
- **Digital or analogue, 12-hour or 24-hour, horizontal or vertical, light or dark.** Every choice
  applies at once, with no Save step. The theme can follow Windows.
- **Keeps your clocks in one readable file.** `%APPDATA%\TimeStrip\settings.json`, written whole
  or not at all. A damaged file is kept aside under another name and never overwritten; one clock
  that cannot be read leaves the others working. From 1.0.0 the file is a promise: every later 1.x
  release reads a 1.0.0 file to the same clocks and choices.
- **Starts with Windows when asked.** Off until you turn it on in Settings or in setup.
- **Help, About and Licence** from the notification-area icon or a right-click on the strip.

## What it does not do

- **It makes no network request.** The only address it knows is the donation page, which it hands
  to your browser when you press the button; TimeStrip itself fetches nothing.
- **Its time zone rules are the ones built into it.** A government that changes its clocks after a
  release is shown correctly only from the next release that carries the new rules.
- **It never changes the Windows clock or time zone.** It reads them.

## Built with

| Part | Choice |
|---|---|
| Backend | Go |
| Desktop shell | Wails v2 over WebView2 |
| Front end | React and TypeScript, built with Vite |
| Time zone rules | the tz database, built into the executable through Go's `time/tzdata` |
| Setup program | a second Wails application in the same module |

## Getting it

Download `TimeStripSetup.exe` and run it. Everything it writes is for your own Windows account, so
it never asks for administrator rights. It installs under `%LOCALAPPDATA%\Programs\TimeStrip` and
offers a Start Menu entry, a Desktop shortcut and Start with Windows.

To remove it, use the Apps list in Windows Settings. Your clocks stay unless you tick **Also forget
my settings**.

## Testing

The gate checks formatting, runs vet and staticcheck, runs every Go test and the front end's lint,
type check and tests, then holds each package to its coverage floor:

```powershell
./test.ps1
```

[TESTING.md](TESTING.md) says what each figure means and what only a person can check.

## Building

```powershell
./build.ps1
```

It runs the gate first, then writes the application to `build/bin/TimeStrip.exe` and the setup
program to `dist-installer/TimeStripSetup.exe`. [DEVELOPMENT.md](DEVELOPMENT.md) sets up a machine
from nothing. [ARCHITECTURE.md](ARCHITECTURE.md) explains the layering and the reasoning behind each
decision. [TECH_DEBT.md](TECH_DEBT.md) lists what is still open, what is deliberately left and what
only looks like debt.

## Supporting the project

TimeStrip is free and stays free: there is no paid tier, no licence key and no feature held back
behind a donation. A donate button sits at the foot of Settings. Pressing it hands the donation page
to your browser; TimeStrip itself sends nothing anywhere.

## Licence

GNU General Public License, version 3: see `LICENSE`. The application shows the same text under
Help, then Licence.
