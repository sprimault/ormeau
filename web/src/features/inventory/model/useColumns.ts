// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useRef, useState } from 'react';

import { ErreurAPI } from '@/shared/api';
import { translate } from '@/shared/i18n';
import type { ColonneSommaire } from '@/shared/model';
import { lireColonnes } from '../api/inventoryApi';

/** Colonnes d'une table, telles que l'arbre les affiche. */
export interface EtatColonnes {
  colonnes: ColonneSommaire[] | undefined;
  enCours: boolean;
  erreur: string | null;
}

/** Ce que l'arbre appelle pour déplier une table. */
export interface Colonnes {
  etat: (cle: string) => EtatColonnes;
  charger: (cle: string, schema: string, table: string) => void;
}

/**
 * Charge les colonnes table par table, et les garde.
 *
 * À la demande, parce que porter les colonnes de quatre cents tables dans
 * l'inventaire coûterait dix fois le transfert pour des lignes que personne
 * n'ouvrira. Gardées ensuite, parce que replier puis rouvrir une table est le
 * geste le plus courant de cet écran, et qu'une table ne change pas de colonnes
 * pendant qu'on la regarde.
 */
export function useColumns(session: string): Colonnes {
  const [parTable, setParTable] = useState<Record<string, EtatColonnes>>({});
  // Les demandes en vol, pour qu'un double repli ne relance pas la requête.
  const enVol = useRef<Set<string>>(new Set());

  const charger = useCallback(
    (cle: string, schema: string, table: string) => {
      if (enVol.current.has(cle)) {
        return;
      }
      enVol.current.add(cle);
      setParTable((precedent) => ({
        ...precedent,
        [cle]: { colonnes: precedent[cle]?.colonnes, enCours: true, erreur: null },
      }));

      lireColonnes(session, schema, table)
        .then((reponse) => {
          setParTable((precedent) => ({
            ...precedent,
            [cle]: { colonnes: reponse.colonnes, enCours: false, erreur: null },
          }));
        })
        .catch((echec: unknown) => {
          setParTable((precedent) => ({
            ...precedent,
            [cle]: {
              colonnes: undefined,
              enCours: false,
              erreur: echec instanceof ErreurAPI ? echec.message : translate('error.unknown'),
            },
          }));
        })
        .finally(() => enVol.current.delete(cle));
    },
    [session],
  );

  const etat = useCallback(
    (cle: string): EtatColonnes =>
      parTable[cle] ?? { colonnes: undefined, enCours: false, erreur: null },
    [parTable],
  );

  return { etat, charger };
}
