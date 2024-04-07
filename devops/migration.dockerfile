FROM migrate/migrate
WORKDIR migrations
COPY migrations .

CMD migrate -path=/migrations -database=mysql://$DB_USER:$DB_PASS@tcp($DB_HOST)/trolly up