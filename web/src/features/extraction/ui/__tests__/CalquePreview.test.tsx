// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ErreurAPI } from '@/shared/api';
import { useLangStore } from '@/shared/i18n';
import type { Extraction, ReponseCalque } from '@/shared/model';
import { lireCalque } from '../../api/extractionApi';
import { ContexteExtractions } from '../../model/contexte';
import { CalquePreview } from '../CalquePreview';

vi.mock('../../api/extractionApi', () => ({ lireCalque: vi.fn() }));

const calque: ReponseCalque = {
  fichier: 'gescom.calque.json',
  extrait_le: '2026-09-11T10:00:00Z',
  empreinte: 'sha256:0',
  contenu: '{"version_ri":1}',
};

/** Calque de deux tables, sérialisé comme le serveur le rend. */
const calqueDeuxTables: ReponseCalque = {
  ...calque,
  contenu: JSON.stringify(
    {
      version_ri: 1,
      tables: [
        { nom: 'clients', schema: 'public', colonnes: [{ nom: 'id', position: 1 }] },
        { nom: 'commandes', schema: 'public', colonnes: [{ nom: 'numero', position: 1 }] },
      ],
    },
    null,
    2,
  ),
};

/** L'aperçu, avec les extractions que le flux aurait rapportées. */
function apercu(extractions: Extraction[] = [], table?: { schema: string; nom: string }) {
  return (
    <ContexteExtractions.Provider value={{ extractions, connecte: true }}>
      <CalquePreview session="session-1" base="gescom" table={table} />
    </ContexteExtractions.Provider>
  );
}

/** Extraction terminée d'une base à un instant donné. */
function terminee(base: string, fin: string): Extraction {
  return { id: fin, base, fichier: `${base}.calque.json`, nb_tables: 0, etat: 'terminee', fin };
}

/** Texte affiché dans le bloc du calque. */
function texteAffiche(container: HTMLElement): string {
  return container.querySelector('pre')?.textContent ?? '';
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

  it('se resserre sur la table cliquée dans l’arbre', async () => {
    vi.mocked(lireCalque).mockResolvedValue(calqueDeuxTables);
    const { container } = render(apercu([], { schema: 'public', nom: 'commandes' }));

    await screen.findByRole('button', { name: 'public.commandes' });
    expect(texteAffiche(container)).toContain('"nom": "commandes"');
    expect(texteAffiche(container)).not.toContain('"nom": "clients"');
  });

  it('montre le fichier complet à la demande, et revient à la table au clic suivant', async () => {
    vi.mocked(lireCalque).mockResolvedValue(calqueDeuxTables);
    const { container, rerender } = render(apercu([], { schema: 'public', nom: 'commandes' }));

    await userEvent.click(await screen.findByRole('button', { name: 'Fichier complet' }));
    expect(texteAffiche(container)).toContain('"nom": "clients"');

    rerender(apercu([], { schema: 'public', nom: 'clients' }));
    expect(texteAffiche(container)).not.toContain('"nom": "commandes"');
  });

  it('dit qu’une table cliquée n’est pas dans le calque', async () => {
    vi.mocked(lireCalque).mockResolvedValue(calqueDeuxTables);
    render(apercu([], { schema: 'public', nom: 'factures' }));

    expect(
      await screen.findByText(/^public\.factures n’est pas dans ce calque/),
    ).toBeInTheDocument();
  });
});
