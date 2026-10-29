# Fit the Meet frontend

React + TypeScript frontend using Mantine as its UI component library.

## Run locally

```bash
npm install
npm run dev
```

Vite proxies `/api` to the existing Go API at `http://127.0.0.1:8080`. Override `VITE_API_BASE_URL` when the API is served elsewhere.

## Existing backend routes used

- `POST /auth/register`
- `POST /auth/login`
- `DELETE /auth/logout`
- `GET /events/`
- `GET /events/{eventID}`
- `GET /events/code/{code}`
- `POST /events/`
- `POST /groups/{groupID}/events/`
- `PATCH /events/{eventID}`
- `PUT /events/{eventID}`

Authenticated event cards are loaded from the API. Opening a card fetches its full availability data, join links resolve through the event code endpoint, and event mutations refresh the detail from the backend. Demo mode still uses local sample data. Poll cards remain local until poll list and detail endpoints are added.
