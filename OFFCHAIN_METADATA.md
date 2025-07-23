# Adding Offchain Metadata to Sukuk

This guide explains how to add offchain business metadata to sukuk that have been synced from the blockchain indexer.

## Overview

The sukuk metadata system works with a two-stage process:

1. **Onchain Data**: Automatically synced from blockchain indexer (contract address, token ID, transaction hash, etc.)
2. **Offchain Metadata**: Manually added business information (tenor, imbal hasil, kuota nasional, etc.)

Only sukuk with both onchain and offchain data that are marked as `metadata_ready: true` will be displayed in the main API responses.

## Step-by-Step Process

### Step 1: Query All Sukuk Metadata

**⚠️ Important:** Before syncing new data from the indexer, always check existing sukuk to avoid duplicate insertion errors. The system uses contract address as unique identifier.

First, get a list of all sukuk including those that are not ready:

```bash
curl -X GET "http://localhost:8080/api/v1/sukuk-metadata"
```

Or filter to see only unready sukuk (those missing offchain metadata):

```bash
curl -X GET "http://localhost:8080/api/v1/sukuk-metadata?ready=false"
```

**Response example:**
```json
[
  {
    "id": 36,
    "sukuk_code": "PTS",
    "sukuk_title": "Ponder Test Sukuk",
    "contract_address": "0xb5b8e5FA5A6DE8Dc19482703cBdD6F4c5ee1d573",
    "transaction_hash": "0x97f55fb87c37dab52248421988c2191da6d600cda2e3ab7fa82e8786cd0458f4",
    "block_number": 28717483,
    "tenor": "",
    "imbal_hasil": "",
    "kuota_nasional": 0,
    "metadata_ready": false
  }
]
```

### Step 2: Add Offchain Metadata

Use the PUT endpoint to add business metadata to a specific sukuk by ID:

```bash
curl -X PUT http://localhost:8080/api/v1/sukuk-metadata/36 \
  -H "Content-Type: application/json" \
  -d '{
    "sukuk_deskripsi": "Sukuk untuk testing dan pengembangan platform blockchain",
    "tenor": "5 Tahun",
    "imbal_hasil": "6.25% / Tahun",
    "periode_pembelian": "1 Agustus - 31 Agustus 2025",
    "jatuh_tempo": "2030-08-01T00:00:00Z",
    "kuota_nasional": 1000000000,
    "penerimaan_kupon": "Bulanan",
    "minimum_pembelian": 1000000,
    "tanggal_bayar_kupon": "1 Setiap Bulan",
    "maksimum_pembelian": 50000000,
    "kupon_pertama": "2025-09-01T00:00:00Z",
    "tipe_kupon": "Fixed Rate",
    "logo_url": "https://example.com/sukuk-logo.png"
  }'
```

**Available offchain metadata fields:**
- `sukuk_deskripsi` (string) - Description of the sukuk
- `tenor` (string) - Duration (e.g., "5 Tahun")
- `imbal_hasil` (string) - Return rate (e.g., "6.25% / Tahun")
- `periode_pembelian` (string) - Purchase period
- `jatuh_tempo` (datetime) - Maturity date
- `kuota_nasional` (number) - National quota
- `penerimaan_kupon` (string) - Coupon frequency (e.g., "Bulanan")
- `minimum_pembelian` (number) - Minimum purchase amount
- `tanggal_bayar_kupon` (string) - Coupon payment date
- `maksimum_pembelian` (number) - Maximum purchase amount
- `kupon_pertama` (datetime) - First coupon date
- `tipe_kupon` (string) - Coupon type (e.g., "Fixed Rate")
- `logo_url` (string) - Logo URL
- `status` (string) - Status (e.g., "berlangsung")

**Note:** All fields are optional. You can update only specific fields as needed.

### Step 3: Mark as Ready

Once you've added all the necessary offchain metadata, mark the sukuk as ready:

```bash
curl -X PUT http://localhost:8080/api/v1/sukuk-metadata/36/ready
```

**Response:**
```json
{
  "id": 36,
  "sukuk_code": "PTS", 
  "sukuk_title": "Ponder Test Sukuk",
  "contract_address": "0xb5b8e5FA5A6DE8Dc19482703cBdD6F4c5ee1d573",
  "tenor": "5 Tahun",
  "imbal_hasil": "6.25% / Tahun",
  "kuota_nasional": 1000000000,
  "metadata_ready": true,
  "created_at": "2025-07-24T02:04:00.741438+07:00",
  "updated_at": "2025-07-24T02:15:45.695844+07:00"
}
```

