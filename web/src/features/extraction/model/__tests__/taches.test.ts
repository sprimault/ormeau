// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { describe, expect, it } from 'vitest';

import type { Extraction } from '@/shared/model';
import { estTerminal, fichierCalque, remplacements, secondesEcoulees } from '../taches';

const debut = '2026-09-11T10:00:00.000Z';

const gescom: Extraction = {
  id: 'tache-1',
  base: 'gescom',
  fichier: 'gescom.calque.json',
  nb_tables: 0,
  etat: 'en_cours',
};

/** Extraction terminée d'une base à un instant donné. */
function terminee(id: string, base: string, fin: string): Extraction {
  return { id, base, fichier: fichierCalque(base), nb_tables: 0, etat: 'terminee', debut, fin };
}

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
    const finie = { ...gescom, etat: 'terminee', debut, fin: '2026-09-11T10:00:42.000Z' };
    expect(secondesEcoulees(finie, Date.parse(debut) + 3_600_000)).toBe(42);
  });

  it('garde la fraction d’une extraction plus courte qu’une seconde', () => {
    const eclair = { ...gescom, etat: 'terminee', debut, fin: '2026-09-11T10:00:00.400Z' };
    expect(secondesEcoulees(eclair, Date.now())).toBe(0.4);
  });

  it('ne rend jamais une durée négative quand les horloges s’écartent', () => {
    expect(secondesEcoulees({ ...gescom, debut }, Date.parse(debut) - 1500)).toBe(0);
  });
});

describe('remplacements', () => {
  it('désigne les calques réécrits depuis, par la plus récente extraction de la base', () => {
    const premiere = terminee('a', 'gescom', '2026-09-11T10:01:00.000Z');
    const seconde = terminee('b', 'gescom', '2026-09-11T10:02:00.000Z');
    const derniere = terminee('c', 'gescom', '2026-09-11T10:03:00.000Z');
    const autreBase = terminee('d', 'paie', '2026-09-11T10:00:30.000Z');

    const remplacees = remplacements([premiere, seconde, derniere, autreBase]);

    expect([...remplacees]).toEqual([
      ['a', derniere.fin],
      ['b', derniere.fin],
    ]);
  });

  it('ne compte que les extractions terminées : une annulée n’a rien écrit', () => {
    const ecrite = terminee('a', 'gescom', '2026-09-11T10:01:00.000Z');
    const annulee = { ...terminee('b', 'gescom', '2026-09-11T10:02:00.000Z'), etat: 'annulee' };

    expect(remplacements([ecrite, annulee]).size).toBe(0);
  });
});
