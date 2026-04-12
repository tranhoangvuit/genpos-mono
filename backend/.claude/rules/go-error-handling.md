# Go Error Handling Rules

All errors returned from services, gateways, and datastores MUST go through the `pkg/errors` wrapper around `samber/oops`.

## Why

`pkg/errors` (backed by `samber/oops`) automatically attaches a stack trace and caller information to every error. This means the failing function is identifiable from the error itself — **the function name is enough context**, so do NOT duplicate it in the message.

## Rules

1. Use `pkg/errors` constructors for domain errors:
   ```go
   import "github.com/genpick/genbus-backend/pkg/errors"

   return nil, errors.NotFound("reason not found")
   return nil, errors.BadRequest("invalid input")
   return nil, errors.Internal("database error")
   ```

2. Wrap underlying errors with `errors.Wrap` to preserve the cause while adding a short domain-level message:
   ```go
   if err != nil {
       return nil, errors.Wrap(err, "list reasons")
   }
   ```

3. Do NOT return bare `fmt.Errorf` / `errors.New` from the service, gateway, or datastore layers. These lose the stack trace and caller metadata.

4. Do NOT prefix messages with the function name:
   - ❌ `errors.Internal("ReasonService.ListReasons: failed to load")`
   - ✅ `errors.Internal("failed to load reasons")`

5. Keep messages short and user-domain-oriented. Handlers are responsible for translating structured errors to HTTP responses — do NOT format HTTP-specific messages inside services.

## Layer boundaries

| Layer | Returns |
|-------|---------|
| Datastore | `errors.Wrap(err, "short-message")` around pgx errors |
| Gateway / Service | `errors.NotFound`, `errors.BadRequest`, `errors.Internal`, or `errors.Wrap` |
| Handler | Translates `pkg/errors` types to HTTP status codes |
