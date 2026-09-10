// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { Monitor, Moon, Sun } from 'lucide-react';

import { useT } from '@/shared/i18n';
import { useThemeStore, type Theme } from '@/shared/model';

/** Les trois états et leur icône, dans l'ordre d'affichage. */
const choix = [
  { valeur: 'clair', Icone: Sun },
  { valeur: 'sombre', Icone: Moon },
  { valeur: 'systeme', Icone: Monitor },
] as const satisfies readonly { valeur: Theme; Icone: typeof Sun }[];

/**
 * Sélecteur de thème.
 *
 * Trois boutons plutôt qu'une liste déroulante : l'état courant se lit sans
 * ouvrir quoi que ce soit, et changer d'avis coûte un clic au lieu de trois.
 * Le libellé reste accessible aux lecteurs d'écran, que l'icône seule laisserait
 * sans rien à annoncer.
 */
export function ThemeToggle() {
  const t = useT();
  const theme = useThemeStore((etat) => etat.theme);
  const setTheme = useThemeStore((etat) => etat.setTheme);

  return (
    <div className="flex items-center gap-0.5" role="group" aria-label={t('theme.label')}>
      {choix.map(({ valeur, Icone }) => (
        <button
          key={valeur}
          type="button"
          onClick={() => setTheme(valeur)}
          aria-pressed={theme === valeur}
          aria-label={t(`theme.${valeur}`)}
          title={t(`theme.${valeur}`)}
          className={`rounded p-1 ${
            theme === valeur
              ? 'bg-slate-200 text-slate-900 dark:bg-slate-700 dark:text-slate-100'
              : 'text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
          }`}
        >
          <Icone size={14} aria-hidden />
        </button>
      ))}
    </div>
  );
}
