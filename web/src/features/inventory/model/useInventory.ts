// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useEffect, useState } from 'react';

import { ErreurAPI } from '@/shared/api';
import { translate } from '@/shared/i18n';
import type { TableSommaire } from '@/shared/model';
import { lireInventaire } from '../api/inventoryApi';

/** Ce que l'arbre reçoit du serveur. */
export interface EtatInventaire {
  tables: TableSommaire[];
  enCours: boolean;
  erreur: string | null;
}

/**
 * Charge l'inventaire d'une session, une fois.
 *
 * L'appel est abandonné si la session change ou si l'écran disparaît : sur une
 * base lente, une réponse arrivée après coup écraserait l'arbre de la connexion
 * suivante.
 */
export function useInventory(session: string, schemas: string[]): EtatInventaire {
  const [tables, setTables] = useState<TableSommaire[]>([]);
  const [enCours, setEnCours] = useState(true);
  const [erreur, setErreur] = useState<string | null>(null);

  // Les schémas arrivent dans un tableau reconstruit à chaque rendu : le
  // comparer par son contenu évite de relancer l'inventaire sans raison.
  const cleSchemas = schemas.join(',');

  useEffect(() => {
    const abandon = new AbortController();
    setEnCours(true);
    setErreur(null);

    lireInventaire(session, cleSchemas ? cleSchemas.split(',') : [], abandon.signal)
      .then((reponse) => {
        setTables(reponse.tables);
        setEnCours(false);
      })
      .catch((echec: unknown) => {
        if (abandon.signal.aborted) {
          return;
        }
        setErreur(echec instanceof ErreurAPI ? echec.message : translate('error.unknown'));
        setEnCours(false);
      });

    return () => abandon.abort();
  }, [session, cleSchemas]);

  return { tables, enCours, erreur };
}
