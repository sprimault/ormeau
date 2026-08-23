-- Base de référence PostgreSQL.
--
-- Ce fichier est le document de spécification le plus utile du dépôt : il
-- couvre délibérément les cas tordus qu'un générateur naïf rate. À tenir à jour
-- en même temps que les heuristiques.

CREATE SCHEMA IF NOT EXISTS gescom;
SET search_path TO gescom;

-- Cas nominal : entité simple, identité, commentaires, contrainte de contrôle
-- qui doit devenir une énumération.
CREATE TABLE t_commercial (
    com_id     int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    com_nom    varchar(80) NOT NULL,
    com_actif  boolean     NOT NULL DEFAULT true
);
COMMENT ON TABLE t_commercial IS 'Force de vente';

CREATE TABLE t_client (
    cli_id      int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    cli_nom     varchar(120) NOT NULL,
    cli_siret   char(14),
    cli_statut  varchar(20)  NOT NULL DEFAULT 'ACTIF',
    cli_com_id  int,
    cli_ca_ttc  numeric(12, 2),
    -- colonne générée : la propriété doit être ni insérable ni modifiable
    cli_ca_ht   numeric(12, 2) GENERATED ALWAYS AS (cli_ca_ttc / 1.2) STORED,
    created_at  timestamptz  NOT NULL DEFAULT now(),
    updated_at  timestamptz,
    CONSTRAINT ck_cli_statut CHECK (cli_statut IN ('ACTIF', 'SUSPENDU', 'ARCHIVE')),
    CONSTRAINT uq_cli_siret UNIQUE (cli_siret),
    CONSTRAINT fk_client_commercial FOREIGN KEY (cli_com_id)
        REFERENCES t_commercial (com_id) ON DELETE SET NULL
);
COMMENT ON COLUMN t_client.cli_siret IS 'Nul tant que la fiche n''est pas validée';

CREATE INDEX ix_cli_nom ON t_client (cli_nom);
-- index partiel : information perdue par information_schema
CREATE INDEX ix_cli_actifs ON t_client (cli_com_id) WHERE cli_statut = 'ACTIF';
-- classe d'opérateurs explicite : sans elle, le DDL n'est pas reconstructible,
-- et deux index de comportements différents deviennent indistinguables.
CREATE INDEX ix_cli_nom_prefixe ON t_client (cli_nom text_pattern_ops);

-- Table de jointure pure : doit produire une association, pas une entité.
CREATE TABLE t_client_tag (
    cli_id int NOT NULL REFERENCES t_client (cli_id) ON DELETE CASCADE,
    tag_id int NOT NULL,
    PRIMARY KEY (cli_id, tag_id)
);

-- Table de liaison portant une donnée propre : doit rester une entité.
CREATE TABLE t_client_contact (
    cli_id     int NOT NULL REFERENCES t_client (cli_id),
    ctc_id     int NOT NULL,
    role       varchar(30) NOT NULL,
    PRIMARY KEY (cli_id, ctc_id)
);

-- Héritage : la clé primaire est aussi une clé étrangère.
CREATE TABLE t_client_grand_compte (
    cli_id       int PRIMARY KEY REFERENCES t_client (cli_id),
    remise_taux  numeric(4, 2) NOT NULL
);

-- Auto-référence.
CREATE TABLE t_categorie (
    cat_id     int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    cat_libelle varchar(60) NOT NULL,
    cat_parent  int REFERENCES t_categorie (cat_id)
);

-- Aucune clé primaire : doit produire un avertissement, pas une exception.
CREATE TABLE t_log_import (
    horodatage timestamptz NOT NULL,
    message    text
);

-- Clé étrangère implicite : aucune contrainte déclarée, mais toutes les valeurs
-- existent dans t_client. Détectable seulement avec --echantillonner.
CREATE TABLE t_facture (
    fac_id     int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    fac_cli_id int NOT NULL,
    fac_total  numeric(12, 2) NOT NULL
);

-- Type énuméré natif.
CREATE TYPE canal AS ENUM ('web', 'telephone', 'agence');
CREATE TABLE t_commande (
    cmd_id    int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    cmd_canal canal NOT NULL
);

-- Identifiants réservés et accents, pour éprouver l'échappement.
CREATE TABLE "t_référence" (
    "id"    int GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "order" int,
    "select" varchar(10)
);

CREATE VIEW v_client_actif AS
SELECT cli_id, cli_nom FROM t_client WHERE cli_statut = 'ACTIF';

INSERT INTO t_commercial (com_nom) VALUES ('Durand'), ('Nguyen');
INSERT INTO t_client (cli_nom, cli_statut, cli_com_id, cli_ca_ttc)
VALUES ('Alpha', 'ACTIF', 1, 1200.00),
       ('Beta', 'SUSPENDU', 1, 0.00),
       ('Gamma', 'ARCHIVE', 2, 45000.00);
INSERT INTO t_facture (fac_cli_id, fac_total) VALUES (1, 120.00), (1, 80.00), (3, 900.00);
