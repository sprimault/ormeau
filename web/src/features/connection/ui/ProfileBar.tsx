// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';

import { useT } from '@/shared/i18n';
import type { Profil, ProfilResume } from '@/shared/model';
import { Button, ErrorBanner, Field } from '@/shared/ui';

/** Ce que la barre reçoit du formulaire. */
interface ProprietesProfileBar {
  profils: ProfilResume[];
  avertissement: string;
  erreur: string | null;
  /** Nom du profil choisi, vide pour une connexion neuve. */
  choisi: string;
  onChoisir: (nom: string) => void;
  /** Ce que le formulaire porte, à enregistrer sous le nom donné. */
  aEnregistrer: () => Omit<Profil, 'nom'>;
  onEnregistrer: (profil: Profil, remplacer: boolean, avecMotDePasse: boolean) => Promise<boolean>;
  onSupprimer: (nom: string) => void;
  /** Vrai quand le profil choisi porte déjà un mot de passe. */
  motDePasseEnregistre: boolean;
}

/**
 * Choix d'une connexion enregistrée, au-dessus du formulaire.
 *
 * Le sélecteur n'a **pas de sélection par défaut** : s'il pré-choisissait le
 * premier profil, l'écran se remplirait tout seul et quelqu'un qui voulait
 * taper une connexion neuve devrait d'abord comprendre pourquoi les champs sont
 * pleins. La connexion ponctuelle reste le cas le plus fréquent sur un outil de
 * reprise.
 *
 * Le formulaire ne disparaît pas quand des profils existent : un écran qui
 * change de visage après un premier enregistrement donne l'impression d'avoir
 * déclenché autre chose que ce qu'on voulait.
 */
export function ProfileBar({
  profils,
  avertissement,
  erreur,
  choisi,
  onChoisir,
  aEnregistrer,
  onEnregistrer,
  onSupprimer,
  motDePasseEnregistre,
}: ProprietesProfileBar) {
  const t = useT();
  const [nom, setNom] = useState<string | null>(null);
  const [aConfirmer, setAConfirmer] = useState(false);
  const [avecMotDePasse, setAvecMotDePasse] = useState(false);

  function ouvrir() {
    setNom(choisi);
    setAConfirmer(false);
    // Jamais cochée d'elle-même, pas même sur un profil qui porte déjà un mot
    // de passe : enregistrer un secret est une action, et une action se décide.
    setAvecMotDePasse(false);
  }

  function fermer() {
    setNom(null);
    setAConfirmer(false);
  }

  async function enregistrer(remplacer: boolean) {
    if (nom === null) {
      return;
    }
    const pris = await onEnregistrer({ ...aEnregistrer(), nom }, remplacer, avecMotDePasse);
    if (pris) {
      onChoisir(nom);
      fermer();
      return;
    }
    // Refusé parce que le nom est déjà pris : on demande plutôt que d'écraser
    // en silence un profil de production.
    setAConfirmer(true);
  }

  return (
    <div className="flex flex-col gap-2 rounded border border-slate-200 p-3 dark:border-slate-800">
      <div className="flex items-end gap-2">
        <label className="flex flex-1 flex-col gap-1">
          <span className="text-xs font-medium text-slate-600 dark:text-slate-400">
            {t('profile.label')}
          </span>
          <select
            value={choisi}
            onChange={(evenement) => onChoisir(evenement.target.value)}
            className="rounded border border-slate-300 bg-white px-2 py-1.5 text-sm dark:border-slate-700 dark:bg-slate-900"
          >
            <option value="">{t('profile.none')}</option>
            {profils.map(({ profil }) => (
              <option key={profil.nom} value={profil.nom}>
                {profil.nom}
              </option>
            ))}
          </select>
        </label>

        {nom === null ? (
          <>
            <button
              type="button"
              onClick={ouvrir}
              className="py-1.5 text-sm underline underline-offset-2"
            >
              {t('profile.save')}
            </button>
            {choisi ? (
              <button
                type="button"
                onClick={() => onSupprimer(choisi)}
                className="py-1.5 text-sm text-slate-500 underline underline-offset-2"
              >
                {t('profile.delete')}
              </button>
            ) : null}
          </>
        ) : null}
      </div>

      {nom !== null ? (
        <div className="flex flex-col gap-2">
          <div className="flex items-end gap-2">
            <div className="flex-1">
              <Field
                label={t('profile.save.name')}
                value={nom}
                autoFocus
                onChange={(evenement) => {
                  setNom(evenement.target.value);
                  setAConfirmer(false);
                }}
                onKeyDown={(evenement) => evenement.key === 'Escape' && fermer()}
              />
            </div>
            <Button type="button" disabled={!nom.trim()} onClick={() => void enregistrer(false)}>
              {t('profile.save.confirm')}
            </Button>
            <button
              type="button"
              onClick={fermer}
              className="py-1.5 text-sm text-slate-500 underline underline-offset-2"
            >
              {t('profile.save.cancel')}
            </button>
          </div>

          {/* La case vit ici et non dans le formulaire de connexion : elle
              n'agit qu'à l'enregistrement, et placée là-bas elle laissait croire
              qu'elle faisait quelque chose au moment de se connecter. */}
          <label className="flex items-start gap-2 text-sm">
            <input
              type="checkbox"
              checked={avecMotDePasse}
              onChange={(evenement) => setAvecMotDePasse(evenement.target.checked)}
              className="mt-1"
            />
            <span>
              {t('profile.password.save')}
              <span className="block text-xs text-slate-500 dark:text-slate-400">
                {t('profile.password.save.hint')}
              </span>
            </span>
          </label>

          {motDePasseEnregistre && !avecMotDePasse ? (
            <p className="text-xs text-amber-700 dark:text-amber-500">
              {t('profile.password.erase')}
            </p>
          ) : null}

          {aConfirmer ? (
            <div
              role="alert"
              className="flex items-center gap-3 rounded bg-amber-50 p-2 text-sm dark:bg-amber-950"
            >
              <span className="flex-1">{t('profile.replace')}</span>
              <Button type="button" onClick={() => void enregistrer(true)}>
                {t('profile.replace.confirm')}
              </Button>
            </div>
          ) : null}
        </div>
      ) : null}

      {avertissement ? (
        <p className="text-xs text-amber-700 dark:text-amber-500">{avertissement}</p>
      ) : null}
      {erreur ? <ErrorBanner message={erreur} /> : null}
    </div>
  );
}
