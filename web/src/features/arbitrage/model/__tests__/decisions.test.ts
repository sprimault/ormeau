// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { describe, expect, it } from 'vitest';

import type { Decisions } from '@/shared/model';
import { fichierDecisions, forcerType, nommerCas, renommer, retirerEnumeration } from '../decisions';

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

  it('nomme les cas d’une énumération sur toutes ses colonnes, sans perdre les cas déjà décidés', () => {
    const d: Decisions = {
      enumerations: [
        { colonne: 'public.facture.prefix', nom: 'PrefixeFacture', cas: { AV: 'Avoir' } },
        { colonne: 'public.autre.genre', nom: 'Genre' },
      ],
    };

    const nommee = nommerCas(
      d,
      { nom: 'Prefix', colonnes: ['public.facture.prefix', 'public.devis.prefix'] },
      { FA: 'Facture' },
    );

    const cas = { AV: 'Avoir', FA: 'Facture' };
    expect(nommee.enumerations).toEqual([
      { colonne: 'public.autre.genre', nom: 'Genre' },
      { colonne: 'public.facture.prefix', nom: 'PrefixeFacture', cas },
      { colonne: 'public.devis.prefix', nom: 'PrefixeFacture', cas },
    ]);
  });

  it('retire la décision d’énumération d’une colonne, sans laisser de liste vide', () => {
    const d: Decisions = { enumerations: [{ colonne: 'public.facture.prefix', nom: 'Prefix' }] };
    expect(retirerEnumeration(d, 'public.facture.prefix').enumerations).toBeUndefined();
  });

  it('nomme le fichier d’après la base', () => {
    expect(fichierDecisions('gescom')).toBe('gescom.decisions.yaml');
  });
});
