# Fluent Markdown Document Builder Helpers

This is the document workspace for ticket goja-text-document-builder.

## Structure

- **design/**: Design documents and architecture notes
- **reference/**: Reference documentation and API contracts
- **playbooks/**: Operational playbooks and procedures
- **scripts/**: Utility scripts and automation
- **sources/**: External sources and imported documents
- **various/**: Scratch or meeting notes, working notes
- **archive/**: Optional space for deprecated or reference-only artifacts

## Getting Started

Use docmgr commands to manage this workspace:

- Add documents: `docmgr doc add --ticket goja-text-document-builder --doc-type design-doc --title "My Design"`
- Import sources: `docmgr import file --ticket goja-text-document-builder --file /path/to/doc.md`
- Update metadata: `docmgr meta update --ticket goja-text-document-builder --field Status --value review`
