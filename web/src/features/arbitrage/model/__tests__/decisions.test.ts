// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { describe, expect, it } from 'vitest';

import type { Decisions } from '@/shared/model';
import {
  ajouterRelation,
  fichierDecisions,
  forcerType,
  nommerCas,
  renommer,
  retirerEnumeration,
  retirerRelation,
} from '../decisions';

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

  it('relie une colonne à une autre entité, en remplaçant la relation qu’elle portait', () => {
    const premiere = {
      source: 'public.commande.client_ref',
      cible: 'public.client.id',
      genre: 'plusieurs_vers_un',
      nom: 'client',
    };
    const seconde = { ...premiere, cible: 'public.utilisateur.id', nom: 'acheteur' };
    const autre = { ...premiere, source: 'public.commande.vendeur_ref', nom: 'vendeur' };

    const d = ajouterRelation(ajouterRelation(ajouterRelation({}, premiere), autre), seconde);

    expect(d.relations_forcees).toEqual([autre, seconde]);
    expect(retirerRelation(d, autre.source).relations_forcees).toEqual([seconde]);
    expect(retirerRelation({ relations_forcees: [premiere] }, premiere.source).relations_forcees).toBeUndefined();
  });

  it('nomme le fichier d’après la base', () => {
    expect(fichierDecisions('gescom')).toBe('gescom.decisions.yaml');
  });
});
