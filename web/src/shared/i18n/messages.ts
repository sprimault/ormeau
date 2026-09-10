// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

/** Langues servies par l'interface. Le bilinguisme s'arrête ici : il ne remonte
 *  ni dans le calque, ni dans l'inférence, ni dans le code généré. */
export type Lang = 'fr' | 'en';

/**
 * Dictionnaire français, qui fait référence pour le typage des clés.
 *
 * Ce qui ne se traduit jamais n'y figure pas : noms de tables, de colonnes, de
 * schémas et types bruts s'affichent tels qu'ils sont en base, puisque c'est ce
 * que l'utilisateur retrouvera dans son SGBD.
 */
const fr = {
  'app.name': 'Ormeau',
  'app.workdir': 'Répertoire de travail',
  'app.version': 'Version {version}',

  'theme.label': 'Thème',
  'theme.clair': 'Clair',
  'theme.sombre': 'Sombre',
  'theme.systeme': 'Système',

  'lang.label': 'Langue',

  'connection.title': 'Connexion à la base',
  'connection.intro':
    'L’outil se connecte en lecture seule et ne lit aucune donnée tant que l’échantillonnage n’est pas demandé.',
  'connection.mode.composants': 'Composants',
  'connection.mode.dsn': 'Chaîne de connexion',
  'connection.dbms': 'SGBD',
  'connection.dbms.auto': 'Déduit du port',
  'connection.host': 'Hôte',
  'connection.port': 'Port',
  'connection.user': 'Utilisateur',
  'connection.password': 'Mot de passe',
  'connection.database': 'Base',
  'connection.database.hint': 'Laisser vide pour parcourir tout le serveur',
  'connection.dsn': 'DSN',
  'connection.dsn.hint': 'Un DATABASE_URL de Symfony convient : les paramètres inutiles sont retirés.',
  'connection.submit': 'Se connecter',
  'connection.submitting': 'Connexion…',
  'connection.disconnect': 'Se déconnecter',
  'connection.connected': 'Connecté à {catalogue}',
  'connection.server': 'Serveur',
  'connection.schemas': 'Schémas',
  'connection.schemas.empty': 'Aucun schéma exploitable dans cette base.',
  'connection.password.hint': 'Le mot de passe n’est jamais enregistré : il vit le temps de la session.',

  'error.network': 'Le serveur local ne répond pas.',
  'error.unknown': 'Échec inattendu.',
  'error.notJson': 'Réponse inattendue de {path} ({type}).',
  'error.noContentType': 'sans type de contenu',
} as const;

/** Clé du dictionnaire. Une clé absente fait échouer la compilation, pas
 *  seulement l'exécution. */
export type MessageKey = keyof typeof fr;

/** Dictionnaire anglais. Le type impose qu'il couvre exactement les mêmes clés. */
const en: Record<MessageKey, string> = {
  'app.name': 'Ormeau',
  'app.workdir': 'Working directory',
  'app.version': 'Version {version}',

  'theme.label': 'Theme',
  'theme.clair': 'Light',
  'theme.sombre': 'Dark',
  'theme.systeme': 'System',

  'lang.label': 'Language',

  'connection.title': 'Database connection',
  'connection.intro':
    'The tool connects read-only and reads no data unless sampling is requested.',
  'connection.mode.composants': 'Fields',
  'connection.mode.dsn': 'Connection string',
  'connection.dbms': 'DBMS',
  'connection.dbms.auto': 'From the port',
  'connection.host': 'Host',
  'connection.port': 'Port',
  'connection.user': 'User',
  'connection.password': 'Password',
  'connection.database': 'Database',
  'connection.database.hint': 'Leave empty to browse the whole server',
  'connection.dsn': 'DSN',
  'connection.dsn.hint': 'A Symfony DATABASE_URL works: unusable parameters are stripped.',
  'connection.submit': 'Connect',
  'connection.submitting': 'Connecting…',
  'connection.disconnect': 'Disconnect',
  'connection.connected': 'Connected to {catalogue}',
  'connection.server': 'Server',
  'connection.schemas': 'Schemas',
  'connection.schemas.empty': 'No usable schema in this database.',
  'connection.password.hint': 'The password is never stored: it lives for the session only.',

  'error.network': 'The local server is not responding.',
  'error.unknown': 'Unexpected failure.',
  'error.notJson': 'Unexpected response from {path} ({type}).',
  'error.noContentType': 'no content type',
};

/** Dictionnaires servis à l'interface. */
export const messages: Record<Lang, Record<MessageKey, string>> = { fr, en };
