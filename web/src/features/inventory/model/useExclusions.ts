// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useMemo, useState } from 'react';

/** Colonnes retirées des entités, table par table. */
export interface EtatExclusions {
  /** Ce qui partira dans `colonnes_ignorees`, tables et colonnes triées. */
  ignorees: Record<string, string[]>;
  /** Nombre de colonnes écartées, pour l'afficher sans reparcourir. */
  total: number;
  estIgnoree: (table: string, colonne: string) => boolean;
  compte: (table: string) => number;
  basculer: (table: string, colonne: string) => void;
}

/**
 * Tient les colonnes qu'on ne veut pas voir dans les entités.
 *
 * Elles ne touchent pas à l'extraction : le calque physique garde toutes les
 * colonnes, sinon le mode diff les signalerait comme disparues à chaque
 * comparaison avec la base. Ce que l'on décoche ici alimente `colonnes_ignorees`
 * dans le fichier de décisions, qui retire la propriété de l'entité — et se
 * défait six mois plus tard sans rouvrir la connexion.
 */
export function useExclusions(): EtatExclusions {
  const [parTable, setParTable] = useState<Record<string, Set<string>>>({});

  const basculer = useCallback((table: string, colonne: string) => {
    setParTable((precedent) => {
      const suivant = { ...precedent };
      const colonnes = new Set(suivant[table] ?? []);
      if (!colonnes.delete(colonne)) {
        colonnes.add(colonne);
      }
      if (colonnes.size === 0) {
        delete suivant[table];
      } else {
        suivant[table] = colonnes;
      }
      return suivant;
    });
  }, []);

  const estIgnoree = useCallback(
    (table: string, colonne: string) => parTable[table]?.has(colonne) ?? false,
    [parTable],
  );

  const compte = useCallback((table: string) => parTable[table]?.size ?? 0, [parTable]);

  // Trié, pour que deux sessions qui écartent les mêmes colonnes produisent le
  // même fichier de décisions.
  const ignorees = useMemo(() => {
    const sortie: Record<string, string[]> = {};
    for (const table of Object.keys(parTable).sort()) {
      sortie[table] = [...parTable[table]].sort();
    }
    return sortie;
  }, [parTable]);

  const total = useMemo(
    () => Object.values(parTable).reduce((somme, colonnes) => somme + colonnes.size, 0),
    [parTable],
  );

  return { ignorees, total, estIgnoree, compte, basculer };
}
