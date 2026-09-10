// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useT } from '@/shared/i18n';
import { useThemeStore, type Theme } from '@/shared/model';

const ordre: Theme[] = ['clair', 'sombre', 'systeme'];

/** Sélecteur de thème, à trois états plutôt qu'un basculement : « système »
 *  est le défaut et suit le poste en direct. */
export function ThemeToggle() {
  const t = useT();
  const theme = useThemeStore((etat) => etat.theme);
  const setTheme = useThemeStore((etat) => etat.setTheme);

  return (
    <label className="flex items-center gap-1 text-xs text-slate-500 dark:text-slate-400">
      <span className="sr-only">{t('theme.label')}</span>
      <select
        value={theme}
        onChange={(evenement) => setTheme(evenement.target.value as Theme)}
        aria-label={t('theme.label')}
        className="rounded border border-slate-300 bg-white px-1.5 py-1 text-xs dark:border-slate-700 dark:bg-slate-900"
      >
        {ordre.map((valeur) => (
          <option key={valeur} value={valeur}>
            {t(`theme.${valeur}`)}
          </option>
        ))}
      </select>
    </label>
  );
}
