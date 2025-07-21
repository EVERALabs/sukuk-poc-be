# Blockchain Events for Sukuk Platform

## Overview
This document outlines all blockchain events that need to be indexed for the Sukuk platform. The indexer will listen to these events and store them in a database, which the backend will then query to serve data to the frontend.

## Architecture
```
Blockchain → Indexer → Indexed DB → Backend (Read Only) → Frontend
                           ↓
                    Business DB
```

## Event Definitions

### 1. SukukDeployed Event
**When Emitted**: When a new Sukuk series smart contract is deployed

**Event Structure**:
```solidity
event SukukDeployed(
    address indexed tokenAddress,
    string seriesName,
    string symbol,
    uint256 totalSupply,
    uint256 yieldRateProjectile,      // Projected yield rate in basis points (850 = 8.5%)
    uint256 maturityDate,   // Unix timestamp
    uint256 minInvestment,
    address indexed issuer
);
```

**Backend Processing**:
- Find `sukuk_series` record by `seriesName`
- Update `token_address` field
- Validate issuer matches company wallet address
- Update status to `active`

---

### 2. Investment Event
**When Emitted**: When an investor purchases Sukuk tokens

**Event Structure**:
```solidity
event Investment(
    address indexed investor,
    address indexed sukukToken,
    uint256 idrxAmount,                   // Amount paid in IDRX
    uint256 tokenAmount,                  // Sukuk tokens received
    uint256 tokenPrice,                   // Price at time of purchase
    // Previous state for audit trail
    uint256 previousOutstandingSupply,    // Supply before this investment
    uint256 newOutstandingSupply,         // Supply after this investment
    uint256 timestamp
);
```

**Backend Processing**:
- Create new `investment` record
- Link to `sukuk_series` via token address
- Update `outstanding_supply` in sukuk_series
- Store purchase price for future calculations

**Frontend Needs**:
- Show in transaction history
- Update portfolio balance
- Calculate average purchase price

---

### 3. YieldDistributed Event
**When Emitted**: When the company deposits IDRX profits into the vault for investors to claim

**Event Structure**:
```solidity
event YieldDistributed(
    uint256 indexed distributionId,    // Unique ID for this distribution
    address indexed sukukToken,
    uint256 totalYieldAmount,          // Total IDRX amount deposited for distribution
    uint256 periodStart,               // Business/investment period start (e.g., Q1 start date)
    uint256 periodEnd,                 // Business/investment period end (e.g., Q1 end date)
    uint256 yieldPerToken,             // IDRX yield per token (with decimals)
    uint256 timestamp                  // When the distribution was made
);
```

**Backend Processing**:
- Verify IDRX has been deposited to the vault contract
- Create `yield_claim` records for all active investments in this sukuk
- Calculate each investor's portion based on their token holdings
- Note: No claim expiry - investors can claim anytime after distribution

**Frontend Needs**:
- Notify investors of new yield available
- Show claimable amount
- Display payment period details

---

### 4. YieldClaimed Event
**When Emitted**: When an investor claims their yield payment

**Event Structure**:
```solidity
event YieldClaimed(
    address indexed investor,
    address indexed sukukToken,
    uint256 yieldAmount,           // Total IDRX amount claimed
    uint256 fromDistribution,      // Oldest distribution ID included in this claim
    uint256 toDistribution         // Latest distribution ID included in this claim
);
```

**Backend Processing**:
- Record the claim with transaction details
- Update investor's claim history
- Track which distribution range was claimed
- Note: With cumulative tracking, one claim may cover multiple distributions

**Frontend Needs**:
- Update yield history
- Show transaction confirmation
- Update total yield earned

---

### 5. RedemptionRequested Event
**When Emitted**: When an investor requests early redemption

**Event Structure**:
```solidity
event RedemptionRequested(
    bytes32 indexed redemptionId,
    address indexed investor,
    address indexed sukukToken,
    uint256 tokenAmount,
    string reason,
    uint256 timestamp
);
```

**Backend Processing**:
- Create `redemption` record with "requested" status
- Note: Tokens are locked in smart contract

**Frontend Needs**:
- Show pending redemption request
- Display locked token amount
- Track request status

