// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { translate } from '@/shared/i18n';
import type { ReponseErreur } from '@/shared/model';

/**
 * Client HTTP de l'interface.
 *
 * Toutes les URL sont relatives, sans exception : le binaire choisit un port
 * libre à chaque lancement, et la moindre adresse absolue casserait l'interface
 * dès que ce port change.
 *
 * Aucune chaîne SQL ne part d'ici. Le front choisit un point d'entrée et lui
 * passe des paramètres structurés ; il ne décrit jamais ce qu'il faut exécuter.
 */

/** Échec d'un appel d'API, porteur du message que le serveur a rédigé. */
export class ErreurAPI extends Error {
  constructor(
    public readonly statut: number,
    message: string,
  ) {
    super(message);
    this.name = 'ErreurAPI';
  }
}

/** Exécute la requête et transforme tout échec en ErreurAPI. */
async function appeler(chemin: string, init?: RequestInit): Promise<Response> {
  let reponse: Response;
  try {
    reponse = await fetch(chemin, init);
  } catch {
    // Le serveur local est mort, ou l'onglet a survécu à l'arrêt du binaire.
    throw new ErreurAPI(0, translate('error.network'));
  }

  if (!reponse.ok) {
    throw new ErreurAPI(reponse.status, await messageDErreur(reponse));
  }
  return reponse;
}

/**
 * Extrait le message d'un échec.
 *
 * L'API répond toujours en JSON, mais un 404 servi par le repli du front rend
 * du HTML : sans ce contrôle, l'appel échouerait au décodage avec un message
 * qui ne dirait rien de la cause.
 */
async function messageDErreur(reponse: Response): Promise<string> {
  const type = reponse.headers.get('Content-Type') ?? '';
  if (!type.includes('json')) {
    return translate('error.notJson', {
      path: new URL(reponse.url, window.location.origin).pathname,
      type: type || translate('error.noContentType'),
    });
  }

  try {
    const corps = (await reponse.json()) as ReponseErreur;
    return corps.erreur || translate('error.unknown');
  } catch {
    return translate('error.unknown');
  }
}

/** Lit un point d'entrée et décode sa réponse. */
export async function getJSON<T>(chemin: string, signal?: AbortSignal): Promise<T> {
  const reponse = await appeler(chemin, { signal });
  return (await reponse.json()) as T;
}

/** Poste un corps JSON et décode la réponse. */
export async function postJSON<Requete, Reponse>(
  chemin: string,
  corps: Requete,
  signal?: AbortSignal,
): Promise<Reponse> {
  const reponse = await appeler(chemin, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(corps),
    signal,
  });
  return (await reponse.json()) as Reponse;
}

/** Envoie une suppression. Le serveur répond sans corps. */
export async function supprimer<Requete>(
  chemin: string,
  corps: Requete,
  signal?: AbortSignal,
): Promise<void> {
  await appeler(chemin, {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(corps),
    signal,
  });
}
