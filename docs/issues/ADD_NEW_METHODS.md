# TinyTime - New Methods Proposal

## Current State Analysis

### Existing API (3 methods)

**Interface:** `TimeProvider`

```go
type TimeProvider interface {
    // Returns current Unix timestamp in nanoseconds
    UnixNano() int64
    
    // Converts Unix seconds to formatted date string: "2021-01-01 00:00"
    UnixSecondsToDate(int64) string
    
    // Converts Unix nanoseconds to time string: "15:32:14"
    // Supports: int64, int, float64, string
    UnixNanoToTime(any) string
}
```

**Current Implementations:**
- `backStlib.go` - Uses Go standard library `time` package (backend/native)
- `frontWasm.go` - Uses JavaScript Date API via syscall/js (WASM frontend)
- Both implementations share test suite in `data_test.go`

**Dependencies:**
- `webtyp.com/fmt` - For type conversions and formatting

---

## Problem Statement

### Use Case: Medical Platform with TinyBin Serialization

**Context:**
- Backend: Go with standard library
- Frontend: WASM compiled with TinyGo
- Serialization: TinyBin (binary protocol)
- Time storage: `int64` UnixNano timestamps (8 bytes, native TinyBin support)

**Current Gaps:**

1. **No date-only formatting** - Only "2021-01-01 00:00" (date+time combined)
2. **No parsing capabilities** - Cannot convert user input strings to UnixNano
3. **No minute-based time** - Work schedules need "HH:MM" without seconds
4. **No date/time parsing** - Cannot parse "2024-01-15" or "08:30" strings
5. **No date utilities** - Cannot check if timestamp is today/past/future
6. **No time arithmetic** - Cannot calculate days between dates

**Real Database Schema Example:**
```sql
-- reservation table
reservation_at   BIGINT  -- UnixNano (1624397134562544800)
reserved_until   BIGINT  -- UnixNano or 0 (nullable)
updated_at       BIGINT  -- UnixNano (audit)

-- workcalendar table  
work_start       SMALLINT -- Minutes since midnight (480 = 08:00)
work_finish      SMALLINT -- Minutes since midnight (1020 = 17:00)
```

---

## Proposed New Methods

### Priority 1: Display Formatting (Frontend WASM)

#### 1.1 `FormatDate(value any) string`
**Purpose:** Format date from multiple input types

```go
// From UnixNano (int64)
date := tp.FormatDate(1705315200000000000)  // "2024-01-15"

// From string (passthrough validation)
date := tp.FormatDate("2024-01-15")  // "2024-01-15"

// From zero value
date := tp.FormatDate(0)  // "1970-01-01"
```

**Implementation Notes:**
- **Type switch on `value`:**
  - `int64`: Convert UnixNano → date string
  - `string`: Validate format, return as-is or error
  - Other: Return empty string
- Backend: Use `time.Unix(nano/1e9, 0).Format("2006-01-02")`
- Frontend: Use JS `new Date(nano/1e6).toISOString().slice(0,10)`
- Format: "YYYY-MM-DD" (ISO 8601)

**Use Cases:**
- Display birthday field
- Display reservation date without time
- Group records by date

---

#### 1.2 `FormatTime(value any) string`
**Purpose:** Format time from multiple input types

```go
// From minutes (int16)
timeStr := tp.FormatTime(510)   // "08:30"
timeStr := tp.FormatTime(1020)  // "17:00"

// From UnixNano (int64) - extracts time portion
timeStr := tp.FormatTime(1705306200000000000)  // "08:30:00"

// From string (passthrough validation)
timeStr := tp.FormatTime("08:30")  // "08:30"
```

**Implementation Notes:**
- ✅ **Shared logic:** Type switch + formatting
- **Type switch on `value`:**
  - `int16`: Minutes since midnight → "HH:MM"
  - `int64`: UnixNano → "HH:MM:SS"
  - `string`: Validate format, return as-is
  - Other: Return empty string
- Formula (int16): `hours := minutes/60; mins := minutes%60; format("%02d:%02d")`
- Range validation: int16 must be 0-1439
- Use tinystring.Fmt for formatting

**Use Cases:**
- Display work schedule hours
- Display appointment time slots
- Display break times in calendar

---

### Priority 2: Parsing (User Input → Numeric)

