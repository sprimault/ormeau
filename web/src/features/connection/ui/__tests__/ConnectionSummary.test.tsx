// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import { ConnectionSummary } from '../ConnectionSummary';

const serveur = {
  session: 'jeton-de-session',
  sgbd: 'mariadb',
  version: '11.4.2',
  catalogue: 'gescom',
  schemas: ['public', 'Compta_2019'],
};

describe('ConnectionSummary', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
  });

  it('affiche la variante annoncée par le serveur, pas celle qui a été saisie', () => {
    render(<ConnectionSummary serveur={serveur} onFermer={vi.fn()} />);
    expect(screen.getByText('mariadb 11.4.2')).toBeInTheDocument();
  });

  it('rend les noms de schémas tels qu’ils sont en base', () => {
    render(<ConnectionSummary serveur={serveur} onFermer={vi.fn()} />);
    expect(screen.getByText('public, Compta_2019')).toBeInTheDocument();
  });

  it('ne montre jamais l’identifiant de session', () => {
    const { container } = render(<ConnectionSummary serveur={serveur} onFermer={vi.fn()} />);
    expect(container.textContent).not.toContain('jeton-de-session');
  });

  it('signale une base sans schéma exploitable', () => {
    render(
      <ConnectionSummary
        serveur={{ ...serveur, schemas: [] }}

        onFermer={vi.fn()}
      />,
    );
    expect(screen.getByText('Aucun schéma exploitable dans cette base.')).toBeInTheDocument();
  });

  it('permet de refermer la connexion', async () => {
    const onFermer = vi.fn();
    render(
      <ConnectionSummary
        serveur={serveur}

        onFermer={onFermer}
      />,
    );
    await userEvent.click(screen.getByRole('button', { name: 'Se déconnecter' }));
    expect(onFermer).toHaveBeenCalledOnce();
  });
});
