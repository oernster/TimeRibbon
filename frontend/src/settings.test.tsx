import { act, fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { installBridge, snapshot } from './fakeBridge'
import { Settings } from './Settings'

async function open(startAdding = false) {
  const bridge = installBridge()
  const reload = vi.fn()
  const onClose = vi.fn()
  await act(async () => {
    render(<Settings snapshot={snapshot()} startAdding={startAdding} reload={reload} onClose={onClose} />)
  })
  return { bridge, reload, onClose }
}

describe('Settings', () => {
  it('applies a choice at once with no Save step (FR-602)', async () => {
    const { bridge, reload } = await open()
    await act(async () => fireEvent.click(screen.getByLabelText('Small')))
    expect(bridge.SetSize).toHaveBeenCalledWith('small')
    await act(async () => fireEvent.click(screen.getByLabelText('12-hour')))
    expect(bridge.SetFormat).toHaveBeenCalledWith('12h')
    await act(async () => fireEvent.click(screen.getByLabelText('Always on top')))
    expect(bridge.Choose).toHaveBeenCalledWith('always-on-top')
    expect(reload).toHaveBeenCalled()
  })

  it('offers every choice the menus do, ticked as Go says, each carried out by Go (FR-624)', async () => {
    const { bridge } = await open()
    for (const legend of ['Style', 'Colour', 'Orientation', 'Position']) {
      expect(screen.getByText(legend).tagName).toBe('LEGEND')
    }
    expect((screen.getByLabelText('Digital') as HTMLInputElement).checked).toBe(true)
    expect((screen.getByLabelText('Pin ribbon') as HTMLInputElement).checked).toBe(true)
    expect((screen.getByLabelText('Sun map') as HTMLInputElement).checked).toBe(false)
    const picks: [HTMLElement, string][] = [
      [screen.getByLabelText('Analogue'), 'analogue'],
      [screen.getByLabelText('Neon'), 'colour-neon'],
      [screen.getByLabelText('Vertical'), 'vertical'],
      [screen.getByText('Centre on bottom edge'), 'bottom-edge'],
      [screen.getByLabelText('Pin ribbon'), 'pin'],
      [screen.getByLabelText('Sun map'), 'sun-map'],
    ]
    for (const [control, action] of picks) {
      await act(async () => fireEvent.click(control))
      expect(bridge.Choose).toHaveBeenLastCalledWith(action)
    }
  })

  it('greys a Position choice that would leave the ribbon where it stands, so pressing it does nothing (FR-408)', async () => {
    const bridge = installBridge()
    const shown = snapshot()
    shown.choices = shown.choices.map((choice) =>
      choice.label !== 'Position' ? choice : { ...choice, children: choice.children.map((item) => ({ ...item, disabled: item.action === 'top-edge' })) },
    )
    await act(async () => {
      render(<Settings snapshot={shown} startAdding={false} reload={vi.fn()} onClose={vi.fn()} />)
    })
    const top = screen.getByText('Centre on top edge') as HTMLButtonElement
    expect(top.disabled).toBe(true)
    expect((screen.getByText('Centre on bottom edge') as HTMLButtonElement).disabled).toBe(false)
    await act(async () => fireEvent.click(top))
    expect(bridge.Choose).not.toHaveBeenCalled()
  })

  it('offers every date format with the current one chosen, applying one at once (FR-612)', async () => {
    const { bridge } = await open()
    const group = screen.getByText('Date format').closest('fieldset') as HTMLElement
    const offered = Array.from(group.querySelectorAll('label'), (label) => label.textContent)
    expect(offered).toEqual(['28 September', 'September 28', 'DD/MM/YYYY', 'MM/DD/YYYY', 'YYYY/MM/DD'])
    expect((screen.getByLabelText('28 September') as HTMLInputElement).checked).toBe(true)
    await act(async () => fireEvent.click(screen.getByLabelText('DD/MM/YYYY')))
    expect(bridge.SetDateFormat).toHaveBeenCalledWith('dmy')
  })

  it('names the sign-in switch in the words the snapshot gives, so no platform is named here (FR-605)', async () => {
    const { bridge } = await open()
    await act(async () => fireEvent.click(screen.getByLabelText('Start at sign-in')))
    expect(bridge.SetStartWithWindows).toHaveBeenCalledWith(true)
    expect(screen.queryByText(/Windows/)).toBeNull()
  })

  it('asks before removing, naming the clock; Cancel removes nothing (FR-305)', async () => {
    const { bridge } = await open()
    fireEvent.click(screen.getByLabelText('Remove Sydney'))
    expect(screen.getByRole('alertdialog').textContent).toContain('Remove the clock for Sydney?')
    fireEvent.click(screen.getByText('Cancel'))
    expect(bridge.RemoveClock).not.toHaveBeenCalled()
    fireEvent.click(screen.getByLabelText('Remove Sydney'))
    const confirm = screen.getAllByText('Remove')
    await act(async () => fireEvent.click(confirm[confirm.length - 1]))
    expect(bridge.RemoveClock).toHaveBeenCalledWith('syd')
  })

  it('lists the clocks in the order the ribbon shows them, with nothing to move them by hand (FR-102)', async () => {
    await open()
    const rows = screen.getByRole('list', { name: 'Clocks' }).querySelectorAll('li')
    expect(Array.from(rows, (row) => (row.querySelector('input') as HTMLInputElement).value)).toEqual(['New York', 'Sydney'])
    expect(screen.queryByLabelText(/^Move /)).toBeNull()
    expect(Array.from(rows).some((row) => row.draggable)).toBe(false)
  })

  it('renames a clock when its label is left (FR-303)', async () => {
    const { bridge } = await open()
    const input = screen.getByLabelText('Label for America/New_York')
    fireEvent.change(input, { target: { value: 'Mum' } })
    await act(async () => fireEvent.blur(input))
    expect(bridge.RenameClock).toHaveBeenCalledWith('ny', 'Mum')
  })

  it('adds the highlighted place on Enter, moving with the arrows (FR-301, NFR-U-3)', async () => {
    const { bridge } = await open(true)
    const search = screen.getByLabelText('Search places') as HTMLInputElement
    expect(document.activeElement).toBe(search)
    await act(async () => fireEvent.change(search, { target: { value: 'o' } }))
    fireEvent.keyDown(search, { key: 'ArrowDown' })
    await act(async () => fireEvent.keyDown(search, { key: 'Enter' }))
    expect(bridge.AddClock).toHaveBeenCalledWith('Europe/Oslo')
    expect(search.value).toBe('')
  })

  it('keeps the place search open with the Add clock picture beside it (FR-626)', async () => {
    const { bridge } = await open()
    const search = screen.getByLabelText('Search places') as HTMLInputElement
    expect(document.activeElement).not.toBe(search)
    expect(screen.getByRole('listbox', { hidden: true }).hidden).toBe(true)
    const picture = screen.getByLabelText('Add a clock for another place: search by city, zone or country')
    expect(picture.closest('.search-row')).toBe(search.closest('.search-row'))
    await act(async () => fireEvent.click(picture))
    expect(bridge.AddClock).not.toHaveBeenCalled()
    expect(document.activeElement).toBe(search)
    await act(async () => fireEvent.change(search, { target: { value: 'k' } }))
    expect(screen.getByRole('listbox').hidden).toBe(false)
    await act(async () => fireEvent.click(picture))
    expect(bridge.AddClock).toHaveBeenCalledWith('Asia/Kolkata')
    expect(screen.queryByText('Cancel')).toBeNull()
  })

  it('changes a clock\'s place through the same search (FR-304)', async () => {
    const { bridge } = await open()
    fireEvent.click(screen.getAllByText('Change place')[0])
    const kolkata = await screen.findByText('Kolkata')
    await act(async () => fireEvent.click(kolkata))
    expect(bridge.RezoneClock).toHaveBeenCalledWith('ny', 'Asia/Kolkata')
  })

  it('closes on Escape; Escape with something typed clears it first; it cancels a change of place', async () => {
    const { onClose } = await open(true)
    const search = screen.getByLabelText('Search places') as HTMLInputElement
    await act(async () => fireEvent.change(search, { target: { value: 'Oslo' } }))
    fireEvent.keyDown(search, { key: 'Escape' })
    expect(search.value).toBe('')
    expect(onClose).not.toHaveBeenCalled()
    await act(async () => fireEvent.click(screen.getAllByText('Change place')[0]))
    await act(async () => fireEvent.keyDown(screen.getByLabelText('Search places'), { key: 'Escape' }))
    expect(screen.queryByText('Change the place of New York')).toBeNull()
    expect(onClose).not.toHaveBeenCalled()
    await act(async () => fireEvent.keyDown(screen.getByLabelText('Search places'), { key: 'Escape' }))
    expect(onClose).toHaveBeenCalled()
  })

  it('offers a donation that says it opens the browser and hands the page to Go', async () => {
    const { bridge } = await open()
    const donate = screen.getByLabelText('Buy the author a drink (opens your browser)')
    expect(donate.getAttribute('title')).toBe('Buy the author a drink (opens your browser)')
    await act(async () => fireEvent.click(donate))
    expect(bridge.OpenDonation).toHaveBeenCalledTimes(1)
  })

  it('says why when the browser could not be opened on the donation page', async () => {
    const { bridge } = await open()
    bridge.OpenDonation.mockRejectedValueOnce('your browser could not be opened on the donation page')
    await act(async () => fireEvent.click(screen.getByLabelText('Buy the author a drink (opens your browser)')))
    expect(screen.getByRole('alert').textContent).toBe('your browser could not be opened on the donation page')
  })

  it('says why when a call is refused', async () => {
    const { bridge } = await open()
    bridge.SetTheme.mockRejectedValueOnce('settings could not be saved')
    await act(async () => fireEvent.click(screen.getByLabelText('Dark')))
    expect(screen.getByRole('alert').textContent).toBe('settings could not be saved')
  })
})
