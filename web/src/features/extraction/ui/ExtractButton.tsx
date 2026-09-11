// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';

import { ErreurAPI } from '@/shared/api';
import { translate, useT } from '@/shared/i18n';
import type { Portee } from '@/shared/model';
import { Button } from '@/shared/ui';
import { lancerExtraction } from '../api/extractionApi';
import { useExtractions } from '../model/contexte';
import { estTerminal, fichierCalque } from '../model/taches';

/** Propriétés du bouton. */
interface ProprietesBouton {
  session: string;
  /** Base de la session, qui nomme le calque. */
  base: string;
  portee: Portee;
}

/**
 * Lance l'extraction de la portée affichée, et rend la main aussitôt.
 *
 * Rien n'attend la fin : la tâche part en fond, son avancement se lit dans
 * l'en-tête, et l'écran reste disponible pour parcourir ou extraire une autre
 * base. Le bouton ne se bloque que pour la base dont une extraction tourne
 * déjà, puisque les deux écriraient le même fichier.
 */
export function ExtractButton({ session, base, portee }: ProprietesBouton) {
  const t = useT();
  const { extractions } = useExtractions();
  const [envoi, setEnvoi] = useState(false);
  const [erreur, setErreur] = useState<string | null>(null);

  const occupee = extractions.some((e) => e.base === base && !estTerminal(e.etat));

  async function extraire() {
    setEnvoi(true);
    setErreur(null);
    try {
      await lancerExtraction(session, portee);
    } catch (echec) {
      setErreur(echec instanceof ErreurAPI ? echec.message : translate('error.unknown'));
    } finally {
      setEnvoi(false);
    }
  }

  return (
    <div className="flex items-center gap-3">
      <Button onClick={() => void extraire()} disabled={envoi || occupee}>
        {occupee ? t('extraction.running') : t('extraction.launch')}
      </Button>
      <span className="font-mono text-xs text-slate-500" title={t('extraction.target')}>
        {fichierCalque(base)}
      </span>
      {erreur ? (
        <span role="alert" className="text-xs text-red-700 dark:text-red-400">
          {erreur}
        </span>
      ) : null}
    </div>
  );
}
