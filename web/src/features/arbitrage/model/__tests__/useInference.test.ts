// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ErreurAPI } from '@/shared/api';
import { useLangStore } from '@/shared/i18n';
import { CodeCalqueModifie, type Decisions, type ReponseInference } from '@/shared/model';
import { inferer } from '../../api/arbitrageApi';
import { useInference } from '../useInference';

vi.mock('../../api/arbitrageApi', () => ({ inferer: vi.fn() }));

/** Réponse d'inférence vide, pour un calque d'empreinte donnée. */
function reponse(empreinte = 'sha256:1'): ReponseInference {
  return {
    empreinte_physique: empreinte,
    avertissements: [],
    propositions: [],
    enumerations: [],
    entites: [],
    types_doctrine: [],
  };
}

// Une référence stable : un objet recréé à chaque rendu relancerait l'inférence
// sans fin, ce que le contexte du brouillon ne fait pas.
const aucune: Decisions = {};

describe('useInference', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
    vi.mocked(inferer).mockReset();
  });

  it('apprend l’empreinte du calque, puis la renvoie avec le brouillon', async () => {
    vi.mocked(inferer).mockResolvedValue(reponse());
    const { result, rerender } = renderHook(
      ({ decisions }) => useInference('gescom', decisions, ''),
      { initialProps: { decisions: aucune } },
    );
    await waitFor(() => expect(result.current.resultat).not.toBeNull());
    expect(vi.mocked(inferer).mock.calls[0][0]).toEqual({ base: 'gescom', decisions: {} });

    const renommees: Decisions = { renommages: { 'public.clients': 'Client' } };
    rerender({ decisions: renommees });

    await waitFor(() => expect(result.current.decisionsJugees).toEqual(renommees));
    expect(vi.mocked(inferer).mock.calls[1][0]).toEqual({
      base: 'gescom',
      decisions: renommees,
      empreinte_physique: 'sha256:1',
    });
  });

  it('signale un calque réécrit dès la fin d’une extraction, sans perdre ce qui est affiché', async () => {
    vi.mocked(inferer)
      .mockResolvedValueOnce(reponse())
      .mockRejectedValueOnce(new ErreurAPI(409, 'le calque a changé', CodeCalqueModifie));
    const { result, rerender } = renderHook(
      ({ version }) => useInference('gescom', aucune, version),
      { initialProps: { version: '' } },
    );
    await waitFor(() => expect(result.current.resultat).not.toBeNull());

    rerender({ version: '2026-09-11T10:00:00Z' });

    await waitFor(() => expect(result.current.calqueModifie).toBe(true));
    expect(result.current.resultat?.empreinte_physique).toBe('sha256:1');
    expect(result.current.erreur).toBeNull();
  });

  it('recharge sans empreinte, et repart du nouveau calque', async () => {
    vi.mocked(inferer)
      .mockResolvedValueOnce(reponse('sha256:1'))
      .mockRejectedValueOnce(new ErreurAPI(409, 'le calque a changé', CodeCalqueModifie))
      .mockResolvedValueOnce(reponse('sha256:2'));
    const { result, rerender } = renderHook(
      ({ version }) => useInference('gescom', aucune, version),
      { initialProps: { version: '' } },
    );
    await waitFor(() => expect(result.current.resultat).not.toBeNull());
    rerender({ version: '2026-09-11T10:00:00Z' });
    await waitFor(() => expect(result.current.calqueModifie).toBe(true));

    act(() => result.current.recharger());

    await waitFor(() => expect(result.current.resultat?.empreinte_physique).toBe('sha256:2'));
    expect(vi.mocked(inferer).mock.calls[2][0].empreinte_physique).toBeUndefined();
    expect(result.current.calqueModifie).toBe(false);
  });

  it('rend le message du serveur quand la base n’a pas de calque', async () => {
    vi.mocked(inferer).mockRejectedValue(
      new ErreurAPI(404, 'aucun gescom.calque.json dans le répertoire de travail'),
    );
    const { result } = renderHook(() => useInference('gescom', aucune, ''));

    await waitFor(() =>
      expect(result.current.erreur).toBe('aucun gescom.calque.json dans le répertoire de travail'),
    );
    expect(result.current.resultat).toBeNull();
  });
});
