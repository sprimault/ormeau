// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { create } from 'zustand';

import type { Preferences } from './config';

/** Les trois panneaux dont la taille se règle. */
export type ClePreference = 'largeur_arbre' | 'hauteur_apercu' | 'largeur_entites';

/**
 * Délai avant d'enregistrer, en millisecondes.
 *
 * Une poignée qu'on tire produit un événement par image : sans ce délai, tirer
 * un séparateur sur deux cents pixels écrirait deux cents fois le fichier.
 */
const DELAI_ENREGISTREMENT = 400;

/** Ce que le store expose. */
interface EtatPreferences {
  preferences: Preferences;
  /** Met à jour les valeurs données et programme leur enregistrement. */
  regler: (partiel: Partial<Preferences>) => void;
}

/**
 * Préférences d'affichage, dont le serveur est la seule source de vérité.
 *
 * Elles ne vivent pas dans le navigateur, et ce n'est pas un choix de
 * commodité : le binaire écoute sur un port tiré à chaque lancement, l'origine
 * de la page change donc à chaque fois, et tout `localStorage` repart vide.
 * Un thème choisi ne survivait pas à la fermeture de l'onglet.
 *
 * L'état initial est lu sur le document, où le serveur l'a posé avant le
 * premier rendu — pas demandé par un appel d'API, qui arriverait après.
 */
export const usePreferencesStore = create<EtatPreferences>((set, get) => ({
  preferences: preferencesInitiales(),
  regler: (partiel) => {
    const suivantes = { ...get().preferences, ...partiel };
    set({ preferences: suivantes });
    programmerEnregistrement(suivantes);
  },
}));

/**
 * Relit ce que le serveur a injecté dans la page.
 *
 * Tout est facultatif : servie par Vite en développement, la page ne porte
 * aucune de ces valeurs, et l'interface doit s'ouvrir sur ses défauts plutôt
 * que d'échouer. Même repli pour un `index.html` servi en statique par erreur.
 */
function preferencesInitiales(): Preferences {
  const racine = document.documentElement;

  const preferences: Preferences = {
    theme: racine.dataset.theme || 'systeme',
    langue: racine.lang || 'fr',
  };

  const apercu = racine.dataset.apercuOuvert;
  if (apercu === '0' || apercu === '1') {
    preferences.apercu_ouvert = apercu === '1';
  }

  const styles = window.getComputedStyle(racine);
  for (const cle of ['largeur_arbre', 'hauteur_apercu', 'largeur_entites'] as const) {
    const valeur = Number.parseInt(styles.getPropertyValue(variableCSS(cle)), 10);
    if (Number.isFinite(valeur) && valeur > 0) {
      preferences[cle] = valeur;
    }
  }
  return preferences;
}

/** Nom de la variable CSS qui porte une taille. */
export function variableCSS(cle: ClePreference): string {
  return `--ormeau-${cle.replace('_', '-')}`;
}

/**
 * Enregistreur branché par la couche application.
 *
 * Le store ne connaît pas le client HTTP : `shared/api` importe déjà des types
 * de `shared/model`, et l'appeler d'ici fermerait le cercle. C'est donc `app`
 * qui câble les deux, ce qui laisse aussi les tests observer sans réseau.
 */
let enregistrer: ((preferences: Preferences) => void) | null = null;
let minuterie: ReturnType<typeof setTimeout> | undefined;

/** Branche l'enregistrement. À appeler une fois, au démarrage. */
export function brancherEnregistrement(fonction: (preferences: Preferences) => void): void {
  enregistrer = fonction;
}

/** Repousse l'enregistrement tant que les réglages s'enchaînent. */
function programmerEnregistrement(preferences: Preferences): void {
  clearTimeout(minuterie);
  minuterie = setTimeout(() => enregistrer?.(preferences), DELAI_ENREGISTREMENT);
}
