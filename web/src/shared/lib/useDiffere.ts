// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useEffect, useState } from 'react';

/**
 * Rend une valeur avec un temps de retard, remis à zéro à chaque changement.
 *
 * La première valeur passe sans attendre. Les suivantes ne passent qu'après
 * `delai` millisecondes sans nouveau changement : un nom de classe tapé lettre
 * à lettre ne rejoue l'inférence qu'une fois.
 */
export function useDiffere<T>(valeur: T, delai: number): T {
  const [differee, setDifferee] = useState(valeur);

  useEffect(() => {
    const minuterie = window.setTimeout(() => setDifferee(valeur), delai);
    return () => window.clearTimeout(minuterie);
  }, [valeur, delai]);

  return differee;
}
