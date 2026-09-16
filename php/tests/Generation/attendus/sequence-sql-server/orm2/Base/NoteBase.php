<?php

// Généré par Ormeau depuis la base sequence-sql-server et réécrit à chaque génération : le code propre à Note va dans Note.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class NoteBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'SEQUENCE')]
    #[ORM\SequenceGenerator(sequenceName: 'sq_note', initialValue: 1000)]
    #[ORM\Column(name: 'id', type: 'bigint')]
    protected ?string $id = null;

    public function getId(): ?string
    {
        return $this->id;
    }
}
