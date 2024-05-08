# go-url-shortener

use-case

- CreateLink. Create using `href`.
- GetLinkByShortID. Redirect using `short_id`
- IncreamentLinkCounter. (update usage_at, usage_count++)


## run database

`docker run --name url_shortener_db --rm -p 5432:5432 -e POSTGRES_PASSWORD=dbpassword -e POSTGRES_USER=dbuser -e POSTGRES_DB=dbname postgres:16`

## create migration

`migrate create -ext sql -dir ./migrations -seq -digits 4 init`

## up migration

`migrate -database pgx5://dbuser:dbpassword@localhost:5432/dbname?sslmode=disable -path ./migrations up`

## version migration

`migrate -database pgx5://dbuser:dbpassword@localhost:5432/dbname?sslmode=disable -path ./migrations version`

## down migration

`migrate -database pgx5://dbuser:dbpassword@localhost:5432/dbname?sslmode=disable -path ./migrations down`