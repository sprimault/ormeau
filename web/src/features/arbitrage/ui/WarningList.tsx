// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useId, useState } from 'react';

import { estCle, useT } from '@/shared/i18n';
import {
  CodeCasEnumerationOpaque,
  CodeTexteUnicodeJSONPropose,
  CodeTypeNonReconnu,
  type Avertissement,
  type EnumerationInferee,
} from '@/shared/model';
import { Button, HelpTip, TextInput } from '@/shared/ui';
import { estATraiter, lieu } from '../model/avertissements';

/** Propriétés de la liste. */
interface ProprietesAvertissements {
  /** Nom qualifié de la table, retiré des cibles pour ne laisser que la colonne. */
  qualifiee: string;
  avertissements: Avertissement[];
  typesDoctrine: string[];
  /** Type Doctrine que l'inférence a retenu, par colonne. */
  typesRetenus: Record<string, string>;
  enumerations: EnumerationInferee[];
  onForcer: (colonne: string, type: string) => void;
  onNommerCas: (enumeration: EnumerationInferee, cas: Record<string, string>) => void;
}

/**
 * Les avertissements d'une entité, au-dessus de son détail.
 *
 * Le libellé se traduit depuis le code ; le message, rédigé en français par
 * l'inférence, suit tel quel. Un code sans traduction s'affiche brut : il se
 * voit, il ne disparaît pas.
 *
 * Trois avertissements se traitent sur place : un type non reconnu se force,
 * champ prérempli ; les cas d'une énumération se nomment ; un texte Unicode
 * déclaré JSON se force en json d'un clic. L'inférence signale chaque cas
 * opaque à part ; le formulaire, lui, vient une fois par colonne, sous le
 * dernier de ses avertissements. Les autres disent en tête où ils se traitent.
 */
export function WarningList({
  qualifiee,
  avertissements,
  typesDoctrine,
  typesRetenus,
  enumerations,
  onForcer,
  onNommerCas,
}: ProprietesAvertissements) {
  const t = useT();
  const idTypes = useId();

  if (avertissements.length === 0) {
    return null;
  }

  const dernierParColonne = new Map<string, number>();
  avertissements.forEach((avertissement, rang) => {
    if (avertissement.code === CodeCasEnumerationOpaque) {
      dernierParColonne.set(avertissement.cible, rang);
    }
  });

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
          const enumeration =
            avertissement.code === CodeCasEnumerationOpaque &&
            dernierParColonne.get(avertissement.cible) === rang
              ? enumerations.find((e) => e.colonnes.includes(avertissement.cible))
              : undefined;

          const ou = lieu(avertissement);

          return (
            <li key={`${avertissement.code}|${avertissement.cible}|${rang}`} className="flex flex-col gap-1">
              <p className={estATraiter(avertissement) ? 'text-amber-700 dark:text-amber-500' : 'text-slate-500'}>
                {ou === 'ecran' ? null : <span className="font-semibold">{t(`arbitrage.lieu.${ou}` as const)} </span>}
                <span className="font-medium">{estCle(libelle) ? t(libelle) : avertissement.code}</span>
                {colonne ? (
                  <>
                    {' '}
                    <span className="font-mono">{colonne}</span>
                  </>
                ) : null}
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
              {avertissement.code === CodeTexteUnicodeJSONPropose && colonne && typesRetenus[colonne] !== 'json' ? (
                <div>
                  <Button variante="discret" taille="petite" onClick={() => onForcer(avertissement.cible, 'json')}>
                    {t('arbitrage.forceJson')}
                  </Button>
                </div>
              ) : null}
              {enumeration ? (
                <NommerCas enumeration={enumeration} onNommer={(cas) => onNommerCas(enumeration, cas)} />
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

/** Propriétés du nommage des cas. */
interface ProprietesNommerCas {
  enumeration: EnumerationInferee;
  onNommer: (cas: Record<string, string>) => void;
}

/**
 * Noms lisibles des cas d'une énumération, un champ par valeur stockée.
 *
 * Chaque champ part du nom que l'inférence a donné : on remplace ceux qui ne
 * disent rien, et un clic confirme l'ensemble. Un nom vide ne part pas, le
 * fichier n'a rien à en faire.
 */
function NommerCas({ enumeration, onNommer }: ProprietesNommerCas) {
  const t = useT();
  const [saisis, setSaisis] = useState<Record<string, string>>({});

  const cas = enumeration.cas.map((c) => {
    const valeur = String(c.valeur);
    return { valeur, nom: saisis[valeur] ?? c.nom };
  });
  const complet = cas.every((c) => c.nom.trim() !== '');

  return (
    <form
      className="flex flex-col gap-1"
      onSubmit={(evenement) => {
        evenement.preventDefault();
        if (complet) {
          onNommer(Object.fromEntries(cas.map((c) => [c.valeur, c.nom.trim()])));
        }
      }}
    >
      <table className="w-auto">
        <tbody>
          {cas.map((c) => (
            <tr key={c.valeur}>
              <td className="pr-3 font-mono text-slate-600 dark:text-slate-400">{c.valeur}</td>
              <td className="py-0.5">
                <TextInput
                  value={c.nom}
                  onChange={(evenement) => {
                    const nom = evenement.target.value;
                    setSaisis((precedents) => ({ ...precedents, [c.valeur]: nom }));
                  }}
                  aria-label={t('arbitrage.nameCases.label', { valeur: c.valeur })}
                  className="w-48 font-mono"
                />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      <div className="flex items-center gap-1">
        <Button type="submit" variante="discret" taille="petite" disabled={!complet}>
          {t('arbitrage.nameCases')}
        </Button>
        <HelpTip texte={t('arbitrage.help.nameCases')} />
      </div>
    </form>
  );
}
