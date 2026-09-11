// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';

import { fragmentOnglet, lireOnglet, useOnglet } from '../onglet';

describe('fragment d’onglet', () => {
  it.each(['gescom', 'ges/com', 'a#b', '50%', 'données clients'])(
    'relit le nom %s tel qu’il a été écrit',
    (base) => {
      const fragment = fragmentOnglet({ ecran: 'arbitrage', base });

      expect(fragment.slice('#arbitrage/'.length)).not.toMatch(/[/# ]/);
      expect(lireOnglet(fragment)).toEqual({ ecran: 'arbitrage', base });
    },
  );

  it.each(['', '#', '#selection', '#arbitrage/', '#arbitrage/%E0%A4%A'])(
    'ramène %j à la sélection',
    (fragment) => {
      expect(lireOnglet(fragment)).toEqual({ ecran: 'selection' });
    },
  );

  it('n’écrit rien pour la sélection', () => {
    expect(fragmentOnglet({ ecran: 'selection' })).toBe('');
  });
});

describe('useOnglet', () => {
  afterEach(() => {
    window.history.replaceState(null, '', '/');
  });

  it('part du fragment de l’adresse, ce qui rouvre l’arbitrage après un rechargement', () => {
    window.history.replaceState(null, '', '/#arbitrage/gescom');

    const { result } = renderHook(() => useOnglet());

    expect(result.current[0]).toEqual({ ecran: 'arbitrage', base: 'gescom' });
  });

  it('écrit l’onglet choisi dans l’adresse', () => {
    const { result } = renderHook(() => useOnglet());

    act(() => result.current[1]({ ecran: 'arbitrage', base: 'ges/com' }));
    expect(window.location.hash).toBe('#arbitrage/ges%2Fcom');
    expect(result.current[0]).toEqual({ ecran: 'arbitrage', base: 'ges/com' });

    act(() => result.current[1]({ ecran: 'selection' }));
    expect(window.location.hash).toBe('');
    expect(result.current[0]).toEqual({ ecran: 'selection' });
  });

  it('suit la navigation dans l’historique', () => {
    const { result } = renderHook(() => useOnglet());

    act(() => {
      window.history.pushState(null, '', '/#arbitrage/paie');
      window.dispatchEvent(new PopStateEvent('popstate'));
    });

    expect(result.current[0]).toEqual({ ecran: 'arbitrage', base: 'paie' });
  });
});
