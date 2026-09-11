// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { getJSON } from '@/shared/api';
import type { ReponseDecisions } from '@/shared/model';

/**
 * Lit le fichier de décisions d'une base, dans le répertoire de travail.
 *
 * Le nom de la base part, jamais un chemin : le serveur le valide et compose le
 * fichier lui-même. Un fichier absent n'est pas une erreur, c'est une base
 * encore jamais arbitrée.
 */
export function lireDecisions(base: string, signal?: AbortSignal): Promise<ReponseDecisions> {
  return getJSON<ReponseDecisions>(`/api/decisions?${new URLSearchParams({ base })}`, signal);
}
