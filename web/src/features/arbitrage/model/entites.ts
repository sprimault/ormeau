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
  /** Avertissements qui se règlent dans l'écran. */
  aTraiter: number;
  /**
   * Avertissements qui se traitent ailleurs, ou n'informent que. Comptés à
   * part : les additionner aux premiers ferait croire à une action possible
   * ici.
   */
  autres: number;
}

/** Rend une ligne par entité, dans l'ordre de l'inférence. */
export function lignesEntites(
  entites: ResumeEntite[],
  parTable: Map<string, Avertissement[]>,
): LigneEntite[] {
  return entites.map((entite) => {
    const qualifiee = qualifier(entite.table.schema, entite.table.nom);
    const avertissements = parTable.get(qualifiee) ?? [];
    const aTraiter = avertissements.filter(estATraiter).length;
    return {
      qualifiee,
      table: entite.table,
      nom: entite.nom,
      aTraiter,
      autres: avertissements.length - aTraiter,
    };
  });
}
