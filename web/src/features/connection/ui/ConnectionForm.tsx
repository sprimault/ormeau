// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useState, type FormEvent } from 'react';

import { useT } from '@/shared/i18n';
import { Button, ErrorBanner, Field } from '@/shared/ui';
import type { Profil, RequeteConnexion } from '@/shared/model';

import { useProfils } from '../model/useProfils';
import { ProfileBar } from './ProfileBar';

/**
 * SGBD proposés au choix explicite.
 *
 * Seuls ceux dont le pilote est embarqué y figurent : offrir une option qui
 * échouera à la connexion ferait passer un manque annoncé pour une panne. La
 * liste grandit avec les pilotes.
 */
const sgbdDisponibles = ['postgres'];

/** Les deux formes de saisie. */
type Mode = 'composants' | 'dsn';

/** Propriétés du formulaire. */
interface ProprietesFormulaire {
  enCours: boolean;
  erreur: string | null;
  onConnecter: (requete: RequeteConnexion) => void;
}

/**
 * Formulaire de connexion.
 *
 * Les composants sont le mode par défaut : personne ne tape une URL dans un
 * formulaire, et celui qui découvre une base connaît un hôte et un identifiant,
 * pas une chaîne de connexion. Le SGBD n'est pas obligatoire — le port le
 * désigne, et le serveur dira ensuite ce qu'il est vraiment.
 */
