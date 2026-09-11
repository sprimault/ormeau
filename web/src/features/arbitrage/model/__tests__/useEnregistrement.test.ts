// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ErreurAPI } from '@/shared/api';
import { useLangStore } from '@/shared/i18n';
import { CodeContenuManuel, CodeDecisionsModifiees, type ReponseDecisions } from '@/shared/model';
import { ecrireDecisions } from '../../api/arbitrageApi';
import { useEnregistrement } from '../useEnregistrement';

vi.mock('../../api/arbitrageApi', () => ({ ecrireDecisions: vi.fn() }));

const decisions = { tables_ignorees: ['public.migrations'] };

/** Brouillon parti d'un fichier lu avec l'empreinte sha256:f1. */
function brouillon(manuel = false) {
  const fichier: ReponseDecisions = {
    existe: true,
    decisions: {},
    empreinte_fichier: 'sha256:f1',
    manuel,
  };
  return { base: 'gescom', decisions, fichier, enregistre: vi.fn() };
}

describe('useEnregistrement', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
    vi.mocked(ecrireDecisions).mockReset();
  });

  it('écrit le brouillon avec les deux empreintes, puis retient le fichier écrit', async () => {
    vi.mocked(ecrireDecisions).mockResolvedValue({ empreinte_fichier: 'sha256:f2' });
    const b = brouillon();
    const { result } = renderHook(() => useEnregistrement(b, 'sha256:c'));

    act(() => result.current.enregistrer());

    await waitFor(() => expect(b.enregistre).toHaveBeenCalledWith(decisions, 'sha256:f2'));
    expect(ecrireDecisions).toHaveBeenCalledWith({
      base: 'gescom',
      decisions,
      empreinte_physique: 'sha256:c',
      empreinte_fichier: 'sha256:f1',
    });
    await waitFor(() => expect(result.current.enCours).toBe(false));
  });

  it('demande confirmation avant d’écraser un fichier écrit à la main', async () => {
    vi.mocked(ecrireDecisions).mockResolvedValue({ empreinte_fichier: 'sha256:f2' });
    const b = brouillon(true);
    const { result } = renderHook(() => useEnregistrement(b, 'sha256:c'));

    act(() => result.current.enregistrer());
    expect(result.current.refus).toBe(CodeContenuManuel);
    expect(ecrireDecisions).not.toHaveBeenCalled();

    act(() => result.current.enregistrer(true));
    await waitFor(() => expect(b.enregistre).toHaveBeenCalled());
    expect(vi.mocked(ecrireDecisions).mock.calls[0][0].ecraser_manuel).toBe(true);
    expect(result.current.refus).toBeNull();
  });

  it('retient le refus du serveur jusqu’à ce que l’écran y réponde', async () => {
    vi.mocked(ecrireDecisions).mockRejectedValue(
      new ErreurAPI(409, 'le fichier a changé', CodeDecisionsModifiees),
    );
    const { result } = renderHook(() => useEnregistrement(brouillon(), 'sha256:c'));

    act(() => result.current.enregistrer());

    await waitFor(() => expect(result.current.refus).toBe(CodeDecisionsModifiees));
    expect(result.current.erreur).toBeNull();
    act(() => result.current.abandonner());
    expect(result.current.refus).toBeNull();
  });

  it('rend le message d’un échec ordinaire', async () => {
    vi.mocked(ecrireDecisions).mockRejectedValue(new ErreurAPI(500, 'disque plein'));
    const { result } = renderHook(() => useEnregistrement(brouillon(), 'sha256:c'));

    act(() => result.current.enregistrer());

    await waitFor(() => expect(result.current.erreur).toBe('disque plein'));
    expect(result.current.refus).toBeNull();
  });
});
