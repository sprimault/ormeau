// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwind from '@tailwindcss/vite';

/** Port du backend Go en développement, aligné sur ORMEAU_DEV_PORT du Makefile. */
const portBackend = process.env.ORMEAU_DEV_PORT ?? '7777';

/** Origine que le backend attend sur ses points d'entrée d'API. */
const origineBackend = `http://127.0.0.1:${portBackend}`;

/**
 * Configuration Vite de l'interface.
 *
 * La sortie va directement dans `internal/interface/embarque`, où `go:embed`
 * la lit : `embed` ne remonte pas au-dessus du répertoire de son paquet, et une
 * copie intermédiaire serait une étape de plus à oublier — l'oubli produisant
 * un binaire à interface blanche que rien ne signale à la compilation.
 *
 * Aucune URL absolue côté front : le port du binaire est dynamique par
 * construction, tous les appels sont relatifs et passent par ce proxy en
 * développement.
 *
 * Deux réglages du proxy ne vont pas de soi. `/entrer` y passe comme `/api`,
 * sans quoi le cookie de session se poserait sur le port du backend et ne
 * reviendrait jamais depuis celui du HMR. Et l'en-tête `Origin` est réécrit :
 * le navigateur annonce 5173, le backend n'accepte que la sienne, et sans cette
 * réécriture tous les appels seraient refusés en développement seulement.
 *
 * En pratique : `make dev` imprime l'URL jetonnée du backend, à rouvrir sur
 * 5173 — même chemin, même jeton, l'autre port.
 */
export default defineConfig({
  plugins: [react(), tailwind()],
  resolve: {
    alias: {
      '@': new URL('./src', import.meta.url).pathname,
    },
  },
  server: {
    host: '127.0.0.1',
    port: 5173,
    proxy: {
      '/api': { target: origineBackend, headers: { Origin: origineBackend } },
      '/entrer': { target: origineBackend, headers: { Origin: origineBackend } },
    },
  },
  build: {
    outDir: '../internal/interface/embarque',
    emptyOutDir: true,
  },
});
