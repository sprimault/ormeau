// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { ConnectionForm, ConnectionSummary, useConnection } from '@/features/connection';
import { InventoryScreen } from '@/features/inventory';
import { ErrorBanner } from '@/shared/ui';
import { Header } from '@/widgets/header';

/**
 * Racine de l'interface.
 *
 * Pas de routeur tant que l'écran affiché découle de l'état de la connexion :
 * la session vit en mémoire, donc recharger une adresse profonde ramènerait au
 * formulaire de toute façon. Il viendra avec l'arbitrage, qui aura de vraies
 * adresses à partager.
 *
 * Hauteur fixée à l'écran plutôt que laissée au contenu : l'arbre et le détail
 * défilent chacun de leur côté, et la portée reste visible en bas.
 */
export function App() {
  const { serveur, bases, enCours, erreur, ouvrir, fermer, changerBase } = useConnection();

  if (!serveur) {
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
      {/* Remonté à chaque session : ouvrir une autre base repart d'une sélection
          vide, puisqu'elle désignait les tables de la précédente. */}
      <InventoryScreen
        key={serveur.session}
        serveur={serveur}
        bases={bases}
        enCours={enCours}
        onOuvrirBase={(base) => void changerBase(base)}
      />
    </div>
  );
}
