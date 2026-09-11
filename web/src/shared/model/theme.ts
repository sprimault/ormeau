// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { create } from 'zustand';

import { usePreferencesStore } from './preferences';

/** Les trois états du thème. « Système » est le défaut et suit le poste en
 *  direct, ce qu'attend quelqu'un dont l'écran bascule le soir. */
export type Theme = 'clair' | 'sombre' | 'systeme';

/** État du thème courant. */
interface EtatTheme {
  theme: Theme;
  setTheme: (theme: Theme) => void;
}

/**
 * Store du thème.
 *
 * Il porte ce que l'interface affiche ; ce qui le fait survivre à la fermeture
 * est ailleurs, dans les préférences que le serveur enregistre. La dépendance
 * ne va que dans ce sens — le thème se règle ici et s'enregistre là, jamais
 * l'inverse —, sans quoi deux copies d'une même valeur finiraient par diverger.
 *
 * Le navigateur ne peut rien retenir de son côté : le port d'écoute est tiré à
 * chaque lancement, l'origine de la page change avec lui, et tout stockage
 * local repart vide.
 */
export const useThemeStore = create<EtatTheme>((set) => ({
  theme: 'systeme',
  setTheme: (theme) => {
    usePreferencesStore.getState().regler({ theme });
    appliquer(theme);
    set({ theme });
  },
}));

/**
 * Aligne le store sur le thème que le serveur a posé sur la page, et branche
 * le suivi du réglage système.
 *
 * La page arrive déjà dans le bon thème, injecté dans la balise html avant que
 * le bundle ne soit chargé : cet appel ne corrige donc rien au premier rendu.
 * Il existe pour l'écoute, et pour le cas où la page n'a pas été injectée —
 * servie par Vite en développement, l'attribut est absent et le défaut
 * s'applique.
 *
 * L'écoute est installée une fois pour toutes plutôt qu'ajoutée et retirée au
 * gré des changements : elle ne fait rien tant que le thème n'est pas
 * « système », et un abonnement qu'on n'a pas à défaire ne peut pas fuir.
 */
export function initTheme(): void {
  const theme = themeInitial();
  useThemeStore.setState({ theme });
  appliquer(theme);

  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    if (useThemeStore.getState().theme === 'systeme') {
      appliquer('systeme');
    }
  });
}

/** Rend le thème injecté dans la page, « système » à défaut. */
export function themeInitial(): Theme {
  const valeur = usePreferencesStore.getState().preferences.theme;
  return valeur === 'clair' || valeur === 'sombre' || valeur === 'systeme' ? valeur : 'systeme';
}

/** Pose ou retire la classe que la variante Tailwind observe. */
function appliquer(theme: Theme): void {
  const sombre =
    theme === 'sombre' ||
    (theme === 'systeme' && window.matchMedia('(prefers-color-scheme: dark)').matches);
  document.documentElement.classList.toggle('dark', sombre);
}
