// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useState, type ReactNode } from 'react';

import { qualifier } from '@/shared/lib';
import type { Portee, ReponseConnexion } from '@/shared/model';
import { SplitPane } from '@/shared/ui';
import { useApercuOuvert } from '../model/useApercuOuvert';
import { useColumns } from '../model/useColumns';
import { useExclusions } from '../model/useExclusions';
import { useInventory } from '../model/useInventory';
import { usePortee } from '../model/usePortee';
import { useSelection } from '../model/useSelection';
import { DatabaseTree } from './DatabaseTree';
import { ScopePreview } from './ScopePreview';
import { TableDetails } from './TableDetails';

/** Propriétés de l'écran. */
interface ProprietesEcran {
  serveur: ReponseConnexion;
  bases: string[];
  enCours: boolean;
  onOuvrirBase: (base: string) => void;
  /** Ce qui agit sur la portée composée ici — le lancement d'une extraction. */
  actions?: (portee: Portee) => ReactNode;
  /** Ce que l'extraction a produit, affiché à côté de la portée. */
  produit?: ReactNode;
}

/**
 * Écran de sélection.
 *
 * Trois zones : l'arbre des bases à gauche, le détail de la table active à
 * droite, et en bas la portée qui partira à côté de ce qui a été produit. La
 * sélection est tenue ici parce que les trois la lisent — l'arbre pour cocher,
 * le détail pour signaler une référence sortante, la portée pour la composer.
 *
 * La zone du bas se règle en hauteur et se replie : on la garde grande pour
 * relire un calque, petite pour cocher sur quatre cents tables.
 *
 * L'écran compose la portée mais ne la lance pas : l'extraction est une autre
 * feature, que l'application branche par `actions` et `produit`.
 */
export function InventoryScreen({
  serveur,
  bases,
  enCours,
  onOuvrirBase,
  actions,
  produit,
}: ProprietesEcran) {
  const inventaire = useInventory(serveur.session, serveur.schemas);
  const etat = useSelection(inventaire.tables);
  const colonnes = useColumns(serveur.session);
  const exclusions = useExclusions();
  const portee = usePortee(serveur.schemas, inventaire.tables, etat.selection);
  const [apercuOuvert, basculerApercu] = useApercuOuvert();
  const [active, setActive] = useState<string | null>(null);

  const table = inventaire.tables.find(
    (candidate) => qualifier(candidate.schema, candidate.nom) === active,
  );

  return (
    <SplitPane
      sens="vertical"
      cleStockage="ormeau-hauteur-apercu"
      defaut={320}
      replie={!apercuOuvert}
      premier={
        <SplitPane
          cleStockage="ormeau-largeur-arbre"
          premier={
            <DatabaseTree
              bases={bases}
              courante={serveur.catalogue}
              tables={inventaire.tables}
              enCours={enCours || inventaire.enCours}
              erreur={inventaire.erreur}
              etat={etat}
              colonnes={colonnes}
              exclusions={exclusions}
              active={active}
              onActiver={setActive}
              onOuvrirBase={onOuvrirBase}
            />
          }
          second={
            <TableDetails
              table={table ?? null}
              selection={etat.selection}
              colonnes={colonnes}
              exclusions={exclusions}
            />
          }
        />
      }
      second={
        <ScopePreview
          portee={portee}
          exclusions={exclusions}
          ouvert={apercuOuvert}
          onBasculer={basculerApercu}
          actions={actions?.(portee)}
          produit={produit}
        />
      }
    />
  );
}
