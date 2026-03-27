# Phase 1: Assessment

The first step is a thorough investigation of the source application. The AI agent examines the codebase and produces a structured migration inventory.

## The assess-migration Skill

The `assess-migration` skill provides a 6-step investigation framework that works with any technology stack -- Java, .NET, Python, Node.js, PHP, Ruby, or any other platform. It guides the agent through:

1. **Technology stack** -- build files, frameworks, runtime versions
2. **Data model** -- ORM entities, database schema, relationships
3. **Business logic** -- service classes, stored procedures, validation rules
4. **User interface** -- pages, templates, navigation structure
5. **Integrations** -- REST/SOAP clients, message queues, file feeds
6. **Security** -- authentication, authorization, role definitions

## What to Extract

| Category | What to Document | Where to Look |
|----------|-----------------|---------------|
| **Data model** | Entities, attributes, relationships, constraints | JPA entities, Django models, EF DbContext, DB schema |
| **Business logic** | Validation rules, calculations, workflows | Service classes, stored procedures, triggers |
| **Pages / UI** | Screens, forms, dashboards, navigation | React components, Razor views, JSP templates |
| **Integrations** | APIs consumed/exposed, file feeds, queues | REST clients, SOAP services, Kafka topics |
| **Security** | Authentication, roles, data access rules | Spring Security, RBAC policies, row-level security |
| **Scheduled jobs** | Background tasks, timers, batch processing | Cron jobs, Quartz schedulers, Celery tasks |

## Platform-Specific Skills

For common source platforms, dedicated skills provide deeper guidance with precise element mappings:

| Source Platform | Skill | Key Mappings |
|----------------|-------|-------------|
| **Oracle Forms** | `migrate-oracle-forms` | Form -> Page, Block -> Snippet, PL/SQL -> Microflow, LOV -> Enumeration |
| **K2 / Nintex** | `migrate-k2-nintex` | SmartForm -> Page, SmartObject -> Entity, Workflow -> Microflow chain |
| **Any stack** | `assess-migration` | Generic framework for any technology |

## Assessment Output

The assessment produces a structured report with:

- **Executive summary** -- technology stack, application size, complexity rating
- **Categorized inventory** -- counts and details for each category
- **Mendix mapping** -- how each source element maps to Mendix concepts
- **Migration risks** -- complex stored procedures, custom UI components, real-time integrations
- **Recommended phases** -- suggested order of migration work

## Example: Starting an Assessment

Point the AI agent at the source codebase and ask it to assess for migration:

```
Assess the application in /path/to/source for migration to Mendix.
Use the assess-migration skill.
```

The agent will investigate build files, ORM models, service classes, UI templates, security configuration, and API clients. The resulting assessment report becomes the input for Phase 2.
