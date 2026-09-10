// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useEffect, type ReactNode } from 'react';
import { KeyRound } from 'lucide-react';

import { useT } from '@/shared/i18n';
import { qualifier } from '@/shared/lib';
import type { TableSommaire } from '@/shared/model';
import type { Colonnes } from '../model/useColumns';
import type { EtatExclusions } from '../model/useExclusions';

/** Propriétés du panneau. */
interface ProprietesDetails {
  table: TableSommaire | null;
  selection: Set<string>;
  colonnes: Colonnes;
  exclusions: EtatExclusions;
}

/**
 * Panneau de détail de la table active.
 *
 * Trois niveaux de lecture, du plus rapide au plus précis : les compteurs se
 * lisent d'un coup d'œil, la liste des colonnes demande un moment, les
 * références se consultent quand on arbitre une sélection. Tout enfiler dans une
 * seule liste obligerait à chercher le chiffre au milieu du reste.
 *
 * Il ne montre que ce que le catalogue a rendu : structure et volumétrie
 * estimée, jamais une ligne de données. L'interface n'est pas un explorateur.
 *
 * Les colonnes sont chargées à l'activation, par le même cache que l'arbre :
 * ouvrir une table à gauche puis la sélectionner ne redemande rien au serveur.
 */
export function TableDetails({ table, selection, colonnes, exclusions }: ProprietesDetails) {
  const t = useT();
  const cle = table ? qualifier(table.schema, table.nom) : '';

  useEffect(() => {
    if (table) {
      colonnes.charger(cle, table.schema, table.nom);
    }
  }, [cle, table, colonnes]);

  if (!table) {
    return <p className="p-6 text-sm text-slate-500 dark:text-slate-400">{t('details.none')}</p>;
  }

  const references = table.reference_vers ?? [];
  const etat = colonnes.etat(cle);
  const ecartees = exclusions.compte(cle);

  return (
    <article className="flex flex-col gap-5 p-6">
      <header>
        <h2 className="font-mono text-base font-semibold">{table.nom}</h2>
        <p className="font-mono text-xs text-slate-500">{table.schema}</p>
        {table.commentaire ? (
          <p className="mt-2 text-sm text-slate-600 dark:text-slate-400">{table.commentaire}</p>
        ) : null}
      </header>

      <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
        <Compteur libelle={t('details.columns')}>
          {table.nb_colonnes}
          {ecartees > 0 ? (
            <span className="ml-1 text-xs font-normal text-amber-700 dark:text-amber-500">
              −{ecartees}
            </span>
          ) : null}
        </Compteur>

        {/* Une estimation du catalogue, pas un COUNT : le libellé le dit, parce
            qu'un chiffre affiché seul passe pour exact. */}
        <Compteur libelle={t('details.rows')}>
          {table.lignes_estimees.toLocaleString('fr-FR')}
        </Compteur>

        <Compteur libelle={t('details.primaryKey')} alerte={!table.cle_primaire}>
          {table.cle_primaire ? t('details.yes') : t('details.no')}
        </Compteur>

        <Compteur libelle={t('details.references')}>{references.length}</Compteur>
      </div>

      {table.cle_primaire ? null : (
        <p className="rounded border border-amber-300 bg-amber-50 px-3 py-2 text-sm text-amber-900 dark:border-amber-900 dark:bg-amber-950 dark:text-amber-200">
          {t('details.noPrimaryKeyWarning')}
        </p>
      )}

      <section>
        <h3 className="mb-1 text-xs font-medium text-slate-600 uppercase dark:text-slate-400">
          {t('details.columns')}
        </h3>

        {etat.erreur ? (
          <p className="text-sm text-red-700 dark:text-red-400">{etat.erreur}</p>
        ) : !etat.colonnes ? (
          <p className="text-sm text-slate-500">{t('columns.loading')}</p>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-200 text-left text-xs text-slate-500 dark:border-slate-800">
                <th className="w-8 py-1 font-medium">#</th>
                <th className="py-1 font-medium">{t('details.column')}</th>
                <th className="py-1 font-medium">{t('details.type')}</th>
                <th className="w-20 py-1 font-medium">{t('details.nullable')}</th>
              </tr>
            </thead>
            <tbody>
              {etat.colonnes.map((colonne) => {
                const ignoree = exclusions.estIgnoree(cle, colonne.nom);
                return (
                  <tr
                    key={colonne.nom}
                    className="border-b border-slate-100 last:border-0 dark:border-slate-900"
                  >
                    <td className="py-1 font-mono text-xs text-slate-400">{colonne.position}</td>
                    <td className="py-1">
                      <span className={`font-mono ${ignoree ? 'text-slate-400 line-through' : ''}`}>
                        {colonne.nom}
                      </span>
                      {colonne.cle_primaire ? (
                        <KeyRound
                          size={11}
                          aria-label={t('columns.primaryKey')}
                          className="ml-1 inline text-amber-600"
                        />
                      ) : null}
                      {ignoree ? (
                        <span className="ml-2 text-xs text-amber-700 dark:text-amber-500">
                          {t('details.excluded')}
                        </span>
                      ) : null}
                      {colonne.commentaire ? (
                        <span className="ml-2 text-xs text-slate-400">{colonne.commentaire}</span>
                      ) : null}
                    </td>
                    {/* Type verbatim : c'est ce qui distingue un citext d'un
                        text, et c'est ce que l'utilisateur retrouvera en base. */}
                    <td className="py-1 font-mono text-slate-600 dark:text-slate-400">
                      {colonne.type_brut}
                    </td>
                    <td className="py-1 text-slate-500">
                      {colonne.nullable ? t('details.yes') : ''}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </section>

      <section>
        <h3 className="text-xs font-medium text-slate-600 uppercase dark:text-slate-400">
          {t('details.references')}
        </h3>
        {references.length === 0 ? (
          <p className="mt-1 text-sm text-slate-500">{t('details.noReference')}</p>
        ) : (
          <ul className="mt-1 flex flex-col gap-0.5">
            {references.map((cible) => (
              <li key={cible} className="flex items-center gap-2 text-sm">
                <span className="font-mono">{cible}</span>
                {/* Une référence hors sélection est le défaut silencieux que cet
                    écran existe pour rendre visible. */}
                {selection.has(cle) && !selection.has(cible) ? (
                  <span className="text-xs text-amber-700 dark:text-amber-500">
                    {t('details.outOfSelection')}
                  </span>
                ) : null}
              </li>
            ))}
          </ul>
        )}
      </section>
    </article>
  );
}

/** Propriétés d'un compteur. */
interface ProprietesCompteur {
  libelle: string;
  alerte?: boolean;
  children: ReactNode;
}

/** Un chiffre et son libellé, lisibles d'un coup d'œil. */
function Compteur({ libelle, alerte = false, children }: ProprietesCompteur) {
  return (
    <div className="rounded border border-slate-200 px-3 py-2 dark:border-slate-800">
      <p className="text-xs text-slate-500">{libelle}</p>
      <p
        className={`font-mono text-lg font-semibold ${
          alerte ? 'text-amber-700 dark:text-amber-500' : ''
        }`}
      >
        {children}
      </p>
    </div>
  );
}
