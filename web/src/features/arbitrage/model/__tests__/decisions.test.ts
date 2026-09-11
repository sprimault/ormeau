// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { describe, expect, it } from 'vitest';

import type { Decisions } from '@/shared/model';
import { fichierDecisions, forcerType, renommer } from '../decisions';

describe('transformations du brouillon', () => {
  it('renomme, et rend la main à l’inférence sur un nom vide', () => {
    const renommee = renommer({}, 'public.clients', 'Client');
    expect(renommee.renommages).toEqual({ 'public.clients': 'Client' });
    expect(renommer(renommee, 'public.clients', '').renommages).toBeUndefined();
  });

  it('force un type, et le retire sur un type vide', () => {
    const forcee = forcerType({}, 'dbo.T_CLIENTS.CLI_ACTIF', 'boolean');
    expect(forcee.types_forces).toEqual({ 'dbo.T_CLIENTS.CLI_ACTIF': 'boolean' });
    expect(forcerType(forcee, 'dbo.T_CLIENTS.CLI_ACTIF', '').types_forces).toBeUndefined();
  });

  it('ne touche pas aux autres décisions', () => {
    const d: Decisions = { tables_ignorees: ['public.migrations'], espace_de_noms: 'App' };
    expect(renommer(d, 'public.clients', 'Client')).toMatchObject(d);
    expect(forcerType(d, 'public.clients.geo', 'string')).toMatchObject(d);
  });

  it('nomme le fichier d’après la base', () => {
    expect(fichierDecisions('gescom')).toBe('gescom.decisions.yaml');
  });
});
