FROM node:20-alpine

WORKDIR /app

# Copy package.json & package-lock.json để install cache
COPY package*.json ./

RUN npm install

# Copy toàn bộ source code
COPY . .

EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

# Chạy Next.js ở chế độ Development với Hot-Reloading
CMD ["npm", "run", "dev"]