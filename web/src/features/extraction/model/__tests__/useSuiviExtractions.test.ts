// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import type { Extraction } from '@/shared/model';
import { useSuiviExtractions } from '../useSuiviExtractions';

/**
 * Double d'EventSource.
 *
 * jsdom n'en fournit pas, et c'est le serveur qu'on simule ici : ce qu'il émet,
 * et dans quel ordre.
 */
class FluxDouble {
  static ouverts: FluxDouble[] = [];

  readonly url: string;
  ferme = false;
  private readonly ecouteurs = new Map<string, ((evenement: Event) => void)[]>();

  constructor(url: string) {
    this.url = url;
    FluxDouble.ouverts.push(this);
  }

  /** Enregistre un écouteur, comme EventSource. */
  addEventListener(nom: string, ecouteur: (evenement: Event) => void) {
    this.ecouteurs.set(nom, [...(this.ecouteurs.get(nom) ?? []), ecouteur]);
  }

  /** Ferme le flux. */
  close() {
    this.ferme = true;
  }

  /** Émet un événement du serveur, données sérialisées comme sur le fil. */
  emettre(nom: string, donnees?: unknown) {
    const evenement =
      donnees === undefined
        ? new Event(nom)
        : new MessageEvent(nom, { data: JSON.stringify(donnees) });
    act(() => {
      for (const ecouteur of this.ecouteurs.get(nom) ?? []) {
        ecouteur(evenement);
      }
    });
  }
}

/** Tâche minimale dans un état donné. */
function tache(id: string, etat: string, base = 'gescom'): Extraction {
  return { id, base, fichier: `${base}.calque.json`, nb_tables: 0, etat };
}

/** Le flux que le hook a ouvert. */
function flux(): FluxDouble {
  return FluxDouble.ouverts[0];
}

describe('useSuiviExtractions', () => {
  beforeEach(() => {
    FluxDouble.ouverts = [];
    vi.stubGlobal('EventSource', FluxDouble);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('ouvre un seul flux, par un chemin relatif', () => {
    renderHook(() => useSuiviExtractions());

    expect(FluxDouble.ouverts).toHaveLength(1);
    expect(flux().url).toBe('/api/extractions/evenements');
  });

  it('remplace toute la liste par un instantané', () => {
    const { result } = renderHook(() => useSuiviExtractions());

    flux().emettre('extraction', tache('a', 'en_cours'));
    flux().emettre('etat', { extractions: [tache('b', 'terminee', 'paie')] });

    expect(result.current.extractions.map((e) => e.id)).toEqual(['b']);
  });

  it('met une tâche à jour à sa place et ajoute les nouvelles à la fin', () => {
    const { result } = renderHook(() => useSuiviExtractions());

    flux().emettre('etat', {
      extractions: [tache('a', 'en_cours'), tache('b', 'en_attente', 'paie')],
    });
    flux().emettre('extraction', tache('a', 'terminee'));
    flux().emettre('extraction', tache('c', 'en_attente', 'stock'));

    expect(result.current.extractions.map((e) => `${e.id}:${e.etat}`)).toEqual([
      'a:terminee',
      'b:en_attente',
      'c:en_attente',
    ]);
  });

  it('retire la tâche qu’un événement de retrait désigne', () => {
    const { result } = renderHook(() => useSuiviExtractions());

    flux().emettre('etat', { extractions: [tache('a', 'terminee'), tache('b', 'terminee', 'paie')] });
    flux().emettre('retrait', { id: 'a' });

    expect(result.current.extractions.map((e) => e.id)).toEqual(['b']);
  });

  it('signale la coupure du flux et sa reprise', () => {
    const { result } = renderHook(() => useSuiviExtractions());
    expect(result.current.connecte).toBe(false);

    flux().emettre('open');
    expect(result.current.connecte).toBe(true);

    flux().emettre('error');
    expect(result.current.connecte).toBe(false);
  });

  it('ferme le flux quand l’interface disparaît', () => {
    const { unmount } = renderHook(() => useSuiviExtractions());

    unmount();

    expect(flux().ferme).toBe(true);
  });
});
