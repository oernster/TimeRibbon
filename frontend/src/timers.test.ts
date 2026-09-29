// While the ribbon is shown its page schedules no periodic timer more frequent than once per minute;
// the self-reading cycle of a Help panel runs only while that panel is shown (NFR-P-4, FR-609).
//
// The page is every module main.tsx reaches, followed through the source's own imports; that takes
// in auto-scroll.js, which lives beside the setup page and is imported from there. The setup
// program's other scripts are not the ribbon's page: the ribbon never loads them, so the walk never
// reaches them. Test support (the fake bridge, the setup page's test layout, Vitest's setup file) is
// never shipped, so it is named below and held out of the page. Packages (React) and the runtime
// Wails injects are not this repository's source and are not read.
//
// Every setInterval, setTimeout and requestAnimationFrame call found must match a site on the lists
// below, with the delay it is written with. A new call anywhere fails here until it is added with the
// reason it cannot wake the page more often than once a minute.

import { describe, expect, it } from 'vitest'

// The sources are read as text the way setupPage.ts reads the setup page, so the test sees exactly
// the files the build bundles.
const sources = import.meta.glob<string>(['./**/*.{ts,tsx}', '../../installer/frontend/dist/*.js'], {
  query: '?raw',
  import: 'default',
  eager: true,
})

/** ENTRY is the module index.html loads. */
const ENTRY = 'main.tsx'

/** SCRIPT_EXTENSIONS are the forms an import may resolve to; any other import is an asset. */
const SCRIPT_EXTENSIONS = ['.ts', '.tsx', '.js']

/** TEST_SUPPORT is source that only tests import, with why it is not part of the page. */
const TEST_SUPPORT: Record<string, string> = {
  'fakeBridge.ts': 'stands in for Go in the suites',
  'setupPage.ts': 'lays out the setup page for its suites',
  'test-setup.ts': "Vitest's setup file",
}

const TIMER_APIS = ['setInterval', 'setTimeout', 'requestAnimationFrame'] as const
type TimerApi = (typeof TIMER_APIS)[number]

/** FRAME stands for the delay of a requestAnimationFrame call, which takes none. */
const FRAME = 'next frame'

interface TimerSite {
  file: string
  api: TimerApi
  delay: string
}

interface AllowedSite extends TimerSite {
  reason: string
}

/** HELP_CYCLE is the one module allowed a periodic timer: the self-reading cycle (FR-609). */
const HELP_CYCLE = '../../installer/frontend/dist/auto-scroll.js'

/** HELP_CYCLE_USERS are the only page modules that may start the cycle: Help, while it is mounted. */
const HELP_CYCLE_USERS = ['Help.tsx']

/** HELP_CYCLE_HOOK is the module that starts the cycle for a mounted element and ends it on unmount. */
const HELP_CYCLE_HOOK = 'autoScroll.ts'

const ALLOWED: AllowedSite[] = [
  {
    file: HELP_CYCLE,
    api: 'setInterval',
    delay: 'TICK_MS',
    reason: "the Help panel's self-reading cycle; attach starts it for a mounted panel and its cleanup clears it (FR-609)",
  },
  {
    file: 'App.tsx',
    api: 'setTimeout',
    delay: 'snapshot.refreshInMs',
    reason: "one shot to Go's next minute boundary (clock.NextRefresh), armed again only by the snapshot it loads (FR-208)",
  },
  {
    file: 'Ribbon.tsx',
    api: 'requestAnimationFrame',
    delay: FRAME,
    reason: 'first of two frames telling Go an opened ribbon has been painted; runs once per opening (FR-615)',
  },
  {
    file: 'Ribbon.tsx',
    api: 'requestAnimationFrame',
    delay: FRAME,
    reason: 'second of the two frames, scheduled once from the first; it re-arms nothing (FR-615)',
  },
]

/** normalise folds '.' and '..' out of a path relative to frontend/src, keeping leading '..'. */
function normalise(path: string): string {
  const parts: string[] = []
  for (const part of path.split('/')) {
    if (part === '' || part === '.') {
      continue
    }
    if (part === '..' && parts.length > 0 && parts[parts.length - 1] !== '..') {
      parts.pop()
    } else {
      parts.push(part)
    }
  }
  return parts.join('/')
}

const modules = new Map(Object.entries(sources).map(([key, text]) => [normalise(key), text]))

function isTest(file: string): boolean {
  return /\.test\.tsx?$/.test(file)
}

const IMPORT_FORMS = [
  /(?:import|export)\s[^'"]*?\sfrom\s*['"]([^'"]+)['"]/g,
  /import\s*['"]([^'"]+)['"]/g,
  /import\(\s*['"]([^'"]+)['"]\s*\)/g,
]

