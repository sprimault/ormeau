// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useMemo, useState } from 'react';

import { useDecisions } from '@/entities/decisions';
import { useT } from '@/shared/i18n';
import { qualifier } from '@/shared/lib';
import { CodeCalqueModifie, CodeContenuManuel, CodeDecisionsModifiees } from '@/shared/model';
import { Button, ErrorBanner, SplitPane } from '@/shared/ui';
import { avertissementsParTable } from '../model/avertissements';
import { fichierDecisions } from '../model/decisions';
import { lignesEntites } from '../model/entites';
import { useEnregistrement } from '../model/useEnregistrement';
import { useEntite } from '../model/useEntite';
import { useInference } from '../model/useInference';
import { Alerte } from './Alerte';
import { EntityList } from './EntityList';
import { EntityPanel } from './EntityPanel';

/** Propriétés de l'écran. */
interface ProprietesEcran {
  base: string;
  /** Fin de la dernière extraction terminée de la base, qui relance l'inférence. */
  versionCalque: string;
}

/**
 * Écran d'arbitrage d'une base : la liste des entités, et l'entité ouverte avec
 * ses avertissements.
 *
 * Il ne demande aucune connexion : l'inférence se rejoue sur le calque du
 * répertoire de travail à chaque modification, et l'écran montre l'effet réel
 * des décisions. Sa seule sortie est le fichier de décisions.
 *
 * La première entité s'ouvre d'office : un panneau qui attend un clic pour
 * montrer quoi que ce soit n'apprend rien à qui découvre l'écran.
 */
export function ArbitrageScreen({ base, versionCalque }: ProprietesEcran) {
  const t = useT();
  const brouillon = useDecisions();
  const { decisions, fichier, modifie, modifier, relire } = brouillon;
  const inference = useInference(base, decisions, versionCalque);
  const resultat = inference.resultat;
  const empreinte = resultat?.empreinte_physique ?? '';
  const enregistrement = useEnregistrement(brouillon, empreinte);
  const [choisie, setChoisie] = useState<string | null>(null);

  const parTable = useMemo(
    () =>
      avertissementsParTable(
        resultat?.avertissements ?? [],
        (resultat?.entites ?? []).map((e) => qualifier(e.table.schema, e.table.nom)),
      ),
    [resultat],
  );
  const lignes = useMemo(() => lignesEntites(resultat?.entites ?? [], parTable), [resultat, parTable]);
  const active = lignes.find((ligne) => ligne.qualifiee === choisie) ?? lignes[0] ?? null;
  const entite = useEntite(base, inference.decisionsJugees, empreinte, active?.table ?? null);

  const calqueModifie = inference.calqueModifie || enregistrement.refus === CodeCalqueModifie;
  const nomFichier = fichierDecisions(base);
  const etat = modifie
    ? t('arbitrage.state.unsaved')
    : fichier?.existe
      ? t('arbitrage.state.saved')
      : t('arbitrage.state.absent');

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="border-b border-slate-200 px-4 py-2 dark:border-slate-800">
        {/* Pas de titre : l'onglet nomme déjà la base arbitrée. L'état et
            l'écriture viennent en tête, au-dessus de ce qu'ils enregistrent. */}
        <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
          <span
            className={`text-xs ${modifie ? 'text-amber-700 dark:text-amber-500' : 'text-slate-500'}`}
          >
            {etat}
          </span>
          <Button
            onClick={() => enregistrement.enregistrer()}
            disabled={empreinte === '' || enregistrement.enCours || calqueModifie}
          >
            {enregistrement.enCours
              ? t('arbitrage.saving')
              : t('arbitrage.save', { fichier: nomFichier })}
          </Button>
          {inference.enCours && resultat ? (
            <span className="text-xs text-slate-500">{t('arbitrage.inferring')}</span>
          ) : null}
        </div>
        <p className="mt-0.5 text-xs text-slate-500">{t('arbitrage.intro', { fichier: nomFichier })}</p>
      </div>

      {calqueModifie ? (
        <Alerte message={t('arbitrage.layerChanged')}>
          <Button
            variante="discret"
            onClick={() => {
              enregistrement.abandonner();
              inference.recharger();
            }}
          >
            {t('arbitrage.reload')}
          </Button>
        </Alerte>
      ) : null}
      {enregistrement.refus === CodeDecisionsModifiees ? (
        <Alerte message={t('arbitrage.fileChanged')}>
          <Button
            variante="discret"
            onClick={() => {
              enregistrement.abandonner();
              relire();
            }}
          >
            {t('arbitrage.reread')}
          </Button>
        </Alerte>
      ) : null}
      {enregistrement.refus === CodeContenuManuel ? (
        <Alerte message={t('arbitrage.manual')}>
          <Button onClick={() => enregistrement.enregistrer(true)}>{t('arbitrage.overwrite')}</Button>
          <Button variante="discret" onClick={enregistrement.abandonner}>
            {t('arbitrage.cancel')}
          </Button>
        </Alerte>
      ) : null}
      {enregistrement.erreur ? (
        <div className="px-4 pt-2">
          <ErrorBanner message={enregistrement.erreur} />
        </div>
      ) : null}
      {inference.erreur ? (
        <div className="px-4 pt-2">
          <ErrorBanner message={inference.erreur} />
        </div>
      ) : null}

      {resultat ? (
        <SplitPane
          cleStockage="ormeau-largeur-entites"
          defaut={300}
          premier={
            <EntityList lignes={lignes} active={active?.qualifiee ?? ''} onActiver={setChoisie} />
          }
          second={
            active ? (
              <EntityPanel
                ligne={active}
                etat={entite}
                entites={lignes}
                base={base}
                empreinte={empreinte}
                decisionsJugees={inference.decisionsJugees}
                decisions={decisions}
                avertissements={parTable.get(active.qualifiee) ?? []}
                proposition={resultat.propositions.find((p) => p.cible === active.qualifiee)}
                typesDoctrine={resultat.types_doctrine}
                enumerations={resultat.enumerations}
                onModifier={modifier}
              />
            ) : null
          }
        />
      ) : inference.erreur ? null : (
        <p className="px-4 py-3 text-sm text-slate-500">{t('arbitrage.loading')}</p>
      )}
    </div>
  );
}
