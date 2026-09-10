// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ErreurAPI } from '@/shared/api';
import { useLangStore } from '@/shared/i18n';
import type { ReponseConnexion } from '@/shared/model';
import { changerDeBase, connecter, deconnecter, lireBases } from '../../api/connectionApi';
import { useConnection } from '../useConnection';

vi.mock('../../api/connectionApi', () => ({
  connecter: vi.fn(),
  deconnecter: vi.fn(),
  changerDeBase: vi.fn(),
  lireBases: vi.fn(),
}));

const gescom: ReponseConnexion = {
  session: 'session-1',
  sgbd: 'postgres',
  version: '16.14',
  catalogue: 'gescom',
  schemas: ['public'],
};

const paie: ReponseConnexion = { ...gescom, session: 'session-2', catalogue: 'paie' };

describe('useConnection', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
    vi.mocked(connecter).mockReset();
    vi.mocked(deconnecter).mockReset().mockResolvedValue(undefined);
    vi.mocked(changerDeBase).mockReset();
    vi.mocked(lireBases).mockReset().mockResolvedValue({ bases: [] });
  });

  it('garde le serveur atteint', async () => {
    vi.mocked(connecter).mockResolvedValue(gescom);
    const { result } = renderHook(() => useConnection());

    await act(async () => result.current.ouvrir({ hote: 'bdd' }));

    expect(result.current.serveur).toEqual(gescom);
    expect(result.current.erreur).toBeNull();
  });

  it('affiche le message du serveur quand la connexion échoue', async () => {
    vi.mocked(connecter).mockRejectedValue(new ErreurAPI(502, 'mot de passe refusé'));
    const { result } = renderHook(() => useConnection());

    await act(async () => result.current.ouvrir({ hote: 'bdd' }));

    expect(result.current.erreur).toBe('mot de passe refusé');
    expect(result.current.serveur).toBeNull();
  });

  it('ne conserve rien de ce qui a été saisi', async () => {
    vi.mocked(connecter).mockResolvedValue(gescom);
    const { result } = renderHook(() => useConnection());

    await act(async () => result.current.ouvrir({ hote: 'bdd', mot_de_passe: 'tres-secret' }));

    // Le mot de passe part et est oublié : seule la réponse du serveur reste, et
    // elle n'en porte rien.
    expect(JSON.stringify(result.current.serveur)).not.toContain('tres-secret');
  });

  it('charge les bases du serveur après la connexion', async () => {
    vi.mocked(connecter).mockResolvedValue(gescom);
    vi.mocked(lireBases).mockResolvedValue({ bases: ['gescom', 'paie'] });
    const { result } = renderHook(() => useConnection());

    await act(async () => result.current.ouvrir({ hote: 'bdd' }));

    await waitFor(() => expect(result.current.bases).toEqual(['gescom', 'paie']));
  });

  it('se passe du sélecteur quand le dialecte ne sait pas énumérer', async () => {
    vi.mocked(connecter).mockResolvedValue(gescom);
    vi.mocked(lireBases).mockRejectedValue(new ErreurAPI(501, 'non implémenté'));
    const { result } = renderHook(() => useConnection());

    await act(async () => result.current.ouvrir({ hote: 'bdd' }));

    await waitFor(() => expect(result.current.bases).toEqual([]));
    expect(result.current.serveur).toEqual(gescom);
  });

  it('bascule sur une autre base sans ressaisie', async () => {
    vi.mocked(connecter).mockResolvedValue(gescom);
    vi.mocked(changerDeBase).mockResolvedValue(paie);
    const { result } = renderHook(() => useConnection());

    await act(async () => result.current.ouvrir({ hote: 'bdd' }));
    await act(async () => result.current.changerBase('paie'));

    expect(changerDeBase).toHaveBeenCalledWith('session-1', 'paie');
    expect(result.current.serveur?.catalogue).toBe('paie');
  });

  it('garde la base courante quand la bascule échoue', async () => {
    vi.mocked(connecter).mockResolvedValue(gescom);
    vi.mocked(changerDeBase).mockRejectedValue(new ErreurAPI(502, 'base inaccessible'));
    const { result } = renderHook(() => useConnection());

    await act(async () => result.current.ouvrir({ hote: 'bdd' }));
    await act(async () => result.current.changerBase('paie'));

    expect(result.current.serveur?.catalogue).toBe('gescom');
    expect(result.current.erreur).toBe('base inaccessible');
  });

  it('referme et revient au formulaire', async () => {
    vi.mocked(connecter).mockResolvedValue(gescom);
    const { result } = renderHook(() => useConnection());

    await act(async () => result.current.ouvrir({ hote: 'bdd' }));
    await act(async () => result.current.fermer());

    expect(deconnecter).toHaveBeenCalledWith('session-1');
    expect(result.current.serveur).toBeNull();
  });

  it('revient au formulaire même si la fermeture échoue', async () => {
    vi.mocked(connecter).mockResolvedValue(gescom);
    vi.mocked(deconnecter).mockRejectedValue(new ErreurAPI(0, 'serveur mort'));
    const { result } = renderHook(() => useConnection());

    await act(async () => result.current.ouvrir({ hote: 'bdd' }));
    await act(async () => result.current.fermer());

    expect(result.current.serveur).toBeNull();
  });
});
