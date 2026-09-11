// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useEffect, useState } from 'react';

import { ErreurAPI } from '@/shared/api';
import { translate } from '@/shared/i18n';
import {
  CodeCalqueModifie,
  type Decisions,
  type ReferenceTable,
  type ReponseEntite,
} from '@/shared/model';
import { lireEntite } from '../api/arbitrageApi';

/** Le détail d'une entité, tel que le panneau l'affiche. */
export interface EtatEntite {
  detail: ReponseEntite | null;
  enCours: boolean;
  erreur: string | null;
}

/**
 * Demande le détail d'une entité à son ouverture, puis à chaque inférence.
 *
 * Il suit le brouillon que l'inférence a jugé, pas celui qu'on saisit : le
 * détail et la liste décrivent ainsi les mêmes décisions, et un nom tapé lettre
 * à lettre ne coûte qu'un calcul.
 *
 * Un calque réécrit n'est pas une erreur du détail : l'inférence le signale
 * déjà, et le panneau garde ce qu'il montrait.
 */
export function useEntite(
  base: string,
  decisions: Decisions,
  empreinte: string,
  table: ReferenceTable | null,
): EtatEntite {
  const [detail, setDetail] = useState<ReponseEntite | null>(null);
  const [enCours, setEnCours] = useState(false);
  const [erreur, setErreur] = useState<string | null>(null);
  const schema = table?.schema;
  const nom = table?.nom;

  useEffect(() => {
    if (schema === undefined || nom === undefined || empreinte === '') {
      return;
    }
    const abandon = new AbortController();
    setEnCours(true);

    lireEntite(
      { base, decisions, empreinte_physique: empreinte, schema, table: nom },
      abandon.signal,
    )
      .then((reponse) => {
        setDetail(reponse);
        setErreur(null);
        setEnCours(false);
      })
      .catch((echec: unknown) => {
        if (abandon.signal.aborted) {
          return;
        }
        if (!(echec instanceof ErreurAPI && echec.code === CodeCalqueModifie)) {
          setErreur(echec instanceof ErreurAPI ? echec.message : translate('error.unknown'));
        }
        setEnCours(false);
      });

    return () => abandon.abort();
  }, [base, decisions, empreinte, schema, nom]);

  return { detail, enCours, erreur };
}
