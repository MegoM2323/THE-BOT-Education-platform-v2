# Credit Deduction Flow

## Overview

Credit deduction is performed atomically when a student books a lesson. This document describes the complete flow with transaction isolation guarantees.

---

## High-Level Flow

```
Student Requests Booking
        |
        v
Create Credit Deduction Transaction (PENDING)
        |
        v
Update Student Balance (SERIALIZABLE isolation)
        |
        v
Increment Lesson current_students Counter
        |
        v
Mark Booking as ACTIVE
        |
        v
Commit Transaction
        |
        v
Credit Deducted & Booking Confirmed
```

---

## Detailed Flow with Error Handling

```
┌─────────────────────────────────────────────────────────────┐
│ 1. RECEIVE BOOKING REQUEST                                  │
│    - Input: student_id, lesson_id                           │
│    - Validation: student exists, lesson exists, not deleted │
│    - Check: student not already booked                      │
│    - Check: lesson not at capacity                          │
└─────────────────────────────────────────────────────────────┘
                        |
                        v
┌─────────────────────────────────────────────────────────────┐
│ 2. BEGIN SERIALIZABLE TRANSACTION                           │
│    - IsoLevel: pgx.Serializable                             │
│    - Ensures: No phantoms, serialization anomalies          │
│    - Lock: Student and Lesson records                       │
└─────────────────────────────────────────────────────────────┘
                        |
                        v
┌─────────────────────────────────────────────────────────────┐
│ 3. CREATE CREDIT TRANSACTION RECORD                         │
│    - Table: transactions                                     │
│    - Fields:                                                │
│      * id (UUID)                                            │
│      * user_id (student_id)                                 │
│      * type: 'deduction'                                    │
│      * amount (lesson.credits_cost)                         │
│      * description: "Booking: {lesson_name}"                │
│      * status: 'pending'                                    │
│      * created_at: now()                                    │
│    - Purpose: Audit trail and recovery                      │
└─────────────────────────────────────────────────────────────┘
                        |
                        v
┌─────────────────────────────────────────────────────────────┐
│ 4. CHECK STUDENT BALANCE                                    │
│    - Query: SELECT balance FROM credits WHERE user_id = ?   │
│    - Condition: balance >= lesson.credits_cost              │
│    - Action on FAIL: Rollback, return InsufficientCredits   │
│    - FOR UPDATE: Lock row to prevent concurrent changes     │
└─────────────────────────────────────────────────────────────┘
                        |
                        v
┌─────────────────────────────────────────────────────────────┐
│ 5. DEDUCT CREDITS                                           │
│    - Query: UPDATE credits                                  │
│    - Operation: balance = balance - amount                  │
│    - Updated: updated_at = now()                            │
│    - Atomicity: Single SQL operation (no race conditions)   │
│    - Action on FAIL: Rollback, log error                    │
└─────────────────────────────────────────────────────────────┘
                        |
                        v
┌─────────────────────────────────────────────────────────────┐
│ 6. MARK TRANSACTION AS COMPLETED                            │
│    - Query: UPDATE transactions                             │
│    - Change: status = 'completed'                           │
│    - Lock: No additional lock (already locked)              │
└─────────────────────────────────────────────────────────────┘
                        |
                        v
┌─────────────────────────────────────────────────────────────┐
│ 7. INCREMENT LESSON COUNTER                                 │
│    - Query: UPDATE lessons                                  │
│    - Operation: current_students = current_students + 1     │
│    - Condition: current_students < max_students             │
│    - Action on FAIL: Rollback, return LessonFull            │
│    - FOR UPDATE: Lesson row locked to prevent overbooking   │
└─────────────────────────────────────────────────────────────┘
                        |
                        v
┌─────────────────────────────────────────────────────────────┐
│ 8. CREATE BOOKING RECORD                                    │
│    - Table: bookings                                        │
│    - Fields:                                                │
│      * id (UUID)                                            │
│      * student_id                                           │
│      * lesson_id                                            │
│      * status: 'active'                                     │
│      * booked_at: now()                                     │
│      * created_at: now()                                    │
│    - Action on FAIL: Rollback, log error                    │
└─────────────────────────────────────────────────────────────┘
                        |
                        v
┌─────────────────────────────────────────────────────────────┐
│ 9. COMMIT TRANSACTION                                       │
│    - All changes become visible                             │
│    - Locks released                                         │
│    - Action on FAIL: Automatic rollback                     │
└─────────────────────────────────────────────────────────────┘
                        |
                        v
┌─────────────────────────────────────────────────────────────┐
│ 10. RETURN SUCCESS RESPONSE                                 │
│    - BookingID, remaining balance, lesson info             │
│    - Student sees updated balance in UI                     │
└─────────────────────────────────────────────────────────────┘
```

---

## Transaction Isolation Level: SERIALIZABLE

### Why SERIALIZABLE?

