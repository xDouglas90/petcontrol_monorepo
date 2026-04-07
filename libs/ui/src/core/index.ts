export interface UISpacingScale {
  xs: number;
  sm: number;
  md: number;
  lg: number;
  xl: number;
}

export const defaultSpacingScale: UISpacingScale = {
  xs: 4,
  sm: 8,
  md: 12,
  lg: 16,
  xl: 24,
};

export function joinClassNames(
  ...tokens: Array<string | false | null | undefined>
) {
  return tokens.filter(Boolean).join(' ');
}
