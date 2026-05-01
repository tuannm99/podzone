import { Show, createEffect, createSignal, onCleanup } from 'solid-js';
import QRCode from 'qrcode';
import { classes } from '../../shared/utils';

export function QrCode(props: {
  value: string;
  size?: number;
  class?: string;
  panelClass?: string;
}) {
  const [svg, setSvg] = createSignal('');
  const [error, setError] = createSignal('');

  createEffect(() => {
    const nextValue = props.value.trim();
    if (!nextValue) {
      setSvg('');
      setError('');
      return;
    }

    let active = true;

    void QRCode.toString(nextValue, {
      type: 'svg',
      width: props.size ?? 192,
      margin: 1,
      errorCorrectionLevel: 'M',
      color: {
        dark: '#111827',
        light: '#FFFFFFFF',
      },
    })
      .then((value: string) => {
        if (!active) return;
        setSvg(value);
        setError('');
      })
      .catch((nextError: unknown) => {
        if (!active) return;
        setSvg('');
        setError(
          nextError instanceof Error
            ? nextError.message
            : 'Unable to generate QR code'
        );
      });

    onCleanup(() => {
      active = false;
    });
  });

  return (
    <div
      class={classes(
        'inline-flex rounded-3xl border border-gray-200 bg-white p-4 shadow-sm',
        props.class
      )}
    >
      <Show
        when={svg()}
        fallback={
          <div
            class={classes(
              'flex min-h-48 min-w-48 items-center justify-center rounded-2xl bg-gray-50 px-4 py-4 text-center text-sm text-gray-500',
              props.panelClass
            )}
          >
            {error() || 'Add a value to generate a QR code.'}
          </div>
        }
      >
        <div class={props.panelClass} innerHTML={svg()} />
      </Show>
    </div>
  );
}

export function QrCodeCard(props: {
  value: string;
  title?: string;
  copy?: string;
  class?: string;
}) {
  return (
    <section
      class={classes(
        'inline-flex flex-col gap-4 rounded-3xl border border-gray-200 bg-white p-5 shadow-sm',
        props.class
      )}
    >
      <Show when={props.title || props.copy}>
        <div class="space-y-1">
          <Show when={props.title}>
            <h3 class="text-lg font-semibold text-gray-900">{props.title}</h3>
          </Show>
          <Show when={props.copy}>
            <p class="text-sm text-gray-500">{props.copy}</p>
          </Show>
        </div>
      </Show>
      <QrCode value={props.value} />
    </section>
  );
}
