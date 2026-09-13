// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import type { ColonneSommaire } from '@/shared/model';
import type { EtatExclusions } from '../../model/useExclusions';
import { ColumnList } from '../ColumnList';

/** Colonnes affichées sous une table dépliée. */
const colonnes: ColonneSommaire[] = [
  { nom: 'id', position: 1, type_brut: 'integer', nullable: false, cle_primaire: true },
  {
    nom: 'nom',
    position: 2,
    type_brut: 'character varying(120)',
    nullable: false,
    cle_primaire: false,
  },
  { nom: 'photo', position: 3, type_brut: 'bytea', nullable: true, cle_primaire: false },
];

/** Rend un état d'exclusions contrôlable. */
function exclusions(ignorees: string[] = [], basculer = vi.fn()): EtatExclusions {
  return {
    ignorees: {},
    total: ignorees.length,
    estIgnoree: (_table, colonne) => ignorees.includes(colonne),
    compte: () => ignorees.length,
    basculer,
  };
}

describe('ColumnList', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
  });

  it('rend le type verbatim, comme en base', () => {
    render(
      <ColumnList
        cle="public.clients"
        etat={{ colonnes, enCours: false, erreur: null }}
        exclusions={exclusions()}
      />,
    );
    expect(screen.getByText('character varying(120)')).toBeInTheDocument();
  });

  it('coche par défaut : toute colonne devient une propriété', () => {
    render(
      <ColumnList
        cle="public.clients"
        etat={{ colonnes, enCours: false, erreur: null }}
        exclusions={exclusions()}
      />,
    );
    expect(screen.getByRole('checkbox', { name: 'photo' })).toBeChecked();
  });

  it('décoche ce que l’utilisateur a écarté', () => {
    render(
      <ColumnList
        cle="public.clients"
        etat={{ colonnes, enCours: false, erreur: null }}
        exclusions={exclusions(['photo'])}
      />,
    );
    expect(screen.getByRole('checkbox', { name: 'photo' })).not.toBeChecked();
  });

  it('interdit d’écarter une clé primaire, plutôt que de le défaire plus loin', () => {
    render(
      <ColumnList
        cle="public.clients"
        etat={{ colonnes, enCours: false, erreur: null }}
        exclusions={exclusions()}
      />,
    );
    expect(screen.getByRole('checkbox', { name: 'id' })).toBeDisabled();
  });

  it('signale la colonne à écarter', async () => {
    const basculer = vi.fn();
    render(
      <ColumnList
        cle="public.clients"
        etat={{ colonnes, enCours: false, erreur: null }}
        exclusions={exclusions([], basculer)}
      />,
    );

    await userEvent.click(screen.getByRole('checkbox', { name: 'photo' }));
    expect(basculer).toHaveBeenCalledWith('public.clients', 'photo');
  });

  it('annonce la lecture en cours', () => {
    render(
      <ColumnList
        cle="public.clients"
        etat={{ colonnes: undefined, enCours: true, erreur: null }}
        exclusions={exclusions()}
      />,
    );
    expect(screen.getByText('Lecture des colonnes…')).toBeInTheDocument();
  });

  it('affiche l’échec du serveur', () => {
    render(
      <ColumnList
        cle="public.clients"
        etat={{ colonnes: undefined, enCours: false, erreur: 'connexion perdue' }}
        exclusions={exclusions()}
      />,
    );
    expect(screen.getByText('connexion perdue')).toBeInTheDocument();
  });
});
