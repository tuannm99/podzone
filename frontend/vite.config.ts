import { defineConfig } from 'vite'
import solidPlugin from 'vite-plugin-solid'
import tailwindcss from '@tailwindcss/vite'
import federation from '@originjs/vite-plugin-federation'
import path from 'path'

const BACKOFFICE_REMOTE = process.env.VITE_BACKOFFICE_REMOTE_URL ?? 'http://localhost:3001/assets/remoteEntry.js'
const IAM_REMOTE = process.env.VITE_IAM_REMOTE_URL ?? 'http://localhost:3002/assets/remoteEntry.js'
const ONBOARDING_REMOTE = process.env.VITE_ONBOARDING_REMOTE_URL ?? 'http://localhost:3003/assets/remoteEntry.js'

// Which remotes are actually running under the active `make docker-dev
// PROFILE=...` invocation (see deployments/docker/services.yml
// VITE_ACTIVE_PROFILES). __MFE_<NAME>__ must reflect whether the remote
// container is really up, not just whether its URL env var happens to be
// set — services.yml always provides a default URL for all three remotes
// regardless of profile, so checking var presence alone made every route
// try to fetch remotes that were never started under a narrow profile
// (e.g. PROFILE=iam never starts onboarding-remote, but /admin is owned by
// onboarding — this broke the post-login landing page for every narrow
// profile). Falls back to "full" when unset, matching a raw `docker
// compose up` that bypasses the Makefile.
const ACTIVE_PROFILES = (process.env.VITE_ACTIVE_PROFILES ?? 'full').split(',').map((p) => p.trim())
const isRemoteActive = (name: string) => ACTIVE_PROFILES.includes('full') || ACTIVE_PROFILES.includes(name)

export default defineConfig({
    plugins: [
        tailwindcss(),
        solidPlugin(),
        federation({
            name: 'shell',
            remotes: {
                backoffice: BACKOFFICE_REMOTE,
                iam: IAM_REMOTE,
                onboarding: ONBOARDING_REMOTE,
            },
            shared: {
                'solid-js': { requiredVersion: '^1.0.0', singleton: true } as object,
                '@tanstack/solid-router': { singleton: true } as object,
                '@podzone/shared': { singleton: true } as object,
            },
        }),
    ],
    define: {
        __MFE_BACKOFFICE__: JSON.stringify(isRemoteActive('backoffice')),
        __MFE_IAM__: JSON.stringify(isRemoteActive('iam')),
        __MFE_ONBOARDING__: JSON.stringify(isRemoteActive('onboarding')),
    },
    resolve: {
        alias: {
            // IAM-local components moved alongside pages into apps/iam/src/
            '@/modules/iam': path.resolve(__dirname, './apps/iam/src'),
            '@podzone/shared': path.resolve(__dirname, './packages/shared'),
            '@': path.resolve(__dirname, './src'),
            '@backoffice': path.resolve(__dirname, './apps/backoffice/src'),
            '@iam': path.resolve(__dirname, './apps/iam/src'),
            '@onboarding': path.resolve(__dirname, './apps/onboarding/src'),
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
