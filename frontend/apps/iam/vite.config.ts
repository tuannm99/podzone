import { defineConfig } from 'vite'
import solidPlugin from 'vite-plugin-solid'
import federation from '@originjs/vite-plugin-federation'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const pagesRoot = path.resolve(__dirname, './src/pages')

export default defineConfig({
    root: __dirname,
    plugins: [
        solidPlugin(),
        federation({
            name: 'iam',
            filename: 'remoteEntry.js',
            exposes: {
                './AdminIamPage': `${pagesRoot}/AdminIamPage`,
            },
            shared: {
                // solid-js intentionally excluded: keeping it as a singleton causes
                // createSignal (via importShared → HOST's module) and createRenderEffect
                // (static import → remote's bundle) to be in different reactive systems,
                // breaking signal-to-effect subscriptions. Remote uses its own bundle
                // so all reactive code shares one instance. Auth is bridged via window.__pz_auth_value__.
                '@tanstack/solid-router': { singleton: true } as object,
                '@podzone/shared': { singleton: true } as object,
            },
        }),
    ],
    resolve: {
        alias: {
            // IAM-local components live alongside pages in apps/iam/src/
            '@/modules/iam': path.resolve(__dirname, './src'),
            '@podzone/shared': path.resolve(__dirname, '../../packages/shared'),
        },
    },
    build: {
        target: 'esnext',
        minify: false,
        cssCodeSplit: false,
        outDir: 'dist',
        rollupOptions: {
            output: {
                minifyInternalExports: false,
            },
        },
    },
    preview: {
        port: 3002,
        host: '0.0.0.0',
        cors: true,
        // Requests arrive via APISIX (/mfe/iam/*) with a rewritten Host header
        // (e.g. "apisix"), not "localhost:3002". Vite's preview server rejects
        // unrecognized Host headers by default (403) — this is an internal
        // Docker-only dev service, not internet-exposed, so disabling the
        // check is safe.
        allowedHosts: true,
    },
})
