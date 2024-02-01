<h1 align="center">Gost: Go Starter</h1>

<br>

Golang project starter with Fiber Framework, jwt-auth, email service and soft delete schema to build a robust RestAPI Backend Application.

&#xa0;

## :rocket: Technologies And :wrench: Tools

Techs and tools were used in this project:

- [Fiber Framework](https://docs.gofiber.io/) → Framework for routing & HTTP handler.
- [GORM](https://gorm.io/) → Database logics & queries.
- [PostgreSQL @ Supabase](https://www.supabase.com) → Free database.
- [Github CLI](https://cli.github.com/) → Github management for repository's secrets & etc.
- [Github Action](https://github.com/features/actions) → Automated testing and building across multiple versions of Go.
- [Snyk](https://app.snyk.io/) → Dependency scanning.
- [SonarLint as VSCode ext.](https://marketplace.visualstudio.com/items?itemName=SonarSource.sonarlint-vscode) → Detects & highlights issues that can lead to bugs & vulnerabilities.
- [GoLint](https://github.com/golang/lint) → CLI static code analytic for code-styling & many more.

&#xa0;

## Run Project

1. Clone project

```bash
git clone https://github.com/Lukmanern/gost && cd gost
```

2. Rename or copy the file .env.example to .env

3. Fill all the values in the .env file. For a quick setup, I suggest using Supabase for the database and Gmail for the system email.

4. Download all dependencies, turn on Redis, and then test the connections to the databases (DB and Redis).

```bash
go get -v ./... && go test ./database/...
```

## Github Action and Repository

1. Remove the .git directory to prevent repository cloning.

2. Create a repository on GitHub, but don't push initially. Ensure to add the Repository Secrets for GitHub Actions (SNYK_TOKEN and ENV).

3. Log in to Snyk, get the account token, and then add the token in the GitHub Repository Secret (named: SNYK_TOKEN) of the repository you created.

4. Also, add all .env values to the GitHub Repository Secret (named: ENV) for the repository. If you need a different database for GitHub Actions testing, you can modify the values.

5. Before committing and pushing, take a few minutes to review the GitHub Actions workflow at: `./.github/workflows/*.yml`

6. Once you understand each workflow, proceed with the commit and push!

&#xa0;

## Some Tips

1. You can use [Github-CLI](https://cli.github.com/) to set, remove, or update your GitHub Repository Secret.

```bash
> gh secret list
NAME        UPDATED
ENV         about 1 month ago
SNYK_TOKEN  about 3 months ago
```

2. You can receive advice from SonarLint and Golint. You don't always need to activate SonarLint; just enable it after ensuring your code runs normally. Then, commit the changes and do some code-finishing using SonarLint.

```bash
> golint ./...
domain\entity\role.go:6:6: exported type Role should have comment or be unexported
domain\entity\role.go:13:1: exported method Role.TableName should have comment or be unexported
domain\entity\user.go:10:6: exported type User should have comment or be unexported
domain\entity\user.go:20:1: exported method User.TableName should have comment or be unexported

...
```

3. Go to Web Snyk Dashboard, then you can add all your projects from Github and other platforms. Snyk will scan your project for potential security vulnerabilities and dependencies issues. It analyzes the codebase and dependencies, providing insights into known vulnerabilities, outdated packages, and best practices for secure coding.

&#xa0;

## :closed_book: Read List

- [Database Connection Configuration](https://www.alexedwards.net/blog/configuring-sqldb)
- [Go-Fiber Testing](https://dev.to/koddr/go-fiber-by-examples-testing-the-application-1ldf)
- [Production Checklist 1](https://aleksei-kornev.medium.com/production-readiness-checklist-for-backend-applications-8d2b0c57ccec/)
- [Production Checklist 2](https://github.com/gorrion-io/production-readiness-checklist/)
- [Production Checklist 3](https://www.cockroachlabs.com/docs/cockroachcloud/production-checklist/)
- [Deployment Checklist](https://last9.io/blog/deployment-readiness-checklists/)
- [CI with Github Actions](https://www.alexedwards.net/blog/ci-with-go-and-github-actions)
- [RestAPI Security Checklist](https://roadmap.sh/best-practices/api-security/)

&#xa0;

## :memo: License

This project is under license from MIT. For more details, see the [LICENSE](LICENSE) file.

&#xa0;
