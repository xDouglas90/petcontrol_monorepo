export function normalizeUrl(value: string): string {
  return value.replace(/\/$/, '');
}

export function isNonEmptyTrimmed(value: string): boolean {
  return value.trim().length > 0;
}

export function safeLowerCase(value: string): string {
  return value.trim().toLowerCase();
}
