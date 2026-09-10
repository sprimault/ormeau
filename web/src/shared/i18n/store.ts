// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { create } from 'zustand';

import type { Lang } from './messages';

/** Clé de persistance, partagée avec le script de pré-initialisation. */
export const CLE_LANGUE = 'ormeau-lang';

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
 */
export const useLangStore = create<EtatLangue>((set) => ({
  lang: 'fr',
  setLang: (lang) => {
    try {
      window.localStorage.setItem(CLE_LANGUE, lang);
    } catch {
      // Stockage indisponible : la langue vaut pour cette session seulement.
    }
    document.documentElement.lang = lang;
    set({ lang });
  },
}));
