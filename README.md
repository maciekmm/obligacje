# Obligacje

A self-hosted API for valuing Polish government savings bonds (obligacje skarbowe). It periodically fetches bond data published by the Ministry of Finance and exposes a simple HTTP endpoint to calculate the current (or historical) value of a bond.

## Public Instance

A publicly hosted instance is available at **https://obligacje.mionskowski.pl**.

```bash
# Get current valuation of a TOS bond (purchased on the 1st)
curl https://obligacje.mionskowski.pl/v1/bond/TOS112501/valuation

# Get valuation as JSON for a specific date
curl -H "Accept: application/json" \
  "https://obligacje.mionskowski.pl/v1/bond/TOS112501/valuation?valuated_at=2025-12-06"

# Get historical pricing over a date range
curl "https://obligacje.mionskowski.pl/v1/bond/TOS112501/historical?from=2025-12-01&to=2025-12-07"
```

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

---

### `GET /v1/bond/{name}/historical`

Returns daily valuations of a bond over a date range.

#### Path Parameters

| Parameter | Description |
|-----------|-------------|
| `name`    | Bond series name followed by a two-digit purchase day, e.g. `TOS012515` |

#### Query Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `from`    | Yes      | Start date in `YYYY-MM-DD` format |
| `to`      | Yes      | End date in `YYYY-MM-DD` format |

The maximum span between `from` and `to` is **366 days**. Days before the bond's purchase date are omitted from the result.

#### Response

Always returns `application/json`:

```json
{
  "valuations": {
    "2026-02-25": 102.70,
    "2026-02-26": 102.71,
    "2026-02-27": 102.72
  }
}
```

#### Error Responses

| Status | Reason |
|--------|--------|
| `400`  | Missing or invalid `from`/`to`, `to` before `from`, span exceeds 366 days, or invalid bond name |
| `404`  | Bond series not found |
| `500`  | Internal server error |
