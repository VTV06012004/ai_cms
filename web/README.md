# DDD CMS frontend

This frontend is the DDD-local copy of the existing CMS UI. Its Docker Compose
service listens on container port 3000, published to `http://localhost:3001`,
and calls the Go API at `http://localhost:8081/api`.

From the `DDD` directory:

```sh
docker compose up --build -d
```

Open [http://localhost:3001/login](http://localhost:3001/login). After changing
`NEXT_PUBLIC_API_URL`, rebuild the web image because Next.js embeds public
environment variables in the browser bundle at build time.
