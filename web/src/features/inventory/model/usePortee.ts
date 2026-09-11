// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useMemo } from 'react';

import type { Portee } from '@/shared/model';

/**
 * Compose la portée qui partira à l'extraction.
 *
 * Tables triées : deux sélections identiques doivent donner la même portée,
 * sinon l'aperçu change sans que rien n'ait bougé. Une liste vide n'est pas une
 * portée vide — c'est « toutes les tables des schémas retenus », comme en ligne
 * de commande.
 */
export function usePortee(schemas: string[], selection: Set<string>): Portee {
  return useMemo(
    () => ({ schemas, tables_incluses: [...selection].sort() }),
    [schemas, selection],
  );
}
