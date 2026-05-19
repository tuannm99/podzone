import { For, Show, createEffect, createMemo, createSignal } from 'solid-js';
import { Tabs } from './Tabs';
import { Badge, Button, Card, InputField, TextareaField } from './Primitives';

type KeyValueEntry = {
  key: string;
  value: string;
};

function normalizeEntries(raw: string): KeyValueEntry[] {
  try {
    const parsed = JSON.parse(raw || '{}');
    if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') {
      return [];
    }
    return Object.entries(parsed).map(([key, value]) => ({
      key,
      value: typeof value === 'string' ? value : JSON.stringify(value),
    }));
  } catch {
    return [];
  }
}

function serializeEntries(entries: KeyValueEntry[]) {
  return JSON.stringify(
    entries.reduce<Record<string, string>>((acc, entry) => {
      if (!entry.key.trim()) return acc;
      acc[entry.key.trim()] = entry.value;
      return acc;
    }, {}),
    null,
    2
  );
}

export function IamKeyValueBuilder(props: {
  label: string;
  value: string;
  helper?: string;
  emptyKeyPlaceholder?: string;
  emptyValuePlaceholder?: string;
  addLabel?: string;
  badgeLabel?: string;
  onChange: (value: string) => void;
}) {
  const [mode, setMode] = createSignal<'builder' | 'json'>('builder');
  const [entries, setEntries] = createSignal<KeyValueEntry[]>(
    normalizeEntries(props.value)
  );

  createEffect(() => {
    setEntries(normalizeEntries(props.value));
  });

  const count = createMemo(
    () => entries().filter((entry) => entry.key.trim()).length
  );

  const commit = (next: KeyValueEntry[]) => {
    setEntries(next);
    props.onChange(serializeEntries(next));
  };

  const updateEntry = (index: number, patch: Partial<KeyValueEntry>) => {
    commit(
      entries().map((entry, currentIndex) =>
        currentIndex === index ? { ...entry, ...patch } : entry
      )
    );
  };

  const removeEntry = (index: number) => {
    commit(entries().filter((_, currentIndex) => currentIndex !== index));
  };

  const addEntry = () => {
    commit([...entries(), { key: '', value: '' }]);
  };

  return (
    <div class="space-y-3">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p class="text-sm font-medium text-gray-700">{props.label}</p>
          <Show when={props.helper}>
            <p class="mt-1 text-xs text-gray-500">{props.helper}</p>
          </Show>
        </div>
        <Badge
          content={`${count()} ${props.badgeLabel || 'entries'}`}
          color="blue"
        />
      </div>

      <Tabs
        value={mode()}
        items={[
          { value: 'builder', label: 'Builder' },
          { value: 'json', label: 'JSON' },
        ]}
        onChange={(value) => setMode(value as 'builder' | 'json')}
      />

      <Show when={mode() === 'builder'}>
        <div class="space-y-3">
          <For each={entries()}>
            {(entry, entryIndex) => (
              <Card class="space-y-3 border border-gray-200 bg-gray-50 p-4 shadow-none">
                <div class="flex items-center justify-between gap-3">
                  <Badge content={`Entry ${entryIndex() + 1}`} color="dark" />
                  <Button
                    size="xs"
                    color="red"
                    onClick={() => removeEntry(entryIndex())}
                  >
                    Remove
                  </Button>
                </div>
                <div class="grid gap-3 md:grid-cols-2">
                  <InputField
                    label="Key"
                    value={entry.key}
                    placeholder={props.emptyKeyPlaceholder || 'lane'}
                    onInput={(event) =>
                      updateEntry(entryIndex(), {
                        key: event.currentTarget.value,
                      })
                    }
                  />
                  <InputField
                    label="Value"
                    value={entry.value}
                    placeholder={props.emptyValuePlaceholder || 'priority'}
                    onInput={(event) =>
                      updateEntry(entryIndex(), {
                        value: event.currentTarget.value,
                      })
                    }
                  />
                </div>
              </Card>
            )}
          </For>

          <Button size="sm" color="dark" onClick={addEntry}>
            {props.addLabel || 'Add entry'}
          </Button>
        </div>
      </Show>

      <Show when={mode() === 'json'}>
        <TextareaField
          label={`${props.label} JSON`}
          rows={6}
          value={props.value}
          onInput={(event) => props.onChange(event.currentTarget.value)}
        />
      </Show>
    </div>
  );
}
