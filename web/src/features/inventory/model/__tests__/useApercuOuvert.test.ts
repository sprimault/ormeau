// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';

import { usePreferencesStore } from '@/shared/model';
import { useApercuOuvert } from '../useApercuOuvert';

/** Pose le repli que le serveur aurait injecté dans la page. */
function injecter(ouvert?: boolean) {
  usePreferencesStore.setState({
    preferences: { theme: 'systeme', langue: 'fr', apercu_ouvert: ouvert },
  });
}

describe('useApercuOuvert', () => {
  // Le store est un module partagé : sans cette remise à zéro, un test hérite
  // du repli que le précédent a posé et passe pour une autre raison que la
  // sienne.
  beforeEach(() => {
    injecter();
  });

  it('ouvre l’aperçu quand rien n’a été retenu', () => {
    const { result } = renderHook(() => useApercuOuvert());
    expect(result.current[0]).toBe(true);
  });

  it('se replie, et le retient au lancement suivant', () => {
    const { result, unmount } = renderHook(() => useApercuOuvert());

    act(() => result.current[1]());
    expect(result.current[0]).toBe(false);
    expect(usePreferencesStore.getState().preferences.apercu_ouvert).toBe(false);

    unmount();
    const { result: relance } = renderHook(() => useApercuOuvert());
    expect(relance.current[0]).toBe(false);
  });

  it('se rouvre', () => {
    injecter(false);
    const { result } = renderHook(() => useApercuOuvert());
    expect(result.current[0]).toBe(false);

    act(() => result.current[1]());

    expect(result.current[0]).toBe(true);
    expect(usePreferencesStore.getState().preferences.apercu_ouvert).toBe(true);
  });
});
