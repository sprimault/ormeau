// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { describe, expect, it } from 'vitest';

import { masquerDSN, minusculeInitiale, tronquerMilieu } from '../strings';

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
    expect(masquerDSN('postgres://utilisateur@hote/gescom')).toBe(
      'postgres://utilisateur@hote/gescom',
    );
  });

  it('masque tout ce qu’il ne sait pas analyser', () => {
    expect(masquerDSN('host=hote password=secret')).toBe('***');
  });

  it('rend une chaîne vide telle quelle', () => {
    expect(masquerDSN('')).toBe('');
  });

  it('masque un mot de passe vide, comme le Go', () => {
    expect(masquerDSN('postgres://u:@hote/gescom')).toBe('postgres://u:***@hote/gescom');
  });

  it('masque un mot de passe contenant un @ en entier', () => {
    expect(masquerDSN('postgres://u:mot@Secret123@hote/gescom')).toBe(
      'postgres://u:***@hote/gescom',
    );
  });

  it('masque la valeur de tout paramètre qui n’est pas connu pour inoffensif', () => {
    expect(masquerDSN('sqlserver://sa@hote:1433?database=gescom&password=Secret123')).toBe(
      'sqlserver://sa@hote:1433?database=gescom&password=***',
    );
    expect(masquerDSN('postgres://u@hote/gescom?sslmode=disable&options=Secret123')).toBe(
      'postgres://u@hote/gescom?sslmode=disable&options=***',
    );
    expect(masquerDSN('postgres://u@hote/gescom?SSLPASSWORD=Secret123&app+name=x')).toBe(
      'postgres://u@hote/gescom?SSLPASSWORD=***&app+name=x',
    );
  });

  it.each([
    ['mot de passe avec barre oblique', 'postgres://u:123/Secret123@hote:1/gescom'],
    ['mot de passe avec dièse', 'postgres://u:123#Secret123@hote:1/gescom'],
    ['mot de passe avec point d’interrogation', 'postgres://u:123?Secret123@hote:1/gescom'],
    ['paramètre mal encodé', 'postgres://u@hote/gescom?sslpassword=%zzSecret123'],
    ['paramètre sans valeur', 'postgres://u@hote/gescom?Secret123'],
    ['nom de paramètre avec un blanc', 'postgres://u@hote/gescom?password+Secret123=x'],
    ['caractère de contrôle', 'postgres://u:Secret123@hote/gescom'],
    ['séquence de schéma dans une forme clé/valeur', 'host=hote password=Secret123://x'],
  ])('masque en entier ce qui est ambigu : %s', (_, dsn) => {
    expect(masquerDSN(dsn)).toBe('***');
  });
});

describe('minusculeInitiale', () => {
  it('passe d’un nom de classe à un nom de propriété sans toucher aux majuscules internes', () => {
    expect(minusculeInitiale('TenantInvoice')).toBe('tenantInvoice');
    expect(minusculeInitiale('TLog')).toBe('tLog');
  });

  it('rend une chaîne vide telle quelle', () => {
    expect(minusculeInitiale('')).toBe('');
  });
});
