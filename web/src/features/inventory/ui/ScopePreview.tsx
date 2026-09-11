// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useMemo, useState, type ReactNode } from 'react';

import { useT } from '@/shared/i18n';
import { CopyButton } from '@/shared/ui';
import type { EtatExclusions } from '../model/useExclusions';
import { usePortee } from '../model/usePortee';

/** Propriétés de l'aperçu. */
interface ProprietesApercu {
  schemas: string[];
  selection: Set<string>;
  exclusions: EtatExclusions;
  /** Placé au bout de la barre, visible sans rien déplier. */
  actions?: ReactNode;
}

/**
 * Ce que produira l'écran, en deux morceaux qui ne vont pas au même endroit.
 *
 * La portée part à l'extraction et décide de ce que la base est interrogée. Les
 * colonnes écartées, elles, ne la touchent pas : elles vont dans le fichier de
 * décisions, qui se rejoue hors ligne. Les montrer côte à côte est le seul moyen
 * de rendre cette séparation évidente sans l'expliquer.
 *
 * Le type de la portée vient du Go : ce qui s'affiche est exactement ce qui
 * partira, pas une reconstitution qui pourrait s'en écarter.
 *
 * Une liste de tables vide n'est pas une portée vide : c'est « toutes celles des
 * schémas retenus », comme en ligne de commande. Le dire, parce que le JSON seul
 * laisserait croire l'inverse.
 *
 * Le bouton qui lance l'extraction arrive par `actions` : il appartient à une
 * autre feature, que celle-ci ne connaît pas.
 */
export function ScopePreview({ schemas, selection, exclusions, actions }: ProprietesApercu) {
  const t = useT();
  const [ouvert, setOuvert] = useState(false);

  const portee = usePortee(schemas, selection);
  const json = useMemo(() => JSON.stringify(portee, null, 2), [portee]);
  const decisions = useMemo(() => enYaml(exclusions.ignorees), [exclusions.ignorees]);

  return (
    <section className="border-t border-slate-200 bg-slate-100 dark:border-slate-800 dark:bg-slate-900">
      <div className="flex items-center gap-3 px-3 py-1.5">
        <button
          type="button"
          onClick={() => setOuvert((precedent) => !precedent)}
          aria-expanded={ouvert}
          className="text-xs font-semibold"
        >
          {ouvert ? '▾' : '▸'} {t('scope.title')}
        </button>
        <span className="text-xs text-slate-500">
          {selection.size === 0
            ? t('scope.all')
            : t('scope.count', { n: selection.size, schemas: schemas.length })}
        </span>
        {exclusions.total > 0 ? (
          <span className="text-xs text-amber-700 dark:text-amber-500">
            {t('columns.excluded', { n: exclusions.total })}
          </span>
        ) : null}
        {actions ? <div className="ml-auto">{actions}</div> : null}
      </div>

      {ouvert ? (
        <div className="grid max-h-56 grid-cols-2 gap-px overflow-auto border-t border-slate-200 bg-slate-200 dark:border-slate-800 dark:bg-slate-800">
          <div className="bg-slate-100 dark:bg-slate-900">
            <div className="flex items-center gap-2 px-3 py-1">
              <h3 className="text-xs font-medium text-slate-600 dark:text-slate-400">
                {t('scope.extraction')}
              </h3>
              <div className="ml-auto">
                <CopyButton valeur={json} libelle={t('scope.copy')} />
              </div>
            </div>
            <pre className="px-3 pb-2 font-mono text-xs">{json}</pre>
          </div>

          <div className="bg-slate-100 dark:bg-slate-900">
            <div className="flex items-center gap-2 px-3 py-1">
              <h3 className="text-xs font-medium text-slate-600 dark:text-slate-400">
                {t('scope.decisions')}
              </h3>
              <div className="ml-auto">
                <CopyButton valeur={decisions} libelle={t('scope.copyDecisions')} />
              </div>
            </div>
            <pre className="px-3 pb-2 font-mono text-xs">{decisions}</pre>
          </div>
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
