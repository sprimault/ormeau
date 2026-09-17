// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useT } from '@/shared/i18n';

/**
 * Les avertissements d'une entité dans la liste : en orange ce qui se règle
 * dans l'écran, en gris ce qui ne s'y règle pas. Rien quand il n'y a ni l'un
 * ni l'autre.
 *
 * Deux badges et non un seul total : « 1 à traiter » sur une entité où rien ne
 * se règle ici envoyait chercher une action qui n'existe pas.
 */
export function PendingBadge({ aTraiter, autres }: { aTraiter: number; autres: number }) {
  const t = useT();
  if (aTraiter === 0 && autres === 0) {
    return null;
  }
  return (
    <span className="flex shrink-0 flex-col items-end gap-0.5">
      {aTraiter > 0 ? (
        <span className="rounded bg-amber-100 px-1.5 text-xs text-amber-800 dark:bg-amber-900 dark:text-amber-200">
          {t('arbitrage.pending', { n: aTraiter })}
        </span>
      ) : null}
      {autres > 0 ? (
        <span className="rounded bg-slate-100 px-1.5 text-xs text-slate-600 dark:bg-slate-800 dark:text-slate-300">
          {t('arbitrage.warningsCount', { n: autres })}
        </span>
      ) : null}
    </span>
  );
}
