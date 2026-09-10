// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useContexte } from '@/shared/api';
import { useT } from '@/shared/i18n';
import { tronquerMilieu } from '@/shared/lib';
import { LangToggle, ThemeToggle } from '@/shared/ui';

/**
 * En-tête permanent.
 *
 * Le répertoire de travail y est affiché tout le temps, et en entier au survol :
 * on doit savoir où on écrit avant de cliquer.
 */
export function Header() {
  const t = useT();
  const contexte = useContexte();

  return (
    <header className="flex items-center justify-between gap-4 border-b border-slate-200 px-4 py-2 dark:border-slate-800">
      <div className="flex items-baseline gap-3">
        <span className="font-semibold">{t('app.name')}</span>
        {contexte ? (
          <span className="text-xs text-slate-500" title={contexte.repertoire}>
            {t('app.workdir')} : <span className="font-mono">{tronquerMilieu(contexte.repertoire)}</span>
          </span>
        ) : null}
      </div>

      <div className="flex items-center gap-3">
        {contexte ? (
          <span className="text-xs text-slate-500">
            {t('app.version', { version: contexte.version })}
          </span>
        ) : null}
        <LangToggle />
        <ThemeToggle />
      </div>
    </header>
  );
}
