FROM node:22-bookworm AS builder-fe
WORKDIR /app
RUN corepack enable
COPY package.json package-lock.json ./
RUN npm ci
COPY client ./client
COPY package.json ./
RUN npm run build

FROM golang:1.23-bookworm AS builder-be
WORKDIR /app
COPY backend-go ./
RUN go mod tidy
RUN CGO_ENABLED=1 GOOS=linux go build -o boop-cat .

FROM node:22-bookworm
WORKDIR /app

ENV BUN_INSTALL=/usr/local/bun
ENV PATH=/usr/local/bun/bin:$PATH
ENV DENO_INSTALL=/usr/local/deno
ENV PATH=/usr/local/deno/bin:$PATH
ENV NODE_ENV=production

RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    bash \
    unzip \
    python3 \
    make \
    g++ \
  && rm -rf /var/lib/apt/lists/*

RUN corepack enable \
  && corepack prepare pnpm@9.0.0 --activate \
  && corepack prepare yarn@1.22.22 --activate

RUN curl -fsSL https://bun.sh/install | bash \
  && ln -sf /usr/local/bun/bin/bun /usr/local/bin/bun

RUN curl -fsSL https://deno.land/install.sh | sh \
  && ln -sf /usr/local/deno/bin/deno /usr/local/bin/deno

COPY --from=builder-fe /app/client/dist /client/dist

COPY --from=builder-be /app/boop-cat /app/boop-cat

ENV PORT=8788
EXPOSE 8788

CMD ["/app/boop-cat"]
