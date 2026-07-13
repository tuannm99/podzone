// Ported from Navigation.tsx's NavItem type. Solid's `icon?: JSX.Element`
// field is dropped in this port — Angular has no direct per-array-item
// JSX-equivalent slot without a disproportionate amount of TemplateRef
// plumbing for what's an edge-case field; no current consumer sets it.
// Re-add via a TemplateRef input if a real feature needs it.
export type NavItem = {
  label: string;
  href?: string;
  active?: boolean;
  onClick?: () => void;
};
