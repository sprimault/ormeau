// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useT } from '@/shared/i18n';
import type { Extraction } from '@/shared/model';
import { estTerminal } from '../model/taches';
import { useMaintenant } from '../model/useMaintenant';
import { ExtractionItem } from './ExtractionItem';

/** Propriétés du panneau. */
interface ProprietesPanneau {
  extractions: Extraction[];
  connecte: boolean;
}

/**
 * Détail des extractions, la plus récente en tête.
 *
 * Une coupure du flux se signale ici plutôt que dans l'indicateur : les états
 * affichés sont alors ceux d'avant la coupure, et c'est en les lisant qu'on a
 * besoin de le savoir.
 */
export function ExtractionPanel({ extractions, connecte }: ProprietesPanneau) {
  const t = useT();
  const maintenant = useMaintenant(extractions.some((e) => !estTerminal(e.etat)));

  return (
    <section
      role="dialog"
      aria-label={t('extraction.panel.title')}
      className="w-[28rem] max-w-[calc(100vw-2rem)] rounded border border-slate-200 bg-white shadow-lg dark:border-slate-800 dark:bg-slate-950"
    >
      <header className="flex items-center gap-2 border-b border-slate-200 px-3 py-1.5 dark:border-slate-800">
        <h2 className="text-xs font-semibold">{t('extraction.panel.title')}</h2>
        {connecte ? null : (
          <span className="text-xs text-amber-700 dark:text-amber-500">
            {t('extraction.stream.lost')}
          </span>
        )}
      </header>
      <ul className="max-h-[70vh] divide-y divide-slate-200 overflow-y-auto dark:divide-slate-800">
        {[...extractions].reverse().map((extraction) => (
          <ExtractionItem key={extraction.id} extraction={extraction} maintenant={maintenant} />
        ))}
      </ul>
    </section>
  );
}
