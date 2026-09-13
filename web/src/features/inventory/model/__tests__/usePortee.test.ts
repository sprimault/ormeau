// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import type { TableSommaire } from '@/shared/model';
import { usePortee } from '../usePortee';

/** Table d'inventaire réduite à ce que la portée lit. */
function table(schema: string, nom: string): TableSommaire {
  return { schema, nom, nb_colonnes: 1, lignes_estimees: 0, cle_primaire: true };
}

/** Schémas du serveur ; la portée ne doit retenir que ceux des tables cochées. */
const schemas = ['audit', 'public', 'ventes'];
/** Tables réparties sur les trois schémas, deux dans le même. */
const tables = [
  table('audit', 'journal'),
  table('public', 'clients'),
  table('public', 'commandes'),
  table('ventes', 'factures'),
];

describe('usePortee', () => {
  it('prend tous les schémas quand rien n’est coché', () => {
    const { result } = renderHook(() => usePortee(schemas, tables, new Set()));
    expect(result.current).toEqual({ schemas, tables_incluses: [] });
  });

  it('n’envoie que les schémas des tables cochées', () => {
    // Le premier schéma demandé devient celui du calque : laisser « audit » en
    // tête en ferait le schéma d'un calque qui n'en contient aucune table.
    const selection = new Set(['public.commandes', 'public.clients']);
    const { result } = renderHook(() => usePortee(schemas, tables, selection));

    expect(result.current).toEqual({
      schemas: ['public'],
      tables_incluses: ['public.clients', 'public.commandes'],
    });
  });

  it('garde l’ordre du serveur quand la sélection couvre plusieurs schémas', () => {
    const selection = new Set(['ventes.factures', 'audit.journal']);
    const { result } = renderHook(() => usePortee(schemas, tables, selection));

    expect(result.current.schemas).toEqual(['audit', 'ventes']);
    expect(result.current.tables_incluses).toEqual(['audit.journal', 'ventes.factures']);
  });

  it('lit le schéma dans l’inventaire, pas dans la clé qualifiée', () => {
    const pointe = [table('compta.2024', 'ecritures'), table('public', 'clients')];
    const selection = new Set(['compta.2024.ecritures']);
    const { result } = renderHook(() => usePortee(['compta.2024', 'public'], pointe, selection));

    expect(result.current.schemas).toEqual(['compta.2024']);
  });

  it('rend la même portée tant que rien ne change', () => {
    const selection = new Set(['public.clients']);
    const { result, rerender } = renderHook(() => usePortee(schemas, tables, selection));
    const premiere = result.current;

    rerender();

    expect(result.current).toBe(premiere);
  });
});