#### 2.1 `ParseDate(dateStr string) (int64, error)`
**Purpose:** Parse date string to UnixNano (midnight UTC)

```go
// Returns: 1705276800000000000, nil
nano, err := tp.ParseDate("2024-01-15")

// Returns: 0, error
nano, err := tp.ParseDate("invalid")
```

**Implementation Notes:**
- Backend: Use `time.Parse("2006-01-02", dateStr)` → `.UnixNano()`
- Frontend: Use JS `new Date(dateStr).getTime() * 1e6`
- Expected format: "YYYY-MM-DD" (ISO 8601)
- Timezone: Always UTC midnight
- Error handling: Invalid format, invalid date (e.g., "2024-02-30")

**Use Cases:**
- Parse user birthday input
- Parse reservation date form field
- Parse date filters in queries

---

#### 2.2 `ParseTime(timeStr string) (int16, error)`
**Purpose:** Parse "HH:MM" or "HH:MM:SS" to minutes since midnight

```go
// Returns: 510, nil
minutes, err := tp.ParseTime("08:30")

// Returns: 510, nil (ignores seconds)
minutes, err := tp.ParseTime("08:30:45")

// Returns: 0, nil
minutes, err := tp.ParseTime("00:00")

// Returns: 1439, nil
minutes, err := tp.ParseTime("23:59")

// Returns: 0, error
minutes, err := tp.ParseTime("invalid")
```

**Implementation Notes:**
- ✅ **Shared logic:** String parsing with tinystring (no time/Date API needed)
- Parse format: "HH:MM" or "HH:MM:SS" (ignore seconds if present)
- Algorithm: Split by ':', parse hours/minutes, validate range
- Range: 00:00 to 23:59 (0-1439 minutes)
- Error cases: Invalid format, hours >23, minutes >59, negative values

**Use Cases:**
- Parse work schedule input forms
- Parse appointment time input
- Validate time slot availability

---

#### 2.3 `ParseDateTime(dateStr, timeStr string) (int64, error)`
**Purpose:** Combine date + time strings into single UnixNano

```go
// Returns: 1705306200000000000, nil
nano, err := tp.ParseDateTime("2024-01-15", "08:30")

// Returns: 1705276800000000000, nil (midnight)
nano, err := tp.ParseDateTime("2024-01-15", "00:00")
```

**Implementation Notes:**
- Backend: Parse date, parse time, combine: `time.Date(y, m, d, h, min, 0, 0, time.UTC)`
- Frontend: JS `new Date(dateStr + "T" + timeStr + ":00Z").getTime() * 1e6`
- Convenience method (could be implemented using DateToUnix + TimeToMinutes)
- Always UTC timezone

**Use Cases:**
- Create reservation with date + time from separate form fields
- Combine date picker + time picker inputs
- Schedule appointment creation

---

### Priority 3: Date Utilities (Business Logic)

#### 3.1 `IsToday(nano int64) bool`
**Purpose:** Check if timestamp is today (in local/UTC timezone)

```go
// Returns: true if nano represents today's date
isToday := tp.IsToday(someTimestamp)
```

**Implementation Notes:**
- Backend: Compare `time.Unix(nano/1e9, 0).Format("2006-01-02")` with `time.Now().Format("2006-01-02")`
- Frontend: Compare JS `new Date(nano/1e6).toDateString()` with `new Date().toDateString()`
- Timezone consideration: Should this use UTC or local time? Recommend UTC for consistency

**Use Cases:**
- Highlight today's appointments in calendar UI
- Filter "today's reservations"
- Show "Due today" indicators

---

#### 3.2 `IsPast(nano int64) bool`
**Purpose:** Check if timestamp is in the past

```go
// Returns: true if nano < current time
expired := tp.IsPast(reservation.ReservedUntil)
```

**Implementation Notes:**
- ✅ **Shared logic:** Simple comparison with UnixNano()
- Algorithm: `nano < tp.UnixNano()`
- No time/Date API needed beyond existing UnixNano()

**Use Cases:**
- Check if reservation has expired (ReservedUntil < now)
- Check if invoice is overdue (DueDate < now)
- Show "Past due" warnings

---

#### 3.3 `IsFuture(nano int64) bool`
**Purpose:** Check if timestamp is in the future

```go
// Returns: true if nano > current time
upcoming := tp.IsFuture(reservation.ReservationAt)
```

