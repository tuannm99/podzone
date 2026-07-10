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
                'solid-js': { singleton: true } as object,
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
    },
})
