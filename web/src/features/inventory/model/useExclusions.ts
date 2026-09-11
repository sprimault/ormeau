// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useMemo } from 'react';

import { useDecisions } from '@/entities/decisions';

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
 *
 * L'état vit dans le brouillon de décisions de la base, et nulle part ici :
 * c'est celui que l'écran d'arbitrage enregistre, et celui qui arrive du
 * fichier quand la base a déjà été arbitrée. Une copie locale finirait par le
 * contredire.
 */
export function useExclusions(): EtatExclusions {
  const { decisions, modifier } = useDecisions();
  const parTable = decisions.colonnes_ignorees;

  const basculer = useCallback(
    (table: string, colonne: string) => {
      modifier((precedentes) => {
        const suivantes = { ...precedentes.colonnes_ignorees };
        const colonnes = new Set(suivantes[table] ?? []);
        if (!colonnes.delete(colonne)) {
          colonnes.add(colonne);
        }
        if (colonnes.size === 0) {
          delete suivantes[table];
        } else {
          suivantes[table] = [...colonnes].sort();
        }
        // Absente plutôt que vide : un brouillon qui n'écarte plus rien doit se
        // lire comme celui qui n'a jamais rien écarté.
        return {
          ...precedentes,
          colonnes_ignorees: Object.keys(suivantes).length > 0 ? suivantes : undefined,
        };
      });
    },
    [modifier],
  );

  const estIgnoree = useCallback(
    (table: string, colonne: string) => parTable?.[table]?.includes(colonne) ?? false,
    [parTable],
  );

  const compte = useCallback((table: string) => parTable?.[table]?.length ?? 0, [parTable]);

  // Trié, pour que deux sessions qui écartent les mêmes colonnes produisent le
  // même fichier de décisions.
  const ignorees = useMemo(() => {
    const sortie: Record<string, string[]> = {};
    for (const table of Object.keys(parTable ?? {}).sort()) {
      sortie[table] = [...(parTable?.[table] ?? [])].sort();
    }
    return sortie;
  }, [parTable]);

  const total = useMemo(
    () => Object.values(parTable ?? {}).reduce((somme, colonnes) => somme + colonnes.length, 0),
    [parTable],
  );

  return { ignorees, total, estIgnoree, compte, basculer };
}
