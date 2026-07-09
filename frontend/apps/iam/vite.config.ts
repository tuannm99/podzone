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
            name: 'iam',
            filename: 'remoteEntry.js',
            exposes: {
                './AdminIamPage': `${pagesRoot}/AdminIamPage`,
            },
            shared: ['solid-js', '@tanstack/solid-router'],
        }),
    ],
    resolve: {
        alias: {
            // Override: IAM-local code (components/) moved alongside pages
            '@/modules/iam': path.resolve(__dirname, './src'),
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
        port: 3002,
        host: '0.0.0.0',
        cors: true,
    },
})
