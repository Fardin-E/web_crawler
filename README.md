## Web crawler made in Golang

## Create a postgres docker image


## 🚀 Step 1: Launch Your Docker Compose Services

First, ensure your Docker Compose services (including your `db` service) are running. Navigate to the directory containing your `docker-compose.yml` file and run:

```bash
docker compose up -d
```

This command will start your database in the background, and the `init.sql` file will automatically execute to create your schema and table.

-----

## 💻 Step 2: Connect to the PostgreSQL Container

Once the `db` service is running, you can connect to it using the `psql` command-line client, which is typically available inside the PostgreSQL Docker image.

1.  **Find the running container's name or ID**:

    ```bash
    docker ps
    ```

    Look for a container related to `postgres` or `db`. Its name might be something like `yourprojectname-db-1` or similar.

2.  **Execute `psql` inside the container**:
    Use the `docker exec` command to run `psql` within your database container.

      * Replace `[container_name_or_id]` with the actual name or ID you found in the previous step.
      * `freid` is your `POSTGRES_USER` as defined in `docker-compose.yml`.
      * `crawler` is your `POSTGRES_DB` as defined in `docker-compose.yml`.

    <!-- end list -->

    ```bash
    docker exec -it [container_name_or_id] psql -U freid -d crawler
    ```

    You will be prompted for the password (`password` in your `docker-compose.yml`), then you'll be connected to the PostgreSQL prompt.

-----

## ✅ Step 3: Inspect the Database with SQL Commands

Once you're at the `psql` prompt (`crawler=#`), you can run standard SQL commands to verify your database setup.

1.  **List schemas to confirm `crawler_schema` exists**:

    ```sql
    \dn
    ```

    You should see `crawler_schema` in the list.

2.  **Set the search path to your schema (optional but helpful)**:

    ```sql
    SET search_path TO crawler_schema;
    ```

    This allows you to reference tables within `crawler_schema` without explicitly prefixing them (e.g., `pages` instead of `crawler_schema.pages`).

3.  **List tables in the current schema (or specify the schema)**:

    ```sql
    \dt crawler_schema.
    ```

    This command will show you the `pages` table within the `crawler_schema`.

4.  **Describe the `pages` table to view its structure**:

    ```sql
    \d pages
    ```

    Or, if you didn't set the search path:

    ```sql
    \d crawler_schema.pages
    ```

    This command will display the column names, data types, and constraints for your `pages` table.

5.  **Query data (if any)**:
    If you've inserted any data, you can query it:

    ```sql
    SELECT * FROM pages;
    ```

6.  **Exit `psql`**:
    To exit the `psql` prompt, type:

    ```sql
    \q
    ```