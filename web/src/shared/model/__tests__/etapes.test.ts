// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { describe, expect, it } from 'vitest';

import { messages } from '@/shared/i18n';
import * as introspection from '../introspection';

/**
 * Les codes d'étape arrivent générés depuis internal/introspection/sommaire.go.
 * Le test les parcourt tous plutôt que d'en recopier la liste : un code ajouté
 * côté Go apparaît ici à la régénération des types, et échoue tant qu'il n'a
 * pas de texte. L'erreur attrapée n'est pas « le front ignore ce code », c'est
 * « ce code n'a pas de texte », dans l'une ou l'autre langue.
 */
describe('étapes d’extraction', () => {
  const codes = Object.entries(introspection)
    .filter(([nom, valeur]) => nom.startsWith('Etape') && typeof valeur === 'string')
    .map(([, valeur]) => valeur as string);

  it('trouve les codes générés', () => {
    // Sans ce garde-fou, un renommage des constantes viderait la liste et le
    // test suivant passerait sans rien vérifier.
    expect(codes.length).toBeGreaterThan(0);
  });

  it.each(['fr', 'en'] as const)('donne un texte à chaque code en %s', (langue) => {
    const dictionnaire: Record<string, string> = messages[langue];
    for (const code of codes) {
      expect(dictionnaire[`extraction.step.${code}`], `${langue} : ${code}`).toBeTruthy();
    }
  });
});
