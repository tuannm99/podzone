import { defineConfig } from 'vite'
import solidPlugin from 'vite-plugin-solid'
import federation from '@originjs/vite-plugin-federation'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const srcRoot = path.resolve(__dirname, '../../src')

export default defineConfig({
    plugins: [
        solidPlugin(),
        federation({
            name: 'backoffice',
            filename: 'remoteEntry.js',
            exposes: {
                './TenantHomePage': `${srcRoot}/modules/backoffice/pages/TenantHomePage`,
                './TenantOrdersPage': `${srcRoot}/modules/backoffice/pages/TenantOrdersPage`,
                './TenantOrderAuditPage': `${srcRoot}/modules/backoffice/pages/TenantOrderAuditPage`,
                './TenantOrderFinancePage': `${srcRoot}/modules/backoffice/pages/TenantOrderFinancePage`,
                './TenantPartnersPage': `${srcRoot}/modules/backoffice/pages/TenantPartnersPage`,
                './TenantPartnerDetailPage': `${srcRoot}/modules/backoffice/pages/TenantPartnerDetailPage`,
                './TenantProductSetupPage': `${srcRoot}/modules/backoffice/pages/TenantProductSetupPage`,
            },
            shared: ['solid-js', '@tanstack/solid-router'],
        }),
    ],
    resolve: {
        alias: {
            '@': srcRoot,
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
    },
})
