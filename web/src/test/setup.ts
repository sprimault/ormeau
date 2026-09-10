// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import '@testing-library/jest-dom/vitest';

// jsdom n'implémente pas matchMedia, que le thème interroge dès l'initialisation.
// Sans ce complément, tout test qui monte un composant échouerait sur une
// absence sans rapport avec ce qu'il vérifie.
if (!window.matchMedia) {
  window.matchMedia = ((requete: string) => ({
    matches: false,
    media: requete,
    onchange: null,
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => false,
  })) as unknown as typeof window.matchMedia;
}