/** importsOf answers the page modules file imports, failing loudly on a script it cannot find. */
function importsOf(file: string): string[] {
  const text = modules.get(file) ?? ''
  const folder = file.includes('/') ? file.slice(0, file.lastIndexOf('/')) : ''
  const found: string[] = []
  for (const form of IMPORT_FORMS) {
    for (const match of text.matchAll(form)) {
      const specifier = match[1]
      if (!specifier.startsWith('.')) {
        continue
      }
      const bare = normalise(`${folder}/${specifier}`)
      const extension = bare.slice(bare.lastIndexOf('.'))
      const written = bare.lastIndexOf('.') > bare.lastIndexOf('/') && !SCRIPT_EXTENSIONS.includes(extension)
      if (written || specifier.includes('?')) {
        continue
      }
      const resolved = ['', ...SCRIPT_EXTENSIONS].map((tail) => bare + tail).find((path) => modules.has(path))
      if (resolved == null) {
        throw new Error(`${file} imports ${specifier}, which this test does not read; widen its glob`)
      }
      found.push(resolved)
    }
  }
  return found
}

/** page answers every module reachable from the entry. */
function page(): Set<string> {
  const reached = new Set<string>()
  const waiting = [ENTRY]
  while (waiting.length > 0) {
    const file = waiting.pop() as string
    if (reached.has(file)) {
      continue
    }
    reached.add(file)
    waiting.push(...importsOf(file))
  }
  return reached
}

/** delayOf answers the last top-level argument of the call whose '(' is at open. */
function delayOf(text: string, open: number): string {
  let depth = 0
  let start = open + 1
  let last = ''
  for (let at = open; at < text.length; at++) {
    const char = text[at]
    if ('([{'.includes(char)) {
      depth++
    } else if (')]}'.includes(char)) {
      depth--
      if (depth === 0) {
        last = text.slice(start, at)
        break
      }
    } else if (char === ',' && depth === 1) {
      start = at + 1
    }
  }
  return last.trim()
}

/** timerSites answers every timer call in file with the delay it is written with. */
function timerSites(file: string): TimerSite[] {
  const text = modules.get(file) ?? ''
  const sites: TimerSite[] = []
  for (const match of text.matchAll(new RegExp(`\\b(${TIMER_APIS.join('|')})\\s*\\(`, 'g'))) {
    const api = match[1] as TimerApi
    const open = (match.index ?? 0) + match[0].length - 1
    sites.push({ file, api, delay: api === 'requestAnimationFrame' ? FRAME : delayOf(text, open) })
  }
  return sites
}

function key(site: TimerSite): string {
  return `${site.file} ${site.api}(${site.delay})`
}

function count(text: string, pattern: RegExp): number {
  return [...text.matchAll(pattern)].length
}

describe('the ribbon page timers (NFR-P-4)', () => {
  const reached = page()

  it('reads every shipped source file and no test support, so nothing escapes the scan', () => {
    const shipped = [...modules.keys()].filter((file) => !file.startsWith('..') && !isTest(file) && !(file in TEST_SUPPORT))
    expect(shipped.filter((file) => !reached.has(file)), 'source the ribbon never imports; list it in TEST_SUPPORT or remove it').toEqual([])
    expect(Object.keys(TEST_SUPPORT).filter((file) => reached.has(file)), 'test support the page imports').toEqual([])
    expect(reached.has(HELP_CYCLE)).toBe(true)
  })

  it('schedules a periodic timer only in the Help self-reading cycle (FR-609)', () => {
    const periodic = [...reached].flatMap(timerSites).filter((site) => site.api === 'setInterval' && site.file !== HELP_CYCLE)
    expect(periodic.map(key), 'setInterval outside the Help cycle wakes the shown ribbon; NFR-P-4 allows one a minute at most').toEqual([])
  })

  it('starts the self-reading cycle only from Help, which clears it as the panel goes (FR-609)', () => {
    const starters = [...reached].filter((file) => importsOf(file).includes(HELP_CYCLE_HOOK)).sort()
    expect(starters).toEqual(HELP_CYCLE_USERS)
    const cycle = modules.get(HELP_CYCLE) ?? ''
    expect(count(cycle, /\bclearInterval\s*\(/g)).toBe(count(cycle, /\bsetInterval\s*\(/g))
    expect(modules.get(HELP_CYCLE_HOOK) ?? '').toMatch(/useEffect\(\(\) => \(node == null \? undefined : autoScroll\.attach\(node\)\)/)
  })

  it('has every timer call on the allow-list with the delay it is written with', () => {
    const found = [...reached].flatMap(timerSites).map(key).sort()
    const allowed = ALLOWED.map(key).sort()
    expect(found, 'a timer call not on the allow-list: add it to ALLOWED in timers.test.ts with why it cannot wake the page more than once a minute').toEqual(allowed)
  })
})
