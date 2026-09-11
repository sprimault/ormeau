// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useEffect, useState } from 'react';

/**
 * Rend l'heure courante, rafraîchie chaque seconde tant qu'actif est vrai.
 *
 * Arrêtée quand plus rien ne tourne : les durées des tâches finies sont figées,
 * elles n'ont pas à provoquer un rendu par seconde.
 */
export function useMaintenant(actif: boolean): number {
  const [maintenant, setMaintenant] = useState(() => Date.now());

  useEffect(() => {
    if (!actif) {
      return;
    }
    setMaintenant(Date.now());
    const minuterie = window.setInterval(() => setMaintenant(Date.now()), 1000);
    return () => window.clearInterval(minuterie);
  }, [actif]);

  return maintenant;
}
