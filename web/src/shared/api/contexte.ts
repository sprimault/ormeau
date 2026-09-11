// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useState } from 'react';

import type { ReponseContexte, RequeteRepertoire } from '@/shared/model';

import { getJSON, postJSON } from './client';

/** Ce que l'en-tête tient : le contexte, et de quoi le changer. */
interface Contexte {
  contexte: ReponseContexte | null;
  /** Change le répertoire de travail. Rend le chemin résolu par le serveur. */
  changerRepertoire: (repertoire: string) => Promise<string>;
}

/**
 * Lit le répertoire de travail et la version du binaire.
 *
 * Transverse plutôt que rattaché à un écran : le répertoire est affiché en
 * permanence, parce qu'un fichier de décisions posé dans le mauvais projet ne
 * se remarque pas tout de suite.
 */
export function useContexte(): Contexte {
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

  const changerRepertoire = useCallback(async (repertoire: string) => {
    // Le contexte est repris de la réponse et non de la saisie : le serveur
    // résout les liens et les « .. », et c'est ce chemin-là qui recevra les
    // fichiers.
    const suivant = await postJSON<RequeteRepertoire, ReponseContexte>('/api/repertoire', {
      repertoire,
    });
    setContexte(suivant);
    return suivant.repertoire;
  }, []);

  return { contexte, changerRepertoire };
}
