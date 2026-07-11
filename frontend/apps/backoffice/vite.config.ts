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
            name: 'backoffice',
            filename: 'remoteEntry.js',
            exposes: {
                './TenantHomePage': `${pagesRoot}/TenantHomePage`,
                './TenantOrdersPage': `${pagesRoot}/TenantOrdersPage`,
                './TenantOrderAuditPage': `${pagesRoot}/TenantOrderAuditPage`,
                './TenantOrderFinancePage': `${pagesRoot}/TenantOrderFinancePage`,
                './TenantPartnersPage': `${pagesRoot}/TenantPartnersPage`,
                './TenantPartnerDetailPage': `${pagesRoot}/TenantPartnerDetailPage`,
                './TenantProductSetupPage': `${pagesRoot}/TenantProductSetupPage`,
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
        port: 3001,
        host: '0.0.0.0',
        cors: true,
        // Requests arrive via APISIX (/mfe/backoffice/*) with a rewritten Host
        // header (e.g. "apisix"), not "localhost:3001". Vite's preview server
        // rejects unrecognized Host headers by default (403) — this is an
        // internal Docker-only dev service, not internet-exposed, so
        // disabling the check is safe.
        allowedHosts: true,
    },
})
