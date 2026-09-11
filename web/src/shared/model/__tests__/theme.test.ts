// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { beforeEach, describe, expect, it } from 'vitest';

import { usePreferencesStore } from '../preferences';
import { themeInitial, useThemeStore } from '../theme';

describe('thème', () => {
  beforeEach(() => {
    usePreferencesStore.setState({ preferences: { theme: 'systeme', langue: 'fr' } });
    useThemeStore.setState({ theme: 'systeme' });
    document.documentElement.classList.remove('dark');
  });

  it('vaut « système » quand rien n’a été choisi', () => {
    expect(themeInitial()).toBe('systeme');
  });

  it('relit ce que le serveur a injecté', () => {
    usePreferencesStore.setState({ preferences: { theme: 'sombre', langue: 'fr' } });
    expect(themeInitial()).toBe('sombre');
  });

  it('ignore une valeur qui n’est pas un thème', () => {
    usePreferencesStore.setState({ preferences: { theme: 'bleu', langue: 'fr' } });
    expect(themeInitial()).toBe('systeme');
  });

  it('applique la classe que la variante Tailwind observe', () => {
    useThemeStore.getState().setTheme('sombre');
    expect(document.documentElement.classList.contains('dark')).toBe(true);

    useThemeStore.getState().setTheme('clair');
    expect(document.documentElement.classList.contains('dark')).toBe(false);
  });

  it('porte le choix dans les préférences, que le serveur enregistre', () => {
    useThemeStore.getState().setTheme('clair');
    expect(usePreferencesStore.getState().preferences.theme).toBe('clair');
  });
});
