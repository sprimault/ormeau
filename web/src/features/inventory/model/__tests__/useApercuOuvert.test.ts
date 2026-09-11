// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';

import { useApercuOuvert } from '../useApercuOuvert';

describe('useApercuOuvert', () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it('ouvre l’aperçu quand rien n’a été retenu', () => {
    const { result } = renderHook(() => useApercuOuvert());
    expect(result.current[0]).toBe(true);
  });

  it('se replie, et le retient au lancement suivant', () => {
    const { result, unmount } = renderHook(() => useApercuOuvert());

    act(() => result.current[1]());
    expect(result.current[0]).toBe(false);

    unmount();
    const { result: relance } = renderHook(() => useApercuOuvert());
    expect(relance.current[0]).toBe(false);
  });

  it('se rouvre', () => {
    window.localStorage.setItem('ormeau-apercu-ouvert', '0');
    const { result } = renderHook(() => useApercuOuvert());

    act(() => result.current[1]());

    expect(result.current[0]).toBe(true);
  });
});
