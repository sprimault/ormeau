// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { ConnectionForm, ConnectionSummary, useConnection } from '@/features/connection';
import { Header } from '@/widgets/header';

/**
 * Racine de l'interface.
 *
 * Pas de routeur tant qu'il n'y a qu'un écran : l'état de la connexion suffit à
 * choisir ce qui s'affiche. L'arbre des tables et l'écran d'arbitrage
 * l'introduiront, avec de vraies adresses à partager.
 */
export function App() {
  const { serveur, enCours, erreur, ouvrir, fermer } = useConnection();

  return (
    <div className="flex min-h-screen flex-col">
      <Header />
      <main className="flex flex-1 justify-center px-4 py-8">
        {serveur ? (
          <ConnectionSummary serveur={serveur} onFermer={() => void fermer()} />
        ) : (
          <ConnectionForm
            enCours={enCours}
            erreur={erreur}
            onConnecter={(requete) => void ouvrir(requete)}
          />
        )}
      </main>
    </div>
  );
}
