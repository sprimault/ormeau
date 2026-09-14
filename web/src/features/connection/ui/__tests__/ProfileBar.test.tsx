// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import type { ProfilResume } from '@/shared/model';
import { ProfileBar } from '../ProfileBar';

/** Profils de la barre, l'un avec mot de passe enregistré. */
const PROFILS: ProfilResume[] = [
  {
    profil: { nom: 'gescom production', hote: '192.168.1.10', base: 'gescom' },
    mot_de_passe_enregistre: true,
  },
  {
    profil: { nom: 'gescom recette', hote: '192.168.1.10', base: 'recette' },
    mot_de_passe_enregistre: false,
  },
];

/** Monte la barre avec des rappels observables. */
function monter(options: Partial<Parameters<typeof ProfileBar>[0]> = {}) {
  const props = {
    profils: PROFILS,
    avertissement: '',
    erreur: null,
    choisi: '',
    onChoisir: vi.fn(),
    nom: '',
    onNommer: vi.fn(),
    avecMotDePasse: false,
    onAvecMotDePasse: vi.fn(),
    divergent: false,
    onMettreAJour: vi.fn(),
    onSupprimer: vi.fn(),
    ...options,
  } satisfies Parameters<typeof ProfileBar>[0];
  const { unmount } = render(<ProfileBar {...props} />);
  return { ...props, unmount, utilisateur: userEvent.setup() };
}

describe('ProfileBar', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
  });

  it('n’a pas de sélection par défaut', () => {
    monter();
    // Sinon l'écran se remplit tout seul, et qui voulait taper une connexion
    // neuve doit d'abord comprendre pourquoi les champs sont pleins.
    expect(screen.getByRole('combobox')).toHaveValue('');
    expect(screen.getByRole('option', { name: 'Nouvelle connexion' })).toBeInTheDocument();
  });

  it('liste les profils enregistrés', () => {
    monter();
    expect(screen.getByRole('option', { name: 'gescom production' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'gescom recette' })).toBeInTheDocument();
  });

  it('n’a aucun bouton d’enregistrement', () => {
    // Un nom saisi suffit : le profil est créé à la connexion réussie, et il
    // n'y a donc ni bouton à trouver ni ordre à respecter.
    monter();
    expect(screen.queryByRole('button', { name: /Enregistrer/ })).not.toBeInTheDocument();
  });

  it('demande un nom pour une connexion neuve, et rien pour un profil choisi', () => {
    const { unmount } = monter();
    expect(screen.getByLabelText('Nom du profil')).toBeInTheDocument();
    unmount();

    // Un profil déjà choisi n'a pas à être réenregistré : le faire à chaque
    // connexion effacerait son mot de passe quand la case est décochée.
    monter({ choisi: 'gescom production' });
    expect(screen.queryByLabelText('Nom du profil')).not.toBeInTheDocument();
    expect(screen.queryByRole('checkbox')).not.toBeInTheDocument();
  });

  it('signale un nom déjà pris', () => {
    monter({ nom: 'gescom production' });
    expect(screen.getByText(/porte déjà ce nom/)).toBeInTheDocument();
  });

  it('n’ouvre la case du mot de passe qu’avec un nom', async () => {
    const { unmount } = monter();
    expect(screen.getByRole('checkbox')).toBeDisabled();
    unmount();

    monter({ nom: 'nouveau' });
    expect(screen.getByRole('checkbox')).toBeEnabled();
  });

  it('ne propose la mise à jour que si le formulaire diverge', () => {
    const { unmount } = monter({ choisi: 'gescom production' });
    expect(screen.queryByRole('button', { name: 'Mettre à jour ce profil' })).not.toBeInTheDocument();
    unmount();

    monter({ choisi: 'gescom production', divergent: true });
    expect(screen.getByRole('button', { name: 'Mettre à jour ce profil' })).toBeInTheDocument();
  });

  it('supprime le profil choisi', async () => {
    const { utilisateur, onSupprimer } = monter({ choisi: 'gescom production' });

    await utilisateur.click(screen.getByRole('button', { name: 'Supprimer ce profil' }));

    expect(onSupprimer).toHaveBeenCalledWith('gescom production');
  });

  it('affiche l’avertissement du fichier de profils', () => {
    monter({ avertissement: 'cle.bin est absent' });
    expect(screen.getByText(/cle.bin est absent/)).toBeInTheDocument();
  });
});
