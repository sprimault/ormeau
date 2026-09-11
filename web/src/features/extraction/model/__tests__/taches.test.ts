// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { describe, expect, it } from 'vitest';

import type { Extraction } from '@/shared/model';
import { estTerminal, fichierCalque, secondesEcoulees } from '../taches';

const debut = '2026-09-11T10:00:00Z';

const gescom: Extraction = {
  id: 'tache-1',
  base: 'gescom',
  fichier: 'gescom.calque.json',
  nb_tables: 0,
  etat: 'en_cours',
};

describe('estTerminal', () => {
  it('distingue ce qui évolue encore de ce qui a fini', () => {
    expect(['en_attente', 'en_cours'].map((etat) => estTerminal(etat))).toEqual([false, false]);
    expect(['terminee', 'echouee', 'annulee'].map((etat) => estTerminal(etat))).toEqual([
      true,
      true,
      true,
    ]);
  });
});

describe('fichierCalque', () => {
  it('nomme le calque d’après la base', () => {
    expect(fichierCalque('gescom')).toBe('gescom.calque.json');
  });
});

describe('secondesEcoulees', () => {
  it('n’a pas de durée avant le démarrage', () => {
    expect(secondesEcoulees({ ...gescom, etat: 'en_attente' }, Date.now())).toBeNull();
  });

  it('court jusqu’à maintenant pendant l’extraction', () => {
    expect(secondesEcoulees({ ...gescom, debut }, Date.parse(debut) + 65_000)).toBe(65);
  });

  it('s’arrête à la fin, quelle que soit l’heure', () => {
    const finie = { ...gescom, etat: 'terminee', debut, fin: '2026-09-11T10:00:42Z' };
    expect(secondesEcoulees(finie, Date.parse(debut) + 3_600_000)).toBe(42);
  });

  it('ne rend jamais une durée négative quand les horloges s’écartent', () => {
    expect(secondesEcoulees({ ...gescom, debut }, Date.parse(debut) - 1500)).toBe(0);
  });
});
