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
  'app.workdir.copy': 'Copier le chemin complet',
  'app.version': 'Version {version}',
  'app.title.busy': '({n}) Ormeau',

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
  'split.handleVertical': 'Hauteur du panneau — tirer, ou flèches haut et bas',

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

  'extraction.launch': 'Extraire',
  'extraction.running': 'Extraction en cours',
  'extraction.target': 'Écrit dans le répertoire de travail',
  'extraction.indicator.running': '{n} extraction(s) en cours',
  'extraction.indicator.failed': '{n} extraction(s) en échec',
  'extraction.indicator.finished': '{n} extraction(s) finie(s)',
  'extraction.panel.title': 'Extractions',
  'extraction.stream.lost': 'Suivi interrompu, reconnexion…',
  'extraction.scope.all': 'toutes les tables',
  'extraction.scope.tables': '{n} table(s)',
  'extraction.state.en_attente': 'en attente',
  'extraction.state.en_cours': 'en cours',
  'extraction.state.terminee': 'terminée',
  'extraction.state.echouee': 'échouée',
  'extraction.state.annulee': 'annulée',
  'extraction.step.source': 'serveur',
  'extraction.step.tables': 'tables',
  'extraction.step.colonnes': 'colonnes',
  'extraction.step.contraintes': 'contraintes',
  'extraction.step.index': 'index',
  'extraction.step.sequences': 'séquences',
  'extraction.step.types_enumeres': 'types énumérés',
  'extraction.step.vues': 'vues',
  'extraction.progress': 'Lecture : {etape}, étape {rang} sur {total}',
  'extraction.progressLabel': 'Avancement par étape',
  'extraction.waiting': 'En attente d’une place : deux extractions tournent déjà.',
  'extraction.cancel': 'Annuler',
  'extraction.dismiss': 'Retirer',
  'extraction.result': '{tables} table(s), {colonnes} colonne(s)',
  'extraction.anomalies': 'Anomalies du calque : {n}',
  'extraction.at': 'à {heure}',
  'extraction.replaced':
    'Fichier réécrit depuis par l’extraction terminée à {heure} : ce résumé ne le décrit plus.',

  'anomaly.version_inconnue': 'Version de calque inconnue',
  'anomaly.sgbd_inconnu': 'SGBD hors du vocabulaire',
  'anomaly.champ_requis_vide': 'Champ requis vide',
  'anomaly.empreinte_malformee': 'Empreinte mal formée',
  'anomaly.type_hors_vocabulaire': 'Type hors du vocabulaire',
  'anomaly.genre_defaut_inconnu': 'Genre de défaut inconnu',
  'anomaly.action_inconnue': 'Action référentielle inconnue',
  'anomaly.position_invalide': 'Position de colonne invalide',
  'anomaly.table_sans_colonne': 'Table sans colonne',
  'anomaly.table_dupliquee': 'Table déclarée deux fois',
  'anomaly.colonne_dupliquee': 'Colonne déclarée deux fois',
  'anomaly.colonne_introuvable': 'Colonne introuvable',
  'anomaly.table_cible_introuvable': 'Table référencée hors du calque',
  'anomaly.arite_incoherente': 'Nombre de colonnes incohérent',
  'anomaly.type_enumere_introuvable': 'Type énuméré introuvable',
  'anomaly.statistiques_orphelines': 'Statistiques sans table',

  'calque.title': 'Calque produit',
  'calque.extractedAt': 'extrait à {heure}',
  'calque.copy': 'Copier le calque',
  'calque.loading': 'Lecture du calque…',
  'calque.view': 'Affichage du calque',
  'calque.showFile': 'Fichier complet',
  'calque.tableMissing':
    '{table} n’est pas dans ce calque : pas encore extraite, ou hors de la portée de la dernière extraction.',
  'calque.statsHidden':
    'Statistiques retirées : l’interface n’affiche aucune valeur échantillonnée.',

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
  'app.workdir.copy': 'Copy the full path',
  'app.version': 'Version {version}',
  'app.title.busy': '({n}) Ormeau',

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
  'split.handleVertical': 'Panel height — drag, or up and down arrows',

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

  'extraction.launch': 'Extract',
  'extraction.running': 'Extraction running',
  'extraction.target': 'Written to the working directory',
  'extraction.indicator.running': '{n} extraction(s) running',
  'extraction.indicator.failed': '{n} extraction(s) failed',
  'extraction.indicator.finished': '{n} extraction(s) finished',
  'extraction.panel.title': 'Extractions',
  'extraction.stream.lost': 'Tracking interrupted, reconnecting…',
  'extraction.scope.all': 'all tables',
  'extraction.scope.tables': '{n} table(s)',
  'extraction.state.en_attente': 'waiting',
  'extraction.state.en_cours': 'running',
  'extraction.state.terminee': 'done',
  'extraction.state.echouee': 'failed',
  'extraction.state.annulee': 'cancelled',
  'extraction.step.source': 'server',
  'extraction.step.tables': 'tables',
  'extraction.step.colonnes': 'columns',
  'extraction.step.contraintes': 'constraints',
  'extraction.step.index': 'indexes',
  'extraction.step.sequences': 'sequences',
  'extraction.step.types_enumeres': 'enumerated types',
  'extraction.step.vues': 'views',
  'extraction.progress': 'Reading: {etape}, step {rang} of {total}',
  'extraction.progressLabel': 'Progress by step',
  'extraction.waiting': 'Waiting for a slot: two extractions are already running.',
  'extraction.cancel': 'Cancel',
  'extraction.dismiss': 'Dismiss',
  'extraction.result': '{tables} table(s), {colonnes} column(s)',
  'extraction.anomalies': 'Layer anomalies: {n}',
  'extraction.at': 'at {heure}',
  'extraction.replaced':
    'File rewritten since by the extraction finished at {heure}: this summary no longer describes it.',

  'anomaly.version_inconnue': 'Unknown layer version',
  'anomaly.sgbd_inconnu': 'DBMS outside the vocabulary',
  'anomaly.champ_requis_vide': 'Required field empty',
  'anomaly.empreinte_malformee': 'Malformed fingerprint',
  'anomaly.type_hors_vocabulaire': 'Type outside the vocabulary',
  'anomaly.genre_defaut_inconnu': 'Unknown default kind',
  'anomaly.action_inconnue': 'Unknown referential action',
  'anomaly.position_invalide': 'Invalid column position',
  'anomaly.table_sans_colonne': 'Table without columns',
  'anomaly.table_dupliquee': 'Table declared twice',
  'anomaly.colonne_dupliquee': 'Column declared twice',
  'anomaly.colonne_introuvable': 'Column not found',
  'anomaly.table_cible_introuvable': 'Referenced table outside the layer',
  'anomaly.arite_incoherente': 'Column count mismatch',
  'anomaly.type_enumere_introuvable': 'Enumerated type not found',
  'anomaly.statistiques_orphelines': 'Statistics without a table',

  'calque.title': 'Produced layer',
  'calque.extractedAt': 'extracted at {heure}',
  'calque.copy': 'Copy the layer',
  'calque.loading': 'Reading the layer…',
  'calque.view': 'Layer view',
  'calque.showFile': 'Whole file',
  'calque.tableMissing':
    '{table} is not in this layer: not extracted yet, or outside the scope of the last extraction.',
  'calque.statsHidden': 'Statistics removed: the interface shows no sampled value.',

  'error.network': 'The local server is not responding.',
  'error.unknown': 'Unexpected failure.',
  'error.notJson': 'Unexpected response from {path} ({type}).',
  'error.noContentType': 'no content type',
};

/** Dictionnaires servis à l'interface. */
export const messages: Record<Lang, Record<MessageKey, string>> = { fr, en };

/**
 * Dit si une chaîne est une clé du dictionnaire.
 *
 * Sert aux codes venus du serveur — étapes, anomalies —, dont la clé se
 * construit à l'exécution et que le typage ne peut donc pas vérifier.
 */
export function estCle(cle: string): cle is MessageKey {
  return Object.prototype.hasOwnProperty.call(fr, cle);
}