---

### 6. RedemptionApproved Event
**When Emitted**: When the issuer approves a redemption request

**Event Structure**:
```solidity
event RedemptionApproved(
    bytes32 indexed redemptionId,
    address indexed approver,
    uint256 redemptionAmount,   // IDRX to be paid
    string approvalNotes,
    uint256 timestamp
);
```

**Backend Processing**:
- Update redemption status to "approved"
- Store redemption amount and approval details
- Set approved timestamp

**Frontend Needs**:
- Notify investor of approval
- Show redemption amount
- Display approval notes

---

### 7. RedemptionCompleted Event
**When Emitted**: When the redemption is executed and funds transferred

**Event Structure**:
```solidity
event RedemptionCompleted(
    bytes32 indexed redemptionId,
    address indexed investor,
    uint256 tokensBurned,
    uint256 idrxPaid,
    // Previous state for audit trail
    uint256 previousTokenBalance,         // Investor's tokens before redemption
    uint256 newTokenBalance,              // Investor's tokens after redemption
    uint256 previousOutstandingSupply,    // Total supply before redemption
    uint256 newOutstandingSupply,         // Total supply after redemption
    uint256 timestamp
);
```

**Backend Processing**:
- Update redemption status to "completed"
- Update investment status to "redeemed"
- Decrease `outstanding_supply` in sukuk_series
- Record final amounts

**Frontend Needs**:
- Show redemption confirmation
- Update portfolio balance
- Display transaction details

---

### 8. RedemptionRejected Event
**When Emitted**: When the issuer rejects a redemption request

**Event Structure**:
```solidity
event RedemptionRejected(
    bytes32 indexed redemptionId,
    address indexed rejector,
    string rejectionReason,
    uint256 timestamp
);
```

**Backend Processing**:
- Update redemption status to "rejected"
- Store rejection reason
- Note: Tokens are unlocked in smart contract

**Frontend Needs**:
- Notify investor of rejection
- Display rejection reason
- Show tokens are available again

---

### 9. EmergencySuspended Event
**When Emitted**: When a Sukuk is suspended for compliance or regulatory reasons

**Event Structure**:
```solidity
event EmergencySuspended(
    address indexed sukukToken,
    address indexed suspender,
    string reason,
    uint256 timestamp
);
```

**Backend Processing**:
- Update sukuk_series status to "suspended"
- Log suspension details

**Frontend Needs**:
- Alert all investors
- Display suspension reason
- Show restricted functionality

---

### Smart Contract Architecture for Yield Distribution

Given that you're using ERC20Snapshot and have no claim windows, here's the recommended approach:

**1. Snapshot-Based Distribution with Cumulative Tracking**

```solidity
contract SukukToken is ERC20Snapshot {
    struct Distribution {
        uint256 snapshotId;
        uint256 totalAmount;        // Total IDRX for this distribution
        uint256 tokensAtSnapshot;   // Total tokens at snapshot
        uint256 yieldPerToken;      // Yield per token for this distribution
        uint256 timestamp;
    }
    
    mapping(uint256 => Distribution) public distributions;
    uint256 public distributionCounter;
    
    // Track cumulative yield per token
    uint256 public cumulativeYieldPerToken;
    
    // Track what each investor has already claimed
    mapping(address => uint256) public claimedYieldPerToken;
    
    function distributeYield(uint256 amount) external onlyIssuer {
        require(amount > 0, "Amount must be > 0");
        
        // Take snapshot of current holders
        uint256 snapId = _snapshot();
        uint256 totalTokens = totalSupply();
        require(totalTokens > 0, "No tokens in circulation");
        
        // Calculate yield per token (with precision factor)
        uint256 yieldPerToken = (amount * 1e18) / totalTokens;
        
        // Update cumulative yield
        cumulativeYieldPerToken += yieldPerToken;
        
        // Store distribution info
        distributionCounter++;
        distributions[distributionCounter] = Distribution({
            snapshotId: snapId,
            totalAmount: amount,
            tokensAtSnapshot: totalTokens,
            yieldPerToken: yieldPerToken,
            timestamp: block.timestamp
        });
        
        emit YieldDistributed(
            distributionCounter,
            address(this),
            amount,
            periodStart,
            periodEnd,
            yieldPerToken,
            block.timestamp
        );
    }
    
    function claimYield() external {
        address investor = msg.sender;
        
        // Calculate total unclaimed yield
        uint256 unclaimedPerToken = cumulativeYieldPerToken - claimedYieldPerToken[investor];
        uint256 tokens = balanceOf(investor);
        uint256 yieldAmount = (tokens * unclaimedPerToken) / 1e18;
        
        require(yieldAmount > 0, "No yield to claim");
        
        // Update claimed amount
        claimedYieldPerToken[investor] = cumulativeYieldPerToken;
        
        // Transfer IDRX
        IERC20(idrxToken).transfer(investor, yieldAmount);
        
        emit YieldClaimed(
            investor,
            address(this),
            yieldAmount,
            0, // Could track period info
            0,
            distributionCounter
        );
    }
    
    function getClaimableYield(address investor) external view returns (uint256) {
        uint256 unclaimedPerToken = cumulativeYieldPerToken - claimedYieldPerToken[investor];
        uint256 tokens = balanceOf(investor);
        return (tokens * unclaimedPerToken) / 1e18;
    }
}
```

