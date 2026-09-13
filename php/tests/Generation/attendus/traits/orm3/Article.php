<?php

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\ArticleBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'article')]
class Article extends ArticleBase {}
