// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useState, type ReactNode } from 'react';

import { qualifier } from '@/shared/lib';
import type { Portee, ReponseConnexion } from '@/shared/model';
import { SplitPane } from '@/shared/ui';
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
}

/**
 * Écran de sélection.
 *
 * Trois zones fixes : l'arbre des bases à gauche, le détail de la table active à
 * droite, la portée qui partira à l'extraction en bas. La sélection est tenue
 * ici parce que les trois la lisent — l'arbre pour cocher, le détail pour
 * signaler une référence sortante, la portée pour la composer.
 *
 * L'écran compose la portée mais ne la lance pas : l'extraction est une autre
 * feature, que l'application branche par `actions`.
 */
export function InventoryScreen({
  serveur,
  bases,
  enCours,
  onOuvrirBase,
  actions,
}: ProprietesEcran) {
  const inventaire = useInventory(serveur.session, serveur.schemas);
  const etat = useSelection(inventaire.tables);
  const colonnes = useColumns(serveur.session);
  const exclusions = useExclusions();
  const portee = usePortee(serveur.schemas, etat.selection);
  const [active, setActive] = useState<string | null>(null);

  const table = inventaire.tables.find(
    (candidate) => qualifier(candidate.schema, candidate.nom) === active,
  );

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <SplitPane
        cleStockage="ormeau-largeur-arbre"
        gauche={
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
        droite={
          <TableDetails
            table={table ?? null}
            selection={etat.selection}
            colonnes={colonnes}
            exclusions={exclusions}
          />
        }
      />
      <ScopePreview
        schemas={serveur.schemas}
        selection={etat.selection}
        exclusions={exclusions}
        actions={actions?.(portee)}
      />
    </div>
  );
}
