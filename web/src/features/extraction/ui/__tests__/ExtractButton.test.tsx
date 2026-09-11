// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ErreurAPI } from '@/shared/api';
import { useLangStore } from '@/shared/i18n';
import type { Extraction, Portee } from '@/shared/model';
import { lancerExtraction } from '../../api/extractionApi';
import { ContexteExtractions } from '../../model/contexte';
import { ExtractButton } from '../ExtractButton';

vi.mock('../../api/extractionApi', () => ({ lancerExtraction: vi.fn() }));

const portee: Portee = { schemas: ['public'], tables_incluses: ['public.clients'] };

/** Tâche d'une base dans un état donné. */
function tache(base: string, etat: string, id = base): Extraction {
  return { id, base, fichier: `${base}.calque.json`, nb_tables: 0, etat };
}

/** Monte le bouton avec les tâches que le flux aurait rapportées. */
function rendre(extractions: Extraction[] = []) {
  return render(
    <ContexteExtractions.Provider value={{ extractions, connecte: true }}>
      <ExtractButton session="session-1" base="gescom" portee={portee} />
    </ContexteExtractions.Provider>,
  );
}

describe('ExtractButton', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
    vi.mocked(lancerExtraction).mockReset();
  });

  it('lance l’extraction de la session avec la portée affichée', async () => {
    vi.mocked(lancerExtraction).mockResolvedValue(tache('gescom', 'en_cours'));
    rendre();

    await userEvent.click(screen.getByRole('button', { name: 'Extraire' }));

    expect(lancerExtraction).toHaveBeenCalledWith('session-1', portee);
  });

  it('annonce le fichier qui sera écrit', () => {
    rendre();
    expect(screen.getByText('gescom.calque.json')).toBeInTheDocument();
  });

  it('affiche le refus du serveur', async () => {
    vi.mocked(lancerExtraction).mockRejectedValue(
      new ErreurAPI(409, 'une extraction de gescom est déjà en cours'),
    );
    rendre();

    await userEvent.click(screen.getByRole('button', { name: 'Extraire' }));

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'une extraction de gescom est déjà en cours',
    );
  });

  it('se bloque tant qu’une extraction de la même base tourne', () => {
    rendre([tache('gescom', 'en_cours')]);
    expect(screen.getByRole('button', { name: 'Extraction en cours' })).toBeDisabled();
  });

  it('reste disponible pour une autre base, ou après une extraction finie', () => {
    rendre([tache('paie', 'en_cours'), tache('gescom', 'terminee', 'ancienne')]);
    expect(screen.getByRole('button', { name: 'Extraire' })).toBeEnabled();
  });
});
