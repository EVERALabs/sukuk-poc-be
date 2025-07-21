# Sukuk Platform API Documentation

## Overview
This document outlines all API endpoints, their data sources, correlations to blockchain events, and usage scenarios in the Sukuk platform.

---

## Public APIs (No Authentication Required)

### 1. Company APIs

#### GET /api/v1/companies
**Purpose**: List all active companies  
**Data Source**: `companies` table  
**Event Correlation**: None (static data)  
**When Called**: 
- Frontend home page to show available issuers
- Company selection dropdowns
- Public company directory

**Response Data**:
```json
{
  "data": [
    {
      "id": 1,
      "name": "PT Sukuk Indonesia",
      "description": "Leading sukuk issuer",
      "website": "https://example.com",
      "industry": "Financial Services",
      "logo": "/uploads/logos/company_1_logo.png",
      "is_active": true
    }
  ],
  "count": 1,
  "meta": { "total": 1, "page": 1 }
}
```

#### GET /api/v1/companies/{id}
**Purpose**: Get detailed company information with sukuk series  
**Data Source**: `companies` table + `sukuk_series` table  
**Event Correlation**: None (static data)  
**When Called**: 
- Company profile pages
- Detailed company information views

**Response Data**:
```json
{
  "data": {
    "id": 1,
    "name": "PT Sukuk Indonesia",
    "description": "Leading sukuk issuer",
    "sukuk_series": [
      {
        "id": 1,
        "name": "Green Sukuk Series A",
        "symbol": "GSA",
        "status": "active"
      }
    ]
  }
}
```

#### GET /api/v1/companies/{id}/sukuks
**Purpose**: Get all sukuk series for a specific company  
**Data Source**: `sukuk_series` table  
**Event Correlation**: SukukDeployed events update token_address  
**When Called**: 
- Company-specific sukuk listings
- Investment opportunity browsing

**Response Data**:
```json
{
  "data": [
    {
      "id": 1,
      "company_id": 1,
      "name": "Green Sukuk Series A",
      "symbol": "GSA",
      "total_supply": "1000000000000000000000000",
      "yield_rate": 0.085,
      "maturity_date": "2027-12-31T00:00:00Z",
      "token_address": "0xabc123...",
      "status": "active"
    }
  ],
  "count": 1
}
```

### 2. Sukuk APIs

#### GET /api/v1/sukuks
**Purpose**: List all available sukuk series with filtering  
**Data Source**: `sukuk_series` table + `companies` table  
**Event Correlation**: SukukDeployed events update status and token_address  
**When Called**: 
- Main investment dashboard
- Sukuk marketplace browsing
- Investment opportunity discovery

**Query Parameters**:
- `company_id`: Filter by company
- `status`: Filter by status (active, paused, matured)

**Response Data**:
```json
{
  "data": [
    {
      "id": 1,
      "name": "Green Sukuk Series A",
      "symbol": "GSA",
      "company": {
        "id": 1,
        "name": "PT Sukuk Indonesia"
      },
      "total_supply": "1000000000000000000000000",
      "outstanding_supply": "500000000000000000000000",
      "yield_rate": 0.085,
      "maturity_date": "2027-12-31T00:00:00Z",
      "min_investment": "1000000000000000000",
      "max_investment": "100000000000000000000",
      "token_address": "0xabc123...",
      "status": "active"
    }
  ],
  "count": 1
}
```

#### GET /api/v1/sukuks/{id}
**Purpose**: Get detailed sukuk information with related data  
**Data Source**: `sukuk_series` + `companies` + `investments` + `yield_claims` + `redemptions`  
**Event Correlation**: Multiple events update related data  
**When Called**: 
- Sukuk detail pages
- Investment decision pages
- Performance tracking

**Response Data**:
```json
{
  "data": {
    "id": 1,
    "name": "Green Sukuk Series A",
    "company": {
      "id": 1,
      "name": "PT Sukuk Indonesia"
    },
    "investments": [
      {
        "id": 1,
        "investor_address": "0xdef456...",
        "investment_amount": "10000000000000000000",
        "token_amount": "10000000000000000000",
        "status": "active"
      }
    ],
    "yield_claims": [],
    "redemptions": []
  }
}
```

