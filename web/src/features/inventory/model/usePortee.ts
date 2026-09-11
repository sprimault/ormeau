// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useMemo } from 'react';

import { qualifier } from '@/shared/lib';
import type { Portee, TableSommaire } from '@/shared/model';

/**
 * Compose la portée qui partira à l'extraction.
 *
 * Sans sélection, tous les schémas et toutes leurs tables : une liste de tables
 * vide n'est pas une portée vide, comme en ligne de commande. Avec une
 * sélection, seulement les schémas qui en contiennent une. Le pilote lit chaque
 * schéma demandé en entier, séquences et vues comprises, et retient le premier
 * comme schéma du calque : envoyer un schéma dont aucune table n'est cochée en
 * ferait celui d'un calque qui ne le contient pas.
 *
 * Schémas dans l'ordre du serveur, tables triées : deux sélections identiques
 * donnent la même portée. Le schéma d'une table se lit dans l'inventaire et non
 * dans sa clé qualifiée, un nom de schéma entre guillemets pouvant contenir un
 * point.
 */
export function usePortee(
  schemas: string[],
  tables: TableSommaire[],
  selection: Set<string>,
): Portee {
  return useMemo(() => {
    if (selection.size === 0) {
      return { schemas, tables_incluses: [] };
    }
    const retenus = new Set(
      tables
        .filter((table) => selection.has(qualifier(table.schema, table.nom)))
        .map((table) => table.schema),
    );
    return {
      schemas: schemas.filter((schema) => retenus.has(schema)),
      tables_incluses: [...selection].sort(),
    };
  }, [schemas, tables, selection]);
}
