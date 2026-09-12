// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { getJSON, postJSON, supprimer } from '@/shared/api';
import type {
  RequeteConnexion,
  ReponseConnexion,
  ReponseBases,
  RequeteBase,
  RequeteFermeture,
} from '@/shared/model';

/** Ouvre une connexion et rend ce que le serveur dit de lui-même. */
export function connecter(
  requete: RequeteConnexion,
  signal?: AbortSignal,
): Promise<ReponseConnexion> {
  return postJSON<RequeteConnexion, ReponseConnexion>('/api/connexion', requete, signal);
}

/** Referme une connexion ouverte. */
export function deconnecter(session: string, signal?: AbortSignal): Promise<void> {
  return supprimer<RequeteFermeture>('/api/connexion', { session }, signal);
}

/** Liste les bases du serveur atteint. Vide quand le dialecte ne sait pas les
 *  énumérer — l'écran se passe alors du sélecteur. */
export function lireBases(session: string, signal?: AbortSignal): Promise<ReponseBases> {
  return getJSON<ReponseBases>(`/api/bases?session=${encodeURIComponent(session)}`, signal);
}

/**
 * Rebascule la session sur une autre base du même serveur.
 *
 * Les identifiants sont ceux de la connexion en cours : changer de base ne
 * refait pas saisir un mot de passe qu'on vient de donner. Le serveur rend une
 * nouvelle session, ce qui repart d'un état propre.
 */
export function changerDeBase(
  session: string,
  base: string,
  signal?: AbortSignal,
): Promise<ReponseConnexion> {
  return postJSON<RequeteBase, ReponseConnexion>('/api/base', { session, base }, signal);
}
