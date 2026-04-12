# Go Testing Rules

These rules are **mandatory** for every Go test in this repository.

## 1. Test function naming

Test functions MUST use the prefix pattern `Test_<Scope>_<Subject>`, where `<Scope>` identifies the layer and `<Subject>` is the type/feature being tested.

| Layer | Pattern | Example |
|-------|---------|---------|
| HTTP integration (`test/integration/`) | `Test_Integration_<Feature>` | `Test_Integration_Reasons` |
| Service unit (`internal/service/`) | `Test_<TypeName>Service_<Method>` | `Test_ReasonService_ListReasons` |
| Datastore reader (`internal/infrastructure/datastore/`) | `Test_<TypeName>Reader_<Method>` | `Test_ReasonReader_ListReasons` |
| Datastore writer (`internal/infrastructure/datastore/`) | `Test_<TypeName>Writer_<Method>` | `Test_ReasonWriter_CreateReason` |

- Do NOT use bare `TestFoo` names.
- The underscore-separated prefix is required so tests can be filtered by scope via `go test -run`.

## 2. Table-driven tests

Every test MUST be table-driven. Define a map (or slice) of named cases and iterate with `t.Run(name, …)`. Even single-case tests should use the table form for consistency and future extensibility.

```go
func Test_ReasonService_ListReasons(t *testing.T) {
    t.Parallel()

    cases := map[string]struct {
        setup   func(*mock.MockReasonQueriesGateway)
        want    []*entity.Reason
        wantErr bool
    }{
        "returns reasons": {
            setup: func(m *mock.MockReasonQueriesGateway) { /* ... */ },
            want:  []*entity.Reason{ /* ... */ },
        },
        "returns empty list": {
            setup: func(m *mock.MockReasonQueriesGateway) { /* ... */ },
            want:  []*entity.Reason{},
        },
    }

    for name, tc := range cases {
        t.Run(name, func(t *testing.T) {
            t.Parallel()
            // ...
        })
    }
}
```

## 3. Comparisons via go-cmp

Assertions on structs, slices, maps, or entities MUST use `github.com/google/go-cmp/cmp` with `cmp.Diff`.

- Do NOT use `reflect.DeepEqual`.
- Do NOT use `testify/assert.Equal` for structured comparisons.
- Report failures with the diff, using `(-want +got)`:

```go
if diff := cmp.Diff(want, got); diff != "" {
    t.Errorf("result mismatch (-want +got):\n%s", diff)
}
```

Use `cmpopts.IgnoreFields` / `cmpopts.EquateApproxTime` for timestamps, generated IDs, and other non-deterministic fields:

```go
opts := []cmp.Option{
    cmpopts.IgnoreFields(entity.Reason{}, "ID", "CreatedAt", "UpdatedAt"),
    cmpopts.EquateApproxTime(time.Second),
}
if diff := cmp.Diff(want, got, opts...); diff != "" {
    t.Errorf("reason mismatch (-want +got):\n%s", diff)
}
```

## 4. Parallelism

All tests MUST call `t.Parallel()` — both the top-level test function and each subtest inside `t.Run`. Exceptions require a comment explaining why.

## 5. Preprare data

All data for read testing we must put in db/testfixtures
