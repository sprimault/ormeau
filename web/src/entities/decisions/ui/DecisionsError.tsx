// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { ErrorBanner } from '@/shared/ui';
import { useDecisions } from '../model/contexte';

/**
 * Signale un fichier de décisions que l'écran n'a pas pu lire.
 *
 * Sans ce bandeau, les colonnes que le fichier écartait apparaîtraient cochées
 * dans l'arbre, sans que rien ne dise pourquoi.
 */
export function DecisionsError() {
  const { erreur } = useDecisions();
  if (!erreur) {
    return null;
  }
  return (
    <div className="px-4 py-2">
      <ErrorBanner message={erreur} />
    </div>
  );
}
