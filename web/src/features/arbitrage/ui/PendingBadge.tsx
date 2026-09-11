// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useT } from '@/shared/i18n';

/** Nombre d'avertissements à traiter, rien quand il n'y en a pas. */
export function PendingBadge({ n }: { n: number }) {
  const t = useT();
  if (n === 0) {
    return null;
  }
  return (
    <span className="shrink-0 rounded bg-amber-100 px-1.5 text-xs text-amber-800 dark:bg-amber-900 dark:text-amber-200">
      {t('arbitrage.pending', { n })}
    </span>
  );
}
