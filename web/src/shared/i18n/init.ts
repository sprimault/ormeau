// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { usePreferencesStore } from '@/shared/model';

import type { Lang } from './messages';
import { useLangStore } from './store';

/**
 * Détermine la langue du démarrage : celle que le serveur a injectée dans la
 * page, sinon le français à moins que le navigateur n'annonce l'anglais.
 *
 * Le repli sur le navigateur couvre le premier lancement, où rien n'a encore
 * été choisi, et la page servie par Vite en développement, qui ne porte aucune
 * préférence.
 */
export function langueInitiale(): Lang {
  const injectee = usePreferencesStore.getState().preferences.langue;
  if (injectee === 'fr' || injectee === 'en') {
    return injectee;
  }
  return navigator.language?.toLowerCase().startsWith('en') ? 'en' : 'fr';
}

/**
 * Aligne le store sur la langue de la page. À appeler une fois, avant le
 *  premier rendu.
 *
 * Par `setState` et non par `setLang` : ce dernier enregistre, et un démarrage
 * réécrirait le fichier de préférences à chaque lancement pour y reposer ce
 * qu'il contenait déjà. Une langue détectée depuis le navigateur n'est
 * enregistrée que le jour où on la choisit.
 */
export function initI18n(): void {
  const lang = langueInitiale();
  useLangStore.setState({ lang });
  document.documentElement.lang = lang;
}
