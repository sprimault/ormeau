// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useMemo, type ReactNode } from 'react';

import { useT } from '@/shared/i18n';
import type { Portee } from '@/shared/model';
import { CopyButton } from '@/shared/ui';
import type { EtatExclusions } from '../model/useExclusions';

/** Propriétés de l'aperçu. */
interface ProprietesApercu {
  portee: Portee;
  exclusions: EtatExclusions;
  /** Replié, seule la barre reste. */
  ouvert: boolean;
  onBasculer: () => void;
  /** Placé dans la barre, visible même replié. */
  actions?: ReactNode;
  /** Ce que l'extraction a produit, en colonne de droite. */
  produit?: ReactNode;
}

/**
 * Ce que produira l'écran, et ce qu'il a produit.
 *
 * À gauche, ce qui part, en deux morceaux qui ne vont pas au même endroit : la
 * portée décide de ce que la base est interrogée, les colonnes écartées vont
 * dans le fichier de décisions, qui se rejoue hors ligne. À droite, le calque
 * que l'extraction a écrit. Les voir côte à côte rend évident ce qui a été
 * demandé et ce qui est arrivé, sans l'expliquer.
 *
 * La hauteur et le repli sont tenus par l'écran, qui dimensionne la zone : ici,
 * le contenu occupe ce qu'on lui donne.
 *
 * Une liste de tables vide n'est pas une portée vide : c'est « toutes celles des
 * schémas retenus », comme en ligne de commande. Le dire, parce que le JSON seul
 * laisserait croire l'inverse.
 *
 * Le bouton d'extraction et le calque produit arrivent par `actions` et
 * `produit` : ils appartiennent à une autre feature, que celle-ci ne connaît pas.
 */
export function ScopePreview({
  portee,
  exclusions,
  ouvert,
  onBasculer,
  actions,
  produit,
}: ProprietesApercu) {
  const t = useT();

  const tables = portee.tables_incluses ?? [];
  const schemas = portee.schemas ?? [];
  const json = useMemo(() => JSON.stringify(portee, null, 2), [portee]);
  const decisions = useMemo(() => enYaml(exclusions.ignorees), [exclusions.ignorees]);

  return (
    <section className="flex h-full min-h-0 flex-col border-t border-slate-200 bg-slate-100 dark:border-slate-800 dark:bg-slate-900">
      <div className="flex shrink-0 items-center gap-3 px-3 py-1.5">
        <button
          type="button"
          onClick={onBasculer}
          aria-expanded={ouvert}
          className="text-xs font-semibold"
        >
          {ouvert ? '▾' : '▸'} {t('scope.title')}
        </button>
        <span className="text-xs text-slate-500">
          {tables.length === 0
            ? t('scope.all')
            : t('scope.count', { n: tables.length, schemas: schemas.length })}
        </span>
        {/* Juste après le compte et non repoussé au bout de la barre : c'est là
            que l'œil vient de lire ce qui partira. */}
        {actions}
        {exclusions.total > 0 ? (
          <span className="text-xs text-amber-700 dark:text-amber-500">
            {t('columns.excluded', { n: exclusions.total })}
          </span>
        ) : null}
      </div>

      {ouvert ? (
        <div
          className={`grid min-h-0 flex-1 gap-px border-t border-slate-200 bg-slate-200 dark:border-slate-800 dark:bg-slate-800 ${
            produit ? 'grid-cols-2' : 'grid-cols-1'
          }`}
        >
          <div className="flex min-h-0 flex-col gap-px">
            <div className="flex min-h-0 flex-1 flex-col bg-slate-100 dark:bg-slate-900">
              <div className="flex items-center gap-2 px-3 py-1">
                <h3 className="text-xs font-medium text-slate-600 dark:text-slate-400">
                  {t('scope.extraction')}
                </h3>
                <div className="ml-auto">
                  <CopyButton valeur={json} libelle={t('scope.copy')} />
                </div>
              </div>
              <pre className="min-h-0 flex-1 overflow-auto px-3 pb-2 font-mono text-xs">{json}</pre>
            </div>

            <div className="flex max-h-[40%] min-h-0 flex-col bg-slate-100 dark:bg-slate-900">
              <div className="flex items-center gap-2 px-3 py-1">
                <h3 className="text-xs font-medium text-slate-600 dark:text-slate-400">
                  {t('scope.decisions')}
                </h3>
                <div className="ml-auto">
                  <CopyButton valeur={decisions} libelle={t('scope.copyDecisions')} />
                </div>
              </div>
              <pre className="min-h-0 overflow-auto px-3 pb-2 font-mono text-xs">{decisions}</pre>
            </div>
          </div>

          {produit ? (
            <div className="min-h-0 overflow-hidden bg-slate-100 dark:bg-slate-900">{produit}</div>
          ) : null}
        </div>
      ) : null}
    </section>
  );
}

/**
 * Rend les colonnes écartées telles qu'elles s'écriront dans le fichier.
 *
 * Écrit à la main plutôt qu'avec une bibliothèque YAML : c'est deux niveaux et
 * des listes en ligne, et embarquer un sérialiseur pour cela alourdirait un
 * bundle qui finit dans le binaire.
 */
function enYaml(ignorees: Record<string, string[]>): string {
  const tables = Object.keys(ignorees);
  if (tables.length === 0) {
    return 'colonnes_ignorees: {}';
  }
  const lignes = tables.map((table) => `  ${table}: [${ignorees[table].join(', ')}]`);
  return ['colonnes_ignorees:', ...lignes].join('\n');
}
