// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

import './index.css';
import { App } from '@/app/App';
import { initI18n } from '@/shared/i18n';
import { initTheme } from '@/shared/model';

// Langue et thème sont appliqués au document par le script de
// pré-initialisation ; ces deux appels alignent les stores sur ce qui est déjà
// affiché, sans quoi le premier rendu contredirait la page.
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
