// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { qualifier } from '@/shared/lib';
import type { Avertissement, ReferenceTable, ResumeEntite } from '@/shared/model';
import { estATraiter } from './avertissements';

/** Une entité de la liste, avec ce que la ligne en montre. */
export interface LigneEntite {
  qualifiee: string;
  table: ReferenceTable;
  nom: string;
  aTraiter: number;
}

/** Rend une ligne par entité, dans l'ordre de l'inférence. */
export function lignesEntites(
  entites: ResumeEntite[],
  parTable: Map<string, Avertissement[]>,
): LigneEntite[] {
  return entites.map((entite) => {
    const qualifiee = qualifier(entite.table.schema, entite.table.nom);
    return {
      qualifiee,
      table: entite.table,
      nom: entite.nom,
      aTraiter: (parTable.get(qualifiee) ?? []).filter(estATraiter).length,
    };
  });
}
