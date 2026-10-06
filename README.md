# <img width="128" height="128" alt="application-icon" src="https://github.com/user-attachments/assets/fc124b11-f3a8-467c-9922-0abb40e9871e" /> TimeRibbon

A simple ribbon of configurable world clocks, horizontal or vertical, for Windows, macOS and Linux.

> **Commercial licences available.** TimeRibbon is free and open source under the GNU General Public License, version 3 (GPL-3.0). If those terms do not suit what you are building, such as a closed-source product, a commercial licence can be bought from me separately. It covers my own code; third-party libraries keep their own licences. See [commercial licensing](https://ernster.dev/commercial-licensing.html).

What time and what day is it where the people you talk to are? TimeRibbon shows one clock per place
you choose, each with its local time, weekday and date, in a small frameless ribbon you can put
anywhere on any monitor.

## Who it is for

- Anyone who talks with people in other time zones and wants their time and day in view.
- Anyone on Windows 10 or 11, macOS 12 or later on Apple Silicon or a Linux desktop that runs
  Flatpaks.

## Who it is not for

- Anyone wanting a calendar, meeting planner, reminders or alarms. It shows the time; it plans
  nothing.
- Anyone wanting to track people: a clock is a place, never a person.
- Anyone on an Intel Mac.
- Anyone wanting seconds, a 12-hour clock beside a 24-hour one or clocks arranged by hand.

## What it does

- **Each place's own time and date,** with its zone's abbreviation (EDT) or its offset where the zone
  has no letters (UTC+5:45). Daylight saving follows the tz rules; nothing is set by hand.
- **Runs east from Greenwich:** London, Berlin, Tokyo, Melbourne, then New York, since places behind
  UTC come last. The order is worked out at every refresh.
- **Finds places by city, zone or country** among the 418 zones of the tz database, matching the start
  of a word, so `l` finds London but never Adelaide. A label can be anything up to 32 characters, so
  Manchester can wear London's clock.
- **Stays out of the way.** No title bar, border or taskbar button; its icon sits in the notification
  area, menu bar or tray. The icon's menu and the ribbon's right-click menu offer Add clock,
  Settings, style, colour, orientation, Position, Always on top, Pin ribbon, Sun map, Help, Hide and
  Exit; while the ribbon is hidden the icon offers Show in place of Hide. Hiding the ribbon leaves
  TimeRibbon running. Launching it again shows or hides the ribbon, so one
  launcher button (a Stream Deck's, say) does both.
- **Unpinned, it waits as a thin tab** against the edge it stands on, opening when the pointer rests
  on it without taking the keyboard. Dragged away from every edge it stays in full; dropped near one
  it snaps flush.
- **A sun map** slides out from a handle beside the ribbon: lit where it is day, city lights where
  it is night, each clock's city marked in red. The pictures are NASA's Blue Marble and Black Marble,
  built in, so nothing is fetched.
- **Goes where you put it** on any monitor and opens there next time; Position centres it on an
  edge, greying an edge that would leave it where it stands. It re-centres along its length when a
  clock or notice comes or goes.
- **Fits its clocks, then scrolls,** each cell as wide as the widest time and date its font can
  draw, so nothing is cut short.
- **Choices that apply at once:** digital or analogue, large or small, ten colour schemes, 12-hour or
  24-hour, five date formats, light, dark or the system's theme. Settings offers every menu choice
  too. The Opacity slider fades the ribbon's background from 20 to 100 percent while the clocks stay
  solid. The grip in the ribbon's corner resizes the clocks from 75 to 200 percent, growing from the
  ribbon's top-left corner as a window does; a double-click restores them.
- **Keeps your clocks in one readable file,** `settings.json`, written whole or not at all. A damaged
  file is kept aside and never overwritten; a failed save keeps the change with a notice. A file the
  first release wrote still reads to the same clocks.
- **Starts when you sign in,** when asked: Start with Windows, Open at Login on macOS, Start when I
  sign in on Linux.
- **Tells you when a new release is out,** with Download, Skip this version and Later. Help's Check for
  updates asks at any time.

## What it does not do

- **Its one network request is the update check,** sent at start and once a day with nothing about
  you or your clocks. The donation page and downloads are handed to your browser.
- **It never installs an update itself.**
- **On Windows it uses the time zone rules built into it,** since Windows keeps none it can read; a
  government's change of clocks shows correctly from the next release. macOS and Linux read the
  system's own rules first.
- **It never changes the system clock or time zone.**
- **On Linux it draws through X11,** XWayland on a Wayland desktop, since a Wayland window may not
  choose where it stands. WebKit's DMABUF renderer (blank on NVIDIA's own driver) and hardware
  acceleration are off, unless `WEBKIT_DISABLE_DMABUF_RENDERER` is already set.

## Built with

| Part | Choice |
|---|---|
| Backend | Go |
| Desktop shell | Wails v2: WebView2 on Windows, WKWebView on macOS, WebKitGTK on Linux |
| Front end | React and TypeScript, built with Vite |
| Time zone rules | the tz database built in through Go's `time/tzdata`; macOS and Linux read their own first |
| The desktop | Win32 on Windows; AppKit on macOS; GTK 3 with a D-Bus tray icon on Linux |
| The ribbon itself | [ribbonkit](https://github.com/oernster/ribbonkit), the placing, dragging, tab, tray, scaling and setup every ribbon shares, at one tag |
| Delivery | a setup program on Windows, a signed and notarised DMG on macOS, a Flatpak on Linux |

## Getting it

### Windows

Run `TimeRibbonSetup.exe`. It installs for your account alone under
`%LOCALAPPDATA%\Programs\TimeRibbon`, never asks for administrator rights and offers a Start Menu
entry, a Desktop shortcut and Start with Windows. Remove it from the Apps list; your clocks stay
unless you tick **Also forget my settings**.

### macOS

Open `TimeRibbon.dmg` and drag TimeRibbon to Applications. To remove it, turn off Open at Login, then
move it to the Bin.

### Linux

```bash
flatpak install --user timeribbon.flatpak
```

```bash
flatpak run uk.codecrafter.TimeRibbon
```

It runs on the GNOME 50 runtime from Flathub. Exit it before installing a newer release, since a
copy left running keeps the old one. To remove it, turn off Start when I sign in, Exit, then:

```bash
flatpak uninstall --user uk.codecrafter.TimeRibbon
```

### Your settings

| Platform | Folder holding `settings.json` and the log |
|---|---|
| Windows | `%APPDATA%\TimeRibbon` |
| macOS | `~/Library/Application Support/TimeRibbon` |
| Linux (the Flatpak) | `~/.var/app/uk.codecrafter.TimeRibbon/config/TimeRibbon` |

## Known issues

On Linux three problems are still open:

- **KDE Plasma at 150 percent:** the web view sometimes stayed smaller than its window (2 of 6
  launches). The ribbon is now shown only once sized; 16 of 16 launches have come up whole since,
  which does not yet prove the fault gone.
- **Under KWin** Settings comes out a little short and off centre, since KWin reports the work area in
  logical units.
- **With an NVIDIA RTX 3080 Ti** Settings may ghost while it scrolls. Hardware acceleration is now off;
  whether that ends it needs that machine to check.

## Testing

```powershell
./test.ps1
```

The gate checks formatting, vet and staticcheck, runs every Go test plus the front end's lint, type
check and tests, then holds the domain and application to 100 percent coverage and other packages to
measured floors. [TESTING.md](TESTING.md) has the figures.

## Building

| Platform | Command | Output |
|---|---|---|
| Windows | `./build.ps1` | `build/bin/TimeRibbon.exe` and `dist-installer/TimeRibbonSetup.exe` |
| macOS | `bash builddmg.sh` | `TimeRibbon.dmg` |
| Linux | `bash build_flatpak.sh` | `timeribbon.flatpak` and an install for your account |

`build.ps1` runs the gate first. [DEVELOPMENT.md](DEVELOPMENT.md) sets up each machine;
[ARCHITECTURE.md](ARCHITECTURE.md) explains the layering; [DECISIONS-TRADEOFFS.md](DECISIONS-TRADEOFFS.md)
gives the decisions with what each costs; [TECH_DEBT.md](TECH_DEBT.md) lists what is still open, what
is deliberately left and what only looks like debt.

## Supporting TimeRibbon

TimeRibbon is free and stays free: no paid tier, no licence key, no feature held back behind a
donation. If it earns its place on your screen, a donation is welcome. The same button sits at the
foot of Settings.

<a href="https://www.paypal.com/ncp/payment/THUS4KZ5GECH8"><img src="docs/donate.png" alt="Donate to TimeRibbon" width="120"></a>

## Licence

GNU General Public License, version 3: see [LICENSE](LICENSE), also shown under Help, then Licence.

Commercial licences are available: see [commercial licensing](https://ernster.dev/commercial-licensing.html).
