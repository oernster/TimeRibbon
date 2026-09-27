// The setup page's screens, laid out from the very files it ships with a fake setup program behind
// it (setupPage.ts). What is asserted is the screen shown, the words on it, the button holding
// focus and the calls made: FR-801 (the route), FR-805 (the boxes), FR-806 (uninstall), FR-808 (a
// failure says why) and the opening focus of FR-809.

import { beforeEach, describe, expect, it } from 'vitest'
import {
  activeScreen,
  boxes,
  FakeSetup,
  focusedLabel,
  footerButton,
  footerLabels,
  fresh,
  installed,
  layPage,
  pageElement,
  settle,
  type SetupPage,
  type State,
} from './setupPage'

let setup: FakeSetup
let setupPage: SetupPage

beforeEach(() => {
  setup = new FakeSetup()
  setupPage = layPage(setup)
})

describe('each screen opens on the action it leads with (FR-801, FR-809)', () => {
  const screens: [string, State, string, string[], string][] = [
    ['Install', fresh, 'screen-install', ['Cancel', 'Install'], 'Install'],
    ['Update', { ...installed, route: 'update', installedVersion: '0.0.9' }, 'screen-update', ['Uninstall', 'Close', 'Update'], 'Update'],
    ['Go back', { ...installed, route: 'downgrade', installedVersion: '0.2.0' }, 'screen-update', ['Uninstall', 'Close', 'Go back'], 'Go back'],
    ['Installed', installed, 'screen-manage', ['Uninstall', 'Close', 'Reinstall', 'Repair'], 'Repair'],
    ['Uninstall', { ...installed, uninstall: true }, 'screen-uninstall', ['Cancel', 'Uninstall'], 'Uninstall'],
  ]
  it.each(screens)('the %s screen', (_name, state, screen, actions, lead) => {
    setupPage.route(state)
    expect(activeScreen()).toBe(screen)
    expect(footerLabels()).toEqual(actions)
    expect(focusedLabel()).toBe(lead)
  })

  it('the screen that offers to close a running copy, before any file is touched (FR-807)', async () => {
    setup.running = true
    setupPage.route(installed)
    footerButton('Repair').click()
    await settle()
    expect(activeScreen()).toBe('screen-running')
    expect(footerLabels()).toEqual(['Cancel', 'Close it and continue'])
    expect(focusedLabel()).toBe('Close it and continue')
    expect(setup.repairs).toBe(0)
    footerButton('Close it and continue').click()
    await settle()
    await settle()
    expect(setup.repairs).toBe(1)
  })
})

describe('the words on each route (FR-801)', () => {
  it('names an update as a flow from one version to the other, never in the heading', () => {
    setupPage.route({ ...installed, route: 'update', installedVersion: '0.0.9' })
    expect(pageElement('update-title').textContent).not.toMatch(/[0-9]/)
    expect(pageElement('update-from').textContent).toBe('v0.0.9')
    expect(pageElement('update-to').textContent).toBe('v0.1.0')
  })

  it('writes the product name only as the setup program gives it', () => {
    setupPage.route(fresh)
    expect(pageElement('install-title').textContent).toBe('Install Product')
    expect(pageElement('install-version').textContent).toContain('0.1.0')
  })
})

describe('the boxes (FR-805)', () => {
  it('offers the three boxes and the launch box as the reading gives them', () => {
    setupPage.route(fresh)
    expect(boxes('install-options')).toEqual([
      ['Add to the Start Menu', true],
      ['Add a Desktop shortcut', false],
      ['Start with Windows', false],
      ['Start Product when setup closes', true],
    ])
  })

  it('installs with the boxes as they stand, then starts the application and closes', async () => {
    setupPage.route(fresh)
    const inputs = pageElement('install-options').querySelectorAll('input')
    inputs[2].click()
    footerButton('Install').click()
    await settle()
    await settle()
    expect(setup.installs).toEqual([{ startMenu: true, desktop: false, startWithWindows: true }])
    expect(setup.launches).toBe(1)
    expect(setup.quits).toBe(1)
  })

  it('applies a box on the Installed screen the moment it changes', async () => {
    setupPage.route(installed)
    pageElement('manage-options').querySelectorAll('input')[1].click()
    await settle()
    expect(setup.applies).toEqual([{ startMenu: true, desktop: true, startWithWindows: false }])
  })

  it('puts a box back and says why when applying it fails', async () => {
    setup.refusal = 'the Desktop cannot be written'
    setupPage.route(installed)
    const desktop = pageElement('manage-options').querySelectorAll('input')[1]
    desktop.click()
    await settle()
    expect(desktop.checked).toBe(false)
    expect(activeScreen()).toBe('screen-error')
  })
})

