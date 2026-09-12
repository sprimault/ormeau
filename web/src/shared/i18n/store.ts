// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { create } from 'zustand';

import { usePreferencesStore } from '@/shared/model';

import type { Lang } from './messages';

/** État de la langue courante. */
interface EtatLangue {
  lang: Lang;
  setLang: (lang: Lang) => void;
}

/**
 * Store de la langue.
 *
 * Aucun accès à `window` au chargement du module : le store reste pur et
 * testable, la détection vit dans `init.ts` et n'est appelée qu'une fois, avant
 * le premier rendu.
 *
 * Ce qui fait survivre la langue à la fermeture est ailleurs, dans les
 * préférences que le serveur enregistre — le navigateur ne peut rien retenir,
 * son origine changeant à chaque lancement avec le port d'écoute.
 */
export const useLangStore = create<EtatLangue>((set) => ({
  lang: 'fr',
  setLang: (lang) => {
    usePreferencesStore.getState().regler({ langue: lang });
    document.documentElement.lang = lang;
    set({ lang });
  },
}));
