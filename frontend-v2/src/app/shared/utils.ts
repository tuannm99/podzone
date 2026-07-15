export function classes(...values: Array<string | false | null | undefined>) {
  return values.filter(Boolean).join(' ');
}

export function isExternalUrl(url: string): boolean {
  return /^https?:\/\//i.test(url);
}

// Angular has no public equivalent of Solid's createUniqueId() — used by
// field components that need a stable id for label/aria-describedby wiring
// when the consumer doesn't pass one explicitly.
let uniqueIdCounter = 0;
export function createUniqueId(): string {
  uniqueIdCounter += 1;
  return `field-${uniqueIdCounter}`;
}
