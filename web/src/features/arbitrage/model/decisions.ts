// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import type { Decisions, EnumerationInferee, RelationForcee } from '@/shared/model';

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

/**
 * Nomme les cas d'une énumération.
 *
 * Le fichier rattache une énumération à une colonne : portée par plusieurs,
 * elle y figure pour chacune, sous le même nom, pour que la décision vaille
 * partout où l'inférence l'a trouvée. Les cas déjà décidés restent, ceux qu'on
 * nomme les remplacent.
 */
export function nommerCas(
  d: Decisions,
  enumeration: Pick<EnumerationInferee, 'nom' | 'colonnes'>,
  cas: Record<string, string>,
): Decisions {
  const precedente = d.enumerations?.find((e) => enumeration.colonnes.includes(e.colonne));
  const nom = precedente?.nom ?? enumeration.nom;
  const fusion = { ...precedente?.cas, ...cas };
  const autres = (d.enumerations ?? []).filter((e) => !enumeration.colonnes.includes(e.colonne));
  return {
    ...d,
    enumerations: [...autres, ...enumeration.colonnes.map((colonne) => ({ colonne, nom, cas: fusion }))],
  };
}

/** Retire la décision d'énumération d'une colonne, et rend la main à l'inférence. */
export function retirerEnumeration(d: Decisions, colonne: string): Decisions {
  const reste = (d.enumerations ?? []).filter((e) => e.colonne !== colonne);
  return { ...d, enumerations: reste.length > 0 ? reste : undefined };
}

/**
 * Relie une colonne à une autre entité. Une colonne ne porte qu'une relation :
 * celle qu'elle portait déjà est remplacée.
 */
export function ajouterRelation(d: Decisions, relation: RelationForcee): Decisions {
  const autres = (d.relations_forcees ?? []).filter((r) => r.source !== relation.source);
  return { ...d, relations_forcees: [...autres, relation] };
}

/** Retire la relation posée sur une colonne. */
export function retirerRelation(d: Decisions, source: string): Decisions {
  const reste = (d.relations_forcees ?? []).filter((r) => r.source !== source);
  return { ...d, relations_forcees: reste.length > 0 ? reste : undefined };
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