#### GET /api/v1/sukuks/{id}/metrics
**Purpose**: Get performance metrics for sukuk  
**Data Source**: Aggregated data from `investments`, `yield_claims`, `redemptions`  
**Event Correlation**: InvestmentMade, YieldClaimed, RedemptionRequested events  
**When Called**: 
- Dashboard metrics
- Performance analytics
- Admin monitoring

**Response Data**:
```json
{
  "data": {
    "total_investors": 25,
    "total_investment": "250000000000000000000000",
    "pending_yields": 5,
    "pending_redemptions": 2
  }
}
```

#### GET /api/v1/sukuks/{id}/holders
**Purpose**: Get current active holders of sukuk  
**Data Source**: `investments` table (active only)  
**Event Correlation**: InvestmentMade and RedemptionCompleted events  
**When Called**: 
- Holder analytics
- Distribution lists
- Compliance reporting

**Response Data**:
```json
{
  "data": [
    {
      "id": 1,
      "investor_address": "0xdef456...",
      "investment_amount": "10000000000000000000",
      "token_amount": "10000000000000000000",
      "investment_date": "2024-01-15T10:30:00Z",
      "status": "active"
    }
  ],
  "count": 1
}
```

### 3. Investment APIs

#### GET /api/v1/investments
**Purpose**: List investments with filtering  
**Data Source**: `investments` table + `sukuk_series` + `companies`  
**Event Correlation**: InvestmentMade events create records  
**When Called**: 
- User portfolio views
- Investment history
- Admin investment monitoring

**Query Parameters**:
- `investor_address`: Filter by investor
- `sukuk_series_id`: Filter by sukuk
- `status`: Filter by status

**Response Data**:
```json
{
  "data": [
    {
      "id": 1,
      "sukuk_series": {
        "id": 1,
        "name": "Green Sukuk Series A",
        "company": {
          "name": "PT Sukuk Indonesia"
        }
      },
      "investor_address": "0xdef456...",
      "investment_amount": "10000000000000000000",
      "token_amount": "10000000000000000000",
      "token_price": "1000000000000000000",
      "investment_date": "2024-01-15T10:30:00Z",
      "status": "active"
    }
  ],
  "count": 1
}
```

#### GET /api/v1/investments/{investor_address}
**Purpose**: Get all investments for a specific investor  
**Data Source**: `investments` table filtered by investor  
**Event Correlation**: InvestmentMade events  
**When Called**: 
- Personal portfolio dashboard
- Investment tracking
- User account pages

### 4. Yield APIs

#### GET /api/v1/yields
**Purpose**: List yield distributions and claims  
**Data Source**: `yield_claims` table + `sukuk_series`  
**Event Correlation**: YieldDistributed and YieldClaimed events  
**When Called**: 
- Yield history views
- Claim tracking
- Income reporting

**Query Parameters**:
- `investor_address`: Filter by investor
- `sukuk_series_id`: Filter by sukuk
- `status`: Filter by claim status

**Response Data**:
```json
{
  "data": [
    {
      "id": 1,
      "sukuk_series": {
        "id": 1,
        "name": "Green Sukuk Series A"
      },
      "investor_address": "0xdef456...",
      "yield_amount": "850000000000000000",
      "distribution_date": "2024-03-31T00:00:00Z",
      "claim_date": "2024-04-01T10:15:00Z",
      "status": "claimed"
    }
  ],
  "count": 1
}
```

#### GET /api/v1/yields/{investor_address}
**Purpose**: Get yield history for specific investor  
**Data Source**: `yield_claims` table filtered by investor  
**Event Correlation**: YieldDistributed and YieldClaimed events  
**When Called**: 
- Personal yield dashboard
- Income tracking
- Tax reporting

### 5. Redemption APIs

#### GET /api/v1/redemptions
**Purpose**: List redemption requests and completions  
**Data Source**: `redemptions` table + `sukuk_series`  
**Event Correlation**: RedemptionRequested and RedemptionCompleted events  
**When Called**: 
- Redemption tracking
- Exit position monitoring
- Admin redemption management

**Query Parameters**:
- `investor_address`: Filter by investor
- `sukuk_series_id`: Filter by sukuk
- `status`: Filter by redemption status

