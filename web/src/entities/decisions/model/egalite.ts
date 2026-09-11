// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import type { Decisions } from '@/shared/model';

/**
 * Dit si deux brouillons décident la même chose.
 *
 * Comparés sous la forme que le fichier écrit, pas objet par objet : une clé
 * absente, une chaîne vide et une collection vide n'y figurent pas, et l'ordre
 * des clés d'un objet n'y change rien. Sans cela, écarter puis remettre une
 * colonne laisserait l'écran annoncer des modifications non enregistrées.
 *
 * L'ordre des listes compte, lui : le fichier le conserve.
 */
export function decisionsEgales(a: Decisions, b: Decisions): boolean {
  return JSON.stringify(canonique(a)) === JSON.stringify(canonique(b));
}

/** Retire ce qui ne s'écrit pas, et trie les clés des objets. */
function canonique(valeur: unknown): unknown {
  if (Array.isArray(valeur)) {
    return valeur.map(canonique);
  }
  if (valeur === null || typeof valeur !== 'object') {
    return valeur;
  }

  const sortie: Record<string, unknown> = {};
  for (const cle of Object.keys(valeur).sort()) {
    const propre = canonique((valeur as Record<string, unknown>)[cle]);
    if (!vide(propre)) {
      sortie[cle] = propre;
    }
  }
  return sortie;
}

/** Dit si une valeur disparaît à l'écriture du fichier. */
function vide(valeur: unknown): boolean {
  if (valeur === undefined || valeur === '') {
    return true;
  }
  if (Array.isArray(valeur)) {
    return valeur.length === 0;
  }
  return valeur !== null && typeof valeur === 'object' && Object.keys(valeur).length === 0;
}
