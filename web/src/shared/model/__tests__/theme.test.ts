// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { beforeEach, describe, expect, it } from 'vitest';

import { CLE_THEME, themeEnregistre, useThemeStore } from '../theme';

describe('thème', () => {
  beforeEach(() => {
    window.localStorage.clear();
    useThemeStore.setState({ theme: 'systeme' });
    document.documentElement.classList.remove('dark');
  });

  it('vaut « système » quand rien n’a été choisi', () => {
    expect(themeEnregistre()).toBe('systeme');
  });

  it('relit ce qui a été choisi', () => {
    window.localStorage.setItem(CLE_THEME, 'sombre');
    expect(themeEnregistre()).toBe('sombre');
  });

  it('ignore une valeur qui n’est pas un thème', () => {
    window.localStorage.setItem(CLE_THEME, 'bleu');
    expect(themeEnregistre()).toBe('systeme');
  });

  it('applique la classe que la variante Tailwind observe', () => {
    useThemeStore.getState().setTheme('sombre');
    expect(document.documentElement.classList.contains('dark')).toBe(true);

    useThemeStore.getState().setTheme('clair');
    expect(document.documentElement.classList.contains('dark')).toBe(false);
  });

  it('persiste le choix', () => {
    useThemeStore.getState().setTheme('clair');
    expect(window.localStorage.getItem(CLE_THEME)).toBe('clair');
  });
});
