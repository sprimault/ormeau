// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useDiffere } from '../useDiffere';

describe('useDiffere', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('rend la première valeur sans attendre', () => {
    const { result } = renderHook(() => useDiffere('a', 300));

    expect(result.current).toBe('a');
  });

  it('ne suit qu’après un délai sans nouveau changement', () => {
    const { result, rerender } = renderHook(({ valeur }) => useDiffere(valeur, 300), {
      initialProps: { valeur: 'a' },
    });

    rerender({ valeur: 'ab' });
    act(() => vi.advanceTimersByTime(200));
    rerender({ valeur: 'abc' });
    act(() => vi.advanceTimersByTime(200));
    expect(result.current).toBe('a');

    act(() => vi.advanceTimersByTime(100));
    expect(result.current).toBe('abc');
  });
});
