// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useId } from 'react';
import { CircleHelp } from 'lucide-react';

import { useT } from '@/shared/i18n';

/**
 * Petite aide posée à côté d'un libellé : un point d'interrogation qui dit, au
 * survol ou au focus clavier, à quoi sert ce qu'il accompagne.
 *
 * Une bulle et non l'attribut `title` : celui-ci n'apparaît qu'après une
 * seconde d'immobilité, jamais au clavier, et se tronque sur une phrase longue.
 */
export function HelpTip({ texte }: { texte: string }) {
  const t = useT();
  const id = useId();

  return (
    <span className="group relative inline-flex align-middle font-normal">
      <button
        type="button"
        aria-label={t('help.label')}
        aria-describedby={id}
        className="rounded-full text-slate-400 hover:text-slate-700 focus:text-slate-700 focus:outline-none dark:hover:text-slate-200 dark:focus:text-slate-200"
      >
        <CircleHelp size={14} aria-hidden />
      </button>
      <span
        id={id}
        role="tooltip"
        className="pointer-events-none invisible absolute top-full left-0 z-20 mt-1 w-72 rounded border border-slate-200 bg-white p-2 text-xs text-slate-700 shadow-lg group-focus-within:visible group-hover:visible dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200"
      >
        {texte}
      </span>
    </span>
  );
}
