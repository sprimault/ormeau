// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { describe, expect, it } from 'vitest';

import { heure } from '../dates';

describe('heure', () => {
  it('rend l’heure locale sur vingt-quatre heures, secondes comprises', () => {
    // Le fuseau du poste qui lance les tests décide de l'heure affichée : on
    // compare au décompte local, pas à une chaîne écrite en dur.
    const instant = '2026-09-11T21:05:09.123Z';
    const date = new Date(instant);
    const attendu = [date.getHours(), date.getMinutes(), date.getSeconds()]
      .map((valeur) => String(valeur).padStart(2, '0'))
      .join(':');

    expect(heure(instant)).toBe(attendu);
  });

  it('garde deux chiffres partout, y compris la nuit', () => {
    expect(heure('2026-09-11T00:00:00Z')).toMatch(/^\d{2}:\d{2}:\d{2}$/);
  });
});
