// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { describe, expect, it } from 'vitest';

import { fragmentDeTable } from '../fragment';

const clients = {
  nom: 'clients',
  schema: 'public',
  colonnes: [{ nom: 'id', position: 1, type_brut: 'integer', type_normalise: 'entier' }],
};

/**
 * Calque sérialisé comme le serveur le rend : deux espaces d'indentation. La
 * borne bigint d'une séquence est posée dans le texte, comme sur disque — un
 * littéral JavaScript la fausserait avant même la lecture.
 */
const contenu = JSON.stringify(
  {
    version_ri: 1,
    tables: [clients, { ...clients, schema: 'archives' }],
    sequences: [{ nom: 'grande', schema: 'public', maximum: 0 }],
  },
  null,
  2,
).replace('"maximum": 0', '"maximum": 9223372036854775807');

describe('fragmentDeTable', () => {
  it('rend l’entrée de la table, mise en forme comme dans le fichier', () => {
    expect(fragmentDeTable(contenu, 'public', 'clients')).toBe(JSON.stringify(clients, null, 2));
  });

  it('distingue deux tables homonymes par leur schéma', () => {
    const fragment = fragmentDeTable(contenu, 'archives', 'clients');
    expect(fragment).toContain('"schema": "archives"');
  });

  it('rend null pour une table absente du calque', () => {
    expect(fragmentDeTable(contenu, 'public', 'commandes')).toBeNull();
  });
});
