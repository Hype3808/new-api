# Session Invalidation After Database Reset

## Problem

When you delete the database and restart the service, previous cookies/sessions may still remain active. This happens because:

1. **Session data is stored in cookies** - Sessions are encrypted and stored in browser cookies using the `SESSION_SECRET` key
2. **No database validation** - The authentication middleware only checked session data without verifying that the user still exists in the database
3. **Session secret persistence** - If `SESSION_SECRET` remains the same across restarts, old cookies can still be decrypted

## Solutions

### Solution 1: Change SESSION_SECRET (Immediate Fix)

The quickest way to invalidate all existing sessions is to change the `SESSION_SECRET` environment variable:

```bash
# Set a new random session secret
export SESSION_SECRET="your-new-random-secret-here"

# Or on Windows PowerShell
$env:SESSION_SECRET = "your-new-random-secret-here"
```

**Important:** 
- Use a long, random string for `SESSION_SECRET`
- Never use the default value `"random_string"`
- Save this value in your environment configuration to persist across restarts

### Solution 2: Database Validation (Implemented)

The authentication middleware has been updated to validate that users in active sessions still exist in the database. Now when a request is made:

1. Session data is read from the cookie
2. User ID is extracted from the session
3. **NEW:** System queries the database to verify the user still exists
4. If user doesn't exist, the session is cleared and authentication fails
5. If user exists, session data is refreshed with current database values

This ensures that:
- Deleted users can't access the system even with valid cookies
- User role/status changes are reflected immediately
- Database is the single source of truth for authentication

### Solution 3: Use Redis for Session Storage (Advanced)

For production environments, consider using Redis for session storage instead of cookies:

1. Install Redis
2. Update session configuration in `main.go` to use Redis store
3. Sessions will be stored server-side with only a session ID in the cookie
4. Clearing Redis will immediately invalidate all sessions

## Verification

After implementing the fix, you can verify it works by:

1. Login to the application (session is created)
2. Delete the database
3. Try to access a protected endpoint
4. You should now see: `{"success": false, "message": "会话已失效，请重新登录"}`

## Best Practices

1. **Always set SESSION_SECRET** in production environments
2. **Rotate SESSION_SECRET periodically** to invalidate old sessions
3. **Monitor session activity** for suspicious access patterns
4. **Consider session timeout** - Current default is 30 days (`MaxAge: 2592000`)
5. **Use HTTPS in production** - Set `Secure: true` in session options

## Configuration

Session configuration is in `main.go`:

```go
store := cookie.NewStore([]byte(common.SessionSecret))
store.Options(sessions.Options{
    Path:     "/",
    MaxAge:   2592000,  // 30 days - adjust as needed
    HttpOnly: true,      // Prevents JavaScript access
    Secure:   false,     // Set to true in production with HTTPS
    SameSite: http.SameSiteStrictMode,
})
```

Adjust these values based on your security requirements.
