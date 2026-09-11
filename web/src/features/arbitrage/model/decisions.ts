// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import type { Decisions } from '@/shared/model';

// Chaque transformation rend un brouillon neuf et n'y laisse ni clé vide ni
// collection vide : le fichier ne les écrit pas, et un brouillon qui ne décide
// plus rien doit se lire comme celui qui n'a jamais rien décidé.

/** Ce que l'écran passe à ses parties pour modifier le brouillon. */
export type Modifier = (transformation: (decisions: Decisions) => Decisions) => void;

/**
 * Nom du fichier de décisions d'une base, pour l'afficher. C'est le serveur qui
 * compose le chemin.
 */
export function fichierDecisions(base: string): string {
  return `${base}.decisions.yaml`;
}

/** Décide le nom de classe d'une table. Un nom vide rend la main à l'inférence. */
export function renommer(d: Decisions, table: string, nom: string): Decisions {
  return { ...d, renommages: poser(d.renommages, table, nom) };
}

/** Force le type Doctrine d'une colonne. Un type vide rend la main à l'inférence. */
export function forcerType(d: Decisions, colonne: string, type: string): Decisions {
  return { ...d, types_forces: poser(d.types_forces, colonne, type) };
}

/** Pose ou retire une entrée d'un dictionnaire, sans le laisser vide. */
function poser(
  dictionnaire: Record<string, string> | undefined,
  cle: string,
  valeur: string,
): Record<string, string> | undefined {
  const suivant = { ...dictionnaire };
  if (valeur === '') {
    delete suivant[cle];
  } else {
    suivant[cle] = valeur;
  }
  return Object.keys(suivant).length > 0 ? suivant : undefined;
}
