declare module 'qrcode' {
  type ErrorCorrectionLevel = 'L' | 'M' | 'Q' | 'H'

  type RenderOptions = {
    type?: 'svg'
    width?: number
    margin?: number
    errorCorrectionLevel?: ErrorCorrectionLevel
    color?: {
      dark?: string
      light?: string
    }
  }

  const QRCode: {
    toString(text: string, options?: RenderOptions): Promise<string>
  }

  export default QRCode
}
