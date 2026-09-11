// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { createContext, useContext } from 'react';

import type { BrouillonDecisions } from './useBrouillonDecisions';

/** Brouillon de décisions de la base ouverte, fourni par DecisionsProvider. */
export const ContexteDecisions = createContext<BrouillonDecisions | null>(null);

/**
 * Lit le brouillon de décisions de la base ouverte.
 *
 * Échoue hors du fournisseur plutôt que de rendre un brouillon vide : un écran
 * monté sans lui écrirait des décisions que personne n'enregistrerait.
 */
export function useDecisions(): BrouillonDecisions {
  const brouillon = useContext(ContexteDecisions);
  if (!brouillon) {
    throw new Error('useDecisions appelé hors de DecisionsProvider');
  }
  return brouillon;
}
