// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import type { Lang } from './messages';
import { CLE_LANGUE, useLangStore } from './store';

/**
 * Détermine la langue du démarrage : celle qui a été choisie, sinon le
 * français à moins que le navigateur n'annonce l'anglais.
 */
export function langueInitiale(): Lang {
  try {
    const enregistree = window.localStorage.getItem(CLE_LANGUE);
    if (enregistree === 'fr' || enregistree === 'en') {
      return enregistree;
    }
  } catch {
    // Stockage indisponible : on retombe sur la langue du navigateur.
  }
  return navigator.language?.toLowerCase().startsWith('en') ? 'en' : 'fr';
}

/** Aligne le store sur la langue que le script de pré-initialisation a déjà
 *  posée sur le document. À appeler une fois, avant le premier rendu. */
export function initI18n(): void {
  useLangStore.getState().setLang(langueInitiale());
}
