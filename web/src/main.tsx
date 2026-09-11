// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

import './index.css';
import { App } from '@/app/App';
import { enregistrerPreferences } from '@/shared/api';
import { initI18n } from '@/shared/i18n';
import { brancherEnregistrement, initTheme } from '@/shared/model';

// Le store des préférences ne connaît pas le client HTTP : `shared/api` importe
// déjà des types de `shared/model`, et l'appeler de là-bas fermerait le cercle.
// C'est donc ici que les deux se rejoignent.
brancherEnregistrement(enregistrerPreferences);

// Langue et thème sont déjà appliqués à la page, injectés par le serveur avant
// que ce bundle ne soit chargé ; ces deux appels alignent les stores sur ce qui
// est affiché, sans quoi le premier rendu contredirait la page.
initI18n();
initTheme();

const racine = document.getElementById('racine');
if (!racine) {
  throw new Error('élément racine introuvable');
}

createRoot(racine).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
