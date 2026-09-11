// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { describe, expect, it } from 'vitest';

import { compact, duree } from '../numbers';

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

describe('duree', () => {
  it('compte en secondes sous la minute', () => {
    expect(duree(0)).toBe('0 s');
    expect(duree(42)).toBe('42 s');
  });

  it('passe aux minutes, secondes sur deux chiffres', () => {
    expect(duree(65)).toBe('1 min 05 s');
    expect(duree(600)).toBe('10 min 00 s');
  });

  it('laisse tomber les secondes au-delà de l’heure', () => {
    expect(duree(3600)).toBe('1 h 00 min');
    expect(duree(4320)).toBe('1 h 12 min');
  });
});
