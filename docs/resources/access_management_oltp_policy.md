# altr_access_management_oltp_policy (Resource)

Manages an OLTP access management policy.

## Schema

### Required

- `name` (String) Name of the OLTP access management policy. Must be between 1 and 255 characters.
- `repo_name` (String) The name of the repository this policy belongs to. Must be between 1 and 255 characters.
- `case_sensitivity` (String) Case sensitivity for the policy. Must be set to `case_sensitive`.
- `database_type` (Number) Database type ID for the policy. Must be one of `9`, `4`, or `24`.
- `database_type_name` (String) Database type name for the policy. Must be set to `oracle`.
- `rules` (Attributes List) List of rules for the OLTP access management policy. Must contain at least one rule. See [Nested Schema for `rules`](#nested-schema-for-rules).

### Optional

- `description` (String) Description of the OLTP access management policy. Must be between 1 and 255 characters.

### Read-Only

- `id` (String) Unique identifier for the OLTP access management policy.
- `created_at` (String) Creation timestamp.
- `updated_at` (String) Last update timestamp.

---

### Nested Schema for `rules`

#### Required

- `actors` (Attributes List) List of actors for the rule. Must contain at least one actor. See [Nested Schema for `rules.actors`](#nested-schema-for-rulesactors).
- `objects` (Attributes List) List of objects for the rule. Must contain at least one object. See [Nested Schema for `rules.objects`](#nested-schema-for-rulesobjects).

#### Optional

- `type` (String) Type of the rule. Must be one of `read`, `update`, `delete`, or `create`.

---

### Nested Schema for `rules.actors`

#### Required

- `type` (String) Type of the actor. Must be one of `idp_user` or `idp_group`.
- `condition` (String) Condition for the actor. Must be set to `equals`.
- `identifiers` (List of String) List of identifiers for the actor.

---

### Nested Schema for `rules.objects`

#### Required

- `type` (String) Type of the object. Must be one of `database`, `schema`, `table`, or `column`.
- `identifiers` (Attributes List) List of identifiers for the object. Must contain at least one identifier. See [Nested Schema for `rules.objects.identifiers`](#nested-schema-for-rulesobjectsidentifiers).

---

### Nested Schema for `rules.objects.identifiers`

#### Optional

- `database` (Attributes Object) Database identifier part. See [Nested Schema for `rules.objects.identifiers.database`](#nested-schema-for-rulesobjectsidentifiersdatabase).
- `schema` (Attributes Object) Schema identifier part. See [Nested Schema for `rules.objects.identifiers.schema`](#nested-schema-for-rulesobjectsidentifiersschema).
- `table` (Attributes Object) Table identifier part. See [Nested Schema for `rules.objects.identifiers.table`](#nested-schema-for-rulesobjectsidentifierstable).
- `column` (Attributes Object) Column identifier part. See [Nested Schema for `rules.objects.identifiers.column`](#nested-schema-for-rulesobjectsidentifierscolumn).

---

### Nested Schema for `rules.objects.identifiers.database`

#### Optional

- `name` (String) Name of the database. Must be between 1 and 255 characters.
- `wildcard` (Boolean) Wildcard for the database.

---

### Nested Schema for `rules.objects.identifiers.schema`

#### Optional

- `name` (String) Name of the schema. Must be between 1 and 255 characters.
- `wildcard` (Boolean) Wildcard for the schema.

---

### Nested Schema for `rules.objects.identifiers.table`

#### Optional

- `name` (String) Name of the table. Must be between 1 and 255 characters.
- `wildcard` (Boolean) Wildcard for the table.

---

### Nested Schema for `rules.objects.identifiers.column`

#### Optional

- `name` (String) Name of the column. Must be between 1 and 255 characters.
- `wildcard` (Boolean) Wildcard for the column.
