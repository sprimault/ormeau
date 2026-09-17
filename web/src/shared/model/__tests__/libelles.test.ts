// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { describe, expect, it } from 'vitest';

import { messages } from '@/shared/i18n';
import * as api from '../api';
import * as calque from '../calque';
import * as index from '..';

/**
 * Rend les constantes chaîne d'un module généré dont le nom commence par le
 * préfixe, avec leur nom pour le message d'échec.
 *
 * Le module est lu directement, et non par l'index de shared/model : ses
 * réexportations sont écrites à la main, et un code ajouté côté Go n'y
 * figurerait pas.
 */
function constantes(module: object, prefixe: string): [string, string][] {
  return Object.entries(module)
    .filter(([nom, valeur]) => nom.startsWith(prefixe) && typeof valeur === 'string')
    .map(([nom, valeur]) => [nom, valeur as string]);
}

/** Les deux langues, que le type des dictionnaires tient déjà à parité. */
const langues = ['fr', 'en'] as const;

/**
 * Les codes du calque arrivent générés depuis internal/calque : avertissements
 * de logique.go, affichés sous warning., anomalies de validation.go, affichées
 * sous anomaly.. Sans libellé, l'écran montre le code brut. Le balayage attrape
 * un code ajouté côté Go sans son texte ; il ne vérifie pas la famille, que les
 * constantes ne portent pas à l'exécution et qu'une erreur ne toucherait qu'en
 * ajoutant un code. Si elle doit être garantie, c'est par un type par famille
 * côté Go.
 */
describe('libellés des codes du calque (warning., anomaly.)', () => {
  const codes = constantes(calque, 'Code');

  it('trouve des codes libellés dans chacune des deux familles', () => {
    // Sans ce garde-fou, un renommage des constantes viderait la liste, et le
    // test suivant passerait sans rien vérifier.
    const fr: Record<string, string> = messages.fr;
    expect(codes.filter(([, code]) => fr[`warning.${code}`]).length, 'aucun code libellé sous warning.').toBeGreaterThan(0);
    expect(codes.filter(([, code]) => fr[`anomaly.${code}`]).length, 'aucun code libellé sous anomaly.').toBeGreaterThan(0);
  });

  it.each(langues)('donne à chaque code un libellé warning. ou anomaly., et un seul, en %s', (langue) => {
    const dictionnaire: Record<string, string> = messages[langue];
    for (const [nom, code] of codes) {
      const familles = ['warning', 'anomaly'].filter((famille) => dictionnaire[`${famille}.${code}`]);
      expect(familles, `${langue} : ${nom} (« ${code} ») attend warning.${code} ou anomaly.${code}, trouvé ${familles.length}`).toHaveLength(1);
    }
  });
});

/**
 * L'index de shared/model est écrit à la main : un code d'avertissement généré
 * qu'il ne réexporterait pas manquerait aux features, qui ne lisent que lui —
 * au rangement des avertissements par lieu, notamment.
 */
describe('réexportation des codes d’avertissement', () => {
  it('réexporte chaque code libellé sous warning.', () => {
    const fr: Record<string, string> = messages.fr;
    const exportes = new Set(Object.keys(index));
    for (const [nom, code] of constantes(calque, 'Code')) {
      if (fr[`warning.${code}`]) {
        expect(exportes.has(nom), `${nom} n'est pas réexporté par shared/model`).toBe(true);
      }
    }
  });
});

/**
 * Les états d'une tâche d'extraction, générés depuis internal/interface, sous
 * extraction.state..
 */
describe('libellés des états d’extraction (extraction.state.)', () => {
  const etats = constantes(api, 'Etat');

  it('trouve les états générés', () => {
    expect(etats.length).toBeGreaterThan(0);
  });

  it.each(langues)('donne un libellé extraction.state. à chaque état en %s', (langue) => {
    const dictionnaire: Record<string, string> = messages[langue];
    for (const [nom, etat] of etats) {
      expect(dictionnaire[`extraction.state.${etat}`], `${langue} : ${nom} attend extraction.state.${etat}`).toBeTruthy();
    }
  });
});
