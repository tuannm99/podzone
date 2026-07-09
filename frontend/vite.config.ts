import { defineConfig } from 'vite'
import solidPlugin from 'vite-plugin-solid'
import tailwindcss from '@tailwindcss/vite'
import federation from '@originjs/vite-plugin-federation'
import path from 'path'

const BACKOFFICE_REMOTE = process.env.VITE_BACKOFFICE_REMOTE_URL ?? 'http://localhost:3001/assets/remoteEntry.js'

export default defineConfig({
    plugins: [
        tailwindcss(),
        solidPlugin(),
        federation({
            name: 'shell',
            remotes: {
                backoffice: BACKOFFICE_REMOTE,
            },
            shared: {
                'solid-js': { requiredVersion: '^1.0.0', singleton: true } as object,
                '@tanstack/solid-router': { singleton: true } as object,
            },
        }),
    ],
    define: {
        __MFE_BACKOFFICE__: JSON.stringify(!!process.env.VITE_BACKOFFICE_REMOTE_URL),
    },
    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src'),
            '@backoffice': path.resolve(__dirname, './apps/backoffice/src'),
        },
    },
    server: {
        port: 3000,
        proxy: {
            '/api': {
                target: process.env.VITE_API_PROXY_TARGET ?? 'http://localhost:8080',
                changeOrigin: true,
            },
        },
    },
    build: {
        target: 'esnext',
    },
})
