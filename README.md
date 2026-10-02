# <img width="128" height="128" alt="application-icon" src="https://github.com/user-attachments/assets/fc124b11-f3a8-467c-9922-0abb40e9871e" /> TimeRibbon

A simple ribbon of configurable world clocks, horizontal or vertical, for Windows, macOS and Linux.

> **Commercial licences available.** TimeRibbon is free and open source under the GNU General Public License, version 3 (GPL-3.0). If those terms do not suit what you are building, such as a closed-source product, a commercial licence can be bought from me separately. It covers my own code; third-party libraries keep their own licences. See [commercial licensing](https://ernster.dev/commercial-licensing.html).

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
  id or country. What you type matches the start of a word, so `l` finds London but never Adelaide;
  places whose name begins with it come first. Label a clock whatever you like, up to 32 characters
  (a city with no zone of its own, such as Manchester, takes its zone's clock and your label).
  Settings lists your clocks, where each can be renamed, moved to another place or removed; removing
  one asks first. The search stays open beneath them, with the Add clock button beside its box.
- **Stays out of the way.** The ribbon has no title bar, no border and no taskbar or Dock button;
  its icon lives in the notification area on Windows, the menu bar on macOS and the system tray on
  Linux. The icon's menu and the ribbon's own right-click menu both add a clock, open Settings,
  choose digital or analogue, choose a colour scheme, choose horizontal or vertical, centre the
  ribbon on an edge, turn Always on top, Pin ribbon and Sun map on or off, open Help (About, Licence
  and Check for updates) and exit. The icon's menu also shows or hides the ribbon; the ribbon's own
  menu hides it. Hiding it leaves TimeRibbon running; only Exit ends it.
- **Answers the icon the way each desktop expects.** On Windows a left click on the icon shows or
  hides the ribbon and a right click opens the menu. On macOS a click opens the menu, as every menu
  bar icon does. On Linux the tray decides: on Ubuntu a click opens the menu and a double click
  shows or hides the ribbon. On Windows, Alt+F4 on the ribbon hides it too. Launching TimeRibbon
  again while it runs shows or hides the ribbon in the same way, so a single launcher button, such
  as a Stream Deck's Open action pointed at TimeRibbon, does both.
- **Gets out of the way when you want it to.** Untick Pin ribbon in either menu and the ribbon
  shrinks to a thin tab in the scheme's accent a second after the pointer leaves it. Rest the pointer
  on the tab for a moment and the ribbon opens again, without taking the keyboard from what you are
  typing in. Unpinned, it stays above other windows so the tab is never lost. The tab only happens
  against an edge: dragged away from every edge the ribbon stays in full; dropped near an edge it
  snaps flush and waits as a tab again. It is pinned unless you choose otherwise.
- **Shows where it is day.** Tick Sun map in either menu and the ribbon gains a handle in a lane of
  its own along its inner side, so the ribbon grows a little deeper and the handle covers no clock.
  Its arrow slides a world map out below a horizontal ribbon (above one at the bottom of the screen)
  or beside a vertical one, then back again. The map is lit where it is day, dark with city lights
  where it is night, each clock's city marked in red. Nearby cities'
  names move aside rather than print over each other. On Windows the desktop shows round the map;
  on macOS and Linux the window stays a rectangle round the ribbon and its map. The pictures are
  NASA's Blue Marble and Black Marble, built in, so nothing is fetched to draw them.
- **Goes where you put it.** Drag the ribbon by any empty part onto any monitor. On Windows it is
  kept wholly on a display while it moves; on macOS and Linux one left partly off every display is
  put back when the drag ends. It opens there next time. When that monitor is gone, it opens on the
  primary one, at the edge its orientation sends it to. Position in either menu puts it flush
  against an edge of its display, centred along that edge: left or right for a vertical ribbon, top
  or bottom for a horizontal one. Choosing horizontal sends it to the top edge; choosing vertical
  sends it to the right.
- **Re-centres when its length changes.** Adding or removing a clock, a notice appearing or going
  and a change of style, size or scale all change the ribbon's length. When that happens the ribbon
  centres itself along its length on its display, keeping its position across; it opens there next
  time. Otherwise only a drag, Position or a change of orientation moves it.
- **Fits its clocks, then scrolls.** The ribbon is as long as its clocks until it reaches the edge
  of the display, then its clocks scroll rather than shrink or wrap. A plain mouse wheel moves a
  horizontal ribbon along. Each clock is as wide as the widest time and date its size, style and
  formats can show in the font it is drawn in, so no time or date is ever cut short.
- **Digital or analogue, large or small, ten colour schemes, 12-hour or 24-hour, five date
  formats, horizontal or vertical, light or dark.** Every choice applies at once, with no Save
  step. The schemes are Classic, Neon (glowing digits and hands when dark), Ocean, Sunset, Forest,
  Amber, Ruby, Indigo, Berry and Contrast (black and white); each follows the light or dark theme.
  The date reads "Monday, 28 September" or "Monday, September 28" in words; in numbers it reads
  "Mon 28/09/2026", "Mon 09/28/2026" or "Mon 2026/09/28". Settings offers every choice the menus
  do, from the same items, so the two always agree; size, time format, date format, theme and
  opacity are in Settings alone. Small clocks suit a small screen such as a 13 inch laptop. The
  theme can follow the system's. Settings opens wide with its choices side by side; it is as tall as
  its content, up to the size of the display, where it scrolls instead. The Opacity slider draws
  the whole window from 20 to 100 percent opaque, so the ribbon can sit over other work without
  hiding it. The grip in the ribbon's corner resizes the clocks by hand, from 75 to 200 percent of
  either size, everything in them together; a double-click on it returns them to their own size.
- **Keeps your clocks in one readable file,** `settings.json`, written whole or not at all (where
  it lives is in [Your settings](#your-settings)). A damaged file is kept aside under another name
  and never overwritten; a notice on the ribbon says so. A save that fails keeps the change in
  effect with a notice until a later save succeeds. One clock that cannot be read leaves the others
  working. The file is a promise: a file 1.0.0 wrote still reads to the same clocks and choices. A
  later release may add keys but never renames, drops or changes the meaning of one 1.0.0 wrote.
- **Starts when you sign in, when asked.** Off until you turn it on in Settings (or in setup on
  Windows). Each platform names it in its own words: Start with Windows, Open at Login on macOS,
  Start when I sign in on Linux.
- **Help, About and Licence.** About names the version, the author and every component this
  platform's build ships, with its licence. Licence shows the GPL-3.0 exactly as written, never
  wrapped again, its type sized so the widest line fits. Either reads itself slowly when it holds
  more than fits.
- **Tells you when a new release is out.** A few seconds after it starts, then once a day, it asks
  GitHub for the latest release. Only when that is newer than yours does the ribbon show it, with
  Download (the file for your platform opened in your browser, else the release's page), Skip this
  version and Later. A skipped version is never offered again unasked. Help's Check for updates
  asks at any time and always answers, including when GitHub cannot be reached.

## What it does not do

- **Its one network request is the update check.** It asks GitHub for the latest release when it
  starts and once a day, sending nothing about you or your clocks. Nothing else it does touches the
  network: the donation page and a release's download are handed to your browser, never fetched by
  TimeRibbon itself.
- **It never installs an update by itself.** Download opens the file in your browser; installing it
  is yours to do.
- **On Windows its time zone rules are the ones built into it.** Windows keeps no rules TimeRibbon
  can read, so a government that changes its clocks after a release is shown correctly there only
  from the next release carrying the new rules. On macOS and Linux it reads the system's own rules
  first, which the system's updates keep current; the built-in rules stand in only for a zone the
  system lacks.
- **It never changes the system clock or time zone.** It reads them.
- **On Linux it draws through X11, never Wayland directly.** A window on Wayland may not choose
  where it stands, which the ribbon must; on a Wayland desktop it runs through XWayland. It also
  turns off WebKit's DMABUF renderer, which draws a blank window on NVIDIA's own driver, unless
  `WEBKIT_DISABLE_DMABUF_RENDERER` is already set, in which case that value is left as it is.

## Built with

| Part | Choice |
|---|---|
| Backend | Go |
| Desktop shell | Wails v2: over WebView2 on Windows, WKWebView on macOS, WebKitGTK on Linux |
| Front end | React and TypeScript, built with Vite |
| Time zone rules | the tz database, built into the executable through Go's `time/tzdata`; on macOS and Linux the system's own copy is read first |
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

Choose Exit from its menu before installing a newer release over it. A copy left running keeps
running the old release; launching the new one then only shows or hides that old ribbon.

To remove it, turn off Start when I sign in in its Settings and choose Exit first, then:

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
end's lint, type check and tests, then holds the domain and application layers to 100 percent
coverage and most other packages to a floor at the coverage each reaches:

```powershell
./test.ps1
```

[TESTING.md](TESTING.md) says what each figure means, how the macOS and Linux code is tested on its
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
only looks like debt. [`DECISIONS-TRADEOFFS.md`](DECISIONS-TRADEOFFS.md) sets out the decisions
TimeRibbon rests on, with what each one gains and what it costs.

## Supporting TimeRibbon

TimeRibbon is free and stays free: there is no paid tier, no licence key and no feature held back
behind a donation. If it earns its place on your screen, a donation is welcome. The same button sits
at the foot of Settings in the application; pressing it hands the donation page to your browser.

<a href="https://www.paypal.com/ncp/payment/THUS4KZ5GECH8"><img src="docs/donate.png" alt="Donate to TimeRibbon" width="120"></a>

## Licence

GNU General Public License, version 3: see [LICENSE](LICENSE). The application shows the same text
under Help, then Licence.

Commercial licences are available: see [commercial licensing](https://ernster.dev/commercial-licensing.html).
