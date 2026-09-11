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
 * Durée d'une tâche en secondes : arrêtée à sa fin, courant jusqu'à maintenant
 * sinon, absente tant qu'elle n'a pas démarré.
 *
 * Les deux bornes viennent de deux horloges — le serveur pose le début, le
 * navigateur fournit maintenant. Elles tournent sur le même poste, mais un écart
 * d'une seconde ne doit pas afficher une durée négative.
 */
export function secondesEcoulees(extraction: Extraction, maintenant: number): number | null {
  if (!extraction.debut) {
    return null;
  }
  const fin = extraction.fin ? Date.parse(extraction.fin) : maintenant;
  return Math.max(0, Math.round((fin - Date.parse(extraction.debut)) / 1000));
}
