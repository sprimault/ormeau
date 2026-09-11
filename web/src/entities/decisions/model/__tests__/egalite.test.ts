// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { describe, expect, it } from 'vitest';

import { decisionsEgales } from '../egalite';

describe('decisionsEgales', () => {
  it('ne tient pas compte de l’ordre des clés', () => {
    expect(
      decisionsEgales(
        { renommages: { 'public.a': 'A', 'public.b': 'B' }, tables_ignorees: ['public.x'] },
        { tables_ignorees: ['public.x'], renommages: { 'public.b': 'B', 'public.a': 'A' } },
      ),
    ).toBe(true);
  });

  it('confond ce que le fichier n’écrit pas : absent, vide, collection vide', () => {
    expect(
      decisionsEgales(
        { colonnes_ignorees: {}, prefixes_a_retirer: [], espace_de_noms: '' },
        {},
      ),
    ).toBe(true);
  });

  it('distingue une décision réelle', () => {
    expect(decisionsEgales({ tables_ignorees: ['public.x'] }, {})).toBe(false);
  });

  it('garde l’ordre des listes, que le fichier conserve', () => {
    expect(
      decisionsEgales({ prefixes_a_retirer: ['T_', 'tbl_'] }, { prefixes_a_retirer: ['tbl_', 'T_'] }),
    ).toBe(false);
  });
});
