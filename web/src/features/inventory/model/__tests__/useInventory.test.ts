// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ErreurAPI } from '@/shared/api';
import { useLangStore } from '@/shared/i18n';
import type { TableSommaire } from '@/shared/model';
import { lireInventaire } from '../../api/inventoryApi';
import { useInventory } from '../useInventory';

vi.mock('../../api/inventoryApi', () => ({ lireInventaire: vi.fn() }));

/** Inventaire d'une seule table, ce que le hook doit rendre tel quel. */
const tables: TableSommaire[] = [
  { schema: 'public', nom: 'clients', nb_colonnes: 4, lignes_estimees: 12, cle_primaire: true },
];

describe('useInventory', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
    vi.mocked(lireInventaire).mockReset();
  });

  it('charge l’inventaire de la session', async () => {
    vi.mocked(lireInventaire).mockResolvedValue({ tables });
    const { result } = renderHook(() => useInventory('session-1', ['public']));

    await waitFor(() => expect(result.current.enCours).toBe(false));
    expect(result.current.tables).toEqual(tables);
    expect(lireInventaire).toHaveBeenCalledWith('session-1', ['public'], expect.anything());
  });

  it('affiche le message du serveur', async () => {
    vi.mocked(lireInventaire).mockRejectedValue(new ErreurAPI(502, 'catalogue illisible'));
    const { result } = renderHook(() => useInventory('session-1', ['public']));

    await waitFor(() => expect(result.current.erreur).toBe('catalogue illisible'));
    expect(result.current.enCours).toBe(false);
  });

  it('ne relance rien quand les schémas sont reconstruits à l’identique', async () => {
    vi.mocked(lireInventaire).mockResolvedValue({ tables });
    const { rerender } = renderHook(({ schemas }) => useInventory('session-1', schemas), {
      initialProps: { schemas: ['public'] },
    });

    await waitFor(() => expect(lireInventaire).toHaveBeenCalledTimes(1));
    // Même contenu, tableau différent : le hook compare ce qu'il y a dedans.
    rerender({ schemas: ['public'] });

    expect(lireInventaire).toHaveBeenCalledTimes(1);
  });

  it('recharge quand la session change, et abandonne la précédente', async () => {
    vi.mocked(lireInventaire).mockResolvedValue({ tables });
    const { rerender } = renderHook(({ session }) => useInventory(session, ['public']), {
      initialProps: { session: 'session-1' },
    });

    await waitFor(() => expect(lireInventaire).toHaveBeenCalledTimes(1));
    const premier = vi.mocked(lireInventaire).mock.calls[0][2] as AbortSignal;

    rerender({ session: 'session-2' });

    // Sur une base lente, une réponse arrivée après coup écraserait l'arbre de
    // la connexion suivante.
    expect(premier.aborted).toBe(true);
    await waitFor(() => expect(lireInventaire).toHaveBeenCalledTimes(2));
  });

  it('ignore l’échec d’un appel abandonné', async () => {
    vi.mocked(lireInventaire).mockImplementation((_session, _schemas, signal) =>
      Promise.reject(Object.assign(new ErreurAPI(0, 'abandon'), { signal })).catch(
        (echec: unknown) => {
          throw echec;
        },
      ),
    );

    const { result, unmount } = renderHook(() => useInventory('session-1', ['public']));
    unmount();

    expect(result.current.erreur).toBeNull();
  });
});
