// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import { ErreurAPI, getJSON, postJSON, supprimer } from '../client';

/** Rend une réponse comme le ferait fetch. */
function reponse(corps: unknown, { statut = 200, type = 'application/json' } = {}): Response {
  return new Response(typeof corps === 'string' ? corps : JSON.stringify(corps), {
    status: statut,
    headers: { 'Content-Type': type },
  });
}

describe('client HTTP', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('n’appelle que des chemins relatifs : le port change à chaque lancement', async () => {
    const fetchDouble = vi.fn().mockResolvedValue(reponse({ ok: true }));
    vi.stubGlobal('fetch', fetchDouble);

    await getJSON('/api/contexte');
    expect(fetchDouble.mock.calls[0][0]).toBe('/api/contexte');
  });

  it('rend le message que le serveur a rédigé', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(reponse({ erreur: 'mot de passe refusé' }, { statut: 502 })),
    );

    await expect(getJSON('/api/connexion')).rejects.toThrow('mot de passe refusé');
  });

  it('porte le statut, pour que l’écran distingue ce qu’il peut corriger', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(reponse({ erreur: 'session absente' }, { statut: 401 })),
    );

    await expect(getJSON('/api/inventaire')).rejects.toMatchObject({ statut: 401 });
  });

  it('porte le code d’un refus, pour que l’écran sache quoi faire', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue(
          reponse({ erreur: 'le calque a changé', code: 'calque_modifie' }, { statut: 409 }),
        ),
    );

    await expect(postJSON('/api/inference', {})).rejects.toMatchObject({
      statut: 409,
      code: 'calque_modifie',
      message: 'le calque a changé',
    });
  });

  it('reconnaît une réponse qui n’est pas du JSON', async () => {
    // Le repli du front sert index.html sur tout ce qu'il ne connaît pas : sans
    // ce contrôle, l'appel échouerait au décodage avec un message qui ne dirait
    // rien de la cause.
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue(
          reponse('<!doctype html>', { statut: 404, type: 'text/html; charset=utf-8' }),
        ),
    );

    await expect(getJSON('/api/absent')).rejects.toThrow(/Réponse inattendue/);
  });

  it('traduit l’absence de serveur local plutôt que de laisser fuir l’erreur réseau', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')));

    await expect(getJSON('/api/contexte')).rejects.toThrow('Le serveur local ne répond pas.');
  });

  it('poste le corps en JSON', async () => {
    const fetchDouble = vi.fn().mockResolvedValue(reponse({ session: 'abc' }));
    vi.stubGlobal('fetch', fetchDouble);

    await postJSON('/api/connexion', { hote: 'bdd' });

    const init = fetchDouble.mock.calls[0][1] as RequestInit;
    expect(init.method).toBe('POST');
    expect(init.body).toBe('{"hote":"bdd"}');
  });

  it('accepte une suppression sans corps de réponse', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(null, { status: 204 })));

    await expect(supprimer('/api/connexion', { session: 'abc' })).resolves.toBeUndefined();
  });

  it('rend une ErreurAPI, que l’appelant sait reconnaître', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(reponse({ erreur: 'refus' }, { statut: 409 })),
    );

    await expect(getJSON('/api/base')).rejects.toBeInstanceOf(ErreurAPI);
  });
});
