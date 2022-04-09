## Stock Watcher

### Running Locally

**_Prerequisites_**

- [Docker](https://docker.com/)
- Create `.env` file from `.env.example` in root project folder
- Create `app.env` file from `app.env.example` in server folder

To run locally :

```
docker-compose up
```

To rebuild after changes from docker related files :

```
docker-compose up --build
```

### Migrations

This web application utilize [golang-migrate/migrate](https://github.com/golang-migrate/migrate) to handle it's migrations.

- [CLI Documentation](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- [MongoDB Migrate Documentation](https://github.com/golang-migrate/migrate/tree/master/database/mongodb)

Every migration command must be executed inside `stock_watcher_server` container. to get inside the container run `docker-compose exec stock_watcher_server sh`. You can execute bash scripts inside folder `/server/scripts` to make new migration or execute **golang-migrate/migrate** commands

To create a new migration from bash script:
```
bash scripts/new_migration.sh migration_file_name
```

To execute other **golang-migrate/migrate** commands :
```
bash scripts/migrations.sh OPTIONS COMMAND [arg...]
```

Or You can just execute all available command from **golang-migrate/migrate** directly. These bash scripts was created to make it easier to run migrate commands.
