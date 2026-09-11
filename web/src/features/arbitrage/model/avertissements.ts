// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import {
  CodeClePrimaireComposite,
  CodeColonneIgnoree,
  CodeHeritageDeduit,
  CodeJointurePure,
  CodeTableIgnoree,
  CodeTraitDeduit,
  type Avertissement,
} from '@/shared/model';

/**
 * Codes qui ne font que rappeler un choix fait dans l'onglet de sélection. Le
 * détail montre la colonne écartée ; le répéter en avertissement ferait croire
 * à un problème.
 */
const RAPPELS = new Set<string>([CodeTableIgnoree, CodeColonneIgnoree]);

/**
 * Codes qui informent sans rien demander. Tout autre code compte comme à
 * traiter, y compris un code ajouté plus tard côté Go : il se voit plutôt que
 * de passer inaperçu.
 */
const INFORMATIFS = new Set<string>([
  CodeTraitDeduit,
  CodeJointurePure,
  CodeHeritageDeduit,
  CodeClePrimaireComposite,
]);

/** Dit si un avertissement demande qu'on s'en occupe. */
export function estATraiter(avertissement: Avertissement): boolean {
  return !INFORMATIFS.has(avertissement.code);
}

/**
 * Range les avertissements sous l'entité de la table qu'ils visent.
 *
 * La cible qualifie la table, puis éventuellement la colonne : `public.clients`
 * ou `public.clients.statut`. La table retenue est la plus longue qui la
 * préfixe, pour qu'un point dans un nom ne rattache pas une colonne à la
 * mauvaise table. Les rappels de choix sont écartés, et ce qui ne vise aucune
 * de ces tables n'apparaît pas. Chaque liste va du plus douteux au plus sûr.
 */
export function avertissementsParTable(
  avertissements: Avertissement[],
  tables: string[],
): Map<string, Avertissement[]> {
  const parLongueur = [...tables].sort((a, b) => b.length - a.length);
  const parTable = new Map<string, Avertissement[]>();

  for (const avertissement of trierAvertissements(avertissements)) {
    if (RAPPELS.has(avertissement.code)) {
      continue;
    }
    const { cible } = avertissement;
    const table = parLongueur.find((t) => cible === t || cible.startsWith(`${t}.`));
    if (table !== undefined) {
      parTable.set(table, [...(parTable.get(table) ?? []), avertissement]);
    }
  }
  return parTable;
}

/**
 * Trie les avertissements du plus douteux au plus sûr.
 *
 * Une confiance basse est ce que l'inférence a le moins bien tranché. À
 * confiance égale, code puis cible : l'ordre ne bouge pas d'un recalcul à
 * l'autre.
 */
export function trierAvertissements(avertissements: Avertissement[]): Avertissement[] {
  return [...avertissements].sort(
    (a, b) => a.confiance - b.confiance || comparer(a.code, b.code) || comparer(a.cible, b.cible),
  );
}

/** Compare deux chaînes sans dépendre de la langue du navigateur. */
function comparer(a: string, b: string): number {
  return a < b ? -1 : a > b ? 1 : 0;
}