**2. Key Benefits of This Approach:**
- **No unclaimed IDRX confusion**: Each distribution adds to cumulative yield
- **Investors can claim anytime**: They claim their total accumulated yield
- **Works with token transfers**: New holders can't claim past yields
- **Simple accounting**: One claim gets all accumulated yields

**3. Alternative: Multi-Distribution Tracking**

If you need to track individual distributions:

```solidity
mapping(address => mapping(uint256 => bool)) public claimed;  // investor => distributionId => claimed

function claimYield(uint256[] calldata distributionIds) external {
    uint256 totalYield = 0;
    
    for (uint i = 0; i < distributionIds.length; i++) {
        uint256 distId = distributionIds[i];
        require(!claimed[msg.sender][distId], "Already claimed");
        
        Distribution memory dist = distributions[distId];
        uint256 tokensHeld = balanceOfAt(msg.sender, dist.snapshotId);
        
        if (tokensHeld > 0) {
            uint256 yield = (tokensHeld * dist.yieldPerToken) / 1e18;
            totalYield += yield;
            claimed[msg.sender][distId] = true;
        }
    }
    
    require(totalYield > 0, "No yield to claim");
    IERC20(idrxToken).transfer(msg.sender, totalYield);
}
```

---

## Additional Events to Consider

### ~~11. Transfer Event (REMOVED)~~
**Why Removed**: Sukuk tokens are not transferable - only invest and redeem operations are allowed.

### 11. PriceUpdated Event
**When Emitted**: When the token price is updated (if not fixed)

**Event Structure**:
```solidity
event PriceUpdated(
    address indexed sukukToken,
    uint256 oldPrice,
    uint256 newPrice,
    uint256 timestamp
);
```

**Considerations**:
- For dynamic pricing models
- Portfolio valuation updates

---

## Data Flow Examples

### Investment Flow
1. Investor calls `invest()` → `Investment` event emitted
2. Indexer captures event and stores in database
3. Backend queries new investment events
4. Creates/updates investment records
5. Frontend shows updated portfolio

### Yield Claim Flow
1. Company calls `distributeYield()` → `YieldDistributed` event
2. Backend creates yield_claim records for all investors
3. Investor calls `claimYield()` → `YieldClaimed` event
4. Backend updates yield_claim status
5. Frontend shows claimed yield in history

### Redemption Flow
1. Investor calls `requestRedemption()` → `RedemptionRequested` event
2. Company reviews and calls `approveRedemption()` → `RedemptionApproved` event
3. System executes redemption → `RedemptionCompleted` event
4. Backend updates all related records
5. Frontend shows completed redemption

---

## State-Based Handling (No Events Required)

### Maturity Status
Since maturity is time-based and deterministic, no event is needed:

```solidity
// Smart contract just stores the date
uint256 public maturityDate;

function isMatured() public view returns (bool) {
    return block.timestamp >= maturityDate;
}
```

