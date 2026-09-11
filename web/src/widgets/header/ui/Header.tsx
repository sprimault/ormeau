// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { ExtractionIndicator } from '@/features/extraction';
import { useContexte } from '@/shared/api';
import { useT } from '@/shared/i18n';
import { LangToggle, ThemeToggle } from '@/shared/ui';

import { WorkdirField } from './WorkdirField';

/**
 * En-tête permanent.
 *
 * Le répertoire de travail y est affiché et réglé — voir WorkdirField, qui
 * porte les raisons de sa forme.
 *
 * Les extractions s'y suivent aussi. C'est le seul élément présent sur tous les
 * écrans, formulaire de connexion compris, et une extraction lancée continue
 * après la déconnexion.
 */
export function Header() {
  const t = useT();
  const { contexte, changerRepertoire } = useContexte();

  return (
    <header className="flex items-center justify-between gap-4 border-b border-slate-200 px-4 py-2 dark:border-slate-800">
      <div className="flex items-baseline gap-3">
        <span className="font-semibold">{t('app.name')}</span>
        {contexte ? (
          <WorkdirField repertoire={contexte.repertoire} changer={changerRepertoire} />
        ) : null}
      </div>

      <div className="flex items-center gap-3">
        <ExtractionIndicator />
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
