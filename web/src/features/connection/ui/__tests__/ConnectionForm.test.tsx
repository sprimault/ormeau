// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import type { ProfilResume } from '@/shared/model';
import { ConnectionForm } from '../ConnectionForm';

const { profils } = vi.hoisted(() => ({
  profils: [
    {
      profil: {
        nom: 'gescom production',
        sgbd: 'postgres',
        hote: '192.168.0.184',
        port: 30432,
        utilisateur: 'postgres',
        base: 'cadensio_main',
      },
      mot_de_passe_enregistre: true,
    },
  ] satisfies ProfilResume[],
}));

// Le serveur n'existe pas sous test : la liste est posée d'emblée, et les
// écritures ne partent nulle part.
vi.mock('../../api/profilsApi', () => ({
  lireProfils: () => Promise.resolve({ profils }),
  enregistrerProfil: () => Promise.resolve({ profils }),
  supprimerProfil: () => Promise.resolve(),
}));

/**
 * Monte le formulaire et attend que la liste des profils soit arrivée.
 *
 * L'attente n'est pas une précaution : la liste se charge après le montage, et
 * un test qui rendrait la main avant laisserait React mettre l'état à jour hors
 * de `act` — un avertissement qui finirait par masquer un vrai signal.
 */
async function monter(props: Partial<Parameters<typeof ConnectionForm>[0]> = {}) {
  render(
    <ConnectionForm enCours={false} erreur={null} onConnecter={vi.fn()} {...props} />,
  );
  await screen.findByRole('option', { name: 'gescom production' });
}

describe('ConnectionForm', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
  });

  it('poste les composants saisis, sans SGBD quand le port suffit', async () => {
    const onConnecter = vi.fn();
    await monter({ onConnecter });

    await userEvent.clear(screen.getByLabelText('Hôte'));
    await userEvent.type(screen.getByLabelText('Hôte'), 'bdd');
    await userEvent.type(screen.getByLabelText('Utilisateur'), 'gescom');
    await userEvent.type(screen.getByLabelText('Mot de passe'), 'secret');
    await userEvent.type(screen.getByLabelText('Base'), 'facturation');
    await userEvent.click(screen.getByRole('button', { name: 'Se connecter' }));

    expect(onConnecter).toHaveBeenCalledWith({
      sgbd: undefined,
      hote: 'bdd',
      port: 5432,
      utilisateur: 'gescom',
      mot_de_passe: 'secret',
      base: 'facturation',
    });
  });

  it('poste la chaîne de connexion dans l’autre mode', async () => {
    const onConnecter = vi.fn();
    await monter({ onConnecter });

    await userEvent.click(screen.getByRole('tab', { name: 'Chaîne de connexion' }));
    await userEvent.type(screen.getByLabelText('DSN'), 'postgres://u@h/gescom');
    await userEvent.click(screen.getByRole('button', { name: 'Se connecter' }));

    expect(onConnecter).toHaveBeenCalledWith({ dsn: 'postgres://u@h/gescom' });
  });

  it('dit qu’un mot de passe est enregistré, sous le champ', async () => {
    await monter();

    await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Profil' }), [
      'gescom production',
    ]);

    // Sous le champ mot de passe : c'est lui que la phrase explique.
    expect(screen.getByText(/laissez le champ vide pour l’utiliser/)).toBeInTheDocument();
    // La case n'est pas là : elle n'agit qu'à l'enregistrement, et vit donc
    // dans le bloc qui enregistre.
    expect(screen.queryByRole('checkbox')).not.toBeInTheDocument();
  });

  it('reprend les champs du profil choisi, jamais son mot de passe', async () => {
    const onConnecter = vi.fn();
    await monter({ onConnecter });

    await userEvent.selectOptions(await screen.findByRole('combobox', { name: 'Profil' }), [
      'gescom production',
    ]);

    expect(screen.getByLabelText('Hôte')).toHaveValue('192.168.0.184');
    expect(screen.getByLabelText('Base')).toHaveValue('cadensio_main');
    expect(screen.getByLabelText('Mot de passe')).toHaveValue('');

    await userEvent.click(screen.getByRole('button', { name: 'Se connecter' }));

    // Le nom suffit : le serveur relit le profil et va chercher le mot de passe.
    expect(onConnecter).toHaveBeenCalledWith(
      expect.objectContaining({ profil: 'gescom production', mot_de_passe: '' }),
    );
  });

  it('passer à la chaîne de connexion désélectionne le profil', async () => {
    // Un onglet, un usage : la chaîne décrit à elle seule la connexion voulue,
    // et un profil qui resterait choisi laisserait croire qu'il s'applique
    // alors que le champ est vide.
    await monter();

    await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Profil' }), [
      'gescom production',
    ]);
    await userEvent.click(screen.getByRole('tab', { name: 'Chaîne de connexion' }));

    expect(screen.getByRole('combobox', { name: 'Profil' })).toHaveValue('');
    expect(screen.queryByText(/laissez le champ vide/)).not.toBeInTheDocument();
  });

  it('revient à l’onglet composants quand on choisit un profil', async () => {
    await monter();

    await userEvent.click(screen.getByRole('tab', { name: 'Chaîne de connexion' }));
    await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Profil' }), [
      'gescom production',
    ]);

    // Un profil se lit en composants : une chaîne ne se reconstituerait pas
    // sans le mot de passe, qui ne descend pas dans la page.
    expect(screen.getByRole('tab', { name: 'Composants' })).toHaveAttribute(
      'aria-selected',
      'true',
    );
    expect(screen.getByLabelText('Hôte')).toHaveValue('192.168.0.184');
  });

  it('n’expose pas le mot de passe en clair', async () => {
    await monter();
    expect(screen.getByLabelText('Mot de passe')).toHaveAttribute('type', 'password');
  });

  it('annonce l’échec et laisse rejouer', async () => {
    await monter({ erreur: 'mot de passe refusé' });
    expect(screen.getByRole('alert')).toHaveTextContent('mot de passe refusé');
    expect(screen.getByRole('button', { name: 'Se connecter' })).toBeEnabled();
  });

  it('bloque la soumission pendant la connexion', async () => {
    await monter({ enCours: true });
    expect(screen.getByRole('button', { name: 'Connexion…' })).toBeDisabled();
  });
});
