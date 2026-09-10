// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useLangStore, useT, type Lang } from '@/shared/i18n';

const langues: Lang[] = ['fr', 'en'];

/** Sélecteur de langue, permanent et à côté du thème. */
export function LangToggle() {
  const t = useT();
  const lang = useLangStore((etat) => etat.lang);
  const setLang = useLangStore((etat) => etat.setLang);

  return (
    <div className="flex items-center gap-0.5" role="group" aria-label={t('lang.label')}>
      {langues.map((valeur) => (
        <button
          key={valeur}
          type="button"
          onClick={() => setLang(valeur)}
          aria-pressed={lang === valeur}
          className={`rounded px-1.5 py-1 text-xs uppercase ${
            lang === valeur
              ? 'bg-slate-200 font-semibold text-slate-900 dark:bg-slate-700 dark:text-slate-100'
              : 'text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
          }`}
        >
          {valeur}
        </button>
      ))}
    </div>
  );
}
