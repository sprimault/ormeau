// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useMemo, useState } from 'react';

import { useT } from '@/shared/i18n';
import { heure, qualifier } from '@/shared/lib';
import { CopyButton } from '@/shared/ui';
import { useExtractions } from '../model/contexte';
import { fragmentDeTable } from '../model/fragment';
import { fichierCalque } from '../model/taches';
import { derniereFin, useCalque } from '../model/useCalque';

/** Propriétés de l'aperçu du calque. */
interface ProprietesCalque {
  session: string;
  /** Base de la session, qui nomme le calque. */
  base: string;
  /** Table cliquée dans l'arbre : la vue se resserre sur son entrée. */
  table?: { schema: string; nom: string } | null;
}

/**
 * Le calque de la base, tel qu'il est écrit dans le répertoire de travail.
 *
 * Ce que l'extraction a produit, à côté de ce qu'on lui demande : la portée dit
 * ce qui partira, ce fichier dit ce qui est arrivé. Il se relit quand le flux
 * annonce la fin d'une extraction de cette base, et montre sinon le calque
 * qu'une session précédente a laissé.
 *
 * Quand une table est cliquée dans l'arbre, la vue se resserre sur son entrée :
 * sur quatre cents tables, le fichier complet ne se parcourt pas. Le fichier
 * reste à un clic, et cliquer une autre table revient à son entrée.
 */
export function CalquePreview({ session, base, table }: ProprietesCalque) {
  const t = useT();
  const { extractions } = useExtractions();
  const { calque, message, enCours } = useCalque(session, derniereFin(extractions, base));

  const schema = table?.schema;
  const nom = table?.nom;
  const cle = schema !== undefined && nom !== undefined ? qualifier(schema, nom) : null;

  // Le fichier complet se demande table par table : un autre clic dans l'arbre
  // ramène à l'entrée de la nouvelle table plutôt qu'au fichier.
  const [completPour, setCompletPour] = useState<string | null>(null);
  const complet = cle === null || completPour === cle;

  const fragment = useMemo(
    () =>
      calque && schema !== undefined && nom !== undefined
        ? fragmentDeTable(calque.contenu, schema, nom)
        : null,
    [calque, schema, nom],
  );
  const affiche = complet ? (calque?.contenu ?? null) : fragment;

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex items-center gap-2 px-3 py-1">
        <h3 className="text-xs font-medium text-slate-600 dark:text-slate-400">
          {t('calque.title')}
        </h3>
        <span className="font-mono text-xs text-slate-500">
          {calque?.fichier ?? fichierCalque(base)}
        </span>
        {calque?.extrait_le ? (
          <span className="text-xs text-slate-500">
            {t('calque.extractedAt', { heure: heure(calque.extrait_le) })}
          </span>
        ) : null}

        {calque && cle !== null ? (
          <div role="group" aria-label={t('calque.view')} className="flex items-center gap-0.5">
            <button
              type="button"
              aria-pressed={!complet}
              onClick={() => setCompletPour(null)}
              className={bascule(!complet)}
            >
              <span className="font-mono">{cle}</span>
            </button>
            <button
              type="button"
              aria-pressed={complet}
              onClick={() => setCompletPour(cle)}
              className={bascule(complet)}
            >
              {t('calque.showFile')}
            </button>
          </div>
        ) : null}

        <div className="ml-auto">
          {affiche !== null ? <CopyButton valeur={affiche} libelle={t('calque.copy')} /> : null}
        </div>
      </div>

      {calque?.statistiques_retirees ? (
        <p className="px-3 text-xs text-amber-700 dark:text-amber-500">{t('calque.statsHidden')}</p>
      ) : null}

      {!calque ? (
        <p className="px-3 text-xs text-slate-500">{enCours ? t('calque.loading') : message}</p>
      ) : affiche !== null ? (
        <pre className="min-h-0 flex-1 overflow-auto px-3 pb-2 font-mono text-xs">{affiche}</pre>
      ) : (
        <p className="px-3 text-xs text-slate-500">{t('calque.tableMissing', { table: cle ?? '' })}</p>
      )}
    </div>
  );
}

/** Classes d'un bouton de bascule, selon qu'il est actif. */
function bascule(actif: boolean): string {
  return `rounded px-1.5 py-0.5 text-xs ${
    actif
      ? 'bg-slate-200 text-slate-900 dark:bg-slate-700 dark:text-slate-100'
      : 'text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
  }`;
}