**Implementation Notes:**
- ✅ **Shared logic:** Simple comparison with UnixNano()
- Algorithm: `nano > tp.UnixNano()`
- Complement of IsPast

**Use Cases:**
- Filter "upcoming appointments"
- Validate future-dated reservations
- Show "Scheduled for" status

---

#### 3.4 `DaysBetween(nano1, nano2 int64) int`
**Purpose:** Calculate difference in days between two timestamps

```go
// Returns: 7 (if nano2 is 7 days after nano1)
days := tp.DaysBetween(startDate, endDate)

// Returns: -7 (if nano2 is before nano1)
days := tp.DaysBetween(endDate, startDate)
```

**Implementation Notes:**
- ✅ **Shared logic:** Pure arithmetic, no time/Date API needed
- Algorithm: `(nano2 - nano1) / (24 * 60 * 60 * 1e9)`
- Constant: `86400000000000` nanoseconds per day
- Returns integer days (truncates partial days)
- Can return negative if nano2 < nano1

**Use Cases:**
- Calculate age from birthday
- Calculate "days until appointment"
- Calculate "days since last visit"
- Invoice aging reports

---

### Priority 4: Advanced Formatting (Optional)

#### 4.1 `UnixNanoToDateTime(nano int64) string`
**Purpose:** Format full timestamp with seconds precision

```go
// Returns: "2024-01-15 08:30:45"
dateTime := tp.UnixNanoToDateTime(1705306245000000000)
```

**Implementation Notes:**
- Backend: `time.Unix(nano/1e9, 0).Format("2006-01-02 15:04:05")`
- Frontend: Similar to existing UnixSecondsToDate but with seconds
- Essentially combines UnixNanoToDate + UnixNanoToTime

**Use Cases:**
- Audit logs (full timestamp display)
- Medical history records (precise timing)
- Debug/logging output

---

#### 4.2 `FormatMinutesWithSeconds(minutes int16, seconds int16) string`
**Purpose:** Format minutes + seconds as "HH:MM:SS"

```go
// Returns: "08:30:45"
timeStr := tp.FormatMinutesWithSeconds(510, 45)
```

**Implementation Notes:**
- ✅ **Shared logic:** Extension of MinutesToTime with seconds
- Algorithm: `hours := minutes/60; mins := minutes%60; format("%02d:%02d:%02d", hours, mins, seconds)`
- Range: minutes 0-1439, seconds 0-59
- Use tinystring.Fmt for formatting

**Use Cases:**
- Precise time slot display
- Medical procedure duration
- Detailed work log entries

---

## Implementation Strategy

### Architecture Decision: Shared vs Separate

**Shared Logic (Single implementation for both):**
- `MinutesToTime` - Pure math, no Date API
- `TimeToMinutes` - String parsing with tinystring
- `IsPast` / `IsFuture` - Uses existing UnixNano()
- `DaysBetween` - Pure arithmetic
- `FormatMinutesWithSeconds` - Pure math + formatting

**Separate Implementations (Backend vs WASM):**
- `UnixNanoToDate` - Backend uses time.Format, WASM uses JS toISOString
- `DateToUnix` - Backend uses time.Parse, WASM uses Date constructor
- `DateTimeToUnix` - Backend uses time.Date, WASM uses Date string parsing
- `IsToday` - Backend uses time.Format, WASM uses toDateString
- `UnixNanoToDateTime` - Backend uses time.Format, WASM uses JS methods

**Benefits:**
- Less code duplication
- Single test suite for shared methods
- Easier maintenance
- Only duplicate when API differences require it

### Phase 1: Core Display (Week 1)
**Goal:** Enable frontend to display all time fields properly

1. Implement `UnixNanoToDate(int64) string` - Backend & WASM separate
2. Implement `MinutesToTime(int16) string` - ✅ Shared implementation
3. Add tests (shared methods need one test suite)
4. Update README with examples

**Deliverable:** Frontend can display dates and work schedules

---

### Phase 2: Parsing (Week 2)
**Goal:** Enable user input conversion

1. Implement `DateToUnix(string) (int64, error)` - Backend & WASM separate
2. Implement `TimeToMinutes(string) (int16, error)` - ✅ Shared implementation
3. Implement `DateTimeToUnix(string, string) (int64, error)` - Backend & WASM separate
4. Add comprehensive error handling tests (fewer tests for shared methods)
5. Document error cases in README

