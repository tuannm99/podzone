import { ErrorBoundary } from 'solid-js'
import { lazyRouteComponent } from '@tanstack/solid-router'
import { ErrorAlert } from '@podzone/shared/ui/components/common/Feedback'

type ImportFn = Parameters<typeof lazyRouteComponent>[0]

function RemoteLoadError(props: { name: string; reset: () => void }) {
    return (
        <div class="flex flex-col items-start gap-3 p-6">
            <ErrorAlert>
                <span>
                    Module <strong>{props.name}</strong> could not be loaded. The service may be unavailable.
                </span>
            </ErrorAlert>
            <button
                class="text-sm text-gray-600 underline underline-offset-2 hover:text-gray-900"
                onClick={props.reset}
            >
                Retry
            </button>
        </div>
    )
}

export function remotePage(importFn: ImportFn, name: string) {
    const Page = lazyRouteComponent(importFn)
    return () => (
        <ErrorBoundary fallback={(_, reset) => <RemoteLoadError name={name} reset={reset} />}>
            <Page />
        </ErrorBoundary>
    )
}