### Step 4: Verify Ready Sukuk

Check that your sukuk now appears in the ready list:

```bash
curl -X GET "http://localhost:8080/api/v1/sukuk-metadata?ready=true"
```

## API Endpoints Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/sukuk-metadata` | List all sukuk metadata |
| GET | `/api/v1/sukuk-metadata?ready=true` | List only ready sukuk |
| GET | `/api/v1/sukuk-metadata?ready=false` | List only unready sukuk |
| PUT | `/api/v1/sukuk-metadata/{id}` | Update sukuk with offchain metadata |
| PUT | `/api/v1/sukuk-metadata/{id}/ready` | Mark sukuk as ready |
| POST | `/api/v1/sukuk-metadata/sync` | Sync sukuk from blockchain indexer |

## Avoiding Duplicate Sync Errors

### Check Before Syncing

Before attempting to sync from the indexer, check if the sukuk already exists:

```bash
# Check if a specific contract address already exists
curl -s "http://localhost:8080/api/v1/sukuk-metadata" | jq '.[] | select(.contract_address=="0xYourContractAddress")'
```

### Database Constraints

The system enforces uniqueness on:
- ✅ **Contract Address** - Each contract address can only have one sukuk metadata entry
- ❌ **Sukuk Code** - Multiple sukuk can have the same code (but different contracts)

### Sync Error Prevention

If you get this error:
```
ERROR: duplicate key value violates unique constraint "uni_sukuk_metadata_contract_address"
```

It means a sukuk with that contract address already exists. Use the update workflow instead:

1. Find the existing sukuk ID
2. Use `PUT /sukuk-metadata/{id}` to update with new metadata
3. Mark as ready when complete

## Example Workflow Script

Here's a bash script example to automate the process safely:

```bash
#!/bin/bash

CONTRACT_ADDRESS="0xb5b8e5FA5A6DE8Dc19482703cBdD6F4c5ee1d573"

# Step 1: Check if sukuk already exists
echo "Checking if sukuk exists for contract: $CONTRACT_ADDRESS"
EXISTING_SUKUK=$(curl -s "http://localhost:8080/api/v1/sukuk-metadata" | jq ".[] | select(.contract_address==\"$CONTRACT_ADDRESS\")")

if [ "$EXISTING_SUKUK" != "" ]; then
    SUKUK_ID=$(echo "$EXISTING_SUKUK" | jq -r '.id')
    echo "Found existing sukuk with ID: $SUKUK_ID"
else
    echo "No existing sukuk found. Sync from indexer first."
    exit 1
fi

# Step 2: Add metadata to existing sukuk
echo "Adding metadata to sukuk ID: $SUKUK_ID"

curl -X PUT "http://localhost:8080/api/v1/sukuk-metadata/$SUKUK_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "tenor": "5 Tahun",
    "imbal_hasil": "6.25% / Tahun",
    "kuota_nasional": 1000000000,
    "penerimaan_kupon": "Bulanan"
  }'

# Step 3: Mark as ready
echo "Marking sukuk as ready..."
curl -X PUT "http://localhost:8080/api/v1/sukuk-metadata/$SUKUK_ID/ready"

# Step 4: Verify
echo "Ready sukuk count:"
curl -s "http://localhost:8080/api/v1/sukuk-metadata?ready=true" | jq 'length'
```

## Important Notes

1. **Contract address uniqueness**: Each contract address can only have one sukuk metadata entry.

2. **Onchain data is preserved**: When updating with offchain metadata, all blockchain data (contract address, transaction hash, etc.) is preserved.

3. **Partial updates supported**: You don't need to provide all fields - only update the fields you need.

4. **Ready flag**: Only sukuk marked as `metadata_ready: true` will appear in the main API responses by default.

5. **Indexer sync**: New sukuk from the blockchain are automatically synced but start with `metadata_ready: false` until offchain metadata is added.

6. **Check before sync**: Always query existing data before attempting to sync from indexer to avoid duplicate key errors.

## Troubleshooting

- **404 Not Found**: Check that the sukuk ID exists in the database
- **400 Bad Request**: Verify your JSON payload is valid
- **Duplicate key error on contract_address**: The sukuk already exists, use update workflow instead
- **Empty response**: Use the filtering parameters to see unready sukuk
- **Sync failures**: Check server logs for specific database constraint violations