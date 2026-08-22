# Image d'Ormeau, construite depuis les binaires déjà croisés par `make binaries`.
#
# Aucune compilation ici : buildx renseigne TARGETOS et TARGETARCH, et chaque
# plateforme reçoit l'artefact correspondant. C'est ce qui évite l'émulation
# QEMU, et ce qui impose de lancer `make binaries` avant `make image`.
#
# Le conteneur reste un repli, pas le mode nominal : il doit encore atteindre la
# base, ce que le binaire natif fait sans configuration réseau.

FROM alpine:3 AS certificats
RUN apk add --no-cache ca-certificates

FROM scratch

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG REVISION=unknown

# scratch n'embarque aucune distribution, donc aucun certificat racine. Sans
# cette copie, toute connexion TLS vers une base échoue sur une erreur de
# vérification difficile à rattacher à sa cause.
COPY --from=certificats /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY dist/ormeau_${TARGETOS}_${TARGETARCH} /ormeau

# Utilisateur non privilégié : l'outil ne lit que le réseau et écrit un calque
# sur un volume monté, il n'a jamais besoin de root. scratch n'ayant pas
# /etc/passwd, l'UID est donné en clair.
USER 65532:65532

LABEL org.opencontainers.image.title="Ormeau" \
      org.opencontainers.image.description="Rétro-ingénierie de bases relationnelles vers des entités ORM" \
      org.opencontainers.image.source="https://github.com/sprimault/ormeau" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${REVISION}"

ENTRYPOINT ["/ormeau"]
