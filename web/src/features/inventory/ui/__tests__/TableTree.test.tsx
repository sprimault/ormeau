// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import type { Decisions, TableSommaire } from '@/shared/model';
import type { Colonnes } from '../../model/useColumns';
import { useExclusions } from '../../model/useExclusions';
import { useSelection } from '../../model/useSelection';
import { TableTree } from '../TableTree';

// Les exclusions vivent dans le brouillon de décisions, dont la lecture du
// fichier a ses propres tests : un brouillon en mémoire suffit à l'arbre.
vi.mock('@/entities/decisions', async () => {
  const { useCallback, useState } = await import('react');
  return {
    useDecisions: () => {
      const [decisions, setDecisions] = useState<Decisions>({});
      const modifier = useCallback(
        (transformation: (d: Decisions) => Decisions) => setDecisions((d) => transformation(d)),
        [],
      );
      return { decisions, modifier };
    },
  };
});

/**
 * Monte l'arbre avec une sélection réelle.
 *
 * La sélection est tenue par l'écran, qui la partage avec le détail et la
 * portée : la reconstituer ici testerait un montage que personne n'utilise.
 */
function Arbre({
  tables,
  enCours = false,
  erreur = null,
}: {
  tables: TableSommaire[];
  enCours?: boolean;
  erreur?: string | null;
}) {
  const etat = useSelection(tables);
  const exclusions = useExclusions();
  const [active, setActive] = useState<string | null>(null);

  // Les colonnes ne se chargent qu'au dépliement : ce double ne dit rien, ce
  // qui suffit à ces cas. Le dépliement se teste dans ColumnList.test.tsx.
  const colonnes: Colonnes = {
    etat: () => ({ colonnes: undefined, enCours: false, erreur: null }),
    charger: () => {},
  };

  return (
    <TableTree
      tables={tables}
      enCours={enCours}
      erreur={erreur}
      etat={etat}
      colonnes={colonnes}
      exclusions={exclusions}
      active={active}
      onActiver={setActive}
    />
  );
}

/** Fabrique un sommaire de table pour l'affichage. */
function table(nom: string, reste: Partial<TableSommaire> = {}, schema = 'public'): TableSommaire {
  return {
    schema,
    nom,
    nb_colonnes: 4,
    lignes_estimees: 0,
    cle_primaire: true,
    ...reste,
  };
}

const inventaire = [
  table('commandes', {
    reference_vers: ['public.clients'],
    lignes_estimees: 48210,
  }),
  table('clients', { commentaire: 'fiches signalétiques' }),
  table('t_log_import', { cle_primaire: false }),
  table('parametres', {}, 'compta'),
];

describe('TableTree', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
  });

  it('groupe les tables par schéma', () => {
    render(<Arbre tables={inventaire} />);
    expect(screen.getByText(/public/)).toBeInTheDocument();
    expect(screen.getByText(/compta/)).toBeInTheDocument();
  });

  it('affiche les noms tels qu’ils sont en base', () => {
    render(<Arbre tables={inventaire} />);
    expect(screen.getByText('t_log_import')).toBeInTheDocument();
  });

  it('signale une table sans clé primaire', () => {
    render(<Arbre tables={inventaire} />);
    expect(screen.getByTitle('sans clé primaire')).toBeInTheDocument();
  });

  it('sépare consulter de retenir : cliquer le nom ne coche pas', async () => {
    render(<Arbre tables={inventaire} />);

    // Le bouton porte le nom et la volumétrie : c'est une seule cible de clic.
    await userEvent.click(screen.getByRole('button', { name: /commandes/ }));
    expect(screen.getByRole('checkbox', { name: 'public.commandes' })).not.toBeChecked();
  });

  it('rend la volumétrie sous forme compacte', () => {
    render(<Arbre tables={inventaire} />);
    expect(screen.getByText('48 k lignes')).toBeInTheDocument();
  });

  it('filtre à la frappe, sans rien demander au serveur', async () => {
    render(<Arbre tables={inventaire} />);

    await userEvent.type(screen.getByLabelText('Filtrer les tables'), 'log');
    expect(screen.getByText('t_log_import')).toBeInTheDocument();
    expect(screen.queryByText('commandes')).not.toBeInTheDocument();
  });

  it('cherche aussi dans les commentaires', async () => {
    render(<Arbre tables={inventaire} />);

    await userEvent.type(screen.getByLabelText('Filtrer les tables'), 'signalétiques');
    expect(screen.getByText('clients')).toBeInTheDocument();
  });

  it('annonce qu’aucune table ne correspond', async () => {
    render(<Arbre tables={inventaire} />);

    await userEvent.type(screen.getByLabelText('Filtrer les tables'), 'introuvable');
    expect(screen.getByText(/Aucune table ne correspond/)).toBeInTheDocument();
  });

  it('avertit quand une référence sort de la sélection, et sait la rattraper', async () => {
    render(<Arbre tables={inventaire} />);

    await userEvent.click(screen.getAllByRole('checkbox')[0]);
    expect(screen.getByText(/n’en font pas partie/)).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'Ajouter les dépendances' }));
    expect(screen.queryByText(/n’en font pas partie/)).not.toBeInTheDocument();
  });

  it('garde le compte visible schéma replié', async () => {
    render(<Arbre tables={inventaire} />);

    await userEvent.click(screen.getByRole('button', { name: /public/ }));
    expect(screen.queryByText('commandes')).not.toBeInTheDocument();
    expect(screen.getByText('0 / 3 table(s)')).toBeInTheDocument();
  });

  it('affiche la lecture en cours plutôt qu’un arbre vide', () => {
    render(<Arbre tables={[]} enCours />);
    expect(screen.getByText('Lecture du catalogue…')).toBeInTheDocument();
  });

  it('affiche l’échec du serveur', () => {
    render(<Arbre tables={[]} erreur="connexion perdue" />);
    expect(screen.getByRole('alert')).toHaveTextContent('connexion perdue');
  });

  it('signale une portée sans table', () => {
    render(<Arbre tables={[]} />);
    expect(screen.getByText('Aucune table dans les schémas retenus.')).toBeInTheDocument();
  });
});
