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
    await act(async () => fireEvent.click(screen.getByLabelText('Analogue')))
    expect(bridge.SetStyle).toHaveBeenCalledWith('analogue')
    await act(async () => fireEvent.click(screen.getByLabelText('12-hour')))
    expect(bridge.SetFormat).toHaveBeenCalledWith('12h')
    await act(async () => fireEvent.click(screen.getByLabelText('Vertical')))
    expect(bridge.SetOrientation).toHaveBeenCalledWith('vertical')
    await act(async () => fireEvent.click(screen.getByLabelText('Always on top')))
    expect(bridge.SetAlwaysOnTop).toHaveBeenCalledWith(true)
    expect(reload).toHaveBeenCalled()
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

  it('moves a clock from the keyboard reachable buttons, which stop at the ends (FR-306)', async () => {
    const { bridge } = await open()
    expect((screen.getByLabelText('Move New York up') as HTMLButtonElement).disabled).toBe(true)
    expect((screen.getByLabelText('Move Sydney down') as HTMLButtonElement).disabled).toBe(true)
    await act(async () => fireEvent.click(screen.getByLabelText('Move Sydney up')))
    expect(bridge.MoveClock).toHaveBeenCalledWith('syd', -1)
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
    const search = screen.getByLabelText('Search places')
    fireEvent.keyDown(search, { key: 'ArrowDown' })
    await act(async () => fireEvent.keyDown(search, { key: 'Enter' }))
    expect(bridge.AddClock).toHaveBeenCalledWith('Europe/Oslo')
  })

  it('changes a clock\'s place through the same search (FR-304)', async () => {
    const { bridge } = await open()
    fireEvent.click(screen.getAllByText('Change place')[0])
    const kolkata = await screen.findByText('Kolkata')
    await act(async () => fireEvent.click(kolkata))
    expect(bridge.RezoneClock).toHaveBeenCalledWith('ny', 'Asia/Kolkata')
  })

  it('closes on Escape; Escape in the search closes the search alone', async () => {
    const { onClose } = await open(true)
    fireEvent.keyDown(screen.getByLabelText('Search places'), { key: 'Escape' })
    expect(onClose).not.toHaveBeenCalled()
    fireEvent.keyDown(screen.getByText('Settings'), { key: 'Escape' })
    expect(onClose).toHaveBeenCalled()
  })

  it('says why when a call is refused', async () => {
    const { bridge } = await open()
    bridge.SetTheme.mockRejectedValueOnce('settings could not be saved')
    await act(async () => fireEvent.click(screen.getByLabelText('Dark')))
    expect(screen.getByRole('alert').textContent).toBe('settings could not be saved')
  })
})
