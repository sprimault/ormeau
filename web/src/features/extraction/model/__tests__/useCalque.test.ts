// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ErreurAPI } from '@/shared/api';
import { useLangStore } from '@/shared/i18n';
import type { Extraction, ReponseCalque } from '@/shared/model';
import { lireCalque } from '../../api/extractionApi';
import { derniereFin, useCalque } from '../useCalque';

vi.mock('../../api/extractionApi', () => ({ lireCalque: vi.fn() }));

/** Calque rendu par le serveur ; seul son contenu minimal importe ici. */
const calque: ReponseCalque = {
  fichier: 'gescom.calque.json',
  extrait_le: '2026-09-11T10:00:00Z',
  empreinte: 'sha256:0',
  contenu: '{"version_ri":1}',
};

/** Extraction d'une base dans un état donné, finie à un instant donné. */
function extraction(base: string, etat: string, fin?: string): Extraction {
  return { id: `${base}-${fin}`, base, fichier: `${base}.calque.json`, nb_tables: 0, etat, fin };
}

describe('useCalque', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
    vi.mocked(lireCalque).mockReset();
  });

  it('lit le calque de la session', async () => {
    vi.mocked(lireCalque).mockResolvedValue(calque);
    const { result } = renderHook(() => useCalque('session-1', ''));

    await waitFor(() => expect(result.current.enCours).toBe(false));
    expect(result.current.calque).toEqual(calque);
    expect(lireCalque).toHaveBeenCalledWith('session-1', expect.anything());
  });

  it('rend le message du serveur quand le fichier manque', async () => {
    vi.mocked(lireCalque).mockRejectedValue(
      new ErreurAPI(404, 'aucun gescom.calque.json dans le répertoire de travail'),
    );
    const { result } = renderHook(() => useCalque('session-1', ''));

    await waitFor(() =>
      expect(result.current.message).toBe('aucun gescom.calque.json dans le répertoire de travail'),
    );
    expect(result.current.calque).toBeNull();
  });

  it('relit le calque à chaque version, sans l’effacer pendant la relecture', async () => {
    vi.mocked(lireCalque)
      .mockResolvedValueOnce(calque)
      .mockReturnValueOnce(new Promise<ReponseCalque>(() => undefined));
    const { result, rerender } = renderHook(({ version }) => useCalque('session-1', version), {
      initialProps: { version: '' },
    });
    await waitFor(() => expect(result.current.calque).toEqual(calque));

    rerender({ version: '2026-09-11T10:05:00.000Z' });

    expect(lireCalque).toHaveBeenCalledTimes(2);
    expect(result.current.calque).toEqual(calque);
  });
});

describe('derniereFin', () => {
  it('retient la fin la plus récente des extractions terminées de la base', () => {
    const extractions = [
      extraction('gescom', 'terminee', '2026-09-11T10:01:00.000Z'),
      extraction('gescom', 'terminee', '2026-09-11T10:03:00.000Z'),
      extraction('gescom', 'annulee', '2026-09-11T10:04:00.000Z'),
      extraction('paie', 'terminee', '2026-09-11T10:05:00.000Z'),
    ];

    expect(derniereFin(extractions, 'gescom')).toBe('2026-09-11T10:03:00.000Z');
  });

  it('reste vide tant qu’aucune extraction de la base n’a écrit', () => {
    expect(derniereFin([extraction('gescom', 'en_cours')], 'gescom')).toBe('');
  });
});
