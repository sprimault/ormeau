// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import type { TableSommaire } from '@/shared/model';
import type { Colonnes } from '../../model/useColumns';
import type { EtatExclusions } from '../../model/useExclusions';
import type { EtatSelection } from '../../model/useSelection';
import { DatabaseTree } from '../DatabaseTree';

const tables: TableSommaire[] = [
  { schema: 'public', nom: 'clients', nb_colonnes: 4, lignes_estimees: 12, cle_primaire: true },
];

const selection: EtatSelection = {
  selection: new Set(),
  manquantes: [],
  basculer: vi.fn(),
  basculerSchema: vi.fn(),
  ajouterManquantes: vi.fn(),
  toutEffacer: vi.fn(),
};

const exclusions: EtatExclusions = {
  ignorees: {},
  total: 0,
  estIgnoree: () => false,
  compte: () => 0,
  basculer: vi.fn(),
};

const colonnes: Colonnes = {
  etat: () => ({ colonnes: undefined, enCours: false, erreur: null }),
  charger: vi.fn(),
};

/** Monte l'arbre avec les doubles ci-dessus. */
function monter(proprietes: Partial<Parameters<typeof DatabaseTree>[0]> = {}) {
  const onOuvrirBase = proprietes.onOuvrirBase ?? vi.fn();
  render(
    <DatabaseTree
      bases={['gescom', 'paie']}
      courante="gescom"
      tables={tables}
      enCours={false}
      erreur={null}
      etat={selection}
      colonnes={colonnes}
      exclusions={exclusions}
      active={null}
      onActiver={vi.fn()}
      {...proprietes}
      onOuvrirBase={onOuvrirBase}
    />,
  );
  return onOuvrirBase;
}

describe('DatabaseTree', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
  });

  it('liste toutes les bases du serveur', () => {
    monter();
    expect(screen.getByRole('button', { name: /gescom/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /paie/ })).toBeInTheDocument();
  });

  it('n’ouvre que la base courante', () => {
    monter();
    expect(screen.getByRole('button', { name: /gescom/ })).toHaveAttribute('aria-expanded', 'true');
    expect(screen.getByRole('button', { name: /paie/ })).toHaveAttribute('aria-expanded', 'false');
    expect(screen.getByText('clients')).toBeInTheDocument();
  });

  it('replie la base ouverte quand on la reclique', async () => {
    monter();

    await userEvent.click(screen.getByRole('button', { name: /gescom/ }));

    // Le repli ne ferme pas la connexion : seules les tables disparaissent.
    expect(screen.queryByText('clients')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: /gescom/ })).toHaveAttribute(
      'aria-expanded',
      'false',
    );
  });

  it('déplie de nouveau au clic suivant', async () => {
    monter();
    const base = screen.getByRole('button', { name: /gescom/ });

    await userEvent.click(base);
    await userEvent.click(base);

    expect(screen.getByText('clients')).toBeInTheDocument();
  });

  it('rouvre la connexion sur une autre base', async () => {
    const onOuvrirBase = monter();

    await userEvent.click(screen.getByRole('button', { name: /paie/ }));

    expect(onOuvrirBase).toHaveBeenCalledWith('paie');
  });

  it('garde visible une base courante que l’énumération écarte', () => {
    // Sans base au DSN, on atterrit sur « postgres », que ListerBases exclut.
    monter({ courante: 'postgres' });
    expect(screen.getByRole('button', { name: /postgres/ })).toBeInTheDocument();
  });

  it('empêche de changer de base pendant un chargement', () => {
    monter({ enCours: true });
    expect(screen.getByRole('button', { name: /paie/ })).toBeDisabled();
  });

  it('explique un serveur qui n’expose qu’une base', () => {
    monter({ bases: ['gescom'] });
    expect(screen.getByText(/qu’une base/)).toBeInTheDocument();
  });
});
