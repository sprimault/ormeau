// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { Decisions } from '@/shared/model';
import { useExclusions } from '../useExclusions';

/** Le brouillon simulé : ce qu'il contient au départ, et ce qu'il est devenu. */
const brouillon = vi.hoisted(() => ({
  initial: {} as Decisions,
  courant: {} as Decisions,
}));

// Le brouillon de décisions a son propre fournisseur et ses propres tests : ce
// qui est sous test ici, c'est ce que l'arbre y écrit.
vi.mock('@/entities/decisions', async () => {
  const { useCallback, useState } = await import('react');
  return {
    useDecisions: () => {
      const [decisions, setDecisions] = useState<Decisions>(brouillon.initial);
      const modifier = useCallback(
        (transformation: (d: Decisions) => Decisions) => setDecisions((d) => transformation(d)),
        [],
      );
      brouillon.courant = decisions;
      return { decisions, modifier };
    },
  };
});

describe('useExclusions', () => {
  beforeEach(() => {
    brouillon.initial = {};
    brouillon.courant = {};
  });

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
    expect(brouillon.courant.colonnes_ignorees).toBeUndefined();
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

  it('reprend les colonnes que le fichier de décisions écartait déjà', () => {
    brouillon.initial = { colonnes_ignorees: { 'public.clients': ['photo', 'blob_import'] } };
    const { result } = renderHook(() => useExclusions());

    expect(result.current.estIgnoree('public.clients', 'photo')).toBe(true);
    expect(result.current.total).toBe(2);
  });

  it('écrit dans le brouillon de décisions, sans toucher au reste', () => {
    brouillon.initial = { tables_ignorees: ['public.migrations'] };
    const { result } = renderHook(() => useExclusions());

    act(() => result.current.basculer('public.clients', 'photo'));

    expect(brouillon.courant).toEqual({
      tables_ignorees: ['public.migrations'],
      colonnes_ignorees: { 'public.clients': ['photo'] },
    });
  });
});
