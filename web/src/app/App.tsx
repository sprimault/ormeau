// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { DecisionsError, DecisionsProvider } from '@/entities/decisions';
import { ArbitrageScreen } from '@/features/arbitrage';
import { ConnectionForm, ConnectionSummary, useConnection } from '@/features/connection';
import {
  CalquePreview,
  ExtractButton,
  ExtractionsProvider,
  useVersionCalque,
} from '@/features/extraction';
import { InventoryScreen } from '@/features/inventory';
import { useT } from '@/shared/i18n';
import { useOnglet, type Onglet } from '@/shared/lib';
import { ErrorBanner } from '@/shared/ui';
import { Header } from '@/widgets/header';

/**
 * Racine de l'interface.
 *
 * Le suivi des extractions l'enveloppe entière : il ouvre le seul flux
 * d'événements de la page, et survit au passage d'un écran à l'autre.
 */
export function App() {
  return (
    <ExtractionsProvider>
      <Ecran />
    </ExtractionsProvider>
  );
}

/**
 * Écran courant.
 *
 * Deux onglets, sélection et arbitrage, retenus dans le fragment d'adresse
 * plutôt que par un routeur. La session vit en mémoire et ne survit pas à un
 * rechargement ; l'arbitrage n'en a pas besoin : il désigne sa base par son nom
 * et travaille sur le calque du répertoire de travail. Recharger
 * `#arbitrage/gescom` le rouvre donc, hors ligne.
 *
 * Hauteur fixée à l'écran plutôt que laissée au contenu : les panneaux défilent
 * chacun de leur côté.
 */
function Ecran() {
  const { serveur, bases, enCours, erreur, ouvrir, fermer, changerBase } = useConnection();
  const [onglet, aller] = useOnglet();
  const arbitree = onglet.ecran === 'arbitrage' ? onglet.base : null;

  if (!serveur) {
    if (arbitree !== null) {
      return (
        <div className="flex h-screen flex-col overflow-hidden">
          <Header />
          <Onglets onglet={onglet} base={arbitree} aller={aller} />
          <DecisionsProvider key={arbitree} base={arbitree}>
            <DecisionsError />
            <Arbitrage base={arbitree} />
          </DecisionsProvider>
        </div>
      );
    }

    return (
      <div className="flex min-h-screen flex-col">
        <Header />
        <main className="flex flex-1 justify-center px-4 py-8">
          <ConnectionForm
            enCours={enCours}
            erreur={erreur}
            onConnecter={(requete) => void ouvrir(requete)}
          />
        </main>
      </div>
    );
  }

  const catalogue = serveur.catalogue;

  return (
    <div className="flex h-screen flex-col overflow-hidden">
      <Header />
      <div className="border-b border-slate-200 px-4 py-2 dark:border-slate-800">
        <ConnectionSummary serveur={serveur} onFermer={() => void fermer()} />
      </div>
      {erreur ? (
        <div className="px-4 py-2">
          <ErrorBanner message={erreur} />
        </div>
      ) : null}
      <Onglets onglet={onglet} base={arbitree ?? catalogue} aller={aller} />

      {/* Le brouillon de décisions suit la base : en ouvrir une autre repart de
          son propre fichier. Sélection et arbitrage de la même base partagent
          le même, et la sélection reste montée pendant l'arbitrage : les tables
          cochées ne se perdent pas d'un onglet à l'autre. */}
      <DecisionsProvider key={catalogue} base={catalogue}>
        <DecisionsError />
        <div hidden={arbitree !== null} className="flex min-h-0 flex-1 flex-col">
          {/* Remonté à chaque session : ouvrir une autre base repart d'une
              sélection vide, puisqu'elle désignait les tables de la précédente. */}
          <InventoryScreen
            key={serveur.session}
            serveur={serveur}
            bases={bases}
            enCours={enCours}
            onOuvrirBase={(base) => void changerBase(base)}
            actions={(portee) => (
              <ExtractButton session={serveur.session} base={catalogue} portee={portee} />
            )}
            produit={<CalquePreview session={serveur.session} base={catalogue} />}
          />
        </div>
        {arbitree === catalogue ? <Arbitrage base={catalogue} /> : null}
      </DecisionsProvider>

      {/* Une adresse peut désigner une autre base que celle de la connexion :
          l'arbitrage n'a pas besoin d'elle, il prend son propre brouillon. */}
      {arbitree !== null && arbitree !== catalogue ? (
        <DecisionsProvider key={arbitree} base={arbitree}>
          <DecisionsError />
          <Arbitrage base={arbitree} />
        </DecisionsProvider>
      ) : null}
    </div>
  );
}

/** Propriétés de la barre d'onglets. */
interface ProprietesOnglets {
  onglet: Onglet;
  /** La base que l'onglet d'arbitrage ouvre. */
  base: string;
  aller: (onglet: Onglet) => void;
}

/** Barre des deux écrans. */
function Onglets({ onglet, base, aller }: ProprietesOnglets) {
  const t = useT();
  const classe = (actif: boolean) =>
    `-mb-px border-b-2 px-3 py-1.5 text-sm ${
      actif
        ? 'border-ormeau-600 font-medium'
        : 'border-transparent text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'
    }`;

  return (
    <nav
      role="tablist"
      aria-label={t('tabs.label')}
      className="flex gap-1 border-b border-slate-200 px-4 dark:border-slate-800"
    >
      <button
        type="button"
        role="tab"
        aria-selected={onglet.ecran === 'selection'}
        className={classe(onglet.ecran === 'selection')}
        onClick={() => aller({ ecran: 'selection' })}
      >
        {t('tabs.selection')}
      </button>
      <button
        type="button"
        role="tab"
        aria-selected={onglet.ecran === 'arbitrage'}
        className={classe(onglet.ecran === 'arbitrage')}
        onClick={() => aller({ ecran: 'arbitrage', base })}
      >
        {t('tabs.arbitrage', { base })}
      </button>
    </nav>
  );
}

/** Arbitrage d'une base, relancé quand une extraction de cette base se termine. */
function Arbitrage({ base }: { base: string }) {
  const versionCalque = useVersionCalque(base);
  return <ArbitrageScreen base={base} versionCalque={versionCalque} />;
}
