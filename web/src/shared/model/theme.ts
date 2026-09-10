// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { create } from 'zustand';

/** Clé de persistance, partagée avec le script de pré-initialisation. */
export const CLE_THEME = 'ormeau-theme';

/** Les trois états du thème. « Système » est le défaut et suit le poste en
 *  direct, ce qu'attend quelqu'un dont l'écran bascule le soir. */
export type Theme = 'clair' | 'sombre' | 'systeme';

/** État du thème courant. */
interface EtatTheme {
  theme: Theme;
  setTheme: (theme: Theme) => void;
}

/** Store du thème. Comme celui de la langue, il n'accède à rien au chargement
 *  du module : `initTheme` s'en charge une fois, au démarrage. */
export const useThemeStore = create<EtatTheme>((set) => ({
  theme: 'systeme',
  setTheme: (theme) => {
    try {
      window.localStorage.setItem(CLE_THEME, theme);
    } catch {
      // Stockage indisponible : le thème vaut pour cette session seulement.
    }
    appliquer(theme);
    set({ theme });
  },
}));

/**
 * Aligne le store sur ce que le script de pré-initialisation a déjà appliqué,
 * et branche le suivi du réglage système.
 *
 * L'écoute est installée une fois pour toutes plutôt qu'ajoutée et retirée au
 * gré des changements : elle ne fait rien tant que le thème n'est pas
 * « système », et un abonnement qu'on n'a pas à défaire ne peut pas fuir.
 */
export function initTheme(): void {
  useThemeStore.setState({ theme: themeEnregistre() });
  appliquer(themeEnregistre());

  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    if (useThemeStore.getState().theme === 'systeme') {
      appliquer('systeme');
    }
  });
}

/** Rend le thème choisi précédemment, « système » par défaut. */
export function themeEnregistre(): Theme {
  try {
    const enregistre = window.localStorage.getItem(CLE_THEME);
    if (enregistre === 'clair' || enregistre === 'sombre' || enregistre === 'systeme') {
      return enregistre;
    }
  } catch {
    // Stockage indisponible : le défaut s'applique.
  }
  return 'systeme';
}

/** Pose ou retire la classe que la variante Tailwind observe. */
function appliquer(theme: Theme): void {
  const sombre =
    theme === 'sombre' ||
    (theme === 'systeme' && window.matchMedia('(prefers-color-scheme: dark)').matches);
  document.documentElement.classList.toggle('dark', sombre);
}
