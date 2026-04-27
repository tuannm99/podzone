declare module '@toast-ui/editor' {
  export type PreviewStyle = 'tab' | 'vertical'
  export type EditorType = 'markdown' | 'wysiwyg'

  export type ToolbarItem = string

  export interface LinkAttributes {
    target?: string
    rel?: string
    content?: string
  }

  export interface EditorOptions {
    el: HTMLElement
    height?: string
    minHeight?: string
    initialValue?: string
    previewStyle?: PreviewStyle
    initialEditType?: EditorType
    usageStatistics?: boolean
    hideModeSwitch?: boolean
    placeholder?: string
    linkAttributes?: LinkAttributes
    toolbarItems?: ToolbarItem[][]
    events?: {
      change?: () => void
    }
  }

  export interface ViewerOptions {
    el: HTMLElement
    initialValue?: string
    usageStatistics?: boolean
    linkAttributes?: LinkAttributes
  }

  export class Viewer {
    constructor(options: ViewerOptions)
    setMarkdown(markdown: string): void
    destroy(): void
  }

  export class Editor {
    constructor(options: EditorOptions)
    getMarkdown(): string
    setMarkdown(markdown: string, cursorToEnd?: boolean): void
    changePreviewStyle(style: PreviewStyle): void
    getCurrentPreviewStyle(): PreviewStyle
    destroy(): void
  }

  export default Editor
}

declare module '@toast-ui/editor/viewer' {
  import type { LinkAttributes } from '@toast-ui/editor'

  export interface ViewerOptions {
    el: HTMLElement
    initialValue?: string
    usageStatistics?: boolean
    linkAttributes?: LinkAttributes
  }

  export default class Viewer {
    constructor(options: ViewerOptions)
    setMarkdown(markdown: string): void
    destroy(): void
  }
}
