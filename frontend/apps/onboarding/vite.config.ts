import { defineConfig } from 'vite'
import solidPlugin from 'vite-plugin-solid'
import federation from '@originjs/vite-plugin-federation'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const sharedRoot = path.resolve(__dirname, '../../src')
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
            shared: ['solid-js', '@tanstack/solid-router'],
        }),
    ],
    resolve: {
        alias: {
            '@': sharedRoot,
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
