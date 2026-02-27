# Obligacje

A self-hosted API for valuing Polish government savings bonds (obligacje skarbowe). It periodically fetches bond data published by the Ministry of Finance and exposes a simple HTTP endpoint to calculate the current (or historical) value of a bond.

## Public Instance

A publicly hosted instance is available at **https://obligacje.mionskowski.pl**.

## Self-Hosting

The easiest way to run Obligacje is via Docker. The image requires **LibreOffice Calc** (included in the image) for converting the XLS files.

```bash
docker run -d \
  --name obligacje \
  -p 8080:8080 \
  -v obligacje-data:/data \
  ghcr.io/maciekmm/obligacje:latest
```

The server starts on port **8080** and persists downloaded bond data in the `/data` volume.

## API

### `GET /v1/bond/{name}/valuation`

Returns the current valuation of a bond.

#### Path Parameters

| Parameter | Description |
|-----------|-------------|
| `name`    | Bond series name followed by a two-digit purchase day, e.g. `TOS0125` + day `15` → `TOS012515` |

#### Query Parameters

| Parameter      | Required | Description |
|----------------|----------|-------------|
| `valuated_at`  | No       | Valuation date in `YYYY-MM-DD` format. Defaults to today. |

#### Response Formats

The response format is controlled by the `Accept` header.

##### `text/plain` (default)

Returns the bond price as a plain number:

```
102.72
```

##### `application/json`

Returns a JSON object with full valuation details:

```json
{
  "name": "TOS012515",
  "isin": "PL0000...",
  "valuated_at": "2026-02-27",
  "purchase_day": 15,
  "price": 102.72,
  "currency": "PLN"
}
```

#### Error Responses

| Status | Reason |
|--------|--------|
| `400`  | Invalid bond name or `valuated_at` format, or valuation date is before the bond's purchase date |
| `404`  | Bond series not found |
| `500`  | Internal server error |