**Response Data**:
```json
{
  "data": [
    {
      "id": 1,
      "sukuk_series": {
        "id": 1,
        "name": "Green Sukuk Series A"
      },
      "investor_address": "0xdef456...",
      "token_amount": "5000000000000000000",
      "redemption_amount": "5250000000000000000",
      "request_date": "2024-06-15T14:30:00Z",
      "completed_at": "2024-06-16T09:00:00Z",
      "status": "completed"
    }
  ],
  "count": 1
}
```

#### GET /api/v1/redemptions/{investor_address}
**Purpose**: Get redemption history for specific investor  
**Data Source**: `redemptions` table filtered by investor  
**Event Correlation**: RedemptionRequested and RedemptionCompleted events  
**When Called**: 
- Personal redemption history
- Exit tracking
- Portfolio management

---

## Admin APIs (Authentication Required)

### 1. Company Management

#### POST /api/v1/admin/companies
**Purpose**: Create new partner company  
**Data Source**: Creates record in `companies` table  
**Event Correlation**: None (off-chain operation)  
**When Called**: 
- Adding new sukuk issuers
- Partner onboarding

**Request Body**:
```json
{
  "name": "PT Sukuk Indonesia",
  "description": "Leading sukuk issuer",
  "website": "https://example.com",
  "industry": "Financial Services",
  "email": "contact@example.com",
  "wallet_address": "0x1234567890123456789012345678901234567890"
}
```

#### PUT /api/v1/admin/companies/{id}
**Purpose**: Update company information  
**Data Source**: Updates `companies` table  
**Event Correlation**: None (off-chain operation)  
**When Called**: 
- Company profile updates
- Status changes
- Information corrections

#### POST /api/v1/admin/companies/{id}/upload-logo
**Purpose**: Upload company logo  
**Data Source**: File system + updates `companies.logo`  
**Event Correlation**: None  
**When Called**: 
- Branding updates
- Company profile setup

### 2. Sukuk Management

#### POST /api/v1/admin/sukuks
**Purpose**: Create new sukuk series (off-chain preparation)  
**Data Source**: Creates record in `sukuk_series` table  
**Event Correlation**: Prepares for SukukDeployed event  
**When Called**: 
- Before smart contract deployment
- Sukuk series setup

**Request Body**:
```json
{
  "company_id": 1,
  "name": "Green Sukuk Series A",
  "symbol": "GSA",
  "description": "Sustainable infrastructure financing",
  "total_supply": "1000000000000000000000000",
  "yield_rate": 0.085,
  "maturity_date": "2027-12-31T00:00:00Z",
  "payment_frequency": 4,
  "min_investment": "1000000000000000000",
  "max_investment": "100000000000000000000",
  "is_redeemable": true
}
```

#### PUT /api/v1/admin/sukuks/{id}
**Purpose**: Update sukuk series information  
**Data Source**: Updates `sukuk_series` table  
**Event Correlation**: May sync with SukukDeployed event data  
**When Called**: 
- Post-deployment updates
- Status changes
- Token address updates

#### POST /api/v1/admin/sukuks/{id}/upload-prospectus
**Purpose**: Upload PDF prospectus  
**Data Source**: File system + updates `sukuk_series.prospectus`  
**Event Correlation**: None  
**When Called**: 
- Legal document management
- Compliance requirements

### 3. System Management

#### GET /api/v1/admin/system/sync-status
**Purpose**: Check blockchain sync status  
**Data Source**: `system_states` table  
**Event Correlation**: Tracks last processed event  
**When Called**: 
- System health monitoring
- Sync progress tracking

#### POST /api/v1/admin/system/force-sync
**Purpose**: Trigger manual blockchain sync  
**Data Source**: Processes events from blockchain  
**Event Correlation**: Processes all pending events  
**When Called**: 
- Manual sync operations
- System recovery

---

## Event-to-API Data Flow

### 1. SukukDeployed Event
**Triggered When**: Smart contract deployment completes  
**Updates**: 
- `sukuk_series.token_address`
- `sukuk_series.status` → "active"

**Affected APIs**:
- GET /api/v1/sukuks (shows new active sukuk)
- GET /api/v1/sukuks/{id} (updated with token address)
- GET /api/v1/companies/{id}/sukuks (shows deployed sukuk)

