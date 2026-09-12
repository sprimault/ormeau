// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useT } from '@/shared/i18n';
import type { ProfilResume } from '@/shared/model';
import { Button, ErrorBanner, Field, HelpTip } from '@/shared/ui';

/** Ce que la barre reçoit du formulaire. */
interface ProprietesProfileBar {
  profils: ProfilResume[];
  avertissement: string;
  erreur: string | null;
  /** Nom du profil choisi, vide pour une connexion neuve. */
  choisi: string;
  onChoisir: (nom: string) => void;
  /** Nom sous lequel enregistrer à la connexion ; vide pour ne rien enregistrer. */
  nom: string;
  onNommer: (nom: string) => void;
  /** Le mot de passe part-il dans le profil. */
  avecMotDePasse: boolean;
  onAvecMotDePasse: (avec: boolean) => void;
  /** Le formulaire dit autre chose que le profil choisi. */
  divergent: boolean;
  onMettreAJour: () => void;
  onSupprimer: (nom: string) => void;
}

/**
 * Choix d'une connexion enregistrée, au-dessus du formulaire.
 *
 * **Un seul geste : remplir, nommer, se connecter.** Il n'y a pas de bouton
 * d'enregistrement — un nom saisi suffit, et le profil est créé une fois la
 * connexion réussie. On ne se demande donc ni sur quel bouton cliquer, ni dans
 * quel ordre, et un profil enregistré est forcément un profil qui marche.
 *
 * Le sélecteur n'a **pas de sélection par défaut** : s'il pré-choisissait le
 * premier profil, l'écran se remplirait tout seul et quelqu'un qui voulait
 * taper une connexion neuve devrait d'abord comprendre pourquoi les champs sont
 * pleins. La connexion ponctuelle reste le cas le plus fréquent sur un outil de
 * reprise.
 */
export function ProfileBar({
  profils,
  avertissement,
  erreur,
  choisi,
  onChoisir,
  nom,
  onNommer,
  avecMotDePasse,
  onAvecMotDePasse,
  divergent,
  onMettreAJour,
  onSupprimer,
}: ProprietesProfileBar) {
  const t = useT();
  const dejaPris = nom.trim() !== '' && profils.some(({ profil }) => profil.nom === nom.trim());

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

        {choisi ? (
          <>
            {/* Sans lui, changer un port obligerait à supprimer puis recréer.
                Caché tant que le formulaire dit la même chose que le profil :
                un bouton qui ne ferait rien n'a pas à être là. */}
            {divergent ? (
              <Button type="button" variante="discret" onClick={onMettreAJour}>
                {t('profile.update')}
              </Button>
            ) : null}
            <button
              type="button"
              onClick={() => onSupprimer(choisi)}
              className="py-1.5 text-sm text-slate-500 underline underline-offset-2"
            >
              {t('profile.delete')}
            </button>
          </>
        ) : null}
      </div>

      {/* Le champ de nom n'existe que pour une connexion neuve : un profil déjà
          choisi n'a pas à être réenregistré, et le faire à chaque connexion
          effacerait son mot de passe quand la case est décochée. */}
      {choisi ? null : (
        <>
          <Field
            label={t('profile.save.name')}
            value={nom}
            onChange={(evenement) => onNommer(evenement.target.value)}
            aide={
              <span className="inline-flex items-center gap-1">
                {dejaPris ? t('profile.name.taken') : t('profile.name.optional')}
                <HelpTip texte={t('profile.help.name')} />
              </span>
            }
            aria-invalid={dejaPris}
          />

          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={avecMotDePasse}
              onChange={(evenement) => onAvecMotDePasse(evenement.target.checked)}
              disabled={nom.trim() === ''}
              className="disabled:opacity-50"
            />
            <span className={nom.trim() === '' ? 'text-slate-400 dark:text-slate-600' : undefined}>
              {t('profile.password.save')}
            </span>
            <HelpTip texte={t('profile.password.save.hint')} />
          </label>
        </>
      )}

      {avertissement ? (
        <p className="text-xs text-amber-700 dark:text-amber-500">{avertissement}</p>
      ) : null}
      {erreur ? <ErrorBanner message={erreur} /> : null}
    </div>
  );
}
