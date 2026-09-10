// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Applique le thème et la langue avant que React ne rende quoi que ce soit.
//
// Script séparé plutôt qu'inline pour rester lisible et relu comme le reste,
// mais chargé en synchrone : le bundle est un module, donc différé, et tout ce
// qui l'attend produit une bascule visible à l'écran.
//
// Il n'y a rien à partager avec le store zustand : celui-ci relit les mêmes
// clés au démarrage. Dupliquer trois lignes coûte moins qu'un module importé
// avant le premier rendu.
(function () {
  var CLE_THEME = 'ormeau-theme';
  var CLE_LANGUE = 'ormeau-lang';

  function lire(cle) {
    try {
      return window.localStorage.getItem(cle);
    } catch (e) {
      // Mode privé, quotas, stockage désactivé : on retombe sur les défauts.
      return null;
    }
  }

  var theme = lire(CLE_THEME);
  var sombre =
    theme === 'sombre' ||
    ((theme === null || theme === 'systeme') &&
      window.matchMedia('(prefers-color-scheme: dark)').matches);
  document.documentElement.classList.toggle('dark', sombre);

  var langue = lire(CLE_LANGUE);
  if (langue !== 'fr' && langue !== 'en') {
    langue = navigator.language && navigator.language.toLowerCase().startsWith('en') ? 'en' : 'fr';
  }
  document.documentElement.lang = langue;
})();
