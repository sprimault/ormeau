// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useState, type FormEvent } from 'react';

import { useT } from '@/shared/i18n';
import { Button, ErrorBanner, Field, HelpTip } from '@/shared/ui';
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

/**
 * Modes de sslmode, le vocabulaire de libpq. Le serveur refuse toute autre
 * valeur ; le choix vide laisse au pilote son défaut, prefer.
 */
const modesSSL = ['disable', 'allow', 'prefer', 'require', 'verify-ca', 'verify-full'];

/**
 * Ports par défaut des SGBD proposés, pour comparer une destination comme le
 * serveur le fait : un port laissé vide vaut celui du SGBD.
 */
const portsParDefaut: Record<string, number> = { postgres: 5432 };

/** Ce qu'une connexion vise réellement : SGBD, hôte sans casse ni blanc, port effectif. */
function destination(sgbd: string | undefined, hote: string | undefined, port: number | undefined) {
  const deduit = Object.entries(portsParDefaut).find(([, defaut]) => defaut === port)?.[0] ?? '';
  const sgbdEffectif = (sgbd ?? '').toLowerCase() || deduit;
  return {
    sgbd: sgbdEffectif,
    hote: (hote ?? '').trim().toLowerCase(),
    port: port ?? portsParDefaut[sgbdEffectif],
  };
}

/** Les deux formes de saisie. */
type Mode = 'composants' | 'dsn';

/** Propriétés du formulaire. */
interface ProprietesFormulaire {
  enCours: boolean;
  erreur: string | null;
  onConnecter: (requete: RequeteConnexion) => Promise<boolean>;
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
  const [sslmode, setSslmode] = useState('');
  const [dsn, setDsn] = useState('');
  const [profilChoisi, setProfilChoisi] = useState('');
  const [motDePasseDuProfil, setMotDePasseDuProfil] = useState(false);
  const [nomProfil, setNomProfil] = useState('');
  const [avecMotDePasse, setAvecMotDePasse] = useState(false);
  const { profils, avertissement, enregistrer, retirer, erreur: erreurProfil } = useProfils();

  // Le bouton de mise à jour ne paraît que lorsqu'il a quelque chose à faire.
  const profilRetenu = profils.find(({ profil }) => profil.nom === profilChoisi)?.profil;
  const divergent =
    profilRetenu !== undefined &&
    (profilRetenu.sgbd !== (sgbd || undefined) ||
      profilRetenu.hote !== (hote || undefined) ||
      profilRetenu.port !== (port ? Number(port) : undefined) ||
      profilRetenu.utilisateur !== (utilisateur || undefined) ||
      profilRetenu.base !== (base || undefined) ||
      profilRetenu.sslmode !== (sslmode || undefined));

  // Plus étroit que `divergent` : la base et le sslmode ne changent pas le
  // serveur ni le compte. C'est ce que le serveur compare avant d'envoyer le
  // mot de passe enregistré, et l'aide sous le champ doit dire la même chose.
  const saisie = destination(sgbd || undefined, hote, port ? Number(port) : undefined);
  const retenue =
    profilRetenu && destination(profilRetenu.sgbd, profilRetenu.hote, profilRetenu.port);
  const autreDestination =
    retenue !== undefined &&
    (saisie.sgbd !== retenue.sgbd ||
      saisie.hote !== retenue.hote ||
      saisie.port !== retenue.port ||
      (utilisateur || undefined) !== profilRetenu?.utilisateur);

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
    setSslmode(profil.sslmode ?? '');
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
      // Toujours porté : « Mettre à jour ce profil » remplace le profil entier,
      // et un champ absent ici effacerait en silence un verify-full enregistré.
      sslmode: sslmode || undefined,
    };
  }

  /**
   * Se connecte, puis enregistre le profil si un nom a été donné.
   *
   * Dans cet ordre, et c'est tout l'intérêt : un profil enregistré est
   * forcément un profil qui marche, et il n'y a qu'un bouton à connaître.
   */
  async function soumettre(evenement: FormEvent) {
    evenement.preventDefault();

    const reussie = await onConnecter(
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
            sslmode: sslmode || undefined,
          },
    );

    // Rien à enregistrer sans nom : c'est une connexion ponctuelle. Rien non
    // plus sur un profil déjà choisi — le réenregistrer à chaque connexion
    // effacerait son mot de passe quand la case est décochée.
    if (!reussie || profilChoisi || nomProfil.trim() === '') {
      return;
    }
    await enregistrer(
      { ...champsDuProfil(), nom: nomProfil.trim() },
      motDePasse,
      avecMotDePasse,
      false,
      mode === 'dsn' ? dsn : undefined,
    );
  }

  return (
    <form onSubmit={(evenement) => void soumettre(evenement)} className="flex w-full max-w-xl flex-col gap-4">
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
        nom={nomProfil}
        onNommer={setNomProfil}
        avecMotDePasse={avecMotDePasse}
        onAvecMotDePasse={setAvecMotDePasse}
        divergent={divergent}
        onMettreAJour={() => {
          // En mode DSN, les champs sont vides : c'est la chaîne qui porte la
          // connexion, et le serveur la décompose.
          void enregistrer(
            { ...champsDuProfil(), nom: profilChoisi },
            motDePasse,
            avecMotDePasse,
            true,
            mode === 'dsn' ? dsn : undefined,
          );
        }}
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
            <span className="inline-flex items-center gap-1 text-xs font-medium text-slate-600 dark:text-slate-400">
              {t('connection.dbms')}
              <HelpTip texte={t('connection.help.dbms')} />
            </span>
            {/* Nommée à part : le bouton d'aide dans le libellé en vide le nom
                accessible, et un lecteur d'écran annoncerait une liste muette. */}
            <select
              value={sgbd}
              aria-label={t('connection.dbms')}
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
              !motDePasseDuProfil
                ? t('connection.password.hint')
                : autreDestination
                  ? t('profile.password.otherDestination')
                  : t('profile.password.stored')
            }
            onChange={(evenement) => setMotDePasse(evenement.target.value)}
          />
          <Field
            label={t('connection.database')}
            value={base}
            aide={t('connection.database.hint')}
            onChange={(evenement) => setBase(evenement.target.value)}
          />
          <label className="flex flex-col gap-1">
            <span className="inline-flex items-center gap-1 text-xs font-medium text-slate-600 dark:text-slate-400">
              {t('connection.sslmode')}
              <HelpTip texte={t('connection.help.sslmode')} />
            </span>
            <select
              value={sslmode}
              aria-label={t('connection.sslmode')}
              onChange={(evenement) => setSslmode(evenement.target.value)}
              className="rounded border border-slate-300 bg-white px-2 py-1.5 text-sm dark:border-slate-700 dark:bg-slate-900"
            >
              <option value="">{t('connection.sslmode.default')}</option>
              {modesSSL.map((valeur) => (
                <option key={valeur} value={valeur}>
                  {valeur}
                </option>
              ))}
            </select>
          </label>
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
