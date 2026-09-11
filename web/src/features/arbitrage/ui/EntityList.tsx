// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';

import { useT } from '@/shared/i18n';
import { HelpTip, TextInput } from '@/shared/ui';
import type { LigneEntite } from '../model/entites';
import { PendingBadge } from './PendingBadge';

/** Propriétés de la liste. */
interface ProprietesListe {
  lignes: LigneEntite[];
  /** Nom qualifié de la table de l'entité ouverte. */
  active: string;
  onActiver: (qualifiee: string) => void;
}

/**
 * Les entités que la génération produira, chacune avec sa table et le nombre
 * d'avertissements à traiter. La ligne entière ouvre l'entité.
 */
export function EntityList({ lignes, active, onActiver }: ProprietesListe) {
  const t = useT();
  const [terme, setTerme] = useState('');

  const recherche = terme.trim().toLowerCase();
  const visibles = recherche
    ? lignes.filter(
        (ligne) =>
          ligne.nom.toLowerCase().includes(recherche) ||
          ligne.qualifiee.toLowerCase().includes(recherche),
      )
    : lignes;

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="flex flex-col gap-1.5 p-2">
        <h3 className="flex items-center gap-1 px-1 text-xs font-semibold">
          {t('arbitrage.entities', { n: lignes.length })}
          <HelpTip texte={t('arbitrage.help.entities')} />
        </h3>
        <TextInput
          type="search"
          value={terme}
          onChange={(evenement) => setTerme(evenement.target.value)}
          aria-label={t('arbitrage.entities.filter')}
          placeholder={t('arbitrage.entities.filter')}
          className="w-full"
        />
      </div>

      {lignes.length === 0 ? (
        <p className="px-3 text-xs text-slate-500">{t('arbitrage.entities.none')}</p>
      ) : (
        <ul className="min-h-0 flex-1 overflow-y-auto">
          {visibles.map((ligne) => {
            const estActive = ligne.qualifiee === active;
            return (
              <li key={ligne.qualifiee}>
                <button
                  type="button"
                  onClick={() => onActiver(ligne.qualifiee)}
                  aria-current={estActive ? 'true' : undefined}
                  className={`flex w-full items-center gap-2 px-3 py-1.5 text-left text-xs hover:bg-slate-100 dark:hover:bg-slate-800 ${
                    estActive ? 'bg-slate-200 dark:bg-slate-800' : ''
                  }`}
                >
                  <span className="min-w-0 flex-1">
                    <span className="block truncate font-mono font-medium">{ligne.nom}</span>
                    <span className="block truncate font-mono text-slate-500">{ligne.qualifiee}</span>
                  </span>
                  <PendingBadge n={ligne.aTraiter} />
                </button>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