Backend periodically checks and updates status:
```go
func CheckSukukMaturity(db *gorm.DB) {
    var activeSukuk []SukukSeries
    db.Where("status = ? AND maturity_date <= ?", "active", time.Now()).Find(&activeSukuk)
    
    for _, sukuk := range activeSukuk {
        sukuk.Status = "matured"
        db.Save(&sukuk)
    }
}
```

---

## Implementation Notes

### Indexer Database Schema
```sql
CREATE TABLE blockchain.events (
    id BIGSERIAL PRIMARY KEY,
    event_name VARCHAR(50) NOT NULL,
    tx_hash VARCHAR(66) NOT NULL,
    log_index INTEGER NOT NULL,
    block_number BIGINT NOT NULL,
    block_timestamp TIMESTAMP NOT NULL,
    contract_address VARCHAR(42) NOT NULL,
    -- Event-specific data stored as JSONB
    event_data JSONB NOT NULL,
    -- Prevent duplicates
    UNIQUE(tx_hash, log_index)
);

CREATE INDEX idx_events_block_number ON blockchain.events(block_number);
CREATE INDEX idx_events_contract_address ON blockchain.events(contract_address);
CREATE INDEX idx_events_name ON blockchain.events(event_name);
```

### Backend Sync Service
```go
type EventSyncService struct {
    lastProcessedBlock uint64
}

func (s *EventSyncService) ProcessNewEvents() {
    // Query events since last processed block
    // Process each event type
    // Update business database
    // Store last processed block
}
```

---

## Current Implementation Gaps

1. **Missing Investor Profile Data**
   - Only have wallet addresses
   - No KYC/compliance information storage

2. **No Price History Tracking**
   - Cannot show portfolio value over time
   - No charts or analytics

3. **Limited Yield Management**
   - Manual yield distribution trigger
   - No automatic period calculations

4. **No Secondary Market Support**
   - Transfer events not processed
   - No holder tracking updates

5. **Missing Audit Trail**
   - Need to store all raw events
   - Important for compliance

---

## Recommendations

1. **Event Data Completeness**
   - ✅ Include timestamps in all events (all events have `timestamp` field)
   - ✅ Add event IDs for correlation (distributionId, redemptionId, etc.)
   - Include previous state values for audit trail (see explanation below)

2. **Indexer Reliability**
   - Implement reorg handling
   - Store confirmation counts
   - Handle missing blocks

3. **Performance Optimization**
   - Batch event processing
   - Use materialized views for complex queries
   - Implement caching for frequently accessed data

4. **Security Considerations**
   - Validate event sources
   - Check for event ordering
   - Implement rate limiting on queries

---

## Previous State Values for Audit Trail

**What it means**: Include the "before" and "after" values in events for complete audit trails.

**Examples**:

### Investment Event with Previous State
```solidity
event Investment(
    address indexed investor,
    address indexed sukukToken,
    uint256 idrxAmount,
    uint256 tokenAmount,
    uint256 tokenPrice,
    // Previous state values for audit
    uint256 previousOutstandingSupply,    // Supply before this investment
    uint256 newOutstandingSupply,         // Supply after this investment
    uint256 timestamp
);
```

### Redemption with Previous State
```solidity
event RedemptionCompleted(
    bytes32 indexed redemptionId,
    address indexed investor,
    uint256 tokensBurned,
    uint256 idrxPaid,
    // Previous state for audit
    uint256 previousTokenBalance,         // Investor's tokens before redemption
    uint256 newTokenBalance,              // Investor's tokens after redemption
    uint256 previousOutstandingSupply,    // Total supply before redemption
    uint256 newOutstandingSupply,         // Total supply after redemption
    uint256 timestamp
);
```

**Benefits**:
- Complete audit trail without additional queries
- Can verify state changes are correct
- Helps detect anomalies or attacks
- Regulatory compliance for financial products

**Implementation Consideration**:
- Adds gas costs (more data in events)
- Makes events larger but more complete
- Trade-off between completeness and efficiency

**Recommendation**: ✅ Applied to critical events (Investment and RedemptionCompleted), skipped for others like YieldClaimed and YieldDistributed.