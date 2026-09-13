// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ErreurAPI } from '@/shared/api';
import { useLangStore } from '@/shared/i18n';
import { WorkdirField } from '../WorkdirField';

/** Répertoire affiché au montage. */
const REPERTOIRE = 'C:\\projets\\gescom';

/** Monte le champ avec une fonction de changement observable. */
function monter(changer: (repertoire: string) => Promise<string>) {
  render(<WorkdirField repertoire={REPERTOIRE} changer={changer} />);
  return userEvent.setup();
}

describe('WorkdirField', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
  });

  it('n’ouvre la saisie qu’au clic', async () => {
    const utilisateur = monter(vi.fn());
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument();

    await utilisateur.click(screen.getByRole('button', { name: 'Changer' }));

    expect(screen.getByRole('textbox')).toHaveValue(REPERTOIRE);
  });

  it('envoie le chemin saisi et referme', async () => {
    const changer = vi.fn().mockResolvedValue('D:\\autre\\projet');
    const utilisateur = monter(changer);

    await utilisateur.click(screen.getByRole('button', { name: 'Changer' }));
    await utilisateur.clear(screen.getByRole('textbox'));
    await utilisateur.type(screen.getByRole('textbox'), 'D:\\autre');
    await utilisateur.click(screen.getByRole('button', { name: 'Appliquer' }));

    expect(changer).toHaveBeenCalledWith('D:\\autre');
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
  });

  it('affiche le refus du serveur et garde la saisie', async () => {
    const changer = vi
      .fn()
      .mockRejectedValue(new ErreurAPI(422, 'D:\\absent est introuvable'));
    const utilisateur = monter(changer);

    await utilisateur.click(screen.getByRole('button', { name: 'Changer' }));
    await utilisateur.click(screen.getByRole('button', { name: 'Appliquer' }));

    expect(await screen.findByRole('alert')).toHaveTextContent('D:\\absent est introuvable');
    // La saisie reste ouverte : on corrige le chemin, on ne le retape pas.
    expect(screen.getByRole('textbox')).toBeInTheDocument();
  });

  it('referme sur Échap sans rien changer', async () => {
    const changer = vi.fn();
    const utilisateur = monter(changer);

    await utilisateur.click(screen.getByRole('button', { name: 'Changer' }));
    await utilisateur.keyboard('{Escape}');

    expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
    expect(changer).not.toHaveBeenCalled();
  });
});
