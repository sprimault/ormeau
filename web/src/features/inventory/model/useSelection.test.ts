// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import type { TableSommaire } from '@/shared/model';
import { useSelection } from './useSelection';

/** Fabrique un sommaire, seuls le nom et les références important ici. */
function table(nom: string, reference_vers: string[] = [], schema = 'public'): TableSommaire {
  return {
    schema,
    nom,
    nb_colonnes: 3,
    lignes_estimees: 0,
    cle_primaire: true,
    reference_vers,
  };
}

const inventaire = [
  table('commandes', ['public.clients', 'public.produits']),
  table('clients', ['public.pays']),
  table('produits'),
  table('pays'),
  table('journal'),
];

describe('useSelection', () => {
  it('coche et décoche une table', () => {
    const { result } = renderHook(() => useSelection(inventaire));

    act(() => result.current.basculer('public.journal'));
    expect(result.current.selection.has('public.journal')).toBe(true);

    act(() => result.current.basculer('public.journal'));
    expect(result.current.selection.has('public.journal')).toBe(false);
  });

  it('signale les références qui sortent de la sélection', () => {
    const { result } = renderHook(() => useSelection(inventaire));

    // pays y figure sans être référencé directement par commandes : il vient de
    // clients, et c'est exactement ce que le parcours transitif doit rendre.
    act(() => result.current.basculer('public.commandes'));
    expect(result.current.manquantes).toEqual(['public.clients', 'public.pays', 'public.produits']);
  });

  it('suit les dépendances de proche en proche', () => {
    const { result } = renderHook(() => useSelection(inventaire));

    // commandes -> clients -> pays : le pays doit remonter dès le premier
    // décompte, sinon le total baisserait d'un cran à chaque ajout.
    act(() => result.current.basculer('public.commandes'));
    expect(result.current.manquantes).toContain('public.pays');

    act(() => result.current.ajouterManquantes());
    expect(result.current.manquantes).toEqual([]);
    expect(result.current.selection.size).toBe(4);
  });

  it('ne signale rien pour une table sans référence', () => {
    const { result } = renderHook(() => useSelection(inventaire));

    act(() => result.current.basculer('public.journal'));
    expect(result.current.manquantes).toEqual([]);
  });

  it('ignore une cible absente de l’inventaire', () => {
    const { result } = renderHook(() => useSelection([table('commandes', ['autre.archives'])]));

    act(() => result.current.basculer('public.commandes'));
    expect(result.current.manquantes).toEqual([]);
  });

  it('ne boucle pas sur une clé étrangère auto-référencée', () => {
    const { result } = renderHook(() => useSelection([table('salaries', ['public.salaries'])]));

    act(() => result.current.basculer('public.salaries'));
    expect(result.current.manquantes).toEqual([]);
  });

  it('ne boucle pas sur un cycle entre deux tables', () => {
    const cycle = [table('a', ['public.b']), table('b', ['public.a'])];
    const { result } = renderHook(() => useSelection(cycle));

    act(() => result.current.basculer('public.a'));
    expect(result.current.manquantes).toEqual(['public.b']);
  });

  it('coche puis décoche un schéma entier', () => {
    const { result } = renderHook(() => useSelection(inventaire));

    act(() => result.current.basculerSchema('public', inventaire));
    expect(result.current.selection.size).toBe(inventaire.length);

    act(() => result.current.basculerSchema('public', inventaire));
    expect(result.current.selection.size).toBe(0);
  });

  it('ne touche pas aux autres schémas', () => {
    const melange = [...inventaire, table('parametres', [], 'compta')];
    const { result } = renderHook(() => useSelection(melange));

    act(() => result.current.basculerSchema('compta', melange));
    expect([...result.current.selection]).toEqual(['compta.parametres']);
  });

  it('vide la sélection', () => {
    const { result } = renderHook(() => useSelection(inventaire));

    act(() => result.current.basculerSchema('public', inventaire));
    act(() => result.current.toutEffacer());
    expect(result.current.selection.size).toBe(0);
  });
});
