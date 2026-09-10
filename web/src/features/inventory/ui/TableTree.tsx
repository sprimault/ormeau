// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useMemo, useState } from 'react';
import { ChevronDown, ChevronRight } from 'lucide-react';

import { useT } from '@/shared/i18n';
import { compact, qualifier } from '@/shared/lib';
import type { TableSommaire } from '@/shared/model';
import { Button, ErrorBanner, Field } from '@/shared/ui';
import type { Colonnes } from '../model/useColumns';
import type { EtatExclusions } from '../model/useExclusions';
import type { EtatSelection } from '../model/useSelection';
import { ColumnList } from './ColumnList';

/** Propriétés de l'arbre. */
interface ProprietesArbre {
  tables: TableSommaire[];
  enCours: boolean;
  erreur: string | null;
  etat: EtatSelection;
  colonnes: Colonnes;
  exclusions: EtatExclusions;
  active: string | null;
  onActiver: (cle: string) => void;
}

/** Un schéma et ses tables, après filtrage. */
interface Groupe {
  schema: string;
  tables: TableSommaire[];
  retenues: number;
}

/**
 * Arbre de sélection des tables.
 *
 * Densité d'information élevée, zéro décoration : on l'ouvre une fois par base
 * reprise, on coche, on repart. Le filtrage est local — l'inventaire complet est
 * déjà là, et un aller-retour par frappe rendrait la recherche inutilisable sur
 * une base distante.
 *
 * Deux gestes distincts sur une même ligne : la case retient la table pour
 * l'extraction, le nom l'affiche à droite. Les confondre obligerait à cocher
 * une table pour la consulter.
 */
