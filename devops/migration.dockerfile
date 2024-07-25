FROM galaxyproduction/migrate:latest
WORKDIR migrations
COPY migrations .

ENTRYPOINT []
CMD migrate -path=/migrations/ -database "mysql://${DB_USER}:${DB_PASS}@tcp(${DB_HOST})/trolly" up

# Rebuild galaxyproduction/migrate with arm64 alpine image
# Push galaxyproduction/migrate
# Rebuild this image
# Push this image