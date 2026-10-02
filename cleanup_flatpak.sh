#!/usr/bin/env bash
# cleanup_flatpak.sh: uninstalls the TimeRibbon Flatpak and removes its build artefacts. Run from the
# repo root. Ported from PigeonPost's.
#
# Scoped to the Flatpak's own artefacts plus the Start at sign-in entry it may have written, which
# would otherwise go on trying to start an app that is gone. It deliberately does not touch what
# the other build paths produce, so they stay independent. Settings are left alone: they live in
# ~/.var/app/<id> and are the user's.
set -euo pipefail

eval "$(go run ./tools/identity)"
AUTOSTART="${XDG_CONFIG_HOME:-${HOME}/.config}/autostart/${APP_ID}.desktop"

bold=$(tput bold 2>/dev/null || true)
reset=$(tput sgr0 2>/dev/null || true)
section() { echo; echo "${bold}=== $* ===${reset}"; }

# Uninstalling leaves a running copy running, holding the single-instance lock, so the next
# install's first launch would only toggle it and exit (FR-506). flatpak kill fails when nothing
# runs, hence the check.
section "Stopping a running ${APP_NAME}"
if flatpak ps --columns=application | grep -x "${APP_ID}" > /dev/null; then
    flatpak kill "${APP_ID}"
    echo "  Stopped."
else
    echo "  Not running, skipping."
fi

section "Uninstalling ${APP_ID}"
if flatpak list --user --app --columns=application | grep -qx "${APP_ID}"; then
    flatpak uninstall --user -y "${APP_ID}"
    echo "  Uninstalled."
else
    echo "  Not installed, skipping."
fi

section "Removing the Start at sign-in entry"
if [[ -f "${AUTOSTART}" ]]; then
    rm -f "${AUTOSTART}"
    echo "  Removed ${AUTOSTART}"
else
    echo "  None, skipping."
fi

section "Removing flatpak build artefacts"
rm -f "${BIN_NAME}.flatpak"
# Go writes its module cache read-only. A build that fails leaves its build directory behind with
# that cache in it, before the manifest's own chmod has run, so plain rm cannot remove it.
[[ -d .flatpak-builder ]] && chmod -R u+w .flatpak-builder
rm -rf .flatpak-build .flatpak-repo .flatpak-builder build/linux
rm -f "${APP_ID}.yml"
rm -rf packaging/
echo "  Done."

echo
echo "${bold}Purge complete.${reset}"