### 2. InvestmentMade Event
**Triggered When**: User successfully invests in sukuk  
**Creates**: New record in `investments` table  
**Updates**: 
- `sukuk_series.outstanding_supply` (increases)

**Affected APIs**:
- GET /api/v1/investments (shows new investment)
- GET /api/v1/investments/{investor_address} (personal portfolio)
- GET /api/v1/sukuks/{id}/metrics (updated investor count)
- GET /api/v1/sukuks/{id}/holders (new holder)

### 3. YieldDistributed Event
**Triggered When**: Company deposits yield for distribution  
**Creates**: Records in `yield_claims` table (one per eligible investor)  
**Status**: All claims start as "pending"

**Affected APIs**:
- GET /api/v1/yields (shows new pending claims)
- GET /api/v1/yields/{investor_address} (personal yield tracking)
- GET /api/v1/sukuks/{id}/metrics (updated pending yields count)

### 4. YieldClaimed Event
**Triggered When**: Investor claims their yield  
**Updates**: 
- `yield_claims.status` → "claimed"
- `yield_claims.claim_date`

**Affected APIs**:
- GET /api/v1/yields (updated claim status)
- GET /api/v1/yields/{investor_address} (personal claim history)
- GET /api/v1/sukuks/{id}/metrics (decreased pending yields)

### 5. RedemptionRequested Event
**Triggered When**: Investor requests to redeem tokens  
**Creates**: New record in `redemptions` table  
**Status**: "requested"

**Affected APIs**:
- GET /api/v1/redemptions (shows new request)
- GET /api/v1/redemptions/{investor_address} (personal redemption tracking)
- GET /api/v1/sukuks/{id}/metrics (updated pending redemptions)

### 6. RedemptionCompleted Event
**Triggered When**: Redemption is processed and IDRX transferred  
**Updates**: 
- `redemptions.status` → "completed"
- `redemptions.completed_at`
- `investments.status` → "redeemed" (if full redemption)
- `sukuk_series.outstanding_supply` (decreases)

**Affected APIs**:
- GET /api/v1/redemptions (updated status)
- GET /api/v1/investments (updated investment status)
- GET /api/v1/sukuks/{id}/holders (removed if fully redeemed)
- GET /api/v1/sukuks/{id}/metrics (updated metrics)

---

## API Usage Scenarios

### Frontend User Flows

1. **Browse Investments**:
   - GET /api/v1/companies → GET /api/v1/sukuks → GET /api/v1/sukuks/{id}

2. **Check Portfolio**:
   - GET /api/v1/investments/{investor_address}
   - GET /api/v1/yields/{investor_address}
   - GET /api/v1/redemptions/{investor_address}

3. **Company Research**:
   - GET /api/v1/companies/{id} → GET /api/v1/companies/{id}/sukuks

### Admin Flows

1. **Onboard New Company**:
   - POST /api/v1/admin/companies
   - POST /api/v1/admin/companies/{id}/upload-logo

2. **Launch Sukuk Series**:
   - POST /api/v1/admin/sukuks (prepare off-chain)
   - Deploy smart contract (off-platform)
   - PUT /api/v1/admin/sukuks/{id} (update with token address)
   - POST /api/v1/admin/sukuks/{id}/upload-prospectus

3. **Monitor System**:
   - GET /api/v1/admin/system/sync-status
   - GET /api/v1/sukuks/{id}/metrics (for all active sukuks)

### Integration Scenarios

1. **Blockchain Event Processing**:
   - Events trigger background sync service
   - Database updates happen automatically
   - APIs reflect updated state immediately

2. **External Analytics**:
   - GET /api/v1/sukuks (all sukuk data)
   - GET /api/v1/sukuks/{id}/metrics (performance data)
   - GET /api/v1/investments (investment flow data)

---

## Notes

1. **Real-time Updates**: All APIs reflect the latest blockchain state after event processing
2. **Data Consistency**: Event processing ensures data consistency between blockchain and database
3. **Performance**: Metrics APIs use aggregated data for fast response times
4. **Security**: Admin APIs require authentication, public APIs are read-only
5. **Filtering**: Most list APIs support filtering for efficient data retrieval