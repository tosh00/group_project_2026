# Frontend: build Vite, potem nginx serwuje pliki statyczne i przekazuje /api do backendu.
#   docker build -t planer-web web
# Adres backendu: zmienna API_UPSTREAM (domyślnie http://server:8080, nazwa usługi z docker-compose.yml).

FROM node:22-alpine AS build
WORKDIR /app
RUN corepack enable

# Najpierw zależności, żeby zmiana kodu nie unieważniała cache pnpm.
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .
RUN pnpm build

FROM nginx:1.27-alpine
# Szablony z /etc/nginx/templates przechodzą przez envsubst przy starcie kontenera.
COPY nginx.conf.template /etc/nginx/templates/default.conf.template
COPY --from=build /app/dist /usr/share/nginx/html
ENV API_UPSTREAM=http://server:8080
EXPOSE 80
