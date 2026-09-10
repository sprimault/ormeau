// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { postJSON, supprimer } from '@/shared/api';
import type { RequeteConnexion, ReponseConnexion } from '@/shared/model';

/** Ouvre une connexion et rend ce que le serveur dit de lui-même. */
export function connecter(
  requete: RequeteConnexion,
  signal?: AbortSignal,
): Promise<ReponseConnexion> {
  return postJSON<RequeteConnexion, ReponseConnexion>('/api/connexion', requete, signal);
}

/** Referme une connexion ouverte. */
export function deconnecter(session: string, signal?: AbortSignal): Promise<void> {
  return supprimer('/api/connexion', { session }, signal);
}