export function ConnectionForm({ enCours, erreur, onConnecter }: ProprietesFormulaire) {
  const t = useT();
  const [mode, setMode] = useState<Mode>('composants');
  const [sgbd, setSgbd] = useState('');
  const [hote, setHote] = useState('127.0.0.1');
  const [port, setPort] = useState('5432');
  const [utilisateur, setUtilisateur] = useState('');
  const [motDePasse, setMotDePasse] = useState('');
  const [base, setBase] = useState('');
  const [dsn, setDsn] = useState('');
  const [profilChoisi, setProfilChoisi] = useState('');
  const [motDePasseDuProfil, setMotDePasseDuProfil] = useState(false);
  const { profils, avertissement, enregistrer, retirer, erreur: erreurProfil } = useProfils();

  /**
   * Reprend un profil dans le formulaire.
   *
   * Le mot de passe n'en vient jamais : il ne descend pas dans la page. Le
   * champ reste vide — laissé ainsi, le serveur ira chercher celui du fichier ;
   * retapé, c'est le nouveau qui part.
   */
  function choisirProfil(nom: string) {
    setProfilChoisi(nom);

    const trouve = profils.find(({ profil }) => profil.nom === nom);
    if (!trouve) {
      setMotDePasseDuProfil(false);
      return;
    }
    const { profil } = trouve;
    // Un profil se lit en composants : c'est là qu'on voit ce qu'il contient,
    // et une chaîne ne se reconstituerait pas sans le mot de passe, qui ne
    // descend pas dans la page.
    setMode('composants');
    setSgbd(profil.sgbd ?? '');
    setHote(profil.hote ?? '');
    setPort(profil.port ? String(profil.port) : '');
    setUtilisateur(profil.utilisateur ?? '');
    setBase(profil.base ?? '');
    setMotDePasse('');
    setMotDePasseDuProfil(trouve.mot_de_passe_enregistre);
  }

  /**
   * Change de mode de saisie.
   *
   * Passer à la chaîne de connexion désélectionne le profil : elle décrit à
   * elle seule la connexion voulue, et un profil qui resterait choisi laisserait
   * croire qu'il s'applique alors que le champ est vide.
   */
  function choisirMode(valeur: Mode) {
    setMode(valeur);
    if (valeur === 'dsn' && profilChoisi) {
      choisirProfil('');
    }
  }

  /** Ce que le formulaire porte, hors nom, pour l'enregistrement. */
  function champsDuProfil(): Omit<Profil, 'nom'> {
    return {
      sgbd: sgbd || undefined,
      hote: hote || undefined,
      port: port ? Number(port) : undefined,
      utilisateur: utilisateur || undefined,
      base: base || undefined,
    };
  }

  function soumettre(evenement: FormEvent) {
    evenement.preventDefault();
    onConnecter(
      mode === 'dsn'
        ? // Aucun profil ici : passer sur cet onglet en désélectionne un, et
          // une chaîne tapée décrit à elle seule la connexion voulue.
          { dsn }
        : {
            // Le nom suffit : le serveur relit le profil et va chercher le mot
            // de passe enregistré, qui n'a jamais transité par la page.
            profil: profilChoisi || undefined,
            sgbd: sgbd || undefined,
            hote,
            port: port ? Number(port) : undefined,
            utilisateur,
            mot_de_passe: motDePasse,
            base,
          },
    );
  }

  return (
    <form onSubmit={soumettre} className="flex w-full max-w-xl flex-col gap-4">
      <div>
        <h1 className="text-lg font-semibold">{t('connection.title')}</h1>
        <p className="mt-1 text-sm text-slate-600 dark:text-slate-400">{t('connection.intro')}</p>
      </div>

      <ProfileBar
        profils={profils}
        avertissement={avertissement}
        erreur={erreurProfil}
        choisi={profilChoisi}
        onChoisir={choisirProfil}
        aEnregistrer={champsDuProfil}
        motDePasseEnregistre={motDePasseDuProfil}
        onEnregistrer={(profil, remplacer, avecMotDePasse) =>
          // En mode DSN, les champs sont vides : c'est la chaîne qui porte la
          // connexion, et le serveur la décompose.
          enregistrer(profil, motDePasse, avecMotDePasse, remplacer, mode === 'dsn' ? dsn : undefined)
        }
        onSupprimer={(nom) => {
          void retirer(nom);
          choisirProfil('');
        }}
      />

      <div className="flex gap-1" role="tablist">
        {(['composants', 'dsn'] as Mode[]).map((valeur) => (
          <button
            key={valeur}
            type="button"
            role="tab"
            aria-selected={mode === valeur}
            onClick={() => choisirMode(valeur)}
            className={`rounded px-2 py-1 text-sm ${
              mode === valeur
                ? 'bg-slate-200 font-medium text-slate-900 dark:bg-slate-700 dark:text-slate-100'
                : 'text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200'
            }`}
          >
            {t(`connection.mode.${valeur}`)}
          </button>
        ))}
      </div>

      {mode === 'composants' ? (
        <div className="grid grid-cols-2 gap-3">
          <label className="flex flex-col gap-1">
            <span className="text-xs font-medium text-slate-600 dark:text-slate-400">
              {t('connection.dbms')}
            </span>
            <select
              value={sgbd}
              onChange={(evenement) => setSgbd(evenement.target.value)}
              className="rounded border border-slate-300 bg-white px-2 py-1.5 text-sm dark:border-slate-700 dark:bg-slate-900"
            >
              <option value="">{t('connection.dbms.auto')}</option>
              {sgbdDisponibles.map((valeur) => (
                <option key={valeur} value={valeur}>
                  {valeur}
                </option>
              ))}
            </select>
          </label>

          <Field
            label={t('connection.port')}
            type="number"
            value={port}
            onChange={(evenement) => setPort(evenement.target.value)}
          />
          <Field
            label={t('connection.host')}
            value={hote}
            required
            onChange={(evenement) => setHote(evenement.target.value)}
          />
          <Field
            label={t('connection.user')}
            value={utilisateur}
            autoComplete="username"
            onChange={(evenement) => setUtilisateur(evenement.target.value)}
          />
          <Field
            label={t('connection.password')}
            type="password"
            value={motDePasse}
            autoComplete="current-password"
            aide={
              motDePasseDuProfil ? t('profile.password.stored') : t('connection.password.hint')
            }
            onChange={(evenement) => setMotDePasse(evenement.target.value)}
          />
          <Field
            label={t('connection.database')}
            value={base}
            aide={t('connection.database.hint')}
            onChange={(evenement) => setBase(evenement.target.value)}
          />

        </div>
      ) : (
        <Field
          label={t('connection.dsn')}
          value={dsn}
          // Non requise quand un profil est choisi : c'est lui qui porte alors
          // la connexion, et la chaîne reste vide.
          required
          spellCheck={false}
          aide={t('connection.dsn.hint')}
          onChange={(evenement) => setDsn(evenement.target.value)}
          className="font-mono"
        />
      )}

      {erreur ? <ErrorBanner message={erreur} /> : null}

      <div>
        <Button type="submit" disabled={enCours}>
          {enCours ? t('connection.submitting') : t('connection.submit')}
        </Button>
      </div>
    </form>
  );
}
