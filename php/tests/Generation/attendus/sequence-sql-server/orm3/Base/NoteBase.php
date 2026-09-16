<?php

// Généré par Ormeau depuis la base sequence-sql-server et réécrit à chaque génération : le code propre à Note va dans Note.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class NoteBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'CUSTOM')]
    #[ORM\CustomIdGenerator(class: Generateur\NoteGenerateur::class)]
    #[ORM\Column(name: 'id', type: 'bigint')]
    protected ?int $id = null;

    public function getId(): ?int
    {
        return $this->id;
    }
}
