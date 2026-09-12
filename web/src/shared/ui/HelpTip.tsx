// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useId, useRef, useState } from 'react';
import { CircleHelp } from 'lucide-react';

import { useT } from '@/shared/i18n';

/** Largeur de la bulle, en pixels — celle de la classe w-72. */
const LARGEUR = 288;

/** Marge gardée avec le bord de la fenêtre. */
const MARGE = 8;

/** Où la bulle est posée, en coordonnées de fenêtre. */
interface Position {
  haut: number;
  gauche: number;
}

/**
 * Petite aide posée à côté d'un libellé : un point d'interrogation qui dit, au
 * survol ou au focus clavier, à quoi sert ce qu'il accompagne.
 *
 * Une bulle et non l'attribut `title` : celui-ci n'apparaît qu'après une
 * seconde d'immobilité, jamais au clavier, et se tronque sur une phrase longue.
 *
 * Positionnée en `fixed`, coordonnées calculées à l'ouverture. En `absolute`,
 * elle était découpée par le premier parent défilant — l'arbre des tables la
 * rendait illisible dès que le panneau de gauche était étroit, et le texte le
 * plus utile de l'écran ne se lisait pas.
 */
export function HelpTip({ texte }: { texte: string }) {
  const t = useT();
  const id = useId();
  const bouton = useRef<HTMLButtonElement>(null);
  const [position, setPosition] = useState<Position | null>(null);

  function ouvrir() {
    const cadre = bouton.current?.getBoundingClientRect();
    if (!cadre) {
      return;
    }
    // Débordement à droite rattrapé plutôt que laissé : une bulle à moitié
    // hors de l'écran ne vaut pas mieux qu'une bulle découpée.
    const gauche = Math.max(MARGE, Math.min(cadre.left, window.innerWidth - LARGEUR - MARGE));
    setPosition({ haut: cadre.bottom + 4, gauche });
  }

  return (
    <span className="inline-flex align-middle font-normal">
      <button
        ref={bouton}
        type="button"
        aria-label={t('help.label')}
        aria-describedby={id}
        onPointerEnter={ouvrir}
        onPointerLeave={() => setPosition(null)}
        onFocus={ouvrir}
        onBlur={() => setPosition(null)}
        className="rounded-full text-slate-400 hover:text-slate-700 focus:text-slate-700 focus:outline-none dark:hover:text-slate-200 dark:focus:text-slate-200"
      >
        <CircleHelp size={14} aria-hidden />
      </button>
      {/* Toujours rendu, jamais démonté : c'est lui que `aria-describedby`
          désigne, et un lecteur d'écran doit pouvoir le lire sans survol. */}
      <span
        id={id}
        role="tooltip"
        style={
          position
            ? { position: 'fixed', top: position.haut, left: position.gauche, width: LARGEUR }
            : undefined
        }
        className={`pointer-events-none z-50 rounded border border-slate-200 bg-white p-2 text-xs text-slate-700 shadow-lg dark:border-slate-700 dark:bg-slate-900 dark:text-slate-200 ${
          position ? '' : 'sr-only'
        }`}
      >
        {texte}
      </span>
    </span>
  );
}
