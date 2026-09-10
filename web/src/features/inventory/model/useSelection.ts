// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useMemo, useState } from 'react';

import { qualifier } from '@/shared/lib';
import type { TableSommaire } from '@/shared/model';

/** Ce que l'écran de sélection manipule. */
export interface EtatSelection {
  /** Tables retenues, par nom qualifié. */
  selection: Set<string>;
  /**
   * Références sortantes qui ne sont pas dans la sélection.
   *
   * C'est le piège de cet écran : exclure une table à l'extraction n'est pas
   * l'exclure à l'inférence. Décocher une table référencée laisse une clé
   * étrangère pointant dans le vide, et l'association disparaît sans bruit.
   */
  manquantes: string[];
  basculer: (cle: string) => void;
  basculerSchema: (schema: string, tables: TableSommaire[]) => void;
  ajouterManquantes: () => void;
  toutEffacer: () => void;
}

/**
 * Tient la sélection de tables.
 *
 * En mémoire, et nulle part ailleurs : elle ne sert qu'à l'extraction qui suit,
 * et le stockage local est réservé aux préférences d'affichage, jamais à ce qui
 * touche la base.
 */
export function useSelection(tables: TableSommaire[]): EtatSelection {
  const [selection, setSelection] = useState<Set<string>>(new Set());

  const basculer = useCallback((cle: string) => {
    setSelection((precedente) => {
      const suivante = new Set(precedente);
      if (!suivante.delete(cle)) {
        suivante.add(cle);
      }
      return suivante;
    });
  }, []);

  // Cocher un schéma entier quand il en manque une, tout décocher sinon : c'est
  // ce qu'attend quelqu'un qui clique deux fois de suite sur le même en-tête.
  const basculerSchema = useCallback((schema: string, tablesDuSchema: TableSommaire[]) => {
    setSelection((precedente) => {
      const suivante = new Set(precedente);
      const cles = tablesDuSchema
        .filter((table) => table.schema === schema)
        .map((table) => qualifier(table.schema, table.nom));
      const toutes = cles.every((cle) => suivante.has(cle));
      for (const cle of cles) {
        if (toutes) {
          suivante.delete(cle);
        } else {
          suivante.add(cle);
        }
      }
      return suivante;
    });
  }, []);

  // Recalculé à chaque changement de sélection plutôt que maintenu à côté :
  // deux états qui décrivent la même chose finissent toujours par diverger.
  //
  // Le parcours est transitif : une table dont on ajoute la dépendance amène
  // les siennes. S'arrêter au premier niveau afficherait « 3 manquantes », puis
  // « 2 » après les avoir ajoutées, et ainsi de suite — un décompte qui ne
  // descend jamais d'un coup fait douter de ce qu'il mesure. Le Set des vues
  // ferme les cycles, qu'une clé étrangère auto-référencée suffit à créer.
  const manquantes = useMemo(() => {
    const parCle = new Map(tables.map((table) => [qualifier(table.schema, table.nom), table]));
    const absentes = new Set<string>();
    const vues = new Set<string>();
    const aVisiter = [...selection];

    while (aVisiter.length > 0) {
      const cle = aVisiter.pop() as string;
      if (vues.has(cle)) {
        continue;
      }
      vues.add(cle);

      for (const cible of parCle.get(cle)?.reference_vers ?? []) {
        // Une cible hors inventaire — un schéma qu'on n'a pas demandé — ne se
        // rattrape pas en la cochant : la proposer n'aiderait personne.
        if (!parCle.has(cible) || selection.has(cible)) {
          continue;
        }
        absentes.add(cible);
        aVisiter.push(cible);
      }
    }
    return [...absentes].sort();
  }, [tables, selection]);

  const ajouterManquantes = useCallback(() => {
    setSelection((precedente) => new Set([...precedente, ...manquantes]));
  }, [manquantes]);

  const toutEffacer = useCallback(() => setSelection(new Set()), []);

  return {
    selection,
    manquantes,
    basculer,
    basculerSchema,
    ajouterManquantes,
    toutEffacer,
  };
}
