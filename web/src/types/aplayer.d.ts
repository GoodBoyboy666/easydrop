declare module 'aplayer' {
  interface APlayerAudio {
    name: string
    artist: string
    url: string
    cover: string
    lrc: string
    theme?: string
    type?: string
  }

  interface APlayerOptions {
    container: HTMLElement
    audio: APlayerAudio[]
    autoplay?: boolean
    fixed?: boolean
    mini?: boolean
    theme?: string
    loop?: 'all' | 'one' | 'none'
    order?: 'list' | 'random'
    preload?: 'auto' | 'metadata' | 'none'
    volume?: number
    mutex?: boolean
    listFolded?: boolean
    listMaxHeight?: string
    lrcType?: number
    storageName?: string
  }

  interface APlayerList {
    audios: APlayerAudio[]
    index: number
  }

  export default class APlayer {
    constructor(options: APlayerOptions)
    list: APlayerList
    play(): void
    pause(): void
    theme(color: string, index: number): void
    destroy(): void
  }
}
