// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { describe, expect, it } from 'vitest';

import { masquerDSN, tronquerMilieu } from './strings';

describe('tronquerMilieu', () => {
  it('laisse intacte une chaîne assez courte', () => {
    expect(tronquerMilieu('C:/projets/gescom', 64)).toBe('C:/projets/gescom');
  });

  it('respecte la longueur demandée', () => {
    expect(tronquerMilieu('a'.repeat(200), 40)).toHaveLength(40);
  });

  it('garde le dernier segment, qui porte le nom du projet', () => {
    const tronque = tronquerMilieu('C:/H_DEV/dev/sites/interne/clients/gescom', 24);
    expect(tronque.startsWith('C:/H_DEV')).toBe(true);
    expect(tronque.endsWith('gescom')).toBe(true);
    expect(tronque).toContain('…');
  });
});

describe('masquerDSN', () => {
  it('retire le mot de passe et conserve le reste', () => {
    expect(masquerDSN('postgres://utilisateur:secret@hote:5432/gescom')).toBe(
      'postgres://utilisateur:***@hote:5432/gescom',
    );
  });

  it('laisse passer un DSN sans mot de passe', () => {
    expect(masquerDSN('postgres://utilisateur@hote/gescom')).toBe('postgres://utilisateur@hote/gescom');
  });

  it('masque tout ce qu’il ne sait pas analyser', () => {
    expect(masquerDSN('host=hote password=secret')).toBe('***');
  });

  it('rend une chaîne vide telle quelle', () => {
    expect(masquerDSN('')).toBe('');
  });
});