```
Level           Dirty Reads   Non-repeatable   Phantoms
────────────────────────────────────────────────────────
READ UNCOMMITTED    YES           YES            YES
READ COMMITTED      NO            YES            YES
REPEATABLE READ     NO            NO             YES
SERIALIZABLE        NO            NO             NO        ← Used for credits
```

### What SERIALIZABLE Prevents

1. **Dirty Reads**: Can't see uncommitted changes
2. **Non-repeatable Reads**: Same query returns same result within transaction
3. **Phantom Reads**: No new rows appear between queries in same transaction

### In Credit Context

**Problem without SERIALIZABLE:**
```
T1 (Student A):
1. Check balance: 100 credits
2. Book lesson (-50 credits)
3. Update balance: 50 credits

T2 (Student A) - concurrent:
1. Check balance: 100 credits (didn't see T1 yet)
2. Book lesson (-50 credits)
3. Update balance: 50 credits (WRONG! Should be 0)

Result: Balance mismatch - credits disappeared
```

**Solution with SERIALIZABLE:**
```
T1 (Student A):
BEGIN SERIALIZABLE
1. SELECT ... FOR UPDATE (locks row)
2. Check balance: 100 credits
3. Update balance: 50 credits
COMMIT

T2 (Student A):
BEGIN SERIALIZABLE
1. SELECT ... FOR UPDATE (waits for T1 to release lock)
2. Check balance: 50 credits (correct)
3. Update balance: 0 credits (correct)
COMMIT

Result: No race condition, consistent state
```

---

## Edge Cases & Handling

### Case 1: Insufficient Credits
```
Flow:
1. Booking requested
2. Transaction created (status='pending')
3. Balance check FAILS
4. Complete ROLLBACK
5. Return error to user

Result:
- No credit deduction
- No booking created
- Transaction record marked as failed (optional)
```

### Case 2: Lesson Becomes Full
```
Flow:
1. Credit deducted successfully
2. Increment lesson counter: counter < max (holds true)
3. Lesson counter reaches max_students
4. Booking created

Result:
- Last student to book gets in (by SQL ORDER)
- Future bookings fail immediately (lesson full)
```

### Case 3: Network Failure During Commit
```
Flow:
1. All operations in transaction successful
2. COMMIT command sent to database
3. Network failure occurs
4. Commit status unknown

Recovery:
- Client receives error
- Client should retry GET /booking to check if booking exists
- If booking exists: credit was deducted, resume normally
- If booking missing: credit NOT deducted, safe to retry entire operation
```

### Case 4: Duplicate Booking Request
```
Flow:
1. First request: Creates booking A, deducts credits
2. Second request (retry): Checks if already booked
3. Already booked check returns true
4. Operation fails before deduction

Result:
- Only one booking per lesson per student
- Credits deducted only once
```

---

## Consistency Guarantees

### Before Commit
- No partial state visible
- All or nothing
- Can be rolled back entirely

### After Commit
- Student balance reduced
- Lesson counter incremented
- Booking created
- All changes are durable and visible

### In Case of Crash
- PostgreSQL WAL (Write-Ahead Logging) ensures durability
- In-flight transactions rolled back automatically
- Completed transactions are permanent

---

## Performance Considerations

### Lock Duration
- Locks held: Duration of transaction (typically < 100ms)
- Row locks on: `students.credits` row, `lessons` row
- Range locks: None (point queries only)

### Concurrency Impact
- Multiple students can book different lessons simultaneously
- Same student: Serialized (SERIALIZABLE level)
- Different students: Concurrent (different row locks)

### Typical Timing
```
Operation                   Duration
──────────────────────────  ───────
Begin transaction           1ms
Check balance               2ms
Deduct credits              1ms
Increment counter           1ms
Create booking              2ms
Commit                      3ms
──────────────────────────  ────────
Total                       ~10ms
```

---

## Monitoring & Alerts

### Metrics to Track
1. Booking failure rate (insufficient credits)
2. Transaction commit time
3. Lock wait time
4. Balance-vs-bookings consistency

### SQL Queries for Monitoring

```sql
-- Check for stuck transactions
SELECT * FROM pg_stat_activity WHERE state = 'active';

-- Find transactions holding locks
SELECT * FROM pg_locks WHERE pid = (SELECT pid FROM pg_stat_activity WHERE ...);

-- Check credit consistency
SELECT COUNT(*) FROM (
  SELECT u.id, COUNT(*) as booking_count,
         (SELECT balance FROM credits WHERE user_id = u.id) as balance
  FROM users u
  LEFT JOIN bookings b ON u.id = b.student_id
  WHERE b.status = 'active'
  GROUP BY u.id
  HAVING COUNT(*) > balance
);
```

---

## Recovery Procedures

### If Balance Mismatch Detected
1. Identify affected users
2. Use VerificationService.RepairUserBalance()
3. Recount bookings
4. Update balance to match
5. Audit and alert operations team

---

## Code References

- **File:** `/backend/internal/service/booking_service.go`
- **Method:** `BookingService.Book()`
- **Isolation:** `withSerializableTx()`
- **Transaction:** Database transaction with IsoLevel = Serializable

---
