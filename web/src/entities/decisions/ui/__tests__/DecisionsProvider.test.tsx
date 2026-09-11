// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { act, render, renderHook, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { ErreurAPI } from '@/shared/api';
import { useLangStore } from '@/shared/i18n';
import type { ReponseDecisions } from '@/shared/model';
import { lireDecisions } from '../../api/decisionsApi';
import { useDecisions } from '../../model/contexte';
import { DecisionsProvider } from '../DecisionsProvider';

vi.mock('../../api/decisionsApi', () => ({ lireDecisions: vi.fn() }));

/** Affiche les tables ignorées du brouillon, pour voir ce que le fournisseur transmet. */
function Contenu() {
  const { decisions } = useDecisions();
  return <p>tables : {(decisions.tables_ignorees ?? []).join(', ')}</p>;
}

describe('DecisionsProvider', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
    vi.mocked(lireDecisions).mockReset();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('attend le fichier avant de monter ce qui écrit dans le brouillon', async () => {
    let livrer: (reponse: ReponseDecisions) => void = () => {};
    vi.mocked(lireDecisions).mockReturnValue(
      new Promise((resolve) => {
        livrer = resolve;
      }),
    );

    render(
      <DecisionsProvider base="gescom">
        <Contenu />
      </DecisionsProvider>,
    );
    expect(screen.queryByText(/tables :/)).not.toBeInTheDocument();

    await act(async () => {
      livrer({ existe: true, decisions: { tables_ignorees: ['public.migrations'] }, manuel: false });
    });
    expect(screen.getByText('tables : public.migrations')).toBeInTheDocument();
  });

  it('monte quand même l’écran quand le fichier est illisible', async () => {
    vi.mocked(lireDecisions).mockRejectedValue(new ErreurAPI(422, 'gescom.decisions.yaml illisible'));

    render(
      <DecisionsProvider base="gescom">
        <Contenu />
      </DecisionsProvider>,
    );

    expect(await screen.findByText(/tables :/)).toBeInTheDocument();
  });

  it('refuse d’être lu hors du fournisseur', () => {
    // React journalise l'erreur de rendu avant de la relancer.
    vi.spyOn(console, 'error').mockImplementation(() => {});

    expect(() => renderHook(() => useDecisions())).toThrow('hors de DecisionsProvider');
  });
});
