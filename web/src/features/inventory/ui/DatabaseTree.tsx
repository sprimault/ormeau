// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';
import { ChevronDown, ChevronRight, Database } from 'lucide-react';

import { useT } from '@/shared/i18n';
import type { TableSommaire } from '@/shared/model';
import type { Colonnes } from '../model/useColumns';
import type { EtatExclusions } from '../model/useExclusions';
import type { EtatSelection } from '../model/useSelection';
import { TableTree } from './TableTree';

/** Propriétés de l'arbre. */
interface ProprietesArbre {
  bases: string[];
  courante: string;
  tables: TableSommaire[];
  enCours: boolean;
  erreur: string | null;
  etat: EtatSelection;
  colonnes: Colonnes;
  exclusions: EtatExclusions;
  active: string | null;
  onActiver: (cle: string) => void;
  onOuvrirBase: (base: string) => void;
}

/**
 * Arbre des bases du serveur.
 *
 * Trois niveaux, comme dans un client de base : serveur, base, schéma, tables.
 * Une seule base est ouverte à la fois, et c'est une contrainte du format, pas
 * un choix d'affichage — un calque décrit une base et une seule, donc une
 * sélection ne s'étend pas au-delà.
 *
 * Ouvrir une autre base rouvre la connexion sur elle, avec les mêmes
 * identifiants. La sélection en cours ne la suit pas : elle désignait des tables
 * d'ailleurs.
 */
export function DatabaseTree({
  bases,
  courante,
  tables,
  enCours,
  erreur,
  etat,
  colonnes,
  exclusions,
  active,
  onActiver,
  onOuvrirBase,
}: ProprietesArbre) {
  const t = useT();

  // Repli de la base ouverte, distinct de la connexion : la replier ne doit pas
  // fermer la session ni perdre la sélection, seulement rendre la place aux
  // autres bases quand le serveur en porte vingt.
  const [repliee, setRepliee] = useState(false);

  // La base ouverte peut ne pas figurer dans l'énumération — les bases système
  // en sont écartées, et l'on y atterrit pourtant quand le DSN ne nomme rien.
  const listees = bases.includes(courante) ? bases : [courante, ...bases];

  return (
    <div className="flex h-full min-h-0 flex-col">
      <ul className="min-h-0 flex-1 overflow-y-auto py-1">
        {listees.map((base) => {
          const courant = base === courante;
          const ouverte = courant && !repliee;
          return (
            <li key={base}>
              <button
                type="button"
                onClick={() => {
                  if (courant) {
                    setRepliee((precedent) => !precedent);
                  } else {
                    setRepliee(false);
                    onOuvrirBase(base);
                  }
                }}
                aria-expanded={ouverte}
                disabled={enCours && !courant}
                className={`flex w-full items-center gap-1 px-2 py-1 text-left text-sm hover:bg-slate-100 disabled:opacity-50 dark:hover:bg-slate-900 ${
                  courant ? 'font-semibold' : ''
                }`}
              >
                {ouverte ? (
                  <ChevronDown size={14} aria-hidden />
                ) : (
                  <ChevronRight size={14} aria-hidden />
                )}
                <Database size={14} aria-hidden className="text-ormeau-600 dark:text-ormeau-500" />
                <span className="truncate font-mono">{base}</span>
              </button>

              {ouverte ? (
                <div className="border-l border-slate-200 pl-2 dark:border-slate-800">
                  <TableTree
                    tables={tables}
                    enCours={enCours}
                    erreur={erreur}
                    etat={etat}
                    colonnes={colonnes}
                    exclusions={exclusions}
                    active={active}
                    onActiver={onActiver}
                  />
                </div>
              ) : null}
            </li>
          );
        })}
      </ul>

      {listees.length === 1 ? (
        <p className="border-t border-slate-200 px-2 py-1 text-xs text-slate-500 dark:border-slate-800">
          {t('inventory.singleDatabase')}
        </p>
      ) : null}
    </div>
  );
}
