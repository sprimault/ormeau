// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useT } from '@/shared/i18n';
import { heure } from '@/shared/lib';
import { CopyButton, HelpTip } from '@/shared/ui';
import { fichierCalque } from '../model/taches';
import { useCalque, useVersionCalque } from '../model/useCalque';

/** Propriétés de l'aperçu du calque. */
interface ProprietesCalque {
  session: string;
  /** Base de la session, qui nomme le calque. */
  base: string;
}

/**
 * Le calque de la base, tel qu'il est écrit dans le répertoire de travail.
 *
 * Ce que l'extraction a produit, à côté de ce qu'on lui demande : la portée dit
 * ce qui partira, ce fichier dit ce qui est arrivé. Il se relit quand le flux
 * annonce la fin d'une extraction de cette base, et montre sinon le calque
 * qu'une session précédente a laissé.
 */
export function CalquePreview({ session, base }: ProprietesCalque) {
  const t = useT();
  const { calque, message, enCours } = useCalque(session, useVersionCalque(base));

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex items-center gap-2 px-3 py-1">
        <h3 className="inline-flex items-center gap-1 text-xs font-medium text-slate-600 dark:text-slate-400">
          {t('calque.title')}
          <HelpTip texte={t('calque.help.preview')} />
        </h3>
        <span className="font-mono text-xs text-slate-500">
          {calque?.fichier ?? fichierCalque(base)}
        </span>
        {calque?.extrait_le ? (
          <span className="text-xs text-slate-500">
            {t('calque.extractedAt', { heure: heure(calque.extrait_le) })}
          </span>
        ) : null}
        <div className="ml-auto">
          {calque ? <CopyButton valeur={calque.contenu} libelle={t('calque.copy')} /> : null}
        </div>
      </div>

      {calque?.statistiques_retirees ? (
        <p className="px-3 text-xs text-amber-700 dark:text-amber-500">{t('calque.statsHidden')}</p>
      ) : null}

      {calque ? (
        <pre className="min-h-0 flex-1 overflow-auto px-3 pb-2 font-mono text-xs">
          {calque.contenu}
        </pre>
      ) : (
        <p className="px-3 text-xs text-slate-500">{enCours ? t('calque.loading') : message}</p>
      )}
    </div>
  );
}
