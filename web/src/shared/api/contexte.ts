// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useEffect, useState } from 'react';

import { getJSON } from './client';
import type { ReponseContexte } from '@/shared/model';

/**
 * Lit le répertoire de travail et la version du binaire.
 *
 * Transverse plutôt que rattaché à un écran : le répertoire est affiché en
 * permanence, parce qu'un fichier de décisions posé dans le mauvais projet ne
 * se remarque pas tout de suite.
 */
export function useContexte(): ReponseContexte | null {
  const [contexte, setContexte] = useState<ReponseContexte | null>(null);

  useEffect(() => {
    const abandon = new AbortController();
    getJSON<ReponseContexte>('/api/contexte', abandon.signal)
      .then(setContexte)
      .catch(() => {
        // L'en-tête se passe de ces deux valeurs ; l'écran principal signalera
        // l'indisponibilité du serveur local s'il y a lieu.
      });
    return () => abandon.abort();
  }, []);

  return contexte;
}
