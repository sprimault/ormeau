// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import type { ColonneSommaire, TableSommaire } from '@/shared/model';
import type { Colonnes } from '../../model/useColumns';
import type { EtatExclusions } from '../../model/useExclusions';
import { TableDetails } from '../TableDetails';

/** Table dont le panneau affiche le détail. */
const commandes: TableSommaire = {
  schema: 'public',
  nom: 'commandes',
  nb_colonnes: 3,
  lignes_estimees: 48210,
  cle_primaire: true,
  reference_vers: ['public.clients'],
};

/** Ses colonnes, dont un type utilisateur (`citext`) affiché tel que le catalogue le nomme. */
const colonnesDeCommandes: ColonneSommaire[] = [
  { nom: 'id', position: 1, type_brut: 'integer', nullable: false, cle_primaire: true },
  { nom: 'reference', position: 2, type_brut: 'citext', nullable: false, cle_primaire: false },
  { nom: 'note', position: 3, type_brut: 'text', nullable: true, cle_primaire: false },
];

/** Rend un cache de colonnes figé. */
function colonnes(etat: Partial<ReturnType<Colonnes['etat']>> = {}): Colonnes {
  return {
    etat: () => ({
      colonnes: colonnesDeCommandes,
      enCours: false,
      erreur: null,
      ...etat,
    }),
    charger: vi.fn(),
  };
}

/** Rend un état d'exclusions contrôlable. */
function exclusions(ignorees: string[] = []): EtatExclusions {
  return {
    ignorees: {},
    total: ignorees.length,
    estIgnoree: (_table, colonne) => ignorees.includes(colonne),
    compte: () => ignorees.length,
    basculer: vi.fn(),
  };
}

describe('TableDetails', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
  });

  it('invite à choisir une table quand aucune ne l’est', () => {
    render(
      <TableDetails
        table={null}
        selection={new Set()}
        colonnes={colonnes()}
        exclusions={exclusions()}
      />,
    );
    expect(screen.getByText('Choisir une table pour en voir le détail.')).toBeInTheDocument();
  });

  it('liste les colonnes avec leur type verbatim', () => {
    render(
      <TableDetails
        table={commandes}
        selection={new Set()}
        colonnes={colonnes()}
        exclusions={exclusions()}
      />,
    );
    expect(screen.getByText('reference')).toBeInTheDocument();
    expect(screen.getByText('citext')).toBeInTheDocument();
    expect(screen.getByText('note')).toBeInTheDocument();
  });

  it('demande les colonnes au serveur à l’activation', () => {
    const cache = colonnes();
    render(
      <TableDetails
        table={commandes}
        selection={new Set()}
        colonnes={cache}
        exclusions={exclusions()}
      />,
    );
    expect(cache.charger).toHaveBeenCalledWith('public.commandes', 'public', 'commandes');
  });

  it('marque une colonne écartée sans la cacher', () => {
    render(
      <TableDetails
        table={commandes}
        selection={new Set()}
        colonnes={colonnes()}
        exclusions={exclusions(['note'])}
      />,
    );
    expect(screen.getByText('écartée')).toBeInTheDocument();
    expect(screen.getByText('note')).toBeInTheDocument();
  });

  it('annonce la lecture des colonnes', () => {
    render(
      <TableDetails
        table={commandes}
        selection={new Set()}
        colonnes={colonnes({ colonnes: undefined, enCours: true })}
        exclusions={exclusions()}
      />,
    );
    expect(screen.getByText('Lecture des colonnes…')).toBeInTheDocument();
  });

  it('avertit qu’une table sans clé primaire sera refusée', () => {
    render(
      <TableDetails
        table={{ ...commandes, cle_primaire: false }}
        selection={new Set()}
        colonnes={colonnes()}
        exclusions={exclusions()}
      />,
    );
    expect(screen.getByText(/Doctrine refusera cette entité/)).toBeInTheDocument();
  });

  it('marque une référence qui sort de la sélection', () => {
    render(
      <TableDetails
        table={commandes}
        selection={new Set(['public.commandes'])}
        colonnes={colonnes()}
        exclusions={exclusions()}
      />,
    );
    expect(screen.getByText('hors sélection')).toBeInTheDocument();
  });

  it('ne marque rien quand la cible est retenue', () => {
    render(
      <TableDetails
        table={commandes}
        selection={new Set(['public.commandes', 'public.clients'])}
        colonnes={colonnes()}
        exclusions={exclusions()}
      />,
    );
    expect(screen.queryByText('hors sélection')).not.toBeInTheDocument();
  });

  it('signale l’absence de clé étrangère', () => {
    render(
      <TableDetails
        table={{ ...commandes, reference_vers: [] }}
        selection={new Set()}
        colonnes={colonnes()}
        exclusions={exclusions()}
      />,
    );
    expect(screen.getByText('Aucune clé étrangère déclarée.')).toBeInTheDocument();
  });
});
