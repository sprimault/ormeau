// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';

import { estCle, useT } from '@/shared/i18n';
import { duree } from '@/shared/lib';
import {
  EtatEchouee,
  EtatEnAttente,
  EtatEnCours,
  EtatTerminee,
  type Avancement,
  type Extraction,
} from '@/shared/model';
import { retirerExtraction } from '../api/extractionApi';
import { estTerminal, secondesEcoulees } from '../model/taches';

/** Fonction de traduction, telle que useT la rend. */
type Traduire = ReturnType<typeof useT>;

/** Propriétés d'une ligne. */
interface ProprietesLigne {
  extraction: Extraction;
  maintenant: number;
}

/** Couleur de l'état, affiché en bout de ligne. */
const couleurs: Record<string, string> = {
  [EtatEnAttente]: 'text-slate-500',
  [EtatEnCours]: 'text-ormeau-600 dark:text-ormeau-500',
  [EtatTerminee]: 'text-slate-600 dark:text-slate-300',
  [EtatEchouee]: 'text-red-700 dark:text-red-400',
};

/**
 * Une extraction du panneau.
 *
 * Ce qui s'affiche dépend de l'état : les paliers et l'étape pendant
 * l'extraction, le fichier écrit et ses anomalies ensuite, le message du
 * serveur en cas d'échec. Base, fichier et cibles d'anomalie restent tels
 * qu'ils sont, jamais traduits.
 */
export function ExtractionItem({ extraction, maintenant }: ProprietesLigne) {
  const t = useT();
  const [anomaliesOuvertes, setAnomaliesOuvertes] = useState(false);

  const active = !estTerminal(extraction.etat);
  const secondes = secondesEcoulees(extraction, maintenant);
  const { avancement, resultat } = extraction;
  const anomalies = resultat?.anomalies ?? [];

  return (
    <li className="space-y-1 px-3 py-2 text-xs">
      <div className="flex items-baseline gap-2">
        <span className="font-mono text-sm font-semibold">{extraction.base}</span>
        <span className="text-slate-500">
          {extraction.nb_tables === 0
            ? t('extraction.scope.all')
            : t('extraction.scope.tables', { n: extraction.nb_tables })}
        </span>
        <span className={`ml-auto ${couleurs[extraction.etat] ?? 'text-slate-500'}`}>
          {libelle(t, 'extraction.state.', extraction.etat)}
        </span>
        {secondes === null ? null : (
          <span className="text-slate-500 tabular-nums">{duree(secondes)}</span>
        )}
        <button
          type="button"
          // Aucun message en cas d'échec : le flux dit ce qu'il en est vraiment,
          // et la ligne reflète l'état du serveur, pas celui de la requête.
          onClick={() => void retirerExtraction(extraction.id).catch(() => undefined)}
          className="rounded border border-slate-300 px-1.5 text-slate-700 hover:bg-slate-100 dark:border-slate-700 dark:text-slate-200 dark:hover:bg-slate-800"
        >
          {active ? t('extraction.cancel') : t('extraction.dismiss')}
        </button>
      </div>

      {active ? <Paliers avancement={avancement} /> : null}

      {extraction.etat === EtatEnAttente ? (
        <p className="text-slate-500">{t('extraction.waiting')}</p>
      ) : null}

      {extraction.etat === EtatEnCours && avancement ? (
        <p className="text-slate-600 dark:text-slate-400">
          {t('extraction.progress', {
            etape: libelle(t, 'extraction.step.', avancement.etape),
            rang: avancement.rang,
            total: avancement.total,
          })}
        </p>
      ) : null}

      {extraction.etat === EtatTerminee && resultat ? (
        <div className="space-y-1">
          <p>
            <span className="font-mono">{extraction.fichier}</span>{' '}
            <span className="text-slate-500">
              {t('extraction.result', { tables: resultat.tables, colonnes: resultat.colonnes })}
            </span>
          </p>
          {anomalies.length > 0 ? (
            <button
              type="button"
              aria-expanded={anomaliesOuvertes}
              onClick={() => setAnomaliesOuvertes((precedent) => !precedent)}
              className="text-amber-700 dark:text-amber-500"
            >
              {anomaliesOuvertes ? '▾' : '▸'} {t('extraction.anomalies', { n: anomalies.length })}
            </button>
          ) : null}
          {anomaliesOuvertes ? (
            <ul className="space-y-1 border-l border-amber-300 pl-2 dark:border-amber-800">
              {anomalies.map((anomalie, rang) => (
                <li key={`${anomalie.code}:${anomalie.cible}:${rang}`}>
                  <span>{libelle(t, 'anomaly.', anomalie.code)}</span>{' '}
                  <span className="font-mono text-slate-500">{anomalie.cible}</span>
                  <p className="text-slate-500">{anomalie.message}</p>
                </li>
              ))}
            </ul>
          ) : null}
        </div>
      ) : null}

      {extraction.etat === EtatEchouee && extraction.erreur ? (
        <p className="text-red-700 dark:text-red-400">{extraction.erreur}</p>
      ) : null}
    </li>
  );
}

/** Propriétés des paliers. */
interface ProprietesPaliers {
  avancement?: Avancement;
}

/**
 * Une case par passe.
 *
 * Des cases et non une barre continue : l'avancement se compte en passes, et
 * rien ne dit où en est le serveur à l'intérieur de l'une d'elles. Une barre
 * lissée le prétendrait. Avant la première passe — en attente, ou connexion en
 * cours d'ouverture —, une bande neutre.
 */
function Paliers({ avancement }: ProprietesPaliers) {
  const t = useT();

  if (!avancement) {
    return <div className="h-1.5 rounded-sm bg-slate-200 dark:bg-slate-800" />;
  }

  const faites = avancement.rang - 1;
  return (
    <div
      role="progressbar"
      aria-label={t('extraction.progressLabel')}
      aria-valuemin={0}
      aria-valuemax={avancement.total}
      aria-valuenow={faites}
      className="flex gap-0.5"
    >
      {Array.from({ length: avancement.total }, (_, rang) => (
        <span
          key={rang}
          className={`h-1.5 flex-1 rounded-sm ${
            rang < faites
              ? 'bg-ormeau-600'
              : rang === faites
                ? 'animate-pulse bg-ormeau-500'
                : 'bg-slate-200 dark:bg-slate-800'
          }`}
        />
      ))}
    </div>
  );
}

/**
 * Traduit un code venu du serveur, ou l'affiche tel quel s'il n'a pas de
 * libellé : un code ajouté côté Go sans sa traduction doit se voir, pas
 * s'afficher vide.
 */
function libelle(t: Traduire, prefixe: string, code: string): string {
  const cle = prefixe + code;
  return estCle(cle) ? t(cle) : code;
}
