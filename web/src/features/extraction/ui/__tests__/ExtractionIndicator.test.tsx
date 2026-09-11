// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import type { Extraction } from '@/shared/model';
import { ContexteExtractions } from '../../model/contexte';
import { ExtractionIndicator } from '../ExtractionIndicator';

vi.mock('../../api/extractionApi', () => ({ retirerExtraction: vi.fn() }));

/** Tâche d'une base dans un état donné. */
function tache(base: string, etat: string): Extraction {
  return { id: base, base, fichier: `${base}.calque.json`, nb_tables: 0, etat };
}

/** Monte l'indicateur avec les tâches que le flux aurait rapportées. */
function rendre(extractions: Extraction[], connecte = true) {
  return render(
    <ContexteExtractions.Provider value={{ extractions, connecte }}>
      <ExtractionIndicator />
    </ContexteExtractions.Provider>,
  );
}

describe('ExtractionIndicator', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
    document.title = 'Ormeau';
  });

  it('n’annonce rien tant qu’aucune extraction n’a été lancée', () => {
    const { container } = rendre([]);
    expect(container).toBeEmptyDOMElement();
  });

  it('compte celles qui tournent, et le reprend dans le titre de l’onglet', () => {
    rendre([tache('gescom', 'en_cours'), tache('paie', 'en_attente'), tache('stock', 'terminee')]);

    expect(screen.getByRole('button', { name: '2 extraction(s) en cours' })).toBeInTheDocument();
    expect(document.title).toBe('(2) Ormeau');
  });

  it('signale un échec une fois que plus rien ne tourne', () => {
    rendre([tache('gescom', 'echouee'), tache('paie', 'terminee')]);

    expect(screen.getByRole('button', { name: '1 extraction(s) en échec' })).toBeInTheDocument();
    expect(document.title).toBe('Ormeau');
  });

  it('ouvre le détail au clic et le referme avec Échap', async () => {
    rendre([tache('gescom', 'terminee')]);

    await userEvent.click(screen.getByRole('button', { name: '1 extraction(s) finie(s)' }));
    expect(screen.getByRole('dialog', { name: 'Extractions' })).toBeInTheDocument();

    await userEvent.keyboard('{Escape}');
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('dit dans le détail que le suivi est coupé', async () => {
    rendre([tache('gescom', 'en_cours')], false);

    await userEvent.click(screen.getByRole('button', { name: '1 extraction(s) en cours' }));

    expect(screen.getByText('Suivi interrompu, reconnexion…')).toBeInTheDocument();
  });
});
