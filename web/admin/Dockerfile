FROM rust:1.84 AS builder

RUN cargo install trunk && rustup target add wasm32-unknown-unknown

RUN curl -fsSL https://deb.nodesource.com/setup_18.x | bash - && \
    apt-get install -y nodejs

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm install

COPY . .

ENV BASE_URL=https://gateway.local.com/api
RUN trunk build --release

FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]

