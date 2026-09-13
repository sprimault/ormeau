// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { describe, expect, it } from 'vitest';

import { estCle } from '@/shared/i18n';
import {
  CodeActionInconnue,
  CodeAriteIncoherente,
  CodeChampRequisVide,
  CodeColonneDupliquee,
  CodeColonneIntrouvable,
  CodeEmpreinteMalformee,
  CodeGenreDefautInconnu,
  CodePositionInvalide,
  CodeSGBDInconnu,
  CodeStatistiquesOrphelines,
  CodeTableCibleIntrouvable,
  CodeTableDupliquee,
  CodeTableSansColonne,
  CodeTypeEnumereIntrouvable,
  CodeTypeHorsVocabulaire,
  CodeVersionInconnue,
  EtatAnnulee,
  EtatEchouee,
  EtatEnAttente,
  EtatEnCours,
  EtatTerminee,
} from '@/shared/model';

/**
 * ExtractionItem affiche un code sans libellé tel quel plutôt que vide : rien
 * n'échoue quand le Go en ajoute un sans sa traduction, ni tsc ni le contrôle
 * des types générés. Ces listes sont ce qui le fait échouer.
 *
 * Des listes et non un balayage des exports `Code*` : calque.ts porte aussi
 * les codes d'avertissement, qui se traduisent sous un autre préfixe. Les étapes
 * d'extraction, elles, se balaient : voir shared/model/__tests__/etapes.test.ts.
 */
describe('libellés de l’extraction', () => {
  it('traduit chaque code d’anomalie du calque', () => {
    const codes = [
      CodeActionInconnue,
      CodeAriteIncoherente,
      CodeChampRequisVide,
      CodeColonneDupliquee,
      CodeColonneIntrouvable,
      CodeEmpreinteMalformee,
      CodeGenreDefautInconnu,
      CodePositionInvalide,
      CodeSGBDInconnu,
      CodeStatistiquesOrphelines,
      CodeTableCibleIntrouvable,
      CodeTableDupliquee,
      CodeTableSansColonne,
      CodeTypeEnumereIntrouvable,
      CodeTypeHorsVocabulaire,
      CodeVersionInconnue,
    ];

    for (const code of codes) {
      expect(estCle(`anomaly.${code}`), code).toBe(true);
    }
  });

  it('traduit chaque état d’une tâche', () => {
    for (const etat of [EtatEnAttente, EtatEnCours, EtatTerminee, EtatEchouee, EtatAnnulee]) {
      expect(estCle(`extraction.state.${etat}`), etat).toBe(true);
    }
  });

});
