// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';

import { ErreurAPI } from '@/shared/api';
import { useT } from '@/shared/i18n';
import { tronquerMilieu } from '@/shared/lib';
import { CopyButton, HelpTip } from '@/shared/ui';

/** Ce que le champ reçoit de l'en-tête. */
interface ProprietesWorkdir {
  repertoire: string;
  /** Rend le chemin résolu par le serveur, ou rejette avec son refus. */
  changer: (repertoire: string) => Promise<string>;
}

/**
 * Le répertoire de travail, affiché en permanence et modifiable.
 *
 * Il est affiché tout le temps, et en entier au survol : on doit savoir où on
 * écrit avant de cliquer. Le texte est tronqué au milieu et ne se colle donc
 * nulle part — c'est le bouton de copie qui donne le chemin.
 *
 * La saisie ne s'ouvre qu'au clic sur « Changer », et se referme dès qu'elle a
 * abouti : un champ toujours visible dans un en-tête invite à taper dedans
 * alors que ce geste ne se fait qu'une fois par projet.
 *
 * Aucune boîte « Parcourir » : une page web n'a pas accès aux chemins réels du
 * poste, et un explorateur de dossiers maison ferait lister le disque à
 * l'interface. C'est le serveur qui valide ce qui est tapé.
 */
export function WorkdirField({ repertoire, changer }: ProprietesWorkdir) {
  const t = useT();
  const [saisie, setSaisie] = useState<string | null>(null);
  const [refus, setRefus] = useState<string | null>(null);
  const [encours, setEncours] = useState(false);

  function ouvrir() {
    setSaisie(repertoire);
    setRefus(null);
  }

  function fermer() {
    setSaisie(null);
    setRefus(null);
  }

  async function valider(evenement: React.FormEvent) {
    evenement.preventDefault();
    if (saisie === null || encours) {
      return;
    }

    setEncours(true);
    try {
      await changer(saisie);
      fermer();
    } catch (erreur) {
      // Le message vient du serveur : lui seul sait pourquoi le chemin ne
      // convient pas — absent, fermé en écriture, ou fichier plutôt que
      // répertoire.
      setRefus(erreur instanceof ErreurAPI ? erreur.message : t('error.unknown'));
    } finally {
      setEncours(false);
    }
  }

  if (saisie === null) {
    return (
      <span className="text-xs text-slate-500" title={repertoire}>
        {t('app.workdir')} <HelpTip texte={t('app.help.workdir')} /> :{' '}
        <span className="font-mono">{tronquerMilieu(repertoire)}</span>{' '}
        <CopyButton valeur={repertoire} libelle={t('app.workdir.copy')} />{' '}
        <button
          type="button"
          onClick={ouvrir}
          className="hover:text-ormeau-600 dark:hover:text-ormeau-400 underline underline-offset-2"
        >
          {t('app.workdir.change')}
        </button>
      </span>
    );
  }

  return (
    <form onSubmit={valider} className="flex items-baseline gap-2 text-xs">
      <label htmlFor="repertoire-travail" className="text-slate-500">
        {t('app.workdir')} :
      </label>
      <input
        id="repertoire-travail"
        value={saisie}
        onChange={(evenement) => setSaisie(evenement.target.value)}
        onKeyDown={(evenement) => evenement.key === 'Escape' && fermer()}
        autoFocus
        spellCheck={false}
        aria-invalid={refus !== null}
        aria-describedby={refus ? 'repertoire-refus' : undefined}
        className="w-96 rounded border border-slate-300 px-2 py-1 font-mono dark:border-slate-700 dark:bg-slate-900"
      />
      <button type="submit" disabled={encours} className="underline underline-offset-2">
        {t('app.workdir.apply')}
      </button>
      <button type="button" onClick={fermer} className="text-slate-500 underline underline-offset-2">
        {t('app.workdir.cancel')}
      </button>
      {refus ? (
        <span id="repertoire-refus" role="alert" className="text-red-600 dark:text-red-400">
          {refus}
        </span>
      ) : null}
    </form>
  );
}
