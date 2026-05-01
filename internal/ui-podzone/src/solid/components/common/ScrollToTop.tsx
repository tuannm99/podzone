import { createSignal, onCleanup, onMount } from 'solid-js';
import { classes } from '../../shared/utils';
import { Button } from './Primitives';

export function ScrollToTopButton(props: {
  threshold?: number;
  class?: string;
}) {
  const [visible, setVisible] = createSignal(false);

  onMount(() => {
    const updateVisibility = () => {
      setVisible(window.scrollY > (props.threshold ?? 280));
    };

    updateVisibility();
    window.addEventListener('scroll', updateVisibility, { passive: true });

    onCleanup(() => {
      window.removeEventListener('scroll', updateVisibility);
    });
  });

  return (
    <Button
      pill
      size="sm"
      color="dark"
      aria-label="Scroll to top"
      class={classes(
        'fixed bottom-6 right-6 z-30 shadow-lg transition duration-200',
        visible()
          ? 'translate-y-0 opacity-100'
          : 'pointer-events-none translate-y-3 opacity-0',
        props.class
      )}
      onClick={() => {
        window.scrollTo({
          top: 0,
          left: 0,
          behavior: 'smooth',
        });
      }}
    >
      Top
    </Button>
  );
}
