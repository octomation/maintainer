import { useMDXComponents as themeComponents } from 'nextra-theme-docs'

export function useMDXComponents(components) {
  return { ...themeComponents(), ...components }
}
