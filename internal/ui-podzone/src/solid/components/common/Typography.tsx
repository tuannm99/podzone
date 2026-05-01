import type { ParentProps, JSX } from 'solid-js';
import { Dynamic } from 'solid-js/web';
import { classes } from '../../shared/utils';

type HeadingTag = 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6';

const headingClasses: Record<HeadingTag, string> = {
  h1: 'text-4xl font-semibold tracking-tight sm:text-5xl',
  h2: 'text-3xl font-semibold tracking-tight sm:text-4xl',
  h3: 'text-2xl font-semibold tracking-tight',
  h4: 'text-xl font-semibold tracking-tight',
  h5: 'text-lg font-semibold',
  h6: 'text-base font-semibold uppercase tracking-wide',
};

export function Heading(
  props: ParentProps<{
    as?: HeadingTag;
    class?: string;
  }>
) {
  const tag = () => props.as ?? 'h2';

  return (
    <Dynamic
      component={tag()}
      class={classes('text-gray-900', headingClasses[tag()], props.class)}
    >
      {props.children}
    </Dynamic>
  );
}

export function Paragraph(
  props: ParentProps<{ muted?: boolean; class?: string }>
) {
  return (
    <p
      class={classes(
        'text-base leading-7',
        props.muted ? 'text-gray-500' : 'text-gray-700',
        props.class
      )}
    >
      {props.children}
    </p>
  );
}

export function Blockquote(
  props: ParentProps<{ cite?: string; class?: string }>
) {
  return (
    <figure
      class={classes(
        'rounded-2xl border-s-4 border-blue-600 bg-blue-50 px-5 py-4',
        props.class
      )}
    >
      <blockquote class="text-base leading-7 text-gray-700">
        {props.children}
      </blockquote>
      {props.cite ? (
        <figcaption class="mt-3 text-sm font-medium text-gray-500">
          {props.cite}
        </figcaption>
      ) : null}
    </figure>
  );
}

export function TextLink(
  props: ParentProps<{
    href: string;
    target?: string;
    class?: string;
  }>
) {
  return (
    <a
      href={props.href}
      target={props.target}
      rel={props.target === '_blank' ? 'noreferrer' : undefined}
      class={classes(
        'font-medium text-blue-700 underline decoration-blue-200 underline-offset-4 transition hover:text-blue-800 hover:decoration-blue-400',
        props.class
      )}
    >
      {props.children}
    </a>
  );
}

export function InlineCode(props: ParentProps<{ class?: string }>) {
  return (
    <code
      class={classes(
        'rounded-md bg-gray-100 px-1.5 py-0.5 text-sm text-gray-800',
        props.class
      )}
    >
      {props.children}
    </code>
  );
}

export function Kbd(props: ParentProps<{ class?: string }>) {
  return (
    <kbd
      class={classes(
        'inline-flex min-h-7 items-center rounded-lg border border-gray-200 bg-white px-2.5 text-xs font-semibold uppercase tracking-wide text-gray-600 shadow-sm',
        props.class
      )}
    >
      {props.children}
    </kbd>
  );
}

export function Divider(props: { label?: string; class?: string }) {
  return (
    <div class={classes('flex items-center gap-4', props.class)}>
      <div class="h-px flex-1 bg-gray-200" />
      {props.label ? (
        <>
          <span class="text-xs font-semibold uppercase tracking-[0.2em] text-gray-400">
            {props.label}
          </span>
          <div class="h-px flex-1 bg-gray-200" />
        </>
      ) : null}
    </div>
  );
}

export function ProseList(props: {
  items: Array<string | JSX.Element>;
  ordered?: boolean;
  class?: string;
}) {
  const Component = props.ordered ? 'ol' : 'ul';

  return (
    <Dynamic
      component={Component}
      class={classes(
        'space-y-2 ps-5 text-base leading-7 text-gray-700',
        props.ordered ? 'list-decimal' : 'list-disc',
        props.class
      )}
    >
      {props.items.map((item) => (
        <li>{item}</li>
      ))}
    </Dynamic>
  );
}
