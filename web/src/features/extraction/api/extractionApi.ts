// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { getJSON, postJSON, supprimer } from '@/shared/api';
import type {
  Extraction,
  Portee,
  ReferenceExtraction,
  ReponseCalque,
  RequeteExtraction,
} from '@/shared/model';

/**
 * Lance l'extraction de la base d'une session.
 *
 * Rend la tâche telle qu'elle vient d'être inscrite ; la suite arrive par le
 * flux. Ni la base ni le fichier ne partent d'ici : le serveur les tire de la
 * session.
 */
export function lancerExtraction(
  session: string,
  portee: Portee,
  signal?: AbortSignal,
): Promise<Extraction> {
  return postJSON<RequeteExtraction, Extraction>('/api/extractions', { session, portee }, signal);
}

/** Annule une tâche en cours ou en attente, retire une tâche finie. */
export function retirerExtraction(id: string, signal?: AbortSignal): Promise<void> {
  return supprimer<ReferenceExtraction>('/api/extractions', { id }, signal);
}

/**
 * Ouvre le flux des tâches.
 *
 * EventSource plutôt que fetch : il se reconnecte seul et renvoie Last-Event-ID,
 * ce qui suffit au serveur pour rendre exactement ce qui a été manqué.
 */
export function ouvrirFlux(): EventSource {
  return new EventSource('/api/extractions/evenements');
}

/**
 * Lit le calque de la base d'une session, dans le répertoire de travail.
 *
 * Aucun nom de fichier ne part d'ici : le serveur le compose à partir de la
 * session, comme à l'extraction.
 */
export function lireCalque(session: string, signal?: AbortSignal): Promise<ReponseCalque> {
  return getJSON<ReponseCalque>(`/api/calque?${new URLSearchParams({ session })}`, signal);
}
