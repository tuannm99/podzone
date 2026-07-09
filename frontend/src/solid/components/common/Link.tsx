import { Link as RouterLink } from '@tanstack/solid-router'
import { splitProps, type JSX } from 'solid-js'
import { isExternalUrl } from '../../shared/utils'

export type LinkProps = JSX.AnchorHTMLAttributes<HTMLAnchorElement> & {
    href: string
    class?: string
}

export function Link(props: LinkProps) {
    const [local, rest] = splitProps(props, ['href', 'class', 'children', 'target', 'rel'])

    if (isExternalUrl(local.href)) {
        const rel = local.target === '_blank' ? (local.rel ?? 'noopener noreferrer') : local.rel
        return (
            <a href={local.href} target={local.target} rel={rel} class={local.class} {...rest}>
                {local.children}
            </a>
        )
    }

    return (
        <RouterLink to={local.href} class={local.class} {...(rest as object)}>
            {local.children}
        </RouterLink>
    )
}
