// The window's half of the wire between Go and the page, stated a second time here.
// ribbonkit/ui/window/wire.go is the other statement; a structural test compares the two.

export interface About {
  name: string
  version: string
  author: string
  copyright: string
  credits: Credit[]
}

export interface UpdateStatus {
  current: string
  latest: string
  updateAvailable: boolean
}

export interface Credit {
  name: string
  licence: string
  role: string
}
