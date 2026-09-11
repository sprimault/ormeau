// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import {
  EtatAnnulee,
  EtatEchouee,
  EtatTerminee,
  type EtatExtraction,
  type Extraction,
} from '@/shared/model';

/** Dit si une tâche a fini d'évoluer. */
export function estTerminal(etat: EtatExtraction): boolean {
  return etat === EtatTerminee || etat === EtatEchouee || etat === EtatAnnulee;
}

/**
 * Nom du calque qu'écrira l'extraction d'une base.
 *
 * Pour l'affichage avant lancement seulement : c'est le serveur qui compose le
 * chemin, et la tâche lancée porte le nom qu'il a retenu.
 */
export function fichierCalque(base: string): string {
  return `${base}.calque.json`;
}

/**
 * Durée d'une tâche en secondes, fraction comprise : arrêtée à sa fin, courant
 * jusqu'à maintenant sinon, absente tant qu'elle n'a pas démarré.
 *
 * Les deux bornes viennent de deux horloges — le serveur pose le début, le
 * navigateur fournit maintenant. Elles tournent sur le même poste, mais un léger
 * écart ne doit pas afficher une durée négative.
 */
export function secondesEcoulees(extraction: Extraction, maintenant: number): number | null {
  if (!extraction.debut) {
    return null;
  }
  const fin = extraction.fin ? Date.parse(extraction.fin) : maintenant;
  return Math.max(0, (fin - Date.parse(extraction.debut)) / 1000);
}

/**
 * Associe à chaque extraction terminée dont le fichier a été réécrit depuis la
 * fin de la plus récente qui l'a remplacée.
 *
 * Toutes les extractions d'une base écrivent le même fichier : dès qu'une autre
 * s'est terminée après elle, le résumé d'une ligne ne décrit plus ce qui est sur
 * disque. Les instants du serveur ont tous le même format et le même fuseau, ils
 * se comparent comme des chaînes.
 */
export function remplacements(extractions: Extraction[]): Map<string, string> {
  const dernieres = new Map<string, string>();
  for (const extraction of extractions) {
    if (extraction.etat === EtatTerminee && extraction.fin) {
      if (extraction.fin > (dernieres.get(extraction.fichier) ?? '')) {
        dernieres.set(extraction.fichier, extraction.fin);
      }
    }
  }

  const remplacees = new Map<string, string>();
  for (const extraction of extractions) {
    const derniere = dernieres.get(extraction.fichier);
    if (extraction.etat === EtatTerminee && extraction.fin && derniere && derniere > extraction.fin) {
      remplacees.set(extraction.id, derniere);
    }
  }
  return remplacees;
}
