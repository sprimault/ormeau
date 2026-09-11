// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { describe, expect, it } from 'vitest';

import { estCle } from '@/shared/i18n';
import {
  CodeCasEnumerationOpaque,
  CodeCibleHorsPortee,
  CodeClePrimaireComposite,
  CodeClePrimaireGardee,
  CodeCollision,
  CodeColonneIgnoree,
  CodeDecisionOrpheline,
  CodeDefautIncompatible,
  CodeFKImpliciteProbable,
  CodeHeritageDeduit,
  CodeJointurePure,
  CodeNomNonSingularisable,
  CodePrefixeDetecte,
  CodeTableIgnoree,
  CodeTableSansClePrimaire,
  CodeTraitDeduit,
  CodeTypeNonReconnu,
  type Avertissement,
} from '@/shared/model';
import { avertissementsParTable, estATraiter, trierAvertissements } from '../avertissements';

/** Fabrique un avertissement réduit à ce que le rangement regarde. */
function avertissement(code: string, cible: string, confiance = 1): Avertissement {
  return { code, cible, confiance, message: '', resolution: 'aucune' };
}

describe('avertissements', () => {
  it('trie du plus douteux au plus sûr, puis par code et par cible', () => {
    const tries = trierAvertissements([
      avertissement(CodeTableSansClePrimaire, 'public.t_log', 1),
      avertissement(CodeTypeNonReconnu, 'public.b.geo', 0.3),
      avertissement(CodeCollision, 'public.z', 1),
      avertissement(CodeTypeNonReconnu, 'public.a.geo', 0.3),
    ]);

    expect(tries.map((a) => a.cible)).toEqual(['public.a.geo', 'public.b.geo', 'public.z', 'public.t_log']);
  });

  it('range sous chaque table ce qui la vise, elle ou ses colonnes', () => {
    const parTable = avertissementsParTable(
      [
        avertissement(CodeTableSansClePrimaire, 'public.t_log'),
        avertissement(CodeTypeNonReconnu, 'public.clients.position', 0.3),
        avertissement(CodeTraitDeduit, 'public.clients', 0.8),
        avertissement(CodePrefixeDetecte, 'public'),
      ],
      ['public.clients', 'public.t_log'],
    );

    expect(parTable.get('public.clients')?.map((a) => a.code)).toEqual([CodeTypeNonReconnu, CodeTraitDeduit]);
    expect(parTable.get('public.t_log')?.map((a) => a.code)).toEqual([CodeTableSansClePrimaire]);
    expect(parTable.size).toBe(2);
  });

  it('rattache une colonne à la table la plus longue qui la préfixe', () => {
    const parTable = avertissementsParTable(
      [avertissement(CodeTypeNonReconnu, 'public.client.adresse.rue')],
      ['public.client', 'public.client.adresse'],
    );

    expect(parTable.get('public.client.adresse')).toHaveLength(1);
    expect(parTable.has('public.client')).toBe(false);
  });

  it('ne répète pas en avertissement ce qui a été choisi dans l’onglet de sélection', () => {
    const parTable = avertissementsParTable(
      [
        avertissement(CodeColonneIgnoree, 'public.clients.photo'),
        avertissement(CodeTableIgnoree, 'public.clients'),
      ],
      ['public.clients'],
    );

    expect(parTable.size).toBe(0);
  });

  it('compte comme à traiter tout ce qui n’est pas purement informatif, codes inconnus compris', () => {
    expect(estATraiter(avertissement(CodeTypeNonReconnu, 'public.a.b'))).toBe(true);
    expect(estATraiter(avertissement('code_de_demain', 'public.a'))).toBe(true);
    expect(estATraiter(avertissement(CodeTraitDeduit, 'public.a'))).toBe(false);
    expect(estATraiter(avertissement(CodeJointurePure, 'public.a'))).toBe(false);
  });

  it('traduit chaque code d’avertissement : un code ajouté côté Go appelle son libellé', () => {
    const codes = [
      CodeCasEnumerationOpaque,
      CodeCibleHorsPortee,
      CodeClePrimaireComposite,
      CodeClePrimaireGardee,
      CodeCollision,
      CodeColonneIgnoree,
      CodeDecisionOrpheline,
      CodeDefautIncompatible,
      CodeFKImpliciteProbable,
      CodeHeritageDeduit,
      CodeJointurePure,
      CodeNomNonSingularisable,
      CodePrefixeDetecte,
      CodeTableIgnoree,
      CodeTableSansClePrimaire,
      CodeTraitDeduit,
      CodeTypeNonReconnu,
    ];

    for (const code of codes) {
      expect(estCle(`warning.${code}`), code).toBe(true);
    }
  });
});
