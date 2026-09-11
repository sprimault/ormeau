// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import type { ReactNode } from 'react';

/** Propriétés de l'alerte. */
interface ProprietesAlerte {
  message: string;
  /** Les réponses possibles. */
  children?: ReactNode;
}

/**
 * Situation qui attend une réponse avant d'aller plus loin : calque réécrit,
 * fichier modifié ailleurs, travail manuel à écraser.
 *
 * Distincte du bandeau d'erreur : rien n'a échoué, l'écran demande quoi faire.
 */
export function Alerte({ message, children }: ProprietesAlerte) {
  return (
    <div className="px-4 pt-2">
      <div
        role="status"
        className="flex flex-wrap items-center gap-2 rounded border border-amber-300 bg-amber-50 px-3 py-2 text-sm text-amber-900 dark:border-amber-800 dark:bg-amber-950 dark:text-amber-200"
      >
        <p className="min-w-0 flex-1">{message}</p>
        {children}
      </div>
    </div>
  );
}
