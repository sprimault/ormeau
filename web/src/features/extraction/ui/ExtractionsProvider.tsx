// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import type { ReactNode } from 'react';

import { ContexteExtractions } from '../model/contexte';
import { useSuiviExtractions } from '../model/useSuiviExtractions';

/** Propriétés du fournisseur. */
interface ProprietesFournisseur {
  children: ReactNode;
}

/**
 * Ouvre le flux des extractions pour toute l'interface.
 *
 * À la racine, au-dessus du formulaire de connexion comme de l'écran de
 * sélection : une extraction lancée continue après la déconnexion, et son suivi
 * dans l'en-tête doit continuer avec elle.
 */
export function ExtractionsProvider({ children }: ProprietesFournisseur) {
  const suivi = useSuiviExtractions();
  return <ContexteExtractions.Provider value={suivi}>{children}</ContexteExtractions.Provider>;
}
