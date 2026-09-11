// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useEffect, useState } from 'react';
import { CircleCheck, LoaderCircle, TriangleAlert } from 'lucide-react';

import { useT } from '@/shared/i18n';
import { EtatEchouee } from '@/shared/model';
import { useExtractions } from '../model/contexte';
import { estTerminal } from '../model/taches';
import { ExtractionPanel } from './ExtractionPanel';

/**
 * Indicateur des extractions, dans l'en-tête.
 *
 * Absent tant qu'aucune n'a été lancée : l'en-tête n'annonce pas ce qui n'existe
 * pas. Ensuite, un compte — celles qui tournent d'abord, sinon celles en échec,
 * sinon celles finies — et le détail au clic.
 *
 * Le nombre de tâches actives est repris dans le titre de l'onglet : on lance
 * une extraction longue, on passe à autre chose, et l'onglet dit quand c'est
 * fini.
 */
export function ExtractionIndicator() {
  const t = useT();
  const { extractions, connecte } = useExtractions();
  const [ouvert, setOuvert] = useState(false);

  const actives = extractions.filter((e) => !estTerminal(e.etat)).length;
  const echouees = extractions.filter((e) => e.etat === EtatEchouee).length;

  useEffect(() => {
    document.title = actives > 0 ? t('app.title.busy', { n: actives }) : t('app.name');
  }, [actives, t]);

  // Le panneau se referme quand la dernière tâche est retirée : sans cela, il
  // se rouvrirait tout seul au lancement suivant.
  useEffect(() => {
    if (extractions.length === 0) {
      setOuvert(false);
    }
  }, [extractions.length]);

  useEffect(() => {
    if (!ouvert) {
      return;
    }
    const fermer = (evenement: KeyboardEvent) => {
      if (evenement.key === 'Escape') {
        setOuvert(false);
      }
    };
    window.addEventListener('keydown', fermer);
    return () => window.removeEventListener('keydown', fermer);
  }, [ouvert]);

  if (extractions.length === 0) {
    return null;
  }

  let Icone = CircleCheck;
  let couleur = 'text-slate-500';
  let libelle = t('extraction.indicator.finished', { n: extractions.length });
  if (actives > 0) {
    Icone = LoaderCircle;
    couleur = 'animate-spin text-ormeau-600 dark:text-ormeau-500';
    libelle = t('extraction.indicator.running', { n: actives });
  } else if (echouees > 0) {
    Icone = TriangleAlert;
    couleur = 'text-red-600 dark:text-red-400';
    libelle = t('extraction.indicator.failed', { n: echouees });
  }

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOuvert((precedent) => !precedent)}
        aria-expanded={ouvert}
        aria-haspopup="dialog"
        className="flex items-center gap-1.5 rounded px-2 py-0.5 text-xs hover:bg-slate-100 dark:hover:bg-slate-800"
      >
        <Icone size={14} aria-hidden className={couleur} />
        <span>{libelle}</span>
      </button>

      {ouvert ? (
        <div className="absolute top-full right-0 z-20 mt-1">
          <ExtractionPanel extractions={extractions} connecte={connecte} />
        </div>
      ) : null}
    </div>
  );
}
