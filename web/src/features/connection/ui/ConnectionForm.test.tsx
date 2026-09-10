// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import { ConnectionForm } from './ConnectionForm';

describe('ConnectionForm', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
  });

  it('poste les composants saisis, sans SGBD quand le port suffit', async () => {
    const onConnecter = vi.fn();
    render(<ConnectionForm enCours={false} erreur={null} onConnecter={onConnecter} />);

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
    render(<ConnectionForm enCours={false} erreur={null} onConnecter={onConnecter} />);

    await userEvent.click(screen.getByRole('tab', { name: 'Chaîne de connexion' }));
    await userEvent.type(screen.getByLabelText('DSN'), 'postgres://u@h/gescom');
    await userEvent.click(screen.getByRole('button', { name: 'Se connecter' }));

    expect(onConnecter).toHaveBeenCalledWith({ dsn: 'postgres://u@h/gescom' });
  });

  it('n’expose pas le mot de passe en clair', () => {
    render(<ConnectionForm enCours={false} erreur={null} onConnecter={vi.fn()} />);
    expect(screen.getByLabelText('Mot de passe')).toHaveAttribute('type', 'password');
  });

  it('annonce l’échec et laisse rejouer', () => {
    render(<ConnectionForm enCours={false} erreur="mot de passe refusé" onConnecter={vi.fn()} />);
    expect(screen.getByRole('alert')).toHaveTextContent('mot de passe refusé');
    expect(screen.getByRole('button', { name: 'Se connecter' })).toBeEnabled();
  });

  it('bloque la soumission pendant la connexion', () => {
    render(<ConnectionForm enCours erreur={null} onConnecter={vi.fn()} />);
    expect(screen.getByRole('button', { name: 'Connexion…' })).toBeDisabled();
  });
});
