// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ErreurAPI } from '@/shared/api';
import { useLangStore } from '@/shared/i18n';
import type { Extraction, ReponseCalque } from '@/shared/model';
import { lireCalque } from '../../api/extractionApi';
import { ContexteExtractions } from '../../model/contexte';
import { CalquePreview } from '../CalquePreview';

vi.mock('../../api/extractionApi', () => ({ lireCalque: vi.fn() }));

/** Dernier calque produit, tel que l'aperçu le reçoit. */
const calque: ReponseCalque = {
  fichier: 'gescom.calque.json',
  extrait_le: '2026-09-11T10:00:00Z',
  empreinte: 'sha256:0',
  contenu: '{"version_ri":1}',
};

/** L'aperçu, avec les extractions que le flux aurait rapportées. */
function apercu(extractions: Extraction[] = []) {
  return (
    <ContexteExtractions.Provider value={{ extractions, connecte: true }}>
      <CalquePreview session="session-1" base="gescom" />
    </ContexteExtractions.Provider>
  );
}

/** Extraction terminée d'une base à un instant donné. */
function terminee(base: string, fin: string): Extraction {
  return { id: fin, base, fichier: `${base}.calque.json`, nb_tables: 0, etat: 'terminee', fin };
}

describe('CalquePreview', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
    vi.mocked(lireCalque).mockReset();
  });

  it('montre le fichier écrit, tel quel, et quand il l’a été', async () => {
    vi.mocked(lireCalque).mockResolvedValue(calque);
    render(apercu());

    expect(await screen.findByText('{"version_ri":1}')).toBeInTheDocument();
    expect(screen.getByText('gescom.calque.json')).toBeInTheDocument();
    expect(screen.getByText(/^extrait à \d{2}:\d{2}:\d{2}$/)).toBeInTheDocument();
  });

  it('dit que les statistiques ont été retirées', async () => {
    vi.mocked(lireCalque).mockResolvedValue({ ...calque, statistiques_retirees: true });
    render(apercu());

    expect(await screen.findByText(/^Statistiques retirées/)).toBeInTheDocument();
  });

  it('dit quel fichier manque', async () => {
    vi.mocked(lireCalque).mockRejectedValue(
      new ErreurAPI(404, 'aucun gescom.calque.json dans le répertoire de travail'),
    );
    render(apercu());

    expect(
      await screen.findByText('aucun gescom.calque.json dans le répertoire de travail'),
    ).toBeInTheDocument();
  });

  it('se relit quand une extraction de sa base se termine, pas celle d’une autre', async () => {
    vi.mocked(lireCalque).mockResolvedValue(calque);
    const { rerender } = render(apercu());
    await waitFor(() => expect(lireCalque).toHaveBeenCalledTimes(1));

    rerender(apercu([terminee('paie', '2026-09-11T10:01:00.000Z')]));
    expect(lireCalque).toHaveBeenCalledTimes(1);

    rerender(apercu([terminee('gescom', '2026-09-11T10:02:00.000Z')]));
    await waitFor(() => expect(lireCalque).toHaveBeenCalledTimes(2));
  });
});
