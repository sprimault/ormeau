// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import type { Preferences } from '../config';

/**
 * Charge le module après avoir préparé le document.
 *
 * L'état initial est lu au chargement, comme dans le navigateur où la page
 * arrive déjà injectée : un store importé en tête de fichier aurait figé les
 * valeurs avant que le test ne pose quoi que ce soit.
 */
async function chargerStore() {
  vi.resetModules();
  return await import('../preferences');
}

/** Pose ce que le serveur écrit dans la page. */
function injecter(attributs: Record<string, string>, variables: Record<string, string> = {}) {
  const racine = document.documentElement;
  for (const [nom, valeur] of Object.entries(attributs)) {
    racine.setAttribute(nom, valeur);
  }
  for (const [nom, valeur] of Object.entries(variables)) {
    racine.style.setProperty(nom, valeur);
  }
}

describe('préférences', () => {
  beforeEach(() => {
    const racine = document.documentElement;
    for (const nom of ['data-theme', 'data-apercu-ouvert']) {
      racine.removeAttribute(nom);
    }
    racine.removeAttribute('lang');
    racine.style.cssText = '';
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('lit ce que le serveur a injecté dans la page', async () => {
    injecter(
      { 'data-theme': 'sombre', lang: 'en', 'data-apercu-ouvert': '0' },
      { '--ormeau-largeur-arbre': '320px', '--ormeau-hauteur-apercu': '240px' },
    );

    const { usePreferencesStore } = await chargerStore();
    const preferences = usePreferencesStore.getState().preferences;

    expect(preferences.theme).toBe('sombre');
    expect(preferences.langue).toBe('en');
    expect(preferences.apercu_ouvert).toBe(false);
    expect(preferences.largeur_arbre).toBe(320);
    expect(preferences.hauteur_apercu).toBe(240);
    expect(preferences.largeur_entites).toBeUndefined();
  });

  it('retombe sur ses défauts quand la page n’est pas injectée', async () => {
    const { usePreferencesStore } = await chargerStore();
    const preferences = usePreferencesStore.getState().preferences;

    // Le cas de `make dev`, où Vite sert la page sans passer par le serveur Go.
    expect(preferences.theme).toBe('systeme');
    expect(preferences.langue).toBe('fr');
    expect(preferences.apercu_ouvert).toBeUndefined();
  });

  it('n’enregistre qu’une fois une rafale de réglages', async () => {
    vi.useFakeTimers();
    const { usePreferencesStore, brancherEnregistrement } = await chargerStore();

    const enregistrees: Preferences[] = [];
    brancherEnregistrement((p) => enregistrees.push(p));

    // Ce que produit une poignée qu'on tire : un événement par image.
    for (const largeur of [300, 320, 340, 360]) {
      usePreferencesStore.getState().regler({ largeur_arbre: largeur });
    }
    expect(enregistrees).toHaveLength(0);

    vi.runAllTimers();
    expect(enregistrees).toHaveLength(1);
    expect(enregistrees[0].largeur_arbre).toBe(360);
  });

  it('applique le réglage sans attendre l’enregistrement', async () => {
    vi.useFakeTimers();
    const { usePreferencesStore } = await chargerStore();

    usePreferencesStore.getState().regler({ theme: 'clair' });
    expect(usePreferencesStore.getState().preferences.theme).toBe('clair');
  });

  it('nomme la variable CSS de chaque taille', async () => {
    const { variableCSS } = await chargerStore();

    expect(variableCSS('largeur_arbre')).toBe('--ormeau-largeur-arbre');
    expect(variableCSS('hauteur_apercu')).toBe('--ormeau-hauteur-apercu');
    expect(variableCSS('largeur_entites')).toBe('--ormeau-largeur-entites');
  });
});
