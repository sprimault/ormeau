// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import { useThemeStore } from '@/shared/model';
import { ThemeToggle } from './ThemeToggle';

describe('ThemeToggle', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
    useThemeStore.setState({ theme: 'systeme' });
  });

  it('propose les trois états', () => {
    render(<ThemeToggle />);
    expect(screen.getByRole('button', { name: 'Clair' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Sombre' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Système' })).toBeInTheDocument();
  });

  it('montre l’état courant sans rien ouvrir', () => {
    render(<ThemeToggle />);
    expect(screen.getByRole('button', { name: 'Système' })).toHaveAttribute('aria-pressed', 'true');
  });

  it('change de thème au clic', async () => {
    render(<ThemeToggle />);

    await userEvent.click(screen.getByRole('button', { name: 'Sombre' }));
    expect(useThemeStore.getState().theme).toBe('sombre');
    expect(document.documentElement.classList.contains('dark')).toBe(true);
  });

  it('garde un libellé pour les lecteurs d’écran', () => {
    render(<ThemeToggle />);
    expect(screen.getByRole('group', { name: 'Thème' })).toBeInTheDocument();
  });
});
