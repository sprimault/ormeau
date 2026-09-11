// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { createContext, useContext } from 'react';

import type { SuiviExtractions } from './useSuiviExtractions';

/** Suivi des extractions, fourni une fois à la racine par ExtractionsProvider. */
export const ContexteExtractions = createContext<SuiviExtractions | null>(null);

/**
 * Lit le suivi des extractions.
 *
 * Échoue hors du fournisseur plutôt que de rendre une liste vide : un écran
 * monté sans lui dirait qu'aucune extraction ne tourne pendant qu'elles
 * tournent.
 */
export function useExtractions(): SuiviExtractions {
  const suivi = useContext(ContexteExtractions);
  if (!suivi) {
    throw new Error('useExtractions appelé hors de ExtractionsProvider');
  }
  return suivi;
}