export function TableTree({
  tables,
  enCours,
  erreur,
  etat,
  colonnes,
  exclusions,
  active,
  onActiver,
}: ProprietesArbre) {
  const t = useT();
  const [terme, setTerme] = useState('');
  const [replies, setReplies] = useState<Set<string>>(new Set());
  const [depliees, setDepliees] = useState<Set<string>>(new Set());
  const { selection, manquantes, basculer, basculerSchema, ajouterManquantes, toutEffacer } = etat;

  const filtrees = useMemo(() => {
    const recherche = terme.trim().toLowerCase();
    if (!recherche) {
      return tables;
    }
    return tables.filter(
      (table) =>
        table.nom.toLowerCase().includes(recherche) ||
        table.schema.toLowerCase().includes(recherche) ||
        (table.commentaire ?? '').toLowerCase().includes(recherche),
    );
  }, [tables, terme]);

  const groupes = useMemo(() => regrouper(filtrees, selection), [filtrees, selection]);

  if (erreur) {
    return (
      <div className="p-3">
        <ErrorBanner message={erreur} />
      </div>
    );
  }
  if (enCours) {
    return <p className="p-3 text-sm text-slate-500">{t('inventory.loading')}</p>;
  }
  if (tables.length === 0) {
    return <p className="p-3 text-sm text-slate-500">{t('inventory.empty')}</p>;
  }

  return (
    <div className="flex h-full flex-col gap-2 p-3">
      {/* Pas de titre ici : la base au-dessus dit déjà de quoi il s'agit. Le
          compte, lui, ne se déduit de rien. */}
      <span className="text-xs text-slate-500">
        {t('inventory.selected', { n: selection.size, total: tables.length })}
      </span>

      <Field
        label={t('inventory.search')}
        value={terme}
        spellCheck={false}
        onChange={(evenement) => setTerme(evenement.target.value)}
      />

      {manquantes.length > 0 ? (
        <div className="flex flex-col gap-2 rounded border border-amber-300 bg-amber-50 px-2 py-2 dark:border-amber-900 dark:bg-amber-950">
          <p className="text-xs text-amber-900 dark:text-amber-200">
            {t('inventory.missing', { n: manquantes.length })}
          </p>
          <Button variante="discret" onClick={ajouterManquantes}>
            {t('inventory.addMissing')}
          </Button>
        </div>
      ) : null}

      <div className="min-h-0 flex-1 overflow-y-auto">
        {groupes.length === 0 ? (
          <p className="text-sm text-slate-500">
            {t('inventory.noMatch', { terme: terme.trim() })}
          </p>
        ) : (
          <ul className="flex flex-col gap-2">
            {groupes.map((groupe) => (
              <li key={groupe.schema}>
                <div className="flex items-center gap-2 border-b border-slate-200 pb-1 dark:border-slate-800">
                  <button
                    type="button"
                    onClick={() => setReplies((precedent) => basculerCle(precedent, groupe.schema))}
                    aria-expanded={!replies.has(groupe.schema)}
                    className="font-mono text-xs font-semibold"
                  >
                    {replies.has(groupe.schema) ? '▸' : '▾'} {groupe.schema}
                  </button>
                  {/* Le compte reste lisible schéma replié : c'est ce qui permet
                      de savoir où on en est sans tout déplier. */}
                  <span className="text-xs text-slate-500">
                    {groupe.retenues} / {t('inventory.tables', { n: groupe.tables.length })}
                  </span>
                  <button
                    type="button"
                    onClick={() => basculerSchema(groupe.schema, groupe.tables)}
                    className="text-ormeau-600 dark:text-ormeau-500 ml-auto text-xs hover:underline"
                  >
                    {groupe.retenues === groupe.tables.length
                      ? t('inventory.clear')
                      : t('inventory.selectAll')}
                  </button>
                </div>

                {replies.has(groupe.schema) ? null : (
                  <ul className="mt-0.5 flex flex-col">
                    {groupe.tables.map((table) => {
                      const cle = qualifier(table.schema, table.nom);
                      const depliee = depliees.has(cle);
                      const ecartees = exclusions.compte(cle);
                      return (
                        <li key={cle}>
                          <div
                            className={`flex items-center gap-1 rounded px-1 py-0.5 text-sm ${
                              active === cle
                                ? 'bg-ormeau-50 dark:bg-slate-800'
                                : 'hover:bg-slate-100 dark:hover:bg-slate-900'
                            }`}
                          >
                            <button
                              type="button"
                              onClick={() => {
                                setDepliees((precedent) => basculerCle(precedent, cle));
                                colonnes.charger(cle, table.schema, table.nom);
                              }}
                              aria-expanded={depliee}
                              aria-label={t('columns.expand')}
                              className="shrink-0 text-slate-500"
                            >
                              {depliee ? (
                                <ChevronDown size={12} aria-hidden />
                              ) : (
                                <ChevronRight size={12} aria-hidden />
                              )}
                            </button>
                            <input
                              type="checkbox"
                              checked={selection.has(cle)}
                              aria-label={cle}
                              onChange={() => basculer(cle)}
                            />
                            <button
                              type="button"
                              onClick={() => onActiver(cle)}
                              className="flex min-w-0 flex-1 items-baseline gap-2 text-left"
                            >
                              <span className="truncate font-mono">{table.nom}</span>
                              <span className="ml-auto shrink-0 text-xs text-slate-500">
                                {ecartees > 0
                                  ? t('columns.excluded', { n: ecartees })
                                  : t('inventory.rows', { n: compact(table.lignes_estimees) })}
                              </span>
                            </button>
                            {table.cle_primaire ? null : (
                              <span
                                title={t('inventory.noPrimaryKey')}
                                className="shrink-0 text-xs text-amber-700 dark:text-amber-500"
                              >
                                ⚠
                              </span>
                            )}
                          </div>

                          {depliee ? (
                            <ColumnList
                              cle={cle}
                              etat={colonnes.etat(cle)}
                              exclusions={exclusions}
                            />
                          ) : null}
                        </li>
                      );
                    })}
                  </ul>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>

      {selection.size > 0 ? (
        <Button variante="discret" onClick={toutEffacer}>
          {t('inventory.clear')}
        </Button>
      ) : null}
    </div>
  );
}

/** Regroupe par schéma, dans l'ordre où le serveur les a triés. */
function regrouper(tables: TableSommaire[], selection: Set<string>): Groupe[] {
  const parSchema = new Map<string, Groupe>();
  for (const table of tables) {
    let groupe = parSchema.get(table.schema);
    if (!groupe) {
      groupe = { schema: table.schema, tables: [], retenues: 0 };
      parSchema.set(table.schema, groupe);
    }
    groupe.tables.push(table);
    if (selection.has(qualifier(table.schema, table.nom))) {
      groupe.retenues += 1;
    }
  }
  return [...parSchema.values()];
}

/** Ajoute ou retire une clé d'un ensemble, sans le modifier sur place. */
function basculerCle(ensemble: Set<string>, cle: string): Set<string> {
  const suivant = new Set(ensemble);
  if (!suivant.delete(cle)) {
    suivant.add(cle);
  }
  return suivant;
}
