# Workflow — Private Documents

Upload:
`authorized admin -> metadata validation -> file allow-list/size validation -> safe private storage -> DB metadata -> activity UPLOAD`.

Download:
`authenticated admin -> permission + division scope -> lookup metadata -> stream file -> activity DOWNLOAD`.

Delete:
`permission -> metadata/file operation using selected retention policy -> activity DELETE`.

No public direct file URLs. Do not expose storage paths in UI/API responses beyond safe identifiers/metadata.
