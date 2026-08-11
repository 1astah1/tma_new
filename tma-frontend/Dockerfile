FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .

# Vite подставляет VITE_* на этапе сборки, поэтому это build-arg, а не runtime-env.
ARG VITE_BOT_USERNAME
ARG VITE_API_URL
ENV VITE_BOT_USERNAME=$VITE_BOT_USERNAME
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
