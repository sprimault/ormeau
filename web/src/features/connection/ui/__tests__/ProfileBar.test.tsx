// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import type { ProfilResume } from '@/shared/model';
import { ProfileBar } from '../ProfileBar';

const PROFILS: ProfilResume[] = [
  {
    profil: { nom: 'gescom production', hote: '192.168.0.184', base: 'gescom' },
    mot_de_passe_enregistre: true,
  },
  {
    profil: { nom: 'gescom recette', hote: '192.168.0.184', base: 'recette' },
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
    aEnregistrer: () => ({ hote: '192.168.0.184', base: 'gescom' }),
    onEnregistrer: vi.fn().mockResolvedValue(true),
    onSupprimer: vi.fn(),
    motDePasseEnregistre: false,
    ...options,
  } satisfies Parameters<typeof ProfileBar>[0];
  render(<ProfileBar {...props} />);
  return { ...props, utilisateur: userEvent.setup() };
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

  it('ne propose de supprimer que lorsqu’un profil est choisi', async () => {
    const { utilisateur, onSupprimer } = monter({ choisi: 'gescom production' });

    await utilisateur.click(screen.getByRole('button', { name: 'Supprimer ce profil' }));

    expect(onSupprimer).toHaveBeenCalledWith('gescom production');
  });

  it('cache la suppression sur une connexion neuve', () => {
    monter();
    expect(screen.queryByRole('button', { name: 'Supprimer ce profil' })).not.toBeInTheDocument();
  });

  it('enregistre sous le nom saisi, sans le mot de passe par défaut', async () => {
    const { utilisateur, onEnregistrer, onChoisir } = monter();

    await utilisateur.click(screen.getByRole('button', { name: 'Enregistrer cette connexion' }));
    await utilisateur.type(screen.getByLabelText('Nom du profil'), 'nouveau');
    await utilisateur.click(screen.getByRole('button', { name: 'Enregistrer' }));

    expect(onEnregistrer).toHaveBeenCalledWith(
      { hote: '192.168.0.184', base: 'gescom', nom: 'nouveau' },
      false,
      false,
    );
    expect(onChoisir).toHaveBeenCalledWith('nouveau');
  });

  it('n’ouvre la case du mot de passe qu’avec le bloc d’enregistrement', async () => {
    const { utilisateur } = monter();
    // Elle n'agit qu'à l'enregistrement : ailleurs, elle laisserait croire
    // qu'elle fait quelque chose au moment de se connecter.
    expect(screen.queryByRole('checkbox')).not.toBeInTheDocument();

    await utilisateur.click(screen.getByRole('button', { name: 'Enregistrer cette connexion' }));

    expect(screen.getByRole('checkbox')).not.toBeChecked();
  });

  it('enregistre le mot de passe quand la case est cochée', async () => {
    const { utilisateur, onEnregistrer } = monter();

    await utilisateur.click(screen.getByRole('button', { name: 'Enregistrer cette connexion' }));
    await utilisateur.type(screen.getByLabelText('Nom du profil'), 'nouveau');
    await utilisateur.click(screen.getByRole('checkbox'));
    await utilisateur.click(screen.getByRole('button', { name: 'Enregistrer' }));

    expect(onEnregistrer).toHaveBeenLastCalledWith(expect.anything(), false, true);
  });

  it('prévient qu’un enregistrement effacerait le mot de passe retenu', async () => {
    const { utilisateur } = monter({ choisi: 'gescom production', motDePasseEnregistre: true });

    await utilisateur.click(screen.getByRole('button', { name: 'Enregistrer cette connexion' }));
    expect(screen.getByText(/effacera le mot de passe enregistré/)).toBeInTheDocument();

    // Cochée, l'avertissement n'a plus lieu d'être : le mot de passe est repris.
    await utilisateur.click(screen.getByRole('checkbox'));
    expect(screen.queryByText(/effacera le mot de passe enregistré/)).not.toBeInTheDocument();
  });

  it('demande confirmation quand le nom est déjà pris', async () => {
    // Le serveur refuse tant que le remplacement n'est pas confirmé : écraser
    // un profil de production par erreur n'a rien d'anodin.
    const onEnregistrer = vi
      .fn()
      .mockImplementation((_profil, remplacer: boolean) => Promise.resolve(remplacer));
    const { utilisateur } = monter({ onEnregistrer });

    await utilisateur.click(screen.getByRole('button', { name: 'Enregistrer cette connexion' }));
    await utilisateur.type(screen.getByLabelText('Nom du profil'), 'gescom production');
    await utilisateur.click(screen.getByRole('button', { name: 'Enregistrer' }));

    expect(await screen.findByRole('alert')).toHaveTextContent('porte déjà ce nom');

    await utilisateur.click(screen.getByRole('button', { name: 'Écraser' }));
    expect(onEnregistrer).toHaveBeenLastCalledWith(expect.anything(), true, false);
  });

  it('affiche l’avertissement du fichier de profils', () => {
    monter({ avertissement: 'cle.bin est absent' });
    expect(screen.getByText(/cle.bin est absent/)).toBeInTheDocument();
  });
});
