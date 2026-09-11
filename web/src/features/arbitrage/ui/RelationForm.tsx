// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useId, useState } from 'react';

import { useT } from '@/shared/i18n';
import { minusculeInitiale, qualifier } from '@/shared/lib';
import type { Decisions, RelationForcee } from '@/shared/model';
import { Button, HelpTip, TextInput } from '@/shared/ui';
import type { LigneEntite } from '../model/entites';
import { useEntite } from '../model/useEntite';

/** Propriétés du formulaire. */
interface ProprietesRelation {
  /** Table de l'entité ouverte, qui porte la colonne. */
  qualifiee: string;
  nomEntite: string;
  colonnes: string[];
  entites: LigneEntite[];
  base: string;
  /** Le brouillon jugé par la dernière inférence, pour lire l'entité liée. */
  decisions: Decisions;
  empreinte: string;
  onRelier: (relation: RelationForcee) => void;
}

/**
 * Relie une colonne à une autre entité, quand la base n'a jamais déclaré la
 * clé étrangère.
 *
 * Replié par défaut : c'est un geste rare, qui n'a pas à occuper le panneau.
 */
export function RelationForm(proprietes: ProprietesRelation) {
  const t = useT();
  const [ouvert, setOuvert] = useState(false);

  if (!ouvert) {
    return (
      <div className="flex items-center gap-1">
        <Button variante="discret" taille="petite" onClick={() => setOuvert(true)}>
          {t('arbitrage.relation.open')}
        </Button>
        <HelpTip texte={t('arbitrage.help.relation')} />
      </div>
    );
  }
  return <Formulaire {...proprietes} onFermer={() => setOuvert(false)} />;
}

/** Propriétés du formulaire ouvert. */
interface ProprietesFormulaire extends ProprietesRelation {
  onFermer: () => void;
}

/** Le genre qu'une colonne peut porter : elle désigne une ligne, jamais plusieurs. */
type Genre = 'plusieurs_vers_un' | 'un_vers_un';

/**
 * Le formulaire lui-même, remonté à chaque ouverture pour repartir de zéro.
 *
 * Ce qui se déduit est prérempli : la colonne désignée est la clé primaire de
 * l'entité liée, le nom de la propriété celui de sa classe. Il ne reste qu'à
 * choisir la colonne et l'entité.
 */
function Formulaire({
  qualifiee,
  nomEntite,
  colonnes,
  entites,
  base,
  decisions,
  empreinte,
  onRelier,
  onFermer,
}: ProprietesFormulaire) {
  const t = useT();
  const id = useId();
  const [colonne, setColonne] = useState('');
  const [cible, setCible] = useState('');
  const [colonneChoisie, setColonneChoisie] = useState<string | null>(null);
  const [genre, setGenre] = useState<Genre>('plusieurs_vers_un');
  const [nomSaisi, setNomSaisi] = useState<string | null>(null);

  const ligneCible = entites.find((e) => e.qualifiee === cible);
  const etatCible = useEntite(base, decisions, empreinte, ligneCible?.table ?? null);
  const tableCible =
    etatCible.detail &&
    qualifier(etatCible.detail.table_physique.schema, etatCible.detail.table_physique.nom) === cible
      ? etatCible.detail.table_physique
      : null;

  const colonneDesignee = colonneChoisie ?? tableCible?.cle_primaire?.colonnes[0] ?? '';
  const nom = nomSaisi ?? (ligneCible ? minusculeInitiale(ligneCible.nom) : '');
  const complet = colonne !== '' && cible !== '' && colonneDesignee !== '' && nom.trim() !== '';
  const nomCible = ligneCible?.nom ?? '…';

  return (
    <form
      className="flex flex-col gap-2 rounded border border-slate-200 p-3 dark:border-slate-800"
      onSubmit={(evenement) => {
        evenement.preventDefault();
        if (!complet) {
          return;
        }
        onRelier({
          source: `${qualifiee}.${colonne}`,
          cible: `${cible}.${colonneDesignee}`,
          genre,
          nom: nom.trim(),
        });
        onFermer();
      }}
    >
      <h3 className="flex items-center gap-1 font-semibold">
        {t('arbitrage.relation.title')}
        <HelpTip texte={t('arbitrage.help.relation')} />
      </h3>

      <div className="grid grid-cols-[auto_1fr] items-center gap-x-3 gap-y-1.5">
        <label htmlFor={`${id}-colonne`}>{t('arbitrage.relation.column')}</label>
        <Choix
          id={`${id}-colonne`}
          valeur={colonne}
          onChange={setColonne}
          options={colonnes.map((c) => ({ valeur: c, libelle: c }))}
        />

        <label htmlFor={`${id}-cible`}>{t('arbitrage.relation.target')}</label>
        <Choix
          id={`${id}-cible`}
          valeur={cible}
          onChange={(valeur) => {
            setCible(valeur);
            setColonneChoisie(null);
            setNomSaisi(null);
          }}
          options={entites.map((e) => ({ valeur: e.qualifiee, libelle: `${e.nom} — ${e.qualifiee}` }))}
        />

        <label htmlFor={`${id}-designee`}>{t('arbitrage.relation.targetColumn')}</label>
        <Choix
          id={`${id}-designee`}
          valeur={colonneDesignee}
          onChange={setColonneChoisie}
          options={(tableCible?.colonnes ?? []).map((c) => ({ valeur: c.nom, libelle: c.nom }))}
          desactive={!tableCible}
        />

        <label htmlFor={`${id}-genre`}>{t('arbitrage.relation.kind')}</label>
        <Choix
          id={`${id}-genre`}
          valeur={genre}
          onChange={(valeur) => setGenre(valeur as Genre)}
          options={[
            {
              valeur: 'plusieurs_vers_un',
              libelle: t('arbitrage.relation.manyToOne', { source: nomEntite, cible: nomCible }),
            },
            {
              valeur: 'un_vers_un',
              libelle: t('arbitrage.relation.oneToOne', { source: nomEntite, cible: nomCible }),
            },
          ]}
        />

        <label htmlFor={`${id}-nom`}>{t('arbitrage.relation.name')}</label>
        <TextInput
          id={`${id}-nom`}
          value={nom}
          onChange={(evenement) => setNomSaisi(evenement.target.value)}
          className="w-56 font-mono"
        />
      </div>

      <div className="flex items-center gap-2">
        <Button type="submit" taille="petite" disabled={!complet}>
          {t('arbitrage.relation.submit')}
        </Button>
        <Button type="button" variante="discret" taille="petite" onClick={onFermer}>
          {t('arbitrage.cancel')}
        </Button>
      </div>
    </form>
  );
}

/** Propriétés d'une liste de choix. */
interface ProprietesChoix {
  id: string;
  valeur: string;
  options: { valeur: string; libelle: string }[];
  onChange: (valeur: string) => void;
  desactive?: boolean;
}

/** Liste de choix du formulaire, avec une invite tant que rien n'est choisi. */
function Choix({ id, valeur, options, onChange, desactive = false }: ProprietesChoix) {
  const t = useT();

  return (
    <select
      id={id}
      value={valeur}
      disabled={desactive}
      onChange={(evenement) => onChange(evenement.target.value)}
      className="w-fit max-w-full rounded border border-slate-300 bg-white px-1.5 py-0.5 disabled:opacity-60 dark:border-slate-700 dark:bg-slate-900"
    >
      {valeur === '' ? <option value="">{t('arbitrage.relation.choose')}</option> : null}
      {options.map((option) => (
        <option key={option.valeur} value={option.valeur}>
          {option.libelle}
        </option>
      ))}
    </select>
  );
}
