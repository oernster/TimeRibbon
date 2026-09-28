# <img width="128" height="128" alt="application-icon" src="https://github.com/user-attachments/assets/fc124b11-f3a8-467c-9922-0abb40e9871e" /> TimeRibbon

A simple ribbon of configurable world clocks, horizontal or vertical, for Windows, macOS and Linux.

> **Commercial licences available.** TimeRibbon is free and open source under the GPL-3.0. If those terms do not suit what you are building, such as a closed-source product, a commercial licence can be bought from me separately. It covers my own code; third-party libraries keep their own licences. See [commercial licensing](https://ernster.dev/commercial-licensing.html).

TimeRibbon answers one question at a glance: what time and what day is it where the people you talk
to are? It shows one clock per place you choose, each with its own local time, weekday and date,
in a small frameless ribbon you can put anywhere on any monitor.

## Who it is for

- Anyone who talks with friends, family or colleagues in other time zones and wants their time and
  day in view without opening anything.
- Anyone with several monitors who wants the clocks on one of them, left where they put it.
- Anyone on Windows 10 or 11, on macOS 12 or later with Apple Silicon or on a Linux desktop that
  runs Flatpaks.

## Who it is not for

- Anyone wanting a calendar, a meeting planner, reminders or alarms. It shows the time; it plans
  nothing.
- Anyone wanting to track people. A clock is a place, never a person.
- Anyone on an Intel Mac: the macOS build is for Apple Silicon only.
- Anyone wanting seconds.
- Anyone wanting a 12-hour clock beside a 24-hour one: the format is one choice for every clock.
- Anyone wanting to arrange the clocks by hand: their order follows the time.

## What it does

- **Shows each place's own time and date.** Every clock shows its place, its local time, its
  weekday and date there, plus its zone's abbreviation (such as EDT) or its offset from UTC where
  the zone has no letters (such as UTC+5:45). Daylight saving follows the time zone rules
  themselves; nothing is set by hand.
- **Runs east from Greenwich.** The clocks start at Greenwich and go east round the world: London,
  then Berlin, Tokyo and Melbourne, with New York last, since places behind UTC are reached last.
  The order is worked out afresh at every refresh, so a change of daylight saving moves a clock
  where it moves its offset. Clocks keeping the same time stay in the order they were added.
- **Finds places by city, zone or country.** Search the 418 zones of the tz database by name, zone
  id or country; label a clock whatever you like, up to 32 characters (a city with no zone of its
  own, such as Manchester, takes its zone's clock and your label).
- **Stays out of the way.** The ribbon has no title bar, no border and no taskbar or Dock button;
  its icon lives in the notification area on Windows, the menu bar on macOS and the system tray on
  Linux. The icon's menu and the ribbon's own right-click menu both add a clock, open Settings,
  choose digital or analogue, choose a colour scheme, choose horizontal or vertical, centre the
  ribbon on an edge, turn Always on top on or off, open Help (About and Licence) and exit. The icon's
  menu also shows or hides the ribbon. Hiding it leaves TimeRibbon running; only Exit ends it.
- **Answers the icon the way each desktop expects.** On Windows a left click on the icon shows or
  hides the ribbon and a right click opens the menu. On macOS a click opens the menu, as every menu
  bar icon does. On Linux the tray decides: on Ubuntu a click opens the menu and a double click
  shows or hides the ribbon. On Windows, Alt+F4 on the ribbon hides it too.
- **Goes where you put it.** Drag the ribbon by any empty part onto any monitor. On Windows it is
  kept wholly on a display while it moves; on macOS and Linux one left partly off every display is
  put back when the drag ends. It opens there next time. When that monitor is gone, it opens on the
  primary one, at the edge its orientation sends it to. Position in either menu puts it flush
  against an edge of its display, centred along that edge: left or right for a vertical ribbon, top
  or bottom for a horizontal one. Choosing horizontal sends it to the top edge; choosing vertical
  sends it to the right.
- **Re-centres when its length changes.** Adding or removing a clock, a notice appearing or going
  and a change of style or size all change the ribbon's length. When that happens the ribbon
  centres itself along its length on its display, keeping its position across; it opens there next
  time. Otherwise only a drag, Position or a change of orientation moves it.
- **Fits its clocks, then scrolls.** The ribbon is as long as its clocks until it reaches the edge
  of the display, then its clocks scroll rather than shrink or wrap. A plain mouse wheel moves a
  horizontal ribbon along.
- **Digital or analogue, large or small, five colour schemes, 12-hour or 24-hour, horizontal or
  vertical, light or dark.** Every choice applies at once, with no Save step. The schemes are
  Classic, Neon (dark, with glowing digits and hands), Ocean, Sunset and Forest; each but Neon
  follows the light or dark theme. Style, colour and orientation are in the menus; size, time format
  and theme are in Settings. Small clocks suit a small screen such as a 13 inch laptop. The theme
  can follow the system's.
- **Keeps your clocks in one readable file,** `settings.json`, written whole or not at all (where
  it lives is in [Your settings](#your-settings)). A damaged file is kept aside under another name
  and never overwritten; a notice on the ribbon says so. A save that fails keeps the change in
  effect with a notice until a later save succeeds. One clock that cannot be read leaves the others
  working. The file is a promise: every later release of the same major version reads a file this
  release wrote to the same clocks and choices.
- **Starts when you sign in, when asked.** Off until you turn it on in Settings (or in setup on
  Windows). Each platform names it in its own words: Start with Windows, Open at Login on macOS,
  Start when I sign in on Linux.
- **Help, About and Licence.** About names the version, the author and every component this
  platform's build ships, with its licence. Licence shows the GPL-3.0 exactly as written, never
  wrapped again, its type sized so the widest line fits. Either reads itself slowly when it holds
  more than fits.

## What it does not do

- **It makes no network request.** The only address it knows is the donation page, which it hands
  to your browser when you press the button; TimeRibbon itself fetches nothing.
- **Its time zone rules are the ones built into it.** A government that changes its clocks after a
  release is shown correctly only from the next release that carries the new rules.
- **It never changes the system clock or time zone.** It reads them.
- **On Linux it draws through X11, never Wayland directly.** A window on Wayland may not choose
  where it stands, which the ribbon must; on a Wayland desktop it runs through XWayland.

## Built with

| Part | Choice |
|---|---|
| Backend | Go |
| Desktop shell | Wails v2: over WebView2 on Windows, WKWebView on macOS, WebKitGTK on Linux |
| Front end | React and TypeScript, built with Vite |
| Time zone rules | the tz database, built into the executable through Go's `time/tzdata` |
| The desktop below the window | Win32 on Windows; AppKit on macOS; GTK 3 with a D-Bus tray icon on Linux |
| Delivery | a setup program on Windows, a signed and notarised DMG on macOS, a Flatpak on Linux |

## Getting it

### Windows

Download `TimeRibbonSetup.exe` and run it. Everything it writes is for your own Windows account, so
it never asks for administrator rights. It installs under `%LOCALAPPDATA%\Programs\TimeRibbon` and
offers a Start Menu entry, a Desktop shortcut and Start with Windows.

To remove it, use the Apps list in Windows Settings. Your clocks stay unless you tick **Also forget
my settings**.

### macOS

Download `TimeRibbon.dmg`, open it and drag TimeRibbon to Applications. It is signed with a
Developer ID and notarised by Apple, so Gatekeeper lets it open.

To remove it, turn off Open at Login in its Settings first, then move it from Applications to the
Bin. Your clocks stay in the folder named below until you delete it.

### Linux

Download `timeribbon.flatpak` and install it for your own account. It runs on the GNOME 50 runtime,
which the bundle names Flathub as the source of:

```bash
flatpak install --user timeribbon.flatpak
```

```bash
flatpak run uk.codecrafter.TimeRibbon
```

To remove it, turn off Start when I sign in in its Settings first, then:

```bash
flatpak uninstall --user uk.codecrafter.TimeRibbon
```

### Your settings

| Platform | Folder holding `settings.json` and the log |
|---|---|
| Windows | `%APPDATA%\TimeRibbon` |
| macOS | `~/Library/Application Support/TimeRibbon` |
| Linux (the Flatpak) | `~/.var/app/uk.codecrafter.TimeRibbon/config/TimeRibbon` |

## Testing

On Windows the gate checks formatting, runs vet and staticcheck, runs every Go test and the front
end's lint, type check and tests, then holds each package to its coverage floor:

```powershell
./test.ps1
```

[TESTING.md](TESTING.md) says what each figure means, how the Linux and macOS code is tested on its
own machine and what only a person can check.

## Building

| Platform | Command | Output |
|---|---|---|
| Windows | `./build.ps1` | `build/bin/TimeRibbon.exe` and `dist-installer/TimeRibbonSetup.exe` |
| macOS | `bash builddmg.sh` | `TimeRibbon.dmg` |
| Linux | `bash build_flatpak.sh` | `timeribbon.flatpak` and an install for your account |

`build.ps1` runs the gate first. [DEVELOPMENT.md](DEVELOPMENT.md) sets up each machine from
nothing. [ARCHITECTURE.md](ARCHITECTURE.md) explains the layering and the reasoning behind each
decision. [TECH_DEBT.md](TECH_DEBT.md) lists what is still open, what is deliberately left and what
only looks like debt.

## Supporting TimeRibbon

TimeRibbon is free and stays free: there is no paid tier, no licence key and no feature held back
behind a donation. If it earns its place on your screen, a donation is welcome. The same button sits
at the foot of Settings in the application; pressing it hands the donation page to your browser.

<a href="https://www.paypal.com/ncp/payment/THUS4KZ5GECH8"><img src="docs/donate.png" alt="Donate to TimeRibbon" width="120"></a>

## Licence

GNU General Public License, version 3: see [LICENSE](LICENSE). The application shows the same text
under Help, then Licence.

Commercial licences are available: see [commercial licensing](https://ernster.dev/commercial-licensing.html).