**Deliverable:** Forms can accept date/time inputs and convert to storage format

---

### Phase 3: Utilities (Week 3)
**Goal:** Add business logic helpers

1. Implement `IsToday(int64) bool` - Backend & WASM separate
2. Implement `IsPast(int64) bool` - ✅ Shared implementation
3. Implement `IsFuture(int64) bool` - ✅ Shared implementation
4. Implement `DaysBetween(int64, int64) int` - ✅ Shared implementation
5. Add integration tests with TinyBin (fewer duplicate tests)
6. Performance benchmarks

**Deliverable:** Complete time utilities for medical platform

---

### Phase 4: Advanced (Optional - Week 4)
**Goal:** Convenience methods

1. Implement `UnixNanoToDateTime(int64) string` - Backend & WASM separate
2. Implement `FormatMinutesWithSeconds(int16, int16) string` - ✅ Shared implementation
3. Add locale support (if needed)?
4. Timezone handling improvements?

**Summary:**
- 10 new methods total
- 5 shared implementations (50% code reuse)
- 5 separate implementations (when Date API differences require it)
- Significantly less testing effort for shared methods

---

## Testing Requirements

### Unit Tests

**For Shared Methods** (single test suite):
- MinutesToTime, TimeToMinutes, IsPast, IsFuture, DaysBetween, FormatMinutesWithSeconds
- ✅ Happy path test
- ✅ Zero value test (0)
- ✅ Negative value test (if applicable)
- ✅ Edge case tests (boundary values)
- ✅ Error case tests (for parsing methods)
- Single test file runs for both backend and WASM

**For Separate Methods** (backend + WASM tests):
- UnixNanoToDate, DateToUnix, DateTimeToUnix, IsToday, UnixNanoToDateTime
- ✅ Same test cases for both implementations
- ✅ Cross-implementation consistency verification

### Browser Testing (WASM)

**Execution:** Use `wasmbrowsertest` for automated browser testing

```bash
# Install
go install github.com/agnivade/wasmbrowsertest@latest
mv $(go env GOPATH)/bin/wasmbrowsertest $(go env GOPATH)/bin/go_js_wasm_exec

# Run WASM tests
GOOS=js GOARCH=wasm go test ./wasm_tests/...
```

**See [BROWSER_TEST.md](../BROWSER_TEST.md) for:**
- Complete setup instructions
- CI/CD integration (GitHub Actions, Travis)
- Troubleshooting guide
- Best practices for WASM testing

**Current Coverage:**
- Existing tests in `wasm_tests/wasm_test.go`
- Validates JS Date API integration
- Tests all current TimeProvider methods

**New Methods Testing Strategy:**
- Add WASM tests only for methods with separate implementations
- Shared logic methods don't need duplicate WASM tests
- Focus on JS Date API edge cases and timezone handling

### Integration Tests

- ✅ Round-trip: Parse → Format → Parse (should be identical)
- ✅ TinyBin serialization: Encode struct with timestamps → Decode → Verify
- ✅ Cross-implementation consistency: Backend and WASM return same results
- ✅ Performance: Benchmark common operations

**Testing Effort Reduction:**
- Shared methods: ~50% fewer tests to write/maintain
- Consistency bugs: Eliminated for shared logic
- Browser testing: Automated with wasmbrowsertest

### WASM-Specific Tests

- ✅ Run in browser environment using `wasmbrowsertest`
- ✅ Verify JavaScript Date API calls work correctly
- ✅ Test timezone handling (UTC vs local)
- ✅ Validate against real browser Date implementation
- 📖 Complete guide: [BROWSER_TEST.md](../BROWSER_TEST.md)

---

## Breaking Changes Assessment

### Interface Changes

**Current:**
```go
type TimeProvider interface {
    UnixNano() int64
    UnixSecondsToDate(int64) string
    UnixNanoToTime(any) string
}
```

