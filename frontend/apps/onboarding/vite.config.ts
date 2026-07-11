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
            name: 'onboarding',
            filename: 'remoteEntry.js',
            exposes: {
                './AdminHomePage': `${pagesRoot}/AdminHomePage`,
                './AdminSettingsPage': `${pagesRoot}/AdminSettingsPage`,
                './AdminProvisioningPage': `${pagesRoot}/AdminProvisioningPage`,
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
        port: 3003,
        host: '0.0.0.0',
        cors: true,
    },
})
