// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useId } from 'react';

import { useT } from '@/shared/i18n';
import { qualifier } from '@/shared/lib';
import type { Avertissement, Decisions, Proposition } from '@/shared/model';
import { Button, ErrorBanner, HelpTip, TextInput } from '@/shared/ui';
import { forcerType, renommer, type Modifier } from '../model/decisions';
import type { LigneEntite } from '../model/entites';
import type { EtatEntite } from '../model/useEntite';
import { EntityDetail } from './EntityDetail';
import { WarningList } from './WarningList';

/** Propriétés du panneau. */
interface ProprietesPanneau {
  ligne: LigneEntite;
  etat: EtatEntite;
  decisions: Decisions;
  avertissements: Avertissement[];
  proposition?: Proposition;
  typesDoctrine: string[];
  onModifier: Modifier;
}

/**
 * L'entité ouverte : son nom de classe, ses avertissements, puis l'entité à
 * côté de sa table.
 *
 * Les tables et les colonnes ne se décident pas ici : c'est l'onglet de
 * sélection qui les choisit. On y règle ce que la sélection ne sait pas dire —
 * le nom de la classe, et le type d'une colonne que l'inférence ne reconnaît
 * pas.
 */
export function EntityPanel({
  ligne,
  etat,
  decisions,
  avertissements,
  proposition,
  typesDoctrine,
  onModifier,
}: ProprietesPanneau) {
  const t = useT();
  const idNom = useId();
  const { qualifiee } = ligne;
  const forces = Object.entries(decisions.types_forces ?? {}).filter(([cible]) =>
    cible.startsWith(`${qualifiee}.`),
  );

  // Le détail d'une autre entité reste en mémoire le temps du calcul : il ne
  // s'affiche pas sous le nom de celle qu'on vient d'ouvrir.
  const detail =
    etat.detail &&
    qualifier(etat.detail.table_physique.schema, etat.detail.table_physique.nom) === qualifiee
      ? etat.detail
      : null;
  const typesRetenus: Record<string, string> = {};
  for (const propriete of detail?.entite?.proprietes ?? []) {
    typesRetenus[propriete.colonne] = propriete.type_doctrine;
  }

  return (
    <div className="flex flex-col gap-3 p-4 text-xs">
      <div className="flex flex-wrap items-center gap-2">
        <label htmlFor={idNom} className="text-slate-600 dark:text-slate-400">
          {t('arbitrage.className')}
        </label>
        <HelpTip texte={t('arbitrage.help.className')} />
        <TextInput
          id={idNom}
          value={decisions.renommages?.[qualifiee] ?? ''}
          placeholder={ligne.nom}
          onChange={(evenement) => {
            const nom = evenement.target.value;
            onModifier((d) => renommer(d, qualifiee, nom));
          }}
          className="w-64 font-mono text-sm"
        />
        {proposition ? (
          <span className="flex flex-wrap items-center gap-2 text-slate-500">
            {t('arbitrage.proposal', {
              nom: proposition.nom,
              pct: Math.round(proposition.confiance * 100),
            })}{' '}
            — {proposition.raison}
            <Button
              variante="discret"
              taille="petite"
              onClick={() => onModifier((d) => renommer(d, qualifiee, proposition.nom))}
            >
              {t('arbitrage.useProposal')}
            </Button>
          </span>
        ) : null}
      </div>

      <WarningList
        qualifiee={qualifiee}
        avertissements={avertissements}
        typesDoctrine={typesDoctrine}
        typesRetenus={typesRetenus}
        onForcer={(colonne, type) => onModifier((d) => forcerType(d, colonne, type))}
      />

      {forces.length > 0 ? (
        <section className="flex flex-col gap-1">
          <h3 className="flex items-center gap-1 font-semibold">
            {t('arbitrage.forced')}
            <HelpTip texte={t('arbitrage.help.forced')} />
          </h3>
          <ul className="flex flex-col gap-1">
            {forces.map(([cible, type]) => (
              <li key={cible} className="flex items-center gap-2">
                <span className="font-mono">
                  {cible.slice(qualifiee.length + 1)} → {type}
                </span>
                <Button
                  variante="discret"
                  taille="petite"
                  onClick={() => onModifier((d) => forcerType(d, cible, ''))}
                >
                  {t('arbitrage.forced.undo')}
                </Button>
              </li>
            ))}
          </ul>
        </section>
      ) : null}

      {etat.erreur ? (
        <ErrorBanner message={etat.erreur} />
      ) : detail ? (
        <EntityDetail detail={detail} ecartees={decisions.colonnes_ignorees?.[qualifiee] ?? []} />
      ) : (
        <p className="text-sm text-slate-500">{t('arbitrage.detail.loading')}</p>
      )}
    </div>
  );
}
