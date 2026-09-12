// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useRef, useState, type ReactNode } from 'react';

import { useT } from '@/shared/i18n';
import { usePreferencesStore, type ClePreference } from '@/shared/model';

/** Sens du partage : côte à côte, ou l'un au-dessus de l'autre. */
type Sens = 'horizontal' | 'vertical';

/**
 * Bornes du panneau réglé, en pixels. Côte à côte, en deçà de 240 les noms de
 * tables sont illisibles ; empilés, en deçà de 120 il ne reste qu'une barre. Au
 * delà de 900, l'autre panneau n'a plus la place d'exister.
 */
const BORNES: Record<Sens, { min: number; max: number }> = {
  horizontal: { min: 240, max: 900 },
  vertical: { min: 120, max: 900 },
};

/** Pas du redimensionnement au clavier. */
const PAS = 24;

/** Propriétés du séparateur. */
interface ProprietesSplit {
  premier: ReactNode;
  second: ReactNode;
  sens?: Sens;
  /** Préférence qui porte la taille. Une préférence d'affichage, au même
   *  titre que le thème — jamais rien qui touche à la base. */
  clePreference: ClePreference;
  defaut?: number;
  /** Rend au panneau réglé sa taille naturelle et retire la poignée. Les deux
   *  panneaux restent montés : replier ne perd ni une recherche en cours ni un
   *  défilement. */
  replie?: boolean;
}

/**
 * Deux panneaux séparés par une poignée que l'on tire.
 *
 * Côte à côte, c'est le premier qu'on règle — l'arbre, à gauche. Empilés, c'est
 * le second — l'aperçu, en bas. Dans les deux cas, celui dont on choisit la
 * taille ; l'autre prend le reste.
 *
 * La taille est persistée : on la règle une fois pour son écran, et la
 * retrouver à chaque lancement est ce qui distingue un outil qu'on utilise
 * d'une démonstration.
 *
 * Le séparateur est atteignable au clavier et porte les attributs d'un
 * `separator` ARIA : à la souris seule, il est inutilisable pour qui n'en a pas.
 */
export function SplitPane({
  premier,
  second,
  sens = 'horizontal',
  clePreference,
  defaut = 384,
  replie = false,
}: ProprietesSplit) {
  const t = useT();
  const vertical = sens === 'vertical';
  const { min, max } = BORNES[sens];
  const conteneur = useRef<HTMLDivElement>(null);
  const [taille, setTaille] = useState(() => tailleInjectee(clePreference, defaut, min, max));
  const [glisse, setGlisse] = useState(false);

  const poser = useCallback(
    (valeur: number) => {
      const bornee = Math.min(max, Math.max(min, Math.round(valeur)));
      setTaille(bornee);
      usePreferencesStore.getState().regler({ [clePreference]: bornee });
    },
    [clePreference, min, max],
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
        // Empilés, le panneau réglé est celui du bas : sa hauteur se mesure
        // depuis le bord inférieur.
        poser(vertical ? cadre.bottom - evenement.clientY : evenement.clientX - cadre.left);
      }
    }
    function relacher() {
      setGlisse(false);
    }

    window.addEventListener('pointermove', deplacer);
    window.addEventListener('pointerup', relacher);
    // Sans cela, le glissement sélectionne le texte des deux panneaux.
    document.body.style.userSelect = 'none';
    document.body.style.cursor = vertical ? 'row-resize' : 'col-resize';

    return () => {
      window.removeEventListener('pointermove', deplacer);
      window.removeEventListener('pointerup', relacher);
      document.body.style.userSelect = '';
      document.body.style.cursor = '';
    };
  }, [glisse, poser, vertical]);

  const poignee = replie ? null : (
    <div
      role="separator"
      aria-orientation={vertical ? 'horizontal' : 'vertical'}
      aria-label={t(vertical ? 'split.handleVertical' : 'split.handle')}
      aria-valuenow={taille}
      aria-valuemin={min}
      aria-valuemax={max}
      tabIndex={0}
      onPointerDown={() => setGlisse(true)}
      onDoubleClick={() => poser(defaut)}
      onKeyDown={(evenement) => {
        if (evenement.key === (vertical ? 'ArrowUp' : 'ArrowRight')) {
          poser(taille + PAS);
        } else if (evenement.key === (vertical ? 'ArrowDown' : 'ArrowLeft')) {
          poser(taille - PAS);
        }
      }}
      className={`hover:bg-ormeau-500 focus:bg-ormeau-500 shrink-0 border-slate-200 transition-colors focus:outline-none dark:border-slate-800 ${
        vertical ? 'h-1 cursor-row-resize border-y' : 'w-1 cursor-col-resize border-x'
      } ${glisse ? 'bg-ormeau-500' : 'bg-slate-100 dark:bg-slate-900'}`}
    />
  );

  if (vertical) {
    return (
      <div ref={conteneur} className="flex min-h-0 flex-1 flex-col">
        <div className="flex min-h-0 flex-1 flex-col">{premier}</div>
        {poignee}
        {/* Plafonnée à 80 % du conteneur : une hauteur retenue sur un grand
            écran ne doit pas écraser l'arbre sur un petit. */}
        <div
          style={replie ? undefined : { height: taille }}
          className="flex max-h-[80%] min-h-0 shrink-0 flex-col"
        >
          {second}
        </div>
      </div>
    );
  }

  return (
    <div ref={conteneur} className="flex min-h-0 flex-1">
      <div style={{ width: taille }} className="flex min-w-0 shrink-0 flex-col">
        {premier}
      </div>
      {poignee}
      <div className="min-w-0 flex-1 overflow-y-auto">{second}</div>
    </div>
  );
}

/**
 * Relit la taille que le serveur a injectée, en écartant ce qui n'est pas
 * exploitable.
 *
 * La valeur arrive aussi en variable CSS, appliquée avant même que ce composant
 * ne soit monté : la mise en page est juste dès le premier octet, et cette
 * lecture ne sert qu'à ce que l'état React parte de la même valeur.
 */
function tailleInjectee(cle: ClePreference, defaut: number, min: number, max: number): number {
  const brut = usePreferencesStore.getState().preferences[cle];
  if (typeof brut === 'number' && brut >= min && brut <= max) {
    return brut;
  }
  return defaut;
}
