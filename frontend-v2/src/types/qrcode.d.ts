// Ported verbatim from frontend/src/solid/types/qrcode.d.ts — the
// `qrcode` npm package ships no bundled types.
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
