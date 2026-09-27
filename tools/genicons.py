"""Generate TimeStrip's icons from the master artwork in assets/, ported from BridgeTalk's.

What comes out, every file from one of the masters so the masters stay the one home:

  build/windows/icon.ico               the multi-size Windows icon both executables wear; the
                                       taskbar, the tray and both shortcuts read it out of the binary
  build/appicon.png                    the artwork Wails keeps beside it
  installer/frontend/dist/icon.png     the setup window's header mark
  installer/frontend/dist/light-mode.png, dark-mode.png
                                       the setup window's theme toggle: the sun and the moon
  frontend/src/assets/donate.png       the donate mark at the foot of Settings, from donate.png

The setup page has no build step, so it loads each file as it finds it; shipping the masters there
would put megabytes behind a badge. Each is written at about twice the size it is drawn at, crisp on
a high-density display without carrying detail nothing shows.

Run it when a master changes:

    python tools/genicons.py

It is not part of the build. The output is committed, so a clone needs neither Python nor Pillow
to build.
"""

from __future__ import annotations

import pathlib
import sys

try:
    from PIL import Image
except ImportError:  # pragma: no cover - a plain message beats a traceback
    sys.exit("Pillow is required: python -m pip install pillow")

REPO = pathlib.Path(__file__).resolve().parent.parent
MASTERS = REPO / "assets"
APP_MASTER = MASTERS / "application-icon.png"

# TOGGLE_MASTERS are the sun and the moon, named for the appearance each switches to.
TOGGLE_MASTERS = ("light-mode.png", "dark-mode.png")

# ICO_SIZES are the sizes Windows chooses between: the tray and menu sizes, the taskbar and
# shortcut sizes, then the large one Explorer uses in its biggest view. Leaving one out makes
# Windows scale a neighbour, which looks soft.
ICO_SIZES = [(16, 16), (24, 24), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)]

# APPICON_SIZE is the square Wails expects its build/appicon.png to be.
APPICON_SIZE = 1024

# HEADER_SIZE is about twice the header mark's 126 pixels; TOGGLE_SIZE about twice the toggle's 44.
HEADER_SIZE = 256
TOGGLE_SIZE = 96

# The donate mark is wide artwork drawn at a button's height, not an icon: it is cropped to its
# artwork and scaled by height alone. DONATE_DRAWN is the height the Settings foot draws it at;
# the render is four times that, so it stays crisp under display scaling.
DONATE_MASTER = MASTERS / "donate.png"
DONATE_DRAWN = 32
DONATE_HEIGHT = 4 * DONATE_DRAWN
DONATE_TARGETS = (REPO / "frontend" / "src" / "assets" / "donate.png",)

ICO_TARGET = REPO / "build" / "windows" / "icon.ico"
APPICON_TARGET = REPO / "build" / "appicon.png"
SETUP = REPO / "installer" / "frontend" / "dist"
HEADER_TARGET = SETUP / "icon.png"


def squared(master: pathlib.Path) -> Image.Image:
    """Open a master, trim any transparent margin and centre it on a transparent square."""
    image = Image.open(master).convert("RGBA")
    box = image.getbbox()
    if box is not None:
        image = image.crop(box)
    side = max(image.width, image.height)
    square = Image.new("RGBA", (side, side), (0, 0, 0, 0))
    square.paste(image, ((side - image.width) // 2, (side - image.height) // 2), image)
    return square


def write_png(image: Image.Image, size: int, target: pathlib.Path) -> None:
    """Write image scaled to a size-pixel square and say what was written."""
    target.parent.mkdir(parents=True, exist_ok=True)
    image.resize((size, size), Image.LANCZOS).save(target, "PNG", optimize=True)
    print(f"{target.relative_to(REPO).as_posix():<42} {size:>5} px {target.stat().st_size:>9,} bytes")


def donate_mark() -> Image.Image:
    """The donate artwork cropped to its pixels and scaled to DONATE_HEIGHT, keeping its shape."""
    image = Image.open(DONATE_MASTER).convert("RGBA")
    box = image.getbbox()
    if box is not None:
        image = image.crop(box)
    width = round(image.width * DONATE_HEIGHT / image.height)
    return image.resize((width, DONATE_HEIGHT), Image.LANCZOS)


def main() -> int:
    for master in (APP_MASTER, DONATE_MASTER, *(MASTERS / name for name in TOGGLE_MASTERS)):
        if not master.exists():
            sys.exit(f"no master artwork at {master}")

    app = squared(APP_MASTER)
    ICO_TARGET.parent.mkdir(parents=True, exist_ok=True)
    app.save(ICO_TARGET, "ICO", sizes=ICO_SIZES)
    sizes = ", ".join(str(width) for width, _ in ICO_SIZES)
    print(f"{ICO_TARGET.relative_to(REPO).as_posix():<42} {sizes} {ICO_TARGET.stat().st_size:>9,} bytes")
    write_png(app, APPICON_SIZE, APPICON_TARGET)
    write_png(app, HEADER_SIZE, HEADER_TARGET)
    for name in TOGGLE_MASTERS:
        write_png(squared(MASTERS / name), TOGGLE_SIZE, SETUP / name)
    # One render written to every destination, so no two copies can drift.
    mark = donate_mark()
    for target in DONATE_TARGETS:
        target.parent.mkdir(parents=True, exist_ok=True)
        mark.save(target, "PNG", optimize=True)
        print(f"{target.relative_to(REPO).as_posix():<42} {mark.width}x{mark.height} {target.stat().st_size:>9,} bytes")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
