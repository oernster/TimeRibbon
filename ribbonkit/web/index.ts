// ribbonkit's half of the page: what every ribbon's page does about its window, whatever it shows.
// An application imports it from '@oernster/ribbonkit'; test helpers from '@oernster/ribbonkit/testing'.

export { backgroundReporter, rgbOf, rootId, swatchId, type Rgb } from './background'
export { connect, on, startDrag, windowCalls, type Call, type Refused, type WindowBridge, type WindowCalls } from './bridge'
export { showsTheMenu, useDrag, type Distance } from './drag'
export { opacityProperty, percentOfWhole, showOpacity } from './opacity'
export { naturalHeight, usePanelFit } from './panelFit'
export { watchPixelRatio } from './pixelRatio'
export { scrollbarThickness } from './scrollbar'
export type { About, Credit, UpdateStatus } from './wire'