**Proposed:**
```go
type TimeProvider interface {
    // Existing (unchanged)
    UnixNano() int64
    
    // Display formatting (consolidated with any type)
    FormatDate(value any) string        // Accepts: int64 (UnixNano), string ("2024-01-15")
    FormatTime(value any) string        // Accepts: int64 (UnixNano), int16 (minutes), string ("08:30")
    FormatDateTime(value any) string    // Accepts: int64 (UnixNano), string ("2024-01-15 08:30")
    
    // Parsing (string → numeric)
    ParseDate(dateStr string) (int64, error)              // "2024-01-15" → UnixNano
    ParseTime(timeStr string) (int16, error)              // "08:30" → minutes (510)
    ParseDateTime(dateStr, timeStr string) (int64, error) // Combine to UnixNano
    
    // Utilities
    IsToday(nano int64) bool
    IsPast(nano int64) bool
    IsFuture(nano int64) bool
    DaysBetween(nano1, nano2 int64) int
}
```

**Impact:** 
- ⚠️ **Breaking:** Adds methods to interface (existing implementations will fail to compile)
- ✅ **Mitigation:** All implementations are internal (backStlib.go, frontWasm.go)
- ✅ **Users:** Only call `NewTimeProvider()`, interface is abstraction

**Migration Path:**
- No migration needed for users
- Both implementations must be updated simultaneously
- Tests must be updated

---

## Documentation Updates Needed

1. **README.md**
   - Add new methods to API Reference section
   - Add examples for each new method
   - Update "Quick Start" with parsing example
   - Add "Use Cases" section with medical platform examples

2. **Code Comments**
   - Add godoc comments for each new method
   - Include format specifications
   - Document error cases
   - Add examples in comments

3. **Test Documentation**
   - Update data_test.go header comments
   - Document shared test patterns
   - Add edge case documentation

---

## Questions for Review

### API Design
1. **Method names:** ✅ **DECISION: Use `UnixNanoToDate` / `DateToUnix` pattern**
   - **Rationale:** Consistent with existing API (`UnixSecondsToDate`, `UnixNanoToTime`)
   - **Benefit:** Explicit about input type (UnixNano) and precision (nanoseconds)
   - **Rejected alternatives:** `FormatDate` (ambiguous precision), `DateString` (unclear direction)

2. **Error handling:** ✅ **DECISION: All parsing methods return `(T, error)`**
   - **Rationale:** Standard Go idiom (see `time.Parse`, `strconv.ParseInt`)
   - **Benefit:** Error messages provide context ("invalid format" vs "date out of range")
   - **WASM:** JS errors convert cleanly to Go errors via syscall/js

3. **Zero values and historical dates:**
   
   **UnixTime supports dates before 1970 with negative values:**
   ```go
   // Historical dates have negative UnixNano
   "1941-04-04" → -907,286,400,000,000,000  // 29 years before epoch
   "1900-01-01" → -2,208,988,800,000,000,000
   "1970-01-01" → 0                          // Epoch (not an error!)
   "2024-01-15" → 1,705,276,800,000,000,000  // Modern date
   ```
   
   **Supported range (PostgreSQL BIGINT / Go int64):**
   - Years: **~1677 to ~2262** (584 year range)
   - Medical use case: ✅ Birthdays from 1900+ fully supported
   
   **Zero value ambiguity:**
   - `DateToUnix("1970-01-01")` returns `0, nil` (valid epoch date)
   - `DateToUnix("invalid")` returns `0, error` (parse error)
   - **Solution:** Use error return to disambiguate
   
   **Optional fields strategy:**
   ```go
   // Option A: Use pointer for nullable dates (recommended)
   type Patient struct {
       Name     string
       Birthday *int64  // nil = no birthday set, 0 = epoch, negative = pre-1970
   }
   
   // Option B: Convention that 0 means "not set" (not recommended)
   type Patient struct {
       Name     string
       Birthday int64  // 0 = not set (conflicts with epoch!)
   }
   ```
   
   **Recommendation:** 
   - Use `*int64` pointers for optional date fields
   - Use `int64` for required fields (any value valid: negative, zero, or positive)
   - Keep `(T, error)` return pattern for parsing methods

### Functionality
4. **Timezone handling:** ✅ **DECISION: UTC only, no timezone parameter**
   - **Rationale:** Database stores UnixNano (inherently UTC), timezone conversion at display layer
   - **Benefit:** Simpler API, avoids timezone complexity in WASM (limited TinyGo support)
   - **Alternative:** Users can convert with `time.LoadLocation()` on backend if needed

