// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import type { ReactNode } from 'react';
import { renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import type { Extraction } from '@/shared/model';
import { ContexteExtractions } from '../contexte';
import { useVersionCalque } from '../useCalque';

/** Extraction d'une base dans un état donné, finie à un instant donné. */
function extraction(base: string, etat: string, fin: string): Extraction {
  return { id: `${base}-${fin}`, base, fichier: `${base}.calque.json`, nb_tables: 0, etat, fin };
}

describe('useVersionCalque', () => {
  it('suit la dernière extraction terminée de la base, et seulement elle', () => {
    const extractions = [
      extraction('gescom', 'terminee', '2026-09-11T10:00:00Z'),
      extraction('paie', 'terminee', '2026-09-11T11:00:00Z'),
      extraction('gescom', 'echouee', '2026-09-11T12:00:00Z'),
    ];
    const enveloppe = ({ children }: { children: ReactNode }) => (
      <ContexteExtractions.Provider value={{ extractions, connecte: true }}>
        {children}
      </ContexteExtractions.Provider>
    );

    const { result } = renderHook(() => useVersionCalque('gescom'), { wrapper: enveloppe });

    expect(result.current).toBe('2026-09-11T10:00:00Z');
  });
});
