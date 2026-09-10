// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useCallback } from 'react';

import { messages, type Lang, type MessageKey } from './messages';
import { useLangStore } from './store';

/** Paramètres substitués dans un message, aux emplacements `{nom}`. */
export type Params = Record<string, string | number>;

/**
 * Rend la fonction de traduction et s'abonne à la langue courante.
 *
 * Mémoïsée sur la langue, et c'est nécessaire plutôt que cosmétique : sans
 * cela `t` change d'identité à chaque rendu, et tout effet qui la déclare en
 * dépendance se relance indéfiniment.
 */
export function useT(): (cle: MessageKey, params?: Params) => string {
  const lang = useLangStore((etat) => etat.lang);
  return useCallback((cle: MessageKey, params?: Params) => traduire(lang, cle, params), [lang]);
}

/**
 * Traduit hors de React, pour le code qui n'a pas de hook à sa disposition —
 * le client HTTP, dont les messages d'erreur remontent jusqu'à l'écran.
 *
 * Lit la langue à l'appel plutôt que de s'y abonner : un message d'erreur est
 * produit une fois, il n'a pas à se retraduire ensuite.
 */
export function translate(cle: MessageKey, params?: Params): string {
  return traduire(useLangStore.getState().lang, cle, params);
}

/** Résout une clé et substitue ses paramètres. */
function traduire(lang: Lang, cle: MessageKey, params?: Params): string {
  let message: string = messages[lang][cle];
  if (params) {
    for (const [nom, valeur] of Object.entries(params)) {
      message = message.replace('{' + nom + '}', String(valeur));
    }
  }
  return message;
}
