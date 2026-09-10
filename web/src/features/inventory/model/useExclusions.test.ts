// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { useExclusions } from './useExclusions';

describe('useExclusions', () => {
  it('écarte puis remet une colonne', () => {
    const { result } = renderHook(() => useExclusions());

    act(() => result.current.basculer('public.clients', 'photo'));
    expect(result.current.estIgnoree('public.clients', 'photo')).toBe(true);

    act(() => result.current.basculer('public.clients', 'photo'));
    expect(result.current.estIgnoree('public.clients', 'photo')).toBe(false);
  });

  it('ne laisse pas de table vide dans le fichier de décisions', () => {
    const { result } = renderHook(() => useExclusions());

    act(() => result.current.basculer('public.clients', 'photo'));
    act(() => result.current.basculer('public.clients', 'photo'));

    expect(result.current.ignorees).toEqual({});
  });

  it('trie tables et colonnes, pour que deux sessions identiques se ressemblent', () => {
    const { result } = renderHook(() => useExclusions());

    act(() => result.current.basculer('public.commandes', 'note'));
    act(() => result.current.basculer('public.clients', 'photo'));
    act(() => result.current.basculer('public.clients', 'blob_import'));

    expect(Object.keys(result.current.ignorees)).toEqual(['public.clients', 'public.commandes']);
    expect(result.current.ignorees['public.clients']).toEqual(['blob_import', 'photo']);
  });

  it('compte par table et au total', () => {
    const { result } = renderHook(() => useExclusions());

    act(() => result.current.basculer('public.clients', 'photo'));
    act(() => result.current.basculer('public.commandes', 'note'));

    expect(result.current.compte('public.clients')).toBe(1);
    expect(result.current.compte('public.inconnue')).toBe(0);
    expect(result.current.total).toBe(2);
  });

  it('ne confond pas deux colonnes homonymes de tables différentes', () => {
    const { result } = renderHook(() => useExclusions());

    act(() => result.current.basculer('public.clients', 'note'));

    expect(result.current.estIgnoree('public.commandes', 'note')).toBe(false);
  });
});
