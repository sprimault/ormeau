// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useState } from 'react';

/** Écran demandé par l'adresse. */
export type Onglet = { ecran: 'selection' } | { ecran: 'arbitrage'; base: string };

const PREFIXE_ARBITRAGE = '#arbitrage/';

/**
 * Lit l'onglet porté par le fragment d'adresse.
 *
 * Tout ce qui ne se lit pas — préfixe inconnu, nom vide, encodage cassé —
 * ramène à la sélection plutôt qu'à une erreur : une adresse retouchée à la
 * main ne doit pas laisser un écran blanc.
 */
export function lireOnglet(fragment: string): Onglet {
  if (!fragment.startsWith(PREFIXE_ARBITRAGE)) {
    return { ecran: 'selection' };
  }
  try {
    const base = decodeURIComponent(fragment.slice(PREFIXE_ARBITRAGE.length));
    return base === '' ? { ecran: 'selection' } : { ecran: 'arbitrage', base };
  } catch {
    return { ecran: 'selection' };
  }
}

/**
 * Compose le fragment d'un onglet.
 *
 * Le nom est encodé sans présumer de ce que le serveur accepte : un `/` ou un
 * `#` couperait la lecture, et l'écriture du fragment ne doit pas dépendre de
 * l'expression qui valide les noms de base.
 */
export function fragmentOnglet(onglet: Onglet): string {
  return onglet.ecran === 'arbitrage'
    ? PREFIXE_ARBITRAGE + encodeURIComponent(onglet.base)
    : '';
}

/**
 * Tient l'onglet courant dans le fragment d'adresse.
 *
 * Pas de routeur pour deux écrans dont un seul a un paramètre. Le fragment ne
 * part jamais au serveur, et il retient l'écran à travers un rechargement :
 * c'est ce qui rouvre un arbitrage après F5, sans session. Chaque changement
 * ajoute une entrée d'historique, que le bouton précédent défait.
 */
export function useOnglet(): [Onglet, (onglet: Onglet) => void] {
  const [onglet, setOnglet] = useState(() => lireOnglet(window.location.hash));

  useEffect(() => {
    const relire = () => setOnglet(lireOnglet(window.location.hash));
    window.addEventListener('popstate', relire);
    window.addEventListener('hashchange', relire);
    return () => {
      window.removeEventListener('popstate', relire);
      window.removeEventListener('hashchange', relire);
    };
  }, []);

  const aller = useCallback((suivant: Onglet) => {
    const { pathname, search } = window.location;
    // pushState ne déclenche aucun événement : l'état suit à la main.
    window.history.pushState(null, '', pathname + search + fragmentOnglet(suivant));
    setOnglet(suivant);
  }, []);

  return [onglet, aller];
}
