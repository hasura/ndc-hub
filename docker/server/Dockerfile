ARG DATA_TAG

FROM ghcr.io/hasura/ndc-hub:$DATA_TAG as data 

FROM caddy:alpine

COPY --from=data /assets /srv

EXPOSE 80

CMD ["caddy", "file-server", "--browse", "--root", "/srv"]