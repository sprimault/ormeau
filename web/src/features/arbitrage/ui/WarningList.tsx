// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useId, useState } from 'react';

import { estCle, useT } from '@/shared/i18n';
import { CodeTypeNonReconnu, type Avertissement } from '@/shared/model';
import { Button, HelpTip, TextInput } from '@/shared/ui';
import { estATraiter } from '../model/avertissements';

/** Propriétés de la liste. */
interface ProprietesAvertissements {
  /** Nom qualifié de la table, retiré des cibles pour ne laisser que la colonne. */
  qualifiee: string;
  avertissements: Avertissement[];
  typesDoctrine: string[];
  /** Type Doctrine que l'inférence a retenu, par colonne. */
  typesRetenus: Record<string, string>;
  onForcer: (colonne: string, type: string) => void;
}

/**
 * Les avertissements d'une entité, au-dessus de son détail.
 *
 * Le libellé se traduit depuis le code ; le message, rédigé en français par
 * l'inférence, suit tel quel. Un code sans traduction s'affiche brut : il se
 * voit, il ne disparaît pas. Un type non reconnu se force sur place, parmi les
 * types Doctrine connus ou sous le nom d'un type propre au projet.
 */
export function WarningList({
  qualifiee,
  avertissements,
  typesDoctrine,
  typesRetenus,
  onForcer,
}: ProprietesAvertissements) {
  const t = useT();
  const idTypes = useId();

  if (avertissements.length === 0) {
    return null;
  }

  return (
    <section className="flex flex-col gap-2 rounded border border-slate-200 p-3 dark:border-slate-800">
      <h3 className="flex items-center gap-1 font-semibold">
        {t('arbitrage.warnings')}
        <HelpTip texte={t('arbitrage.help.warnings')} />
      </h3>
      <ul className="flex flex-col gap-2">
        {avertissements.map((avertissement, rang) => {
          const libelle = `warning.${avertissement.code}`;
          const colonne =
            avertissement.cible === qualifiee ? '' : avertissement.cible.slice(qualifiee.length + 1);
          return (
            <li key={`${avertissement.code}|${avertissement.cible}|${rang}`} className="flex flex-col gap-1">
              <p className={estATraiter(avertissement) ? 'text-amber-700 dark:text-amber-500' : 'text-slate-500'}>
                <span className="font-medium">{estCle(libelle) ? t(libelle) : avertissement.code}</span>
                {colonne ? <span className="ml-1 font-mono">{colonne}</span> : null}
                {' — '}
                {avertissement.message}
              </p>
              {avertissement.code === CodeTypeNonReconnu && colonne ? (
                <ForcerType
                  colonne={colonne}
                  defaut={typesRetenus[colonne] ?? ''}
                  idTypes={idTypes}
                  onForcer={(type) => onForcer(avertissement.cible, type)}
                />
              ) : null}
            </li>
          );
        })}
      </ul>
      <datalist id={idTypes}>
        {typesDoctrine.map((type) => (
          <option key={type} value={type} />
        ))}
      </datalist>
    </section>
  );
}

/** Propriétés de la saisie d'un type forcé. */
interface ProprietesForcerType {
  colonne: string;
  /** Le type que l'inférence a retenu faute de mieux. */
  defaut: string;
  idTypes: string;
  onForcer: (type: string) => void;
}

/**
 * Saisie du type Doctrine à imposer à une colonne que l'inférence ne reconnaît
 * pas.
 *
 * Tant que rien n'est saisi, le champ propose le type retenu : le plus souvent,
 * c'est le bon, et un clic suffit à le confirmer. Le défaut arrive avec le
 * détail de l'entité, parfois après le premier rendu : il n'est pas figé à
 * l'ouverture.
 */
function ForcerType({ colonne, defaut, idTypes, onForcer }: ProprietesForcerType) {
  const t = useT();
  const [saisi, setSaisi] = useState<string | null>(null);
  const type = (saisi ?? defaut).trim();

  return (
    <form
      className="flex items-center gap-1"
      onSubmit={(evenement) => {
        evenement.preventDefault();
        if (type !== '') {
          onForcer(type);
        }
      }}
    >
      <TextInput
        list={idTypes}
        value={saisi ?? defaut}
        onChange={(evenement) => setSaisi(evenement.target.value)}
        aria-label={t('arbitrage.forceType.label', { colonne })}
        placeholder={t('arbitrage.forceType.label', { colonne })}
        className="w-56 font-mono"
      />
      <Button type="submit" variante="discret" taille="petite" disabled={type === ''}>
        {t('arbitrage.forceType')}
      </Button>
      <HelpTip texte={t('arbitrage.help.forceType')} />
    </form>
  );
}
