// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ErreurAPI } from '@/shared/api';
import { useLangStore } from '@/shared/i18n';
import type { ColonneSommaire } from '@/shared/model';
import { lireColonnes } from '../../api/inventoryApi';
import { useColumns } from '../useColumns';

vi.mock('../../api/inventoryApi', () => ({ lireColonnes: vi.fn() }));

const colonnes: ColonneSommaire[] = [
  { nom: 'id', position: 1, type_brut: 'integer', nullable: false, cle_primaire: true },
];

describe('useColumns', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
    vi.mocked(lireColonnes).mockReset();
  });

  it('ne demande rien tant qu’aucune table n’est dépliée', () => {
    renderHook(() => useColumns('session-1'));
    expect(lireColonnes).not.toHaveBeenCalled();
  });

  it('charge les colonnes de la table dépliée', async () => {
    vi.mocked(lireColonnes).mockResolvedValue({ colonnes });
    const { result } = renderHook(() => useColumns('session-1'));

    act(() => result.current.charger('public.clients', 'public', 'clients'));

    await waitFor(() => expect(result.current.etat('public.clients').colonnes).toEqual(colonnes));
    expect(lireColonnes).toHaveBeenCalledWith('session-1', 'public', 'clients');
  });

  it('garde ce qui a été chargé : replier puis rouvrir ne redemande rien', async () => {
    vi.mocked(lireColonnes).mockResolvedValue({ colonnes });
    const { result } = renderHook(() => useColumns('session-1'));

    act(() => result.current.charger('public.clients', 'public', 'clients'));
    await waitFor(() => expect(result.current.etat('public.clients').colonnes).toBeDefined());

    act(() => result.current.charger('public.clients', 'public', 'clients'));
    await waitFor(() => expect(result.current.etat('public.clients').enCours).toBe(false));

    expect(lireColonnes).toHaveBeenCalledTimes(2);
  });

  it('ne relance pas une demande déjà en vol', async () => {
    // Un double clic sur le chevron ne doit pas doubler la requête.
    vi.mocked(lireColonnes).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useColumns('session-1'));

    act(() => {
      result.current.charger('public.clients', 'public', 'clients');
      result.current.charger('public.clients', 'public', 'clients');
    });

    expect(lireColonnes).toHaveBeenCalledTimes(1);
  });

  it('sépare les tables : deux clés, deux entrées', async () => {
    vi.mocked(lireColonnes).mockResolvedValue({ colonnes });
    const { result } = renderHook(() => useColumns('session-1'));

    act(() => result.current.charger('public.clients', 'public', 'clients'));
    await waitFor(() => expect(result.current.etat('public.clients').colonnes).toBeDefined());

    expect(result.current.etat('public.commandes').colonnes).toBeUndefined();
  });

  it('rend l’échec du serveur pour la table concernée', async () => {
    vi.mocked(lireColonnes).mockRejectedValue(new ErreurAPI(502, 'table disparue'));
    const { result } = renderHook(() => useColumns('session-1'));

    act(() => result.current.charger('public.clients', 'public', 'clients'));

    await waitFor(() =>
      expect(result.current.etat('public.clients').erreur).toBe('table disparue'),
    );
    expect(result.current.etat('public.clients').enCours).toBe(false);
  });

  it('annonce la lecture en cours', () => {
    vi.mocked(lireColonnes).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useColumns('session-1'));

    act(() => result.current.charger('public.clients', 'public', 'clients'));

    expect(result.current.etat('public.clients').enCours).toBe(true);
  });
});