describe('the progress screen offers nothing', () => {
  it('has no actions and no Licence while work runs, then ends in a verdict', async () => {
    let release = (): void => {}
    setup.hold = new Promise((resolve) => {
      release = resolve
    })
    setupPage.route({ ...installed, uninstall: true })
    footerButton('Uninstall').click()
    await settle()
    expect(activeScreen()).toBe('screen-progress')
    expect(footerLabels()).toEqual([])
    expect(pageElement('licence').hidden).toBe(true)
    release()
    await settle()
    expect(activeScreen()).toBe('screen-done')
    expect(footerLabels()).toEqual(['Close'])
    expect(focusedLabel()).toBe('Close')
    expect(pageElement('licence').hidden).toBe(false)
  })
})

describe('the Uninstall screen (FR-806)', () => {
  it('leaves forgetting the settings unticked', async () => {
    setupPage.route({ ...installed, uninstall: true })
    expect(boxes('uninstall-options')).toEqual([['Also forget my settings', false]])
    footerButton('Uninstall').click()
    await settle()
    expect(setup.uninstalls).toEqual([false])
  })

  it('forgets the settings when the box is ticked', async () => {
    setupPage.route({ ...installed, uninstall: true })
    pageElement('uninstall-options').querySelector('input')?.click()
    footerButton('Uninstall').click()
    await settle()
    expect(setup.uninstalls).toEqual([true])
  })

  it('is reachable from the Installed screen and Cancel returns there', () => {
    setupPage.route(installed)
    footerButton('Uninstall').click()
    expect(activeScreen()).toBe('screen-uninstall')
    footerButton('Cancel').click()
    expect(activeScreen()).toBe('screen-manage')
    expect(setup.quits).toBe(0)
  })

  it('closes setup on Cancel when opened with -uninstall', () => {
    setupPage.route({ ...installed, uninstall: true })
    footerButton('Cancel').click()
    expect(setup.quits).toBe(1)
  })
})

describe('a failure says why (FR-808)', () => {
  const refusal = 'Writing the files: unsafe path in payload: ../escaped.txt'
  const failures: [string, State, string][] = [
    ['an install', fresh, 'Install'],
    ['an update', { ...installed, route: 'update', installedVersion: '0.0.9' }, 'Update'],
    ['a repair', installed, 'Repair'],
    ['an uninstall', { ...installed, uninstall: true }, 'Uninstall'],
  ]
  it.each(failures)('%s that is refused shows the reason, the log and Close alone', async (_name, state, press) => {
    setup.refusal = refusal
    setupPage.route(state)
    footerButton(press).click()
    await settle()
    await settle()
    expect(activeScreen()).toBe('screen-error')
    expect(pageElement('screen-error').querySelector('h1')?.textContent).toBe('Something went wrong')
    expect(pageElement('error-msg').textContent).toBe(refusal)
    expect(pageElement('error-log').textContent).toContain(installed.logPath)
    expect(footerLabels()).toEqual(['Close'])
    footerButton('Close').click()
    expect(setup.quits).toBe(1)
  })
})

describe('the header', () => {
  it('shows the licence as a screen of its own, Back returning where it was', async () => {
    setupPage.route(installed)
    pageElement('licence').click()
    await settle()
    expect(activeScreen()).toBe('screen-licence')
    expect(pageElement('licence-text').textContent).toBe('the terms')
    footerButton('Back').click()
    expect(activeScreen()).toBe('screen-manage')
    expect(footerLabels()).toEqual(['Uninstall', 'Close', 'Reinstall', 'Repair'])
  })

  it('faces the theme toggle with the appearance it switches to', () => {
    setupPage.applyTheme('dark')
    expect(pageElement('theme-icon').getAttribute('src')).toBe('light-mode.png')
    pageElement('theme').click()
    expect(document.documentElement.getAttribute('data-theme')).toBe('light')
    expect(pageElement('theme-icon').getAttribute('src')).toBe('dark-mode.png')
    expect(pageElement('theme').getAttribute('aria-label')).toBe('Switch to dark')
  })
})
