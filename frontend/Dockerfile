FROM node:22-alpine AS builder

WORKDIR /app

RUN npm install -g pnpm

COPY package.json pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile || pnpm install

COPY . .

ARG VITE_API_BASE_URL=""
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

RUN pnpm build

FROM node:22-alpine

WORKDIR /app

COPY --from=builder /app/.output .output

EXPOSE 3032

CMD ["node", ".output/server/index.mjs"]
