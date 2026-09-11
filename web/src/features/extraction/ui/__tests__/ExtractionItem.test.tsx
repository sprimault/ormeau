// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import type { Extraction } from '@/shared/model';
import { retirerExtraction } from '../../api/extractionApi';
import { ExtractionItem } from '../ExtractionItem';

vi.mock('../../api/extractionApi', () => ({ retirerExtraction: vi.fn() }));

const debut = '2026-09-11T10:00:00Z';

const gescom: Extraction = {
  id: 'tache-1',
  base: 'gescom',
  fichier: 'gescom.calque.json',
  nb_tables: 12,
  etat: 'en_cours',
  debut,
};

/** Monte une ligne, une minute et cinq secondes après le début par défaut. */
function rendre(extraction: Extraction, maintenant = Date.parse(debut) + 65_000) {
  return render(
    <ul>
      <ExtractionItem extraction={extraction} maintenant={maintenant} />
    </ul>,
  );
}

describe('ExtractionItem', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
    vi.mocked(retirerExtraction).mockReset().mockResolvedValue(undefined);
  });

  it('montre l’étape en cours, les paliers franchis et la durée', () => {
    rendre({ ...gescom, avancement: { etape: 'colonnes', rang: 3, total: 8 } });

    expect(screen.getByText('Lecture : colonnes, étape 3 sur 8')).toBeInTheDocument();
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '2');
    expect(screen.getByText('1 min 05 s')).toBeInTheDocument();
    expect(screen.getByText('12 table(s)')).toBeInTheDocument();
  });

  it('annule une extraction en cours', async () => {
    rendre(gescom);

    await userEvent.click(screen.getByRole('button', { name: 'Annuler' }));

    expect(retirerExtraction).toHaveBeenCalledWith('tache-1');
  });

  it('dit pourquoi une tâche attend', () => {
    rendre({ id: 'tache-2', base: 'paie', fichier: 'paie.calque.json', nb_tables: 0, etat: 'en_attente' });

    expect(
      screen.getByText('En attente d’une place : deux extractions tournent déjà.'),
    ).toBeInTheDocument();
    expect(screen.getByText('toutes les tables')).toBeInTheDocument();
  });

  it('résume le calque écrit, déplie ses anomalies et se retire', async () => {
    rendre({
      ...gescom,
      etat: 'terminee',
      fin: '2026-09-11T10:00:42Z',
      resultat: {
        tables: 12,
        colonnes: 80,
        empreinte: 'sha256:0',
        anomalies: [
          {
            code: 'table_cible_introuvable',
            cible: 'public.commandes',
            message: 'clé vers public.clients, absente du calque',
          },
        ],
      },
    });

    expect(screen.getByText('gescom.calque.json')).toBeInTheDocument();
    expect(screen.getByText('12 table(s), 80 colonne(s)')).toBeInTheDocument();
    expect(screen.getByText('42 s')).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: /Anomalies du calque : 1/ }));
    expect(screen.getByText('Table référencée hors du calque')).toBeInTheDocument();
    expect(screen.getByText('public.commandes')).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'Retirer' }));
    expect(retirerExtraction).toHaveBeenCalledWith('tache-1');
  });

  it('estompe un calque réécrit depuis par une autre extraction, et le dit', () => {
    const { container } = render(
      <ul>
        <ExtractionItem
          extraction={{
            ...gescom,
            etat: 'terminee',
            fin: '2026-09-11T10:00:00.400Z',
            resultat: { tables: 1, colonnes: 10, empreinte: 'sha256:0', anomalies: [] },
          }}
          maintenant={0}
          remplaceeA="2026-09-11T10:05:00.000Z"
        />
      </ul>,
    );

    expect(screen.getByText(/^Fichier réécrit depuis/)).toBeInTheDocument();
    expect(screen.getByText('< 1 s')).toBeInTheDocument();
    expect(container.querySelector('li')).toHaveClass('opacity-60');
  });

  it('affiche tel quel un code d’étape qui n’a pas de libellé', () => {
    rendre({ ...gescom, avancement: { etape: 'partitions', rang: 1, total: 9 } });

    expect(screen.getByText('Lecture : partitions, étape 1 sur 9')).toBeInTheDocument();
  });

  it('rend le message du serveur quand l’extraction échoue', () => {
    rendre({
      ...gescom,
      etat: 'echouee',
      fin: '2026-09-11T10:00:03Z',
      erreur: 'lecture des colonnes: délai dépassé',
    });

    expect(screen.getByText('lecture des colonnes: délai dépassé')).toBeInTheDocument();
    expect(screen.getByText(/^échouée à \d{2}:\d{2}:\d{2}$/)).toBeInTheDocument();
  });
});
