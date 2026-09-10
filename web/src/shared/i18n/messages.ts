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
  'connection.dsn.hint':
    'Un DATABASE_URL de Symfony convient : les paramètres inutiles sont retirés.',
  'connection.submit': 'Se connecter',
  'connection.submitting': 'Connexion…',
  'connection.disconnect': 'Se déconnecter',
  'connection.connected': 'Connecté à {catalogue}',
  'connection.server': 'Serveur',
  'connection.schemas': 'Schémas',
  'connection.schemas.empty': 'Aucun schéma exploitable dans cette base.',
  'connection.password.hint':
    'Le mot de passe n’est jamais enregistré : il vit le temps de la session.',

  'inventory.title': 'Tables à extraire',
  'inventory.loading': 'Lecture du catalogue…',
  'inventory.empty': 'Aucune table dans les schémas retenus.',
  'inventory.search': 'Filtrer les tables',
  'inventory.noMatch': 'Aucune table ne correspond à « {terme} ».',
  'inventory.selected': '{n} sélectionnée(s) sur {total}',
  'inventory.tables': '{n} table(s)',
  'inventory.columns': '{n} col.',
  'inventory.rows': '{n} lignes',
  'inventory.noPrimaryKey': 'sans clé primaire',
  'inventory.singleDatabase':
    'Ce serveur n’expose qu’une base, ou n’en sait pas énumérer d’autres.',
  'inventory.selectAll': 'Tout cocher',
  'inventory.clear': 'Tout décocher',
  'inventory.advice':
    'Extraire large, filtrer à l’inférence : le calque garde tout, et les tables ignorées se trient hors ligne, sans revenir en base.',
  'inventory.missing':
    '{n} table(s) référencée(s) par votre sélection n’en font pas partie. Sans elles, les clés étrangères pointent dans le vide et les associations disparaissent.',
  'inventory.addMissing': 'Ajouter les dépendances',

  'columns.loading': 'Lecture des colonnes…',
  'columns.primaryKey': 'clé primaire',
  'columns.primaryKeyKept':
    'Une colonne de clé primaire reste mappée : Doctrine refuse une entité sans identifiant.',
  'columns.excluded': '{n} colonne(s) écartée(s)',
  'columns.expand': 'Déplier les colonnes',

  'action.copy': 'Copier',
  'split.handle': 'Largeur du panneau — tirer, ou flèches gauche et droite',

  'scope.title': 'Ce que produira cet écran',
  'scope.extraction': 'Portée envoyée à l’extraction',
  'scope.decisions': 'Décisions — le calque garde tout',
  'scope.copy': 'Copier la portée',
  'scope.copyDecisions': 'Copier les décisions',
  'scope.all': 'toutes les tables des schémas retenus',
  'scope.count': '{n} table(s) sur {schemas} schéma(s)',

  'details.none': 'Choisir une table pour en voir le détail.',
  'details.columns': 'Colonnes',
  'details.rows': 'Lignes estimées',
  'details.primaryKey': 'Clé primaire',
  'details.yes': 'oui',
  'details.no': 'aucune',
  'details.noPrimaryKeyWarning': 'Aucune clé primaire : Doctrine refusera cette entité en l’état.',
  'details.column': 'Colonne',
  'details.type': 'Type',
  'details.nullable': 'Nullable',
  'details.excluded': 'écartée',
  'details.references': 'Références sortantes',
  'details.noReference': 'Aucune clé étrangère déclarée.',
  'details.outOfSelection': 'hors sélection',

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
  'connection.intro': 'The tool connects read-only and reads no data unless sampling is requested.',
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

  'inventory.title': 'Tables to extract',
  'inventory.loading': 'Reading the catalogue…',
  'inventory.empty': 'No table in the selected schemas.',
  'inventory.search': 'Filter tables',
  'inventory.noMatch': 'No table matches “{terme}”.',
  'inventory.selected': '{n} of {total} selected',
  'inventory.tables': '{n} table(s)',
  'inventory.columns': '{n} col.',
  'inventory.rows': '{n} rows',
  'inventory.noPrimaryKey': 'no primary key',
  'inventory.singleDatabase': 'This server exposes a single database, or cannot enumerate others.',
  'inventory.selectAll': 'Select all',
  'inventory.clear': 'Clear selection',
  'inventory.advice':
    'Extract wide, filter at inference: the layer keeps everything, and ignored tables are sorted out offline, without going back to the database.',
  'inventory.missing':
    '{n} table(s) referenced by your selection are not part of it. Without them, foreign keys point nowhere and the associations vanish.',
  'inventory.addMissing': 'Add dependencies',

  'columns.loading': 'Reading columns…',
  'columns.primaryKey': 'primary key',
  'columns.primaryKeyKept':
    'A primary key column stays mapped: Doctrine rejects an entity without an identifier.',
  'columns.excluded': '{n} column(s) left out',
  'columns.expand': 'Expand columns',

  'action.copy': 'Copy',
  'split.handle': 'Panel width — drag, or left and right arrows',

  'scope.title': 'What this screen will produce',
  'scope.extraction': 'Scope sent to extraction',
  'scope.decisions': 'Decisions — the layer keeps everything',
  'scope.copy': 'Copy the scope',
  'scope.copyDecisions': 'Copy the decisions',
  'scope.all': 'every table in the selected schemas',
  'scope.count': '{n} table(s) across {schemas} schema(s)',

  'details.none': 'Pick a table to see its details.',
  'details.columns': 'Columns',
  'details.rows': 'Estimated rows',
  'details.primaryKey': 'Primary key',
  'details.yes': 'yes',
  'details.no': 'none',
  'details.noPrimaryKeyWarning': 'No primary key: Doctrine will reject this entity as is.',
  'details.column': 'Column',
  'details.type': 'Type',
  'details.nullable': 'Nullable',
  'details.excluded': 'left out',
  'details.references': 'Outgoing references',
  'details.noReference': 'No foreign key declared.',
  'details.outOfSelection': 'outside the selection',

  'error.network': 'The local server is not responding.',
  'error.unknown': 'Unexpected failure.',
  'error.notJson': 'Unexpected response from {path} ({type}).',
  'error.noContentType': 'no content type',
};

/** Dictionnaires servis à l'interface. */
export const messages: Record<Lang, Record<MessageKey, string>> = { fr, en };
