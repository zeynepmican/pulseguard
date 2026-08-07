# frontend/Dockerfile

# Aşama 1: Bağımlılıkların yüklenmesi
FROM node:22-alpine AS deps
# Next.js ve Sharp için gerekli sistem kütüphanesi
RUN apk add --no-cache libc6-compat
WORKDIR /app

# Sadece package.json ve varsa package-lock.json dosyasını kopyala
COPY package.json package-lock.json* ./
# En risksiz yöntem olan standart npm install kullanıyoruz
RUN npm install

# Aşama 2: Projenin derlenmesi (Build)
FROM node:22-alpine AS builder
WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules
COPY . .
ENV NEXT_TELEMETRY_DISABLED 1
RUN npm run build

# Aşama 3: Çalıştırma (Runner)
FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV production
ENV NEXT_TELEMETRY_DISABLED 1

COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

EXPOSE 3000
ENV PORT 3000
ENV HOSTNAME "0.0.0.0"

CMD ["node", "server.js"]