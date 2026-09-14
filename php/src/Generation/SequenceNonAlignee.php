<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Séquence dont l'incrément n'est pas 1, laissée avec allocationSize à 1.
 *
 * Sous ORM 2, DBAL 3 compare allocationSize à l'incrément déclaré : écart,
 * schema:update et migrations:diff proposent ALTER SEQUENCE … INCREMENT BY 1.
 * Cet ALTER réussit et modifie une séquence que d'autres applications
 * utilisent peut-être. Aligner allocationSize le ferait taire, mais un
 * incrément supérieur à 1 a deux sens que la base ne distingue pas : des blocs
 * réservés par un ORM, où l'alignement est juste, ou des identifiants décalés
 * entre plusieurs nœuds, où il produit des collisions (essai du 2026-09-14,
 * ORM 2.14.3 et DBAL 3.10.6, deux nœuds en INCREMENT BY 10). allocationSize à
 * 1 lit chaque identifiant par nextval, ce qui est juste dans les deux cas ; le
 * rapport dit pourquoi l'ALTER proposé ne doit pas s'appliquer.
 *
 * Rien de tel sous ORM 3 : la clé y est rendue en IDENTITY, sans générateur.
 */
final class SequenceNonAlignee
{
    /**
     * @param string $entite    entité dont la clé lit la séquence
     * @param string $sequence  nom de la séquence, tel que #[ORM\SequenceGenerator] l'écrit
     * @param int    $increment incrément que la base déclare
     */
    public function __construct(
        public readonly string $entite,
        public readonly string $sequence,
        public readonly int $increment,
    ) {}

    /**
     * Rend l'avertissement en une ligne, dans l'ordre où il faut le lire : ce
     * qui va apparaître, qu'il ne faut pas l'appliquer, et pourquoi l'outil ne
     * tranche pas à la place de l'utilisateur.
     */
    public function message(): string
    {
        return sprintf(
            '%s : schema:update et migrations:diff proposeront « ALTER SEQUENCE %s INCREMENT BY 1 ». Ne pas l\'appliquer : la base incrémente cette séquence de %d, et l\'ALTER la modifierait. Ormeau n\'aligne pas allocationSize, car un incrément de %d réserve des blocs d\'identifiants ou sépare ceux de plusieurs nœuds, la base ne dit pas lequel, et aligner produirait des collisions dans le second cas ; chaque identifiant est lu par nextval, ce qui reste juste dans les deux.',
            $this->entite,
            $this->sequence,
            $this->increment,
            $this->increment,
        );
    }
}
