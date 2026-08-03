# SCOUL ITMS API DOCUMENTATION

Please read this to get directions on how to kickstart and run the api

## RUNNING THE API

The api has a command line tool as part of the `Makefile`
If you want to see all possible commands, run.

```bash
make help
```

Create a `.env` file following `.env.example` at the root of the project

Install all dependencies and tidy up the project

```bash
make tidy
```

Migrate the database

```bash
make migrate
```

Run the project to see if all is well

```bash
make run
```

If you see something like this, Smile :)

```
Building the main application...
Running the main application...
2025/06/11 15:18:06 Mysql Env loaded succesfuly.
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

[GIN-debug] POST   /api/auth/login           --> github.com/othie12/scanner-api/internals/api/hanlders.(*AuthHandler).Login-fm (4 handlers)
Attempting to start server on port: 8080
[GIN-debug] [WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.
Please check https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies for details.
[GIN-debug] Listening and serving HTTP on :8080
```

Stop the project with `ctrl + C` and then create a file named `seed.json` at the root of the project.
This seed file will create the first user in the db. You can later delete this file after the user has been seeded. Below is the format of `seed.json` file

```json
{
  "username": "yesu",
  "role": "approver", // approver | entrant
  "password": "password"
}
```

create a directory named `/public` at the root of the project and add the MRS and GRN excel files in that direcotry in this format.
|Serial No.|Requisition Date|
|--------|--------|
|SGL367|10/01/2024|

_Make sure the `requisition Date` is formatted to excel type `Date`_
_Headings don't matter but make sure you start from the exact first cell and follow the column order_

This is how the file structure must be. Make sure the filenames are exactly matching

```
Project root
|---public
    |-MRS.xlsx
    |-GRN.xlsx
```

Run the project

```bash
make run
```

tadaaa... Now were in
# fiscal-doc-api
