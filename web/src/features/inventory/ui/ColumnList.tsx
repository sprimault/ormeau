// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { KeyRound } from 'lucide-react';

import { useT } from '@/shared/i18n';
import type { EtatColonnes } from '../model/useColumns';
import type { EtatExclusions } from '../model/useExclusions';

/** Propriétés de la liste. */
interface ProprietesColonnes {
  cle: string;
  etat: EtatColonnes;
  exclusions: EtatExclusions;
}

/**
 * Colonnes d'une table dépliée.
 *
 * Une case cochée veut dire « cette colonne devient une propriété ». La
 * décocher ne retire rien de l'extraction — le calque garde toute la table —,
 * elle inscrit la colonne dans `colonnes_ignorees`.
 *
 * Les colonnes de clé primaire ne se décochent pas : Doctrine refuse une entité
 * sans identifiant. Le montrer désactivé plutôt que de l'accepter et le défaire
 * plus loin évite de laisser croire à un arbitrage qui n'aura pas lieu.
 */
export function ColumnList({ cle, etat, exclusions }: ProprietesColonnes) {
  const t = useT();

  if (etat.erreur) {
    return <p className="px-6 py-1 text-xs text-red-700 dark:text-red-400">{etat.erreur}</p>;
  }
  if (etat.enCours && !etat.colonnes) {
    return <p className="px-6 py-1 text-xs text-slate-500">{t('columns.loading')}</p>;
  }
  if (!etat.colonnes) {
    return null;
  }

  return (
    <ul className="flex flex-col border-l border-slate-200 py-0.5 pl-6 dark:border-slate-800">
      {etat.colonnes.map((colonne) => {
        const ignoree = exclusions.estIgnoree(cle, colonne.nom);
        return (
          <li key={colonne.nom} className="flex items-center gap-2 py-0.5 text-xs">
            <input
              type="checkbox"
              checked={!ignoree}
              disabled={colonne.cle_primaire}
              aria-label={colonne.nom}
              title={colonne.cle_primaire ? t('columns.primaryKeyKept') : undefined}
              onChange={() => exclusions.basculer(cle, colonne.nom)}
            />
            <span className={`font-mono ${ignoree ? 'text-slate-400 line-through' : ''}`}>
              {colonne.nom}
            </span>
            {colonne.cle_primaire ? (
              <KeyRound
                size={11}
                aria-label={t('columns.primaryKey')}
                className="shrink-0 text-amber-600"
              />
            ) : null}
            {/* Type verbatim : c'est ce qui distingue un citext d'un text, et
                c'est ce que l'utilisateur retrouvera dans son SGBD. */}
            <span className="ml-auto shrink-0 font-mono text-slate-500">{colonne.type_brut}</span>
            {colonne.nullable ? <span className="shrink-0 text-slate-400">?</span> : null}
          </li>
        );
      })}
    </ul>
  );
}