5. **Locale support:** ✅ **DECISION: 24-hour format only ("08:30", "17:00")**
   - **Rationale:** Medical platform needs unambiguous time (no AM/PM confusion)
   - **Benefit:** Single format = smaller WASM binary, fewer edge cases
   - **Rejected:** 12-hour format adds ~1-2 KB and locale dependencies

6. **Range validation:** ✅ **DECISION: Strict validation, return error on invalid**
   - **Example:** `TimeToMinutes("25:99")` → `0, error("hours must be 0-23")`
   - **Rationale:** Medical data requires correctness, silent clamping hides bugs
   - **Error messages:** Specific ("minutes must be 0-59" vs generic "invalid time")

### Implementation
7. **Priority order:** ✅ **DECISION: Follow 4-phase plan (Core → Parsing → Utilities → Advanced)**
   - **Phase 1 (Week 1):** UnixNanoToDate, MinutesToTime (enables display)
   - **Phase 2 (Week 2):** DateToUnix, TimeToMinutes, DateTimeToUnix (enables input)
   - **Phase 3 (Week 3):** IsToday, IsPast, IsFuture, DaysBetween (business logic)
   - **Phase 4 (Optional):** UnixNanoToDateTime, FormatMinutesWithSeconds (convenience)

8. **Performance requirements:** ✅ **DECISION: Target <100μs per operation (backend), <500μs (WASM)**
   - **Benchmarks required:**
     - `BenchmarkUnixNanoToDate` - Format 1M timestamps
     - `BenchmarkDateToUnix` - Parse 1M date strings
     - `BenchmarkDaysBetween` - Calculate 1M date differences
   - **Acceptance:** No regression vs existing `UnixSecondsToDate` (currently ~80μs)

9. **WASM bundle size:** ✅ **DECISION: Target <5 KB increase (currently 5-8 KB baseline)**
   - **Strategy:** Use shared logic when possible (5 of 10 methods)
   - **JS API calls:** Unavoidable for Date parsing/formatting (only 5 methods need it)
   - **Measurement:** Compare `wasm_exec.js` size before/after with `ls -lh`

### Testing
10. **Test coverage target:** ✅ **DECISION: 100% line coverage required**
   - **Tool:** `go test -cover ./...` must show 100.0%
   - **Rationale:** Small API surface, medical use case requires correctness
   - **Enforcement:** CI/CD fails if coverage drops below 100%

11. **Browser testing:** ✅ **DECISION: Chrome/Chromium only (via wasmbrowsertest)**
   - **Rationale:** 95% of WASM deployments target Chrome (Cloudflare Workers, Vercel Edge)
   - **Setup:** See [BROWSER_TEST.md](../BROWSER_TEST.md) for complete guide
   - **CI/CD:** GitHub Actions runs `GOOS=js GOARCH=wasm go test ./wasm_tests/...`

12. **TinyBin integration:** ✅ **DECISION: Keep separate (TinyBin repo has its own test suite)**
   - **Rationale:** Avoids circular dependencies, TinyBin already tests int64 serialization
   - **Integration test:** Add example in TinyTime README showing TinyBin round-trip
   - **Location:** `examples/tinybin_integration_test.go` (optional, not required for merge)
---

## Estimated Impact

### Binary Size (WASM)
- **Current TinyTime:** ~5-8 KB (minimal)
- **With new methods:** ~8-12 KB estimated
- **Increase:** ~3-4 KB (acceptable for functionality gained)

### Performance
- **Parsing overhead:** Negligible (string operations)
- **Formatting overhead:** Negligible (simple math/string concat)
- **Date utilities:** O(1) operations
- **Overall:** No significant performance impact expected

### Development Effort
- **Phase 1-3:** ~1.5-2 weeks for core functionality (reduced due to shared logic)
- **Phase 4:** ~0.5-1 week for optional features
- **Testing:** ~0.5 week comprehensive test suite (50% shared tests)
- **Documentation:** ~2-3 days
- **Total:** ~3-4 weeks for complete implementation (reduced from 4-5 weeks)

---

## Success Criteria

✅ All new methods implemented for both backend and WASM  
✅ Zero test failures in existing test suite  
✅100% test coverage for new methods  
✅ WASM bundle size increase <5 KB  
✅ No performance regression in existing methods  
✅ README updated with examples  
✅ TinyBin integration validated  
✅ Medical platform can use for all time fields  



