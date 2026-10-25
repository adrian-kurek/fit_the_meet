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
- `POST /events/`
- `POST /groups/{groupID}/events/`
- `PATCH /events/{eventID}`
- `PUT /events/{eventID}`

The frontend does not require or make changes to the Go backend. Because the backend currently has no event-list/detail route and event creation returns no ID, the dashboard cards are preview data and the event drawer accepts an explicit event ID for status and availability mutations.
