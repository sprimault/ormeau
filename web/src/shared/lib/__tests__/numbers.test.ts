// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { describe, expect, it } from 'vitest';

import { compact } from '../numbers';

describe('compact', () => {
  it('laisse les petits nombres entiers', () => {
    expect(compact(0)).toBe('0');
    expect(compact(999)).toBe('999');
  });

  it('abrège les milliers', () => {
    expect(compact(48210)).toBe('48 k');
    expect(compact(1200)).toBe('1,2 k');
  });

  it('garde une décimale sous dix, aucune au-delà', () => {
    expect(compact(4800)).toBe('4,8 k');
    expect(compact(482100)).toBe('482 k');
  });

  it('monte dans les unités supérieures', () => {
    expect(compact(3_400_000)).toBe('3,4 M');
    expect(compact(2_000_000_000)).toBe('2 G');
  });
});
