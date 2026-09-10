// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useRef, useState, type ReactNode } from 'react';

import { useT } from '@/shared/i18n';

/** Bornes du panneau, en pixels. En deçà, les noms de tables sont illisibles ;
 *  au-delà, le détail n'a plus la place d'exister. */
const MIN = 240;
const MAX = 900;

/** Pas du redimensionnement au clavier. */
const PAS = 24;

/** Propriétés du séparateur. */
interface ProprietesSplit {
  gauche: ReactNode;
  droite: ReactNode;
  /** Clé de persistance de la largeur. Une préférence d'affichage, au même
   *  titre que le thème — jamais rien qui touche à la base. */
  cleStockage: string;
  defaut?: number;
}

/**
 * Deux panneaux séparés par une poignée que l'on tire.
 *
 * La largeur est persistée : on la règle une fois pour son écran, et la
 * retrouver à chaque lancement est ce qui distingue un outil qu'on utilise
 * d'une démonstration.
 *
 * Le séparateur est atteignable au clavier et porte les attributs d'un
 * `separator` ARIA : à la souris seule, il est inutilisable pour qui n'en a pas.
 */
export function SplitPane({ gauche, droite, cleStockage, defaut = 384 }: ProprietesSplit) {
  const t = useT();
  const conteneur = useRef<HTMLDivElement>(null);
  const [largeur, setLargeur] = useState(() => largeurEnregistree(cleStockage, defaut));
  const [glisse, setGlisse] = useState(false);

  const poser = useCallback(
    (valeur: number) => {
      const bornee = Math.min(MAX, Math.max(MIN, Math.round(valeur)));
      setLargeur(bornee);
      try {
        window.localStorage.setItem(cleStockage, String(bornee));
      } catch {
        // Stockage indisponible : la largeur vaut pour cette session.
      }
    },
    [cleStockage],
  );

  // Le suivi est posé sur la fenêtre et non sur la poignée : le pointeur va
  // plus vite que le rendu, et sort du séparateur dès qu'on tire un peu franc.
  useEffect(() => {
    if (!glisse) {
      return;
    }
    function deplacer(evenement: PointerEvent) {
      const cadre = conteneur.current?.getBoundingClientRect();
      if (cadre) {
        poser(evenement.clientX - cadre.left);
      }
    }
    function relacher() {
      setGlisse(false);
    }

    window.addEventListener('pointermove', deplacer);
    window.addEventListener('pointerup', relacher);
    // Sans cela, le glissement sélectionne le texte des deux panneaux.
    document.body.style.userSelect = 'none';
    document.body.style.cursor = 'col-resize';

    return () => {
      window.removeEventListener('pointermove', deplacer);
      window.removeEventListener('pointerup', relacher);
      document.body.style.userSelect = '';
      document.body.style.cursor = '';
    };
  }, [glisse, poser]);

  return (
    <div ref={conteneur} className="flex min-h-0 flex-1">
      <div style={{ width: largeur }} className="flex min-w-0 shrink-0 flex-col">
        {gauche}
      </div>

      <div
        role="separator"
        aria-orientation="vertical"
        aria-label={t('split.handle')}
        aria-valuenow={largeur}
        aria-valuemin={MIN}
        aria-valuemax={MAX}
        tabIndex={0}
        onPointerDown={() => setGlisse(true)}
        onDoubleClick={() => poser(defaut)}
        onKeyDown={(evenement) => {
          if (evenement.key === 'ArrowLeft') {
            poser(largeur - PAS);
          } else if (evenement.key === 'ArrowRight') {
            poser(largeur + PAS);
          }
        }}
        className={`hover:bg-ormeau-500 focus:bg-ormeau-500 w-1 shrink-0 cursor-col-resize border-x border-slate-200 transition-colors focus:outline-none dark:border-slate-800 ${
          glisse ? 'bg-ormeau-500' : 'bg-slate-100 dark:bg-slate-900'
        }`}
      />

      <div className="min-w-0 flex-1 overflow-y-auto">{droite}</div>
    </div>
  );
}

/** Relit la largeur choisie, en écartant ce qui n'est pas exploitable. */
function largeurEnregistree(cle: string, defaut: number): number {
  try {
    const brut = Number(window.localStorage.getItem(cle));
    if (Number.isFinite(brut) && brut >= MIN && brut <= MAX) {
      return brut;
    }
  } catch {
    // Stockage indisponible : le défaut s'applique.
  }
  return defaut;
}
