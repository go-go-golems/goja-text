Deployment Notes

# Release 2.4

The release changes the indexing pipeline and adds source-preserving document chunks. Operators should verify that every emitted range points into the original UTF-8 source.

## Migration

Run the schema migration before starting the new workers:

```sh
go run ./cmd/migrate --database production
go run ./cmd/worker --log-level debug
```

The fenced block is atomic in the default Markdown-block strategy. A small budget therefore reports it as oversized instead of cutting through the shell program.

## Verification

- Compare indexed document counts before and after deployment.
- Check that retrieved citations contain the expected Markdown syntax.
- Roll back if the worker reports source-range mismatches.

# Operational Limits

Chunk budgets are application policy. Rune counts are deterministic across models, while token counts must come from the tokenizer used by the embedding or generation model.
