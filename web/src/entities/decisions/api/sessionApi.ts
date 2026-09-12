// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { getJSON, postJSON, supprimer } from '@/shared/api';
import type { ReferenceSession, ReponseSession, Session } from '@/shared/model';

/**
 * Lit le brouillon d'arbitrage d'une base.
 *
 * Le serveur vérifie lui-même qu'il s'applique encore : il a le calque et le
 * fichier de décisions sous la main, la page n'a que ce qu'on lui a dit. Un
 * brouillon écarté revient avec sa raison, pour que l'écran la dise.
 */
export function lireSession(base: string, signal?: AbortSignal): Promise<ReponseSession> {
  return getJSON<ReponseSession>(`/api/session?base=${encodeURIComponent(base)}`, signal);
}

/** Enregistre le brouillon d'arbitrage. */
export function ecrireSession(session: Session, signal?: AbortSignal): Promise<ReponseSession> {
  return postJSON<Session, ReponseSession>('/api/session', session, signal);
}

/** Efface le brouillon, ce que l'écran fait après un enregistrement réussi. */
export function effacerSession(base: string): Promise<void> {
  return supprimer<ReferenceSession>('/api/session', { base });
}
