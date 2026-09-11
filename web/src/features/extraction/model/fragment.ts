// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import type { Physique } from '@/shared/model';

/**
 * Extrait d'un calque sérialisé l'entrée d'une table, mise en forme comme dans
 * le fichier. Rend null quand la table n'y figure pas.
 *
 * Deux espaces d'indentation et l'ordre des clés du document relu : le fragment
 * se lit comme le passage correspondant du fichier, et se compare à l'œil.
 *
 * Seule l'entrée de la table est réécrite. Les bornes des séquences dépassent la
 * précision d'un nombre JavaScript et seraient faussées par un aller-retour,
 * mais elles n'en font pas partie.
 */
export function fragmentDeTable(contenu: string, schema: string, nom: string): string | null {
  const calque = JSON.parse(contenu) as Physique;
  const table = calque.tables.find(
    (candidate) => candidate.schema === schema && candidate.nom === nom,
  );
  return table ? JSON.stringify(table, null, 2) : null;
}
