// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { postJSON } from '@/shared/api';
import type {
  ReponseEcritureDecisions,
  ReponseEntite,
  ReponseInference,
  RequeteEcritureDecisions,
  RequeteEntite,
  RequeteInference,
} from '@/shared/model';

/**
 * Rejoue l'inférence sur le calque de la base, avec le brouillon de l'écran.
 * Aucune base n'est interrogée : le serveur relit le calque du répertoire.
 */
export function inferer(requete: RequeteInference, signal?: AbortSignal): Promise<ReponseInference> {
  return postJSON<RequeteInference, ReponseInference>('/api/inference', requete, signal);
}

/** Calcule une entité avec le brouillon, et rend la table dont elle vient. */
export function lireEntite(requete: RequeteEntite, signal?: AbortSignal): Promise<ReponseEntite> {
  return postJSON<RequeteEntite, ReponseEntite>('/api/inference/entite', requete, signal);
}

/** Écrit le brouillon dans le fichier de décisions de la base. */
export function ecrireDecisions(
  requete: RequeteEcritureDecisions,
): Promise<ReponseEcritureDecisions> {
  return postJSON<RequeteEcritureDecisions, ReponseEcritureDecisions>('/api/decisions', requete);
}
