# Cais

Tempos de espera do metro de Lisboa, estação a estação. App não oficial.

- `main.go`, `server/`, `services/`: o servidor Go, que fala com a API do Metro e serve o frontend.
- `frontend/`: o frontend em Svelte 5 + Vite. O build vai para `web/dist`, que o binário Go embute.
- `cmd/cais-mac/`: a app para macOS (o mesmo servidor e frontend numa janela WebKit nativa).

Precisa de Go 1.27 e Node 22 ou mais recente.

## Correr localmente

```sh
make run
```

Abre http://localhost:8080. Sem token da API do Metro, os dados ao vivo vêm do servidor deployado
(`METRO_BACKEND_URL`). Com token, põe `METRO_API_URL` e `METRO_API_TOKEN` num `.env` na raiz (ver
`.env.example`) ou exporta-os.

Para mexer no frontend com hot reload, em dois terminais:

```sh
make dev-api   # API em :8080
make dev-web   # Vite em http://localhost:5173
```

## App para Mac

```sh
make mac
open build/Cais.app
```

A app arranca o servidor dentro dela própria. Por omissão vai buscar os dados ao servidor deployado;
para usar um token próprio, cria `~/Library/Application Support/Cais/config.json`:

```json
{ "api_url": "https://…", "api_token": "…" }
```

ou `{ "backend_url": "https://…" }` para apontar para outro servidor.

## Deploy

- `caislx.tech`: o frontend, ficheiros estáticos servidos pelo nginx a partir de `/var/www/cais/www`.
- `api.caislx.tech`: o servidor Go (serviço systemd `cais`, porta 8081), atrás do nginx.

No servidor fica tudo em `/var/www/cais`: o binário `cais`, o `.env` (ver `.env.example`), `data/`
(chaves e subscrições das notificações, que têm de sobreviver aos deploys) e `www/`.

Preparar o servidor, uma vez:

```sh
scp deploy/cais.service deploy/nginx/caislx.tech.conf .env.example <user>@<vm>:/tmp/
ssh <user>@<vm> 'sudo bash -s' < scripts/setup-server.sh
```

Depois, cada push para `master` corre os testes, faz o build do frontend (apontado para
`https://api.caislx.tech`) e do binário Linux, copia os dois para `/var/www/cais` e reinicia o serviço
(`.github/workflows/deploy.yml`).
