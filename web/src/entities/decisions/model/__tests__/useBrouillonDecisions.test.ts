// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ErreurAPI } from '@/shared/api';
import { useLangStore } from '@/shared/i18n';
import type { ReponseDecisions } from '@/shared/model';
import { lireDecisions } from '../../api/decisionsApi';
import { useBrouillonDecisions } from '../useBrouillonDecisions';

vi.mock('../../api/decisionsApi', () => ({ lireDecisions: vi.fn() }));

const fichier: ReponseDecisions = {
  existe: true,
  decisions: { colonnes_ignorees: { 'public.clients': ['photo'] } },
  empreinte_fichier: 'sha256:1',
  manuel: false,
};

describe('useBrouillonDecisions', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
    vi.mocked(lireDecisions).mockReset();
  });

  it('part du fichier de la base quand elle a déjà été arbitrée', async () => {
    vi.mocked(lireDecisions).mockResolvedValue(fichier);
    const { result } = renderHook(() => useBrouillonDecisions('gescom'));

    await waitFor(() => expect(result.current.enCours).toBe(false));
    expect(result.current.decisions).toEqual(fichier.decisions);
    expect(result.current.fichier).toEqual(fichier);
    expect(lireDecisions).toHaveBeenCalledWith('gescom', expect.anything());
  });

  it('part d’un brouillon vide quand la base n’a jamais été arbitrée', async () => {
    vi.mocked(lireDecisions).mockResolvedValue({ existe: false, decisions: {}, manuel: false });
    const { result } = renderHook(() => useBrouillonDecisions('gescom'));

    await waitFor(() => expect(result.current.enCours).toBe(false));
    expect(result.current.decisions).toEqual({});
    expect(result.current.erreur).toBeNull();
  });

  it('rend le message du serveur quand le fichier est illisible', async () => {
    vi.mocked(lireDecisions).mockRejectedValue(
      new ErreurAPI(422, 'gescom.decisions.yaml illisible : ligne 3'),
    );
    const { result } = renderHook(() => useBrouillonDecisions('gescom'));

    await waitFor(() => expect(result.current.erreur).toBe('gescom.decisions.yaml illisible : ligne 3'));
    expect(result.current.fichier).toBeNull();
  });

  it('modifie le brouillon sans toucher au fichier lu', async () => {
    vi.mocked(lireDecisions).mockResolvedValue(fichier);
    const { result } = renderHook(() => useBrouillonDecisions('gescom'));
    await waitFor(() => expect(result.current.enCours).toBe(false));

    act(() => result.current.modifier((d) => ({ ...d, tables_ignorees: ['public.migrations'] })));

    expect(result.current.decisions.tables_ignorees).toEqual(['public.migrations']);
    expect(result.current.fichier?.decisions.tables_ignorees).toBeUndefined();
  });

  it('repart du fichier de l’autre base, et abandonne la lecture précédente', async () => {
    vi.mocked(lireDecisions).mockResolvedValue(fichier);
    const { result, rerender } = renderHook(({ base }) => useBrouillonDecisions(base), {
      initialProps: { base: 'gescom' },
    });
    await waitFor(() => expect(lireDecisions).toHaveBeenCalledTimes(1));
    const premiere = vi.mocked(lireDecisions).mock.calls[0][1] as AbortSignal;

    vi.mocked(lireDecisions).mockResolvedValue({ existe: false, decisions: {}, manuel: false });
    rerender({ base: 'paie' });

    expect(premiere.aborted).toBe(true);
    await waitFor(() => expect(result.current.enCours).toBe(false));
    expect(lireDecisions).toHaveBeenLastCalledWith('paie', expect.anything());
    expect(result.current.decisions).toEqual({});
  });

  it('se sait modifié jusqu’à l’enregistrement', async () => {
    vi.mocked(lireDecisions).mockResolvedValue(fichier);
    const { result } = renderHook(() => useBrouillonDecisions('gescom'));
    await waitFor(() => expect(result.current.pret).toBe(true));
    expect(result.current.modifie).toBe(false);

    const suivantes = { ...fichier.decisions, tables_ignorees: ['public.migrations'] };
    act(() => result.current.modifier(() => suivantes));
    expect(result.current.modifie).toBe(true);

    act(() => result.current.enregistre(suivantes, 'sha256:2'));
    expect(result.current.modifie).toBe(false);
    expect(result.current.fichier).toEqual({
      existe: true,
      decisions: suivantes,
      empreinte_fichier: 'sha256:2',
      manuel: false,
    });
  });

  it('relit le fichier et remplace le brouillon', async () => {
    vi.mocked(lireDecisions)
      .mockResolvedValueOnce(fichier)
      .mockResolvedValueOnce({
        ...fichier,
        decisions: { renommages: { 'public.clients': 'Client' } },
        empreinte_fichier: 'sha256:3',
      });
    const { result } = renderHook(() => useBrouillonDecisions('gescom'));
    await waitFor(() => expect(result.current.pret).toBe(true));
    act(() => result.current.modifier((d) => ({ ...d, espace_de_noms: 'App\\Gescom' })));

    act(() => result.current.relire());

    await waitFor(() => expect(result.current.fichier?.empreinte_fichier).toBe('sha256:3'));
    expect(result.current.decisions).toEqual({ renommages: { 'public.clients': 'Client' } });
  });
});
